/*
fsock.go is released under the MIT License <http://www.opensource.org/licenses/mit-license.php
Copyright (C) ITsysCOM. All Rights Reserved.

Provides FreeSWITCH socket communication.

*/

package fsock

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	ErrConnectionPoolTimeout = errors.New("ConnectionPool timeout")
	ErrDisconnected          = errors.New("DISCONNECTED")
)

// NewFSock connects to FS and starts buffering input
func NewFSock(fsaddr, fspaswd string, reconnects int, maxReconnectInterval time.Duration,
	delayFunc func(time.Duration, time.Duration) func() time.Duration,
	eventHandlers map[string][]func(string, int), eventFilters map[string][]string,
	l logger, connIdx int, bgapiSup bool) (fsock *FSock, err error) {
	if l == nil || (reflect.ValueOf(l).Kind() == reflect.Ptr && reflect.ValueOf(l).IsNil()) {
		l = nopLogger{}
	}
	fsock = &FSock{
		fsMutex:              new(sync.RWMutex),
		connIdx:              connIdx,
		fsaddress:            fsaddr,
		fspaswd:              fspaswd,
		eventHandlers:        eventHandlers,
		eventFilters:         eventFilters,
		backgroundChans:      make(map[string]chan string),
		cmdChan:              make(chan string),
		reconnects:           reconnects,
		maxReconnectInterval: maxReconnectInterval,
		delayFunc:            delayFunc,
		logger:               l,
		bgapiSup:             bgapiSup,
	}
	if err = fsock.Connect(); err != nil {
		return nil, err
	}
	return
}

// FSock reperesents the connection to FreeSWITCH Socket
type FSock struct {
	conn                 net.Conn
	fsMutex              *sync.RWMutex
	connIdx              int // Indetifier for the component using this instance of FSock, optional
	buffer               *bufio.Reader
	fsaddress            string
	fspaswd              string
	eventHandlers        map[string][]func(string, int) // eventStr, connId
	eventFilters         map[string][]string
	backgroundChans      map[string]chan string
	cmdChan              chan string
	reconnects           int
	maxReconnectInterval time.Duration
	delayFunc            func(time.Duration, time.Duration) func() time.Duration // used to create/reset the delay function
	stopReadEvents       chan struct{}                                           // Keep a reference towards forkedReadEvents so we can stop them whenever necessary
	errReadEvents        chan error
	logger               logger
	bgapiSup             bool
}

// Connect or reconnect
func (fs *FSock) Connect() error {
	if fs.stopReadEvents != nil {
		close(fs.stopReadEvents) // we have read events already processing, request stop
	}
	// Reinit readEvents channels so we avoid concurrency issues between goroutines
	fs.stopReadEvents = make(chan struct{})
	fs.errReadEvents = make(chan error)
	return fs.connect()
}

func (fs *FSock) connect() (err error) {
	if fs.Connected() {
		fs.Disconnect()
	}

	var conn net.Conn
	if conn, err = net.Dial("tcp", fs.fsaddress); err != nil {
		fs.logger.Err(fmt.Sprintf("<FSock> Attempt to connect to FreeSWITCH, received: %s", err.Error()))
		return
	}
	fs.fsMutex.Lock()
	fs.conn = conn
	fs.fsMutex.Unlock()
	fs.logger.Info("<FSock> Successfully connected to FreeSWITCH!")
	// Connected, init buffer, auth and subscribe to desired events and filters
	fs.fsMutex.RLock()
	fs.buffer = bufio.NewReaderSize(fs.conn, 8192) // reinit buffer
	fs.fsMutex.RUnlock()

	var authChg string
	if authChg, err = fs.readHeaders(); err != nil {
		return fmt.Errorf("Received error<%s> when receiving the auth challenge", err)
	}
	if !strings.Contains(authChg, "auth/request") {
		return errors.New("No auth challenge received")
	}
	if err = fs.auth(); err != nil { // Auth did not succeed
		return
	}

	if err = fs.filterEvents(fs.eventFilters, fs.bgapiSup); err != nil {
		return
	}

	// Subscribe to events handled by event handlers
	if err = fs.eventsPlain(getMapKeys(fs.eventHandlers), fs.bgapiSup); err != nil {
		return
	}
	go fs.readEvents() // Fork read events in it's own goroutine
	return
}

// Connected checks if socket connected. Can be extended with pings
func (fs *FSock) Connected() (ok bool) {
	fs.fsMutex.RLock()
	ok = (fs.conn != nil)
	fs.fsMutex.RUnlock()
	return
}

// Disconnect disconnects from socket
func (fs *FSock) Disconnect() (err error) {
	fs.fsMutex.Lock()
	if fs.conn != nil {
		fs.logger.Info("<FSock> Disconnecting from FreeSWITCH!")
		err = fs.conn.Close()
		fs.conn = nil
	}
	fs.fsMutex.Unlock()
	return
}

// ReconnectIfNeeded if not connected, attempt reconnect if allowed
func (fs *FSock) ReconnectIfNeeded() (err error) {
	if fs.Connected() { // No need to reconnect
		return
	}
	delay := fs.delayFunc(time.Second, fs.maxReconnectInterval)
	for i := 0; fs.reconnects == -1 || i < fs.reconnects; i++ { // Maximum reconnects reached, -1 for infinite reconnects
		if err = fs.connect(); err == nil && fs.Connected() {
			break // No error or unrelated to connection
		}
		time.Sleep(delay())
	}
	if err == nil && !fs.Connected() {
		return errors.New("Not connected to FreeSWITCH")
	}
	return // nil or last error in the loop
}

func (fs *FSock) send(cmd string) (err error) {
	fs.fsMutex.RLock()
	defer fs.fsMutex.RUnlock()
	if fs.conn == nil {
		return ErrDisconnected
	}
	if _, err = fs.conn.Write([]byte(cmd)); err != nil {
		fs.logger.Err(fmt.Sprintf("<FSock> Cannot write command to socket <%s>", err.Error()))
	}
	return
}

// Auth to FS
func (fs *FSock) auth() (err error) {
	if err = fs.send("auth " + fs.fspaswd + "\n\n"); err != nil {
		return
	}
	var rply string
	if rply, err = fs.readHeaders(); err != nil {
		return
	}
	if !strings.Contains(rply, "Reply-Text: +OK accepted") {
		return fmt.Errorf("Unexpected auth reply received: <%s>", rply)
	}
	return
}

func (fs *FSock) sendCmd(cmd string) (rply string, err error) {
	if err = fs.ReconnectIfNeeded(); err != nil {
		return
	}
	if err = fs.send(cmd + "\n"); err != nil {
		return
	}

	rply = <-fs.cmdChan
	if strings.Contains(rply, "-ERR") {
		return "", errors.New(strings.TrimSpace(rply))
	}
	return
}

// Generic proxy for commands
func (fs *FSock) SendCmd(cmdStr string) (string, error) {
	return fs.sendCmd(cmdStr + "\n")
}

func (fs *FSock) SendCmdWithArgs(cmd string, args map[string]string, body string) (string, error) {
	for k, v := range args {
		cmd += k + ": " + v + "\n"
	}
	if len(body) != 0 {
		cmd += "\n" + body + "\n"
	}
	return fs.sendCmd(cmd)
}

// Send API command
func (fs *FSock) SendApiCmd(cmdStr string) (string, error) {
	return fs.sendCmd("api " + cmdStr + "\n")
}

// Send BGAPI command
func (fs *FSock) SendBgapiCmd(cmdStr string) (out chan string, err error) {
	jobUUID := genUUID()
	out = make(chan string)

	fs.fsMutex.Lock()
	fs.backgroundChans[jobUUID] = out
	fs.fsMutex.Unlock()

	_, err = fs.sendCmd("bgapi " + cmdStr + "\nJob-UUID:" + jobUUID + "\n")
	if err != nil {
		return nil, err
	}
	return
}

// SendMsgCmdWithBody command
func (fs *FSock) SendMsgCmdWithBody(uuid string, cmdargs map[string]string, body string) (err error) {
	if len(cmdargs) == 0 {
		return errors.New("Need command arguments")
	}
	_, err = fs.SendCmdWithArgs("sendmsg "+uuid+"\n", cmdargs, body)
	return
}

// SendMsgCmd command
func (fs *FSock) SendMsgCmd(uuid string, cmdargs map[string]string) error {
	return fs.SendMsgCmdWithBody(uuid, cmdargs, "")
}

// SendEventWithBody command
func (fs *FSock) SendEventWithBody(eventSubclass string, eventParams map[string]string, body string) (string, error) {
	// Event-Name is overrided to CUSTOM by FreeSWITCH,
	// so we use Event-Subclass instead
	eventParams["Event-Subclass"] = eventSubclass
	return fs.SendCmdWithArgs("sendevent "+eventSubclass+"\n", eventParams, body)
}

// SendEvent command
func (fs *FSock) SendEvent(eventSubclass string, eventParams map[string]string) (string, error) {
	return fs.SendEventWithBody(eventSubclass, eventParams, "")
}

// ReadEvents reads events from socket, attempt reconnect if disconnected
func (fs *FSock) ReadEvents() (err error) {
	for {
		if err = <-fs.errReadEvents; err == io.EOF { // Disconnected, try reconnect
			if err = fs.ReconnectIfNeeded(); err != nil {
				return
			}
		}
	}
}

func (fs *FSock) LocalAddr() net.Addr {
	if !fs.Connected() {
		return nil
	}
	return fs.conn.LocalAddr()
}

// Reads headers until delimiter reached
func (fs *FSock) readHeaders() (header string, err error) {
	bytesRead := make([]byte, 0)
	var readLine []byte

	for {
		readLine, err = fs.buffer.ReadBytes('\n')
		if err != nil {
			fs.logger.Err(fmt.Sprintf("<FSock> Error reading headers: <%s>", err.Error()))
			fs.Disconnect()
			err = io.EOF // reconnectIfNeeded
			return
		}
		// No Error, add received to localread buffer
		if len(bytes.TrimSpace(readLine)) == 0 {
			break
		}
		bytesRead = append(bytesRead, readLine...)
	}
	return string(bytesRead), nil
}

// Reads the body from buffer, ln is given by content-length of headers
func (fs *FSock) readBody(noBytes int) (body string, err error) {
	bytesRead := make([]byte, noBytes)
	var readByte byte

	for i := 0; i < noBytes; i++ {
		if readByte, err = fs.buffer.ReadByte(); err != nil {
			fs.logger.Err(fmt.Sprintf("<FSock> Error reading message body: <%s>", err.Error()))
			fs.Disconnect()
			err = io.EOF // reconnectIfNeeded
			return
		}
		// No Error, add received to local read buffer
		bytesRead[i] = readByte
	}
	return string(bytesRead), nil
}

// Event is made out of headers and body (if present)
func (fs *FSock) readEvent() (header string, body string, err error) {
	if header, err = fs.readHeaders(); err != nil {
		return
	}
	if !strings.Contains(header, "Content-Length") { //No body
		return
	}
	var cl int
	if cl, err = strconv.Atoi(headerVal(header, "Content-Length")); err != nil {
		err = fmt.Errorf("Cannot extract content length because<%s>", err)
		return
	}
	body, err = fs.readBody(cl)
	return
}

// Read events from network buffer, stop when exitChan is closed, report on errReadEvents on error and exit
// Receive exitChan and errReadEvents as parameters so we avoid concurrency on using fs.
func (fs *FSock) readEvents() {
	for {
		select {
		case <-fs.stopReadEvents:
			return
		default: // Unlock waiting here
		}
		hdr, body, err := fs.readEvent()
		if err != nil {
			fs.errReadEvents <- err
			return
		}
		if strings.Contains(hdr, "api/response") {
			fs.cmdChan <- body
		} else if strings.Contains(hdr, "command/reply") {
			fs.cmdChan <- headerVal(hdr, "Reply-Text")
		} else if body != "" { // We got a body, could be event, try dispatching it
			fs.dispatchEvent(body)
		}
	}
}

// Subscribe to events
func (fs *FSock) eventsPlain(events []string, bgapiSup bool) (err error) {
	eventsCmd := "event plain"
	customEvents := ""
	for _, ev := range events {
		if ev == "ALL" {
			eventsCmd = "event plain all"
			break
		}
		if strings.HasPrefix(ev, "CUSTOM") {
			customEvents += ev[6:] // will capture here also space between CUSTOM and event
			continue
		}
		eventsCmd += " " + ev
	}
	if eventsCmd != "event plain all" {
		if bgapiSup {
			eventsCmd += " BACKGROUND_JOB" // For bgapi
		}
		if len(customEvents) != 0 { // Add CUSTOM events subscribing in the end otherwise unexpected events are received
			eventsCmd += " " + "CUSTOM" + customEvents
		}
	}

	if err = fs.send(eventsCmd + "\n\n"); err != nil {
		fs.Disconnect()
		return
	}
	var rply string
	if rply, err = fs.readHeaders(); err != nil {
		return
	}
	if !strings.Contains(rply, "Reply-Text: +OK") {
		fs.Disconnect()
		return fmt.Errorf("Unexpected events-subscribe reply received: <%s>", rply)
	}
	return
}

// Enable filters
func (fs *FSock) filterEvents(filters map[string][]string, bgapiSup bool) (err error) {
	if len(filters) == 0 {
		return nil
	}
	if bgapiSup {
		filters["Event-Name"] = append(filters["Event-Name"], "BACKGROUND_JOB") // for bgapi
	}
	for hdr, vals := range filters {
		for _, val := range vals {
			if err = fs.send("filter " + hdr + " " + val + "\n\n"); err != nil {
				fs.Disconnect()
				return
			}
			var rply string
			if rply, err = fs.readHeaders(); err != nil {
				return
			}
			if !strings.Contains(rply, "Reply-Text: +OK") {
				fs.Disconnect()
				return fmt.Errorf("Unexpected filter-events reply received: <%s>", rply)
			}
		}
	}
	return nil
}

// Dispatch events to handlers in async mode
func (fs *FSock) dispatchEvent(event string) {
	eventName := headerVal(event, "Event-Name")
	if eventName == "BACKGROUND_JOB" { // for bgapi BACKGROUND_JOB
		go fs.doBackgroundJob(event)
		return
	}

	if eventName == "CUSTOM" {
		eventSubclass := headerVal(event, "Event-Subclass")
		if len(eventSubclass) != 0 {
			eventName += " " + urlDecode(eventSubclass)
		}
	}

	for _, handleName := range []string{eventName, "ALL"} {
		if _, hasHandlers := fs.eventHandlers[handleName]; hasHandlers {
			// We have handlers, dispatch to all of them
			for _, handlerFunc := range fs.eventHandlers[handleName] {
				go handlerFunc(event, fs.connIdx)
			}
			return
		}
	}
	fs.logger.Warning(fmt.Sprintf("<FSock> No dispatcher for event: <%+v> with event name: %s", event, eventName))
}

// bgapi event lisen fuction
func (fs *FSock) doBackgroundJob(event string) { // add mutex protection
	evMap := EventToMap(event)
	jobUUID, has := evMap["Job-UUID"]
	if !has {
		fs.logger.Err("<FSock> BACKGROUND_JOB with no Job-UUID")
		return
	}

	var out chan string
	fs.fsMutex.RLock()
	out, has = fs.backgroundChans[jobUUID]
	fs.fsMutex.RUnlock()
	if !has {
		fs.logger.Err(fmt.Sprintf("<FSock> BACKGROUND_JOB with UUID %s lost!", jobUUID))
		return // not a requested bgapi
	}

	fs.fsMutex.Lock()
	delete(fs.backgroundChans, jobUUID)
	fs.fsMutex.Unlock()

	out <- evMap[EventBodyTag]
}

// Instantiates a new FSockPool
func NewFSockPool(maxFSocks int, fsaddr, fspasswd string, reconnects int, maxWaitConn time.Duration,
	maxReconnectInterval time.Duration, delayFuncConstructor func(time.Duration, time.Duration) func() time.Duration,
	eventHandlers map[string][]func(string, int), eventFilters map[string][]string,
	l logger, connIdx int, bgapiSup bool) *FSockPool {
	if l == nil {
		l = nopLogger{}
	}
	pool := &FSockPool{
		connIdx:              connIdx,
		fsAddr:               fsaddr,
		fsPasswd:             fspasswd,
		reconnects:           reconnects,
		maxReconnectInterval: maxReconnectInterval,
		delayFuncConstructor: delayFuncConstructor,
		maxWaitConn:          maxWaitConn,
		eventHandlers:        eventHandlers,
		eventFilters:         eventFilters,
		logger:               l,
		allowedConns:         make(chan struct{}, maxFSocks),
		fSocks:               make(chan *FSock, maxFSocks),
		bgapiSup:             bgapiSup,
	}
	for i := 0; i < maxFSocks; i++ {
		pool.allowedConns <- struct{}{} // Empty initiate so we do not need to wait later when we pop
	}
	return pool
}

// Connection handler for commands sent to FreeSWITCH
type FSockPool struct {
	connIdx              int
	fsAddr               string
	fsPasswd             string
	reconnects           int
	maxReconnectInterval time.Duration
	delayFuncConstructor func(time.Duration, time.Duration) func() time.Duration
	eventHandlers        map[string][]func(string, int)
	eventFilters         map[string][]string
	logger               logger
	allowedConns         chan struct{} // Will be populated with members allowed
	fSocks               chan *FSock   // Keep here reference towards the list of opened sockets
	maxWaitConn          time.Duration // Maximum duration to wait for a connection to be returned by Pop
	bgapiSup             bool
}

func (fs *FSockPool) PopFSock() (fsock *FSock, err error) {
	if fs == nil {
		return nil, errors.New("Unconfigured ConnectionPool")
	}
	if len(fs.fSocks) != 0 { // Select directly if available, so we avoid randomness of selection
		fsock = <-fs.fSocks
		return
	}
	tm := time.NewTimer(fs.maxWaitConn)
	select { // No fsock available in the pool, wait for first one showing up
	case fsock = <-fs.fSocks:
		tm.Stop()
		return
	case <-fs.allowedConns:
		tm.Stop()
		return NewFSock(fs.fsAddr, fs.fsPasswd, fs.reconnects, fs.maxReconnectInterval, fs.delayFuncConstructor,
			fs.eventHandlers, fs.eventFilters, fs.logger, fs.connIdx, fs.bgapiSup)
	case <-tm.C:
		return nil, ErrConnectionPoolTimeout
	}
}

func (fs *FSockPool) PushFSock(fsk *FSock) {
	if fs == nil { // Did not initialize the pool
		return
	}
	if fsk == nil || !fsk.Connected() {
		fs.allowedConns <- struct{}{}
		return
	}
	fs.fSocks <- fsk
}
