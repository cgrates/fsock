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
	"log/syslog"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	FS        *FSock // Used to share FS connection via package globals
	DelayFunc func() func() int

	ErrConnectionPoolTimeout = errors.New("ConnectionPool timeout")
)

func init() {
	DelayFunc = fib
}

// Connects to FS and starts buffering input
func NewFSock(fsaddr, fspaswd string, reconnects int, eventHandlers map[string][]func(string, int), eventFilters map[string][]string, l *syslog.Writer, connIdx int) (fsock *FSock, err error) {
	fsock = &FSock{
		fsMutex:         new(sync.RWMutex),
		connIdx:         connIdx,
		fsaddress:       fsaddr,
		fspaswd:         fspaswd,
		eventHandlers:   eventHandlers,
		eventFilters:    eventFilters,
		backgroundChans: make(map[string]chan string),
		cmdChan:         make(chan string),
		reconnects:      reconnects,
		delayFunc:       DelayFunc(),
		logger:          l,
	}
	if err = fsock.Connect(); err != nil {
		return nil, err
	}
	return
}

// Connection to FreeSWITCH Socket
type FSock struct {
	conn            net.Conn
	fsMutex         *sync.RWMutex
	connIdx         int // Indetifier for the component using this instance of FSock, optional
	buffer          *bufio.Reader
	fsaddress       string
	fspaswd         string
	eventHandlers   map[string][]func(string, int) // eventStr, connId
	eventFilters    map[string][]string
	backgroundChans map[string]chan string
	cmdChan         chan string
	reconnects      int
	delayFunc       func() int
	stopReadEvents  chan struct{} //Keep a reference towards forkedReadEvents so we can stop them whenever necessary
	errReadEvents   chan error
	logger          *syslog.Writer
}

// Connect or reconnect
func (self *FSock) Connect() error {
	if self.stopReadEvents != nil {
		close(self.stopReadEvents) // we have read events already processing, request stop
	}
	// Reinit readEvents channels so we avoid concurrency issues between goroutines
	self.stopReadEvents = make(chan struct{})
	self.errReadEvents = make(chan error)
	return self.connect()

}

func (self *FSock) connect() error {
	if self.Connected() {
		self.Disconnect()
	}

	conn, err := net.Dial("tcp", self.fsaddress)
	if err != nil {
		if self.logger != nil {
			self.logger.Err(fmt.Sprintf("<FSock> Attempt to connect to FreeSWITCH, received: %s", err.Error()))
		}
		return err
	}
	self.fsMutex.Lock()
	self.conn = conn
	self.fsMutex.Unlock()
	if self.logger != nil {
		self.logger.Info("<FSock> Successfully connected to FreeSWITCH!")
	}
	// Connected, init buffer, auth and subscribe to desired events and filters
	self.fsMutex.RLock()
	self.buffer = bufio.NewReaderSize(self.conn, 8192) // reinit buffer
	self.fsMutex.RUnlock()

	if authChg, err := self.readHeaders(); err != nil || !strings.Contains(authChg, "auth/request") {
		return errors.New("No auth challenge received")
	} else if errAuth := self.auth(); errAuth != nil { // Auth did not succeed
		return errAuth
	}
	// Subscribe to events handled by event handlers
	if err := self.eventsPlain(getMapKeys(self.eventHandlers)); err != nil {
		return err
	}

	if err := self.filterEvents(self.eventFilters); err != nil {
		return err
	}
	go self.readEvents() // Fork read events in it's own goroutine
	return nil
}

// Connected checks if socket connected. Can be extended with pings
func (self *FSock) Connected() (ok bool) {
	self.fsMutex.RLock()
	ok = (self.conn != nil)
	self.fsMutex.RUnlock()
	return
}

// Disconnects from socket
func (self *FSock) Disconnect() (err error) {
	self.fsMutex.Lock()
	if self.conn != nil {
		if self.logger != nil {
			self.logger.Info("<FSock> Disconnecting from FreeSWITCH!")
		}
		err = self.conn.Close()
		self.conn = nil
	}
	self.fsMutex.Unlock()
	return
}

// If not connected, attempt reconnect if allowed
func (self *FSock) ReconnectIfNeeded() (err error) {
	if self.Connected() { // No need to reconnect
		return nil
	}
	for i := 0; self.reconnects == -1 || i < self.reconnects; i++ { // Maximum reconnects reached, -1 for infinite reconnects
		if err = self.connect(); err == nil && self.Connected() {
			self.delayFunc = DelayFunc() // Reset the reconnect delay
			break                        // No error or unrelated to connection
		}
		time.Sleep(time.Duration(self.delayFunc()) * time.Second)
	}
	if err == nil && !self.Connected() {
		return errors.New("Not connected to FreeSWITCH")
	}
	return err // nil or last error in the loop
}

func (self *FSock) send(cmd string) {
	self.fsMutex.RLock()
	// fmt.Fprint(self.conn, cmd)
	w, err := self.conn.Write([]byte(cmd))
	if err != nil {
		if self.logger != nil {
			self.logger.Err(fmt.Sprintf("<FSock> Cannot write command to socket <%s>", err.Error()))
		}
		return
	}
	if w == 0 {
		if self.logger != nil {
			self.logger.Debug("<FSock> Cannot write command to socket: " + cmd)
		}
		return
	}
	self.fsMutex.RUnlock()
}

// Auth to FS
func (self *FSock) auth() error {
	self.send(fmt.Sprintf("auth %s\n\n", self.fspaswd))
	if rply, err := self.readHeaders(); err != nil {
		return err
	} else if !strings.Contains(rply, "Reply-Text: +OK accepted") {
		return fmt.Errorf("Unexpected auth reply received: <%s>", rply)
	}
	return nil
}

func (self *FSock) sendCmd(cmd string) (rply string, err error) {
	if err = self.ReconnectIfNeeded(); err != nil {
		return "", err
	}
	cmd = fmt.Sprintf("%s\n", cmd)
	self.send(cmd)
	rply = <-self.cmdChan
	if strings.Contains(rply, "-ERR") {
		return "", errors.New(strings.TrimSpace(rply))
	}
	return rply, nil
}

// Generic proxy for commands
func (self *FSock) SendCmd(cmdStr string) (string, error) {
	return self.sendCmd(cmdStr + "\n")
}

func (self *FSock) SendCmdWithArgs(cmd string, args map[string]string, body string) (string, error) {
	for k, v := range args {
		cmd += fmt.Sprintf("%s: %s\n", k, v)
	}
	if len(body) != 0 {
		cmd += fmt.Sprintf("\n%s\n", body)
	}
	return self.sendCmd(cmd)
}

// Send API command
func (self *FSock) SendApiCmd(cmdStr string) (string, error) {
	return self.sendCmd("api " + cmdStr + "\n")
}

// Send BGAPI command
func (self *FSock) SendBgapiCmd(cmdStr string) (out chan string, err error) {
	jobUuid := genUUID()
	out = make(chan string)

	self.fsMutex.Lock()
	self.backgroundChans[jobUuid] = out
	self.fsMutex.Unlock()

	_, err = self.sendCmd(fmt.Sprintf("bgapi %s\nJob-UUID:%s\n", cmdStr, jobUuid))
	if err != nil {
		return nil, err
	}
	return
}

// SendMsgCmdWithBody command
func (self *FSock) SendMsgCmdWithBody(uuid string, cmdargs map[string]string, body string) error {
	if len(cmdargs) == 0 {
		return errors.New("Need command arguments")
	}
	_, err := self.SendCmdWithArgs(fmt.Sprintf("sendmsg %s\n", uuid), cmdargs, body)
	return err
}

// SendMsgCmd command
func (self *FSock) SendMsgCmd(uuid string, cmdargs map[string]string) error {
	return self.SendMsgCmdWithBody(uuid, cmdargs, "")
}

// SendEventWithBody command
func (self *FSock) SendEventWithBody(eventSubclass string, eventParams map[string]string, body string) (string, error) {
	// Event-Name is overrided to CUSTOM by FreeSWITCH,
	// so we use Event-Subclass instead
	eventParams["Event-Subclass"] = eventSubclass
	return self.SendCmdWithArgs(fmt.Sprintf("sendevent %s\n", eventSubclass), eventParams, body)
}

// SendEvent command
func (self *FSock) SendEvent(eventSubclass string, eventParams map[string]string) (string, error) {
	return self.SendEventWithBody(eventSubclass, eventParams, "")
}

// ReadEvents reads events from socket, attempt reconnect if disconnected
func (self *FSock) ReadEvents() (err error) {
	var opened bool
	for {
		if err, opened = <-self.errReadEvents; !opened {
			return nil
		} else if err == io.EOF { // Disconnected, try reconnect
			if err = self.ReconnectIfNeeded(); err != nil {
				break
			}
		}
	}
	return err
}

func (self *FSock) LocalAddr() net.Addr {
	if !self.Connected() {
		return nil
	}
	return self.conn.LocalAddr()
}

// Reads headers until delimiter reached
func (self *FSock) readHeaders() (header string, err error) {
	bytesRead := make([]byte, 0)
	var readLine []byte

	for {
		readLine, err = self.buffer.ReadBytes('\n')
		if err != nil {
			if self.logger != nil {
				self.logger.Err(fmt.Sprintf("<FSock> Error reading headers: <%s>", err.Error()))
			}
			self.Disconnect()
			return "", err
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
func (self *FSock) readBody(noBytes int) (body string, err error) {
	bytesRead := make([]byte, noBytes)
	var readByte byte

	for i := 0; i < noBytes; i++ {
		if readByte, err = self.buffer.ReadByte(); err != nil {
			if self.logger != nil {
				self.logger.Err(fmt.Sprintf("<FSock> Error reading message body: <%s>", err.Error()))
			}
			self.Disconnect()
			return "", err
		}
		// No Error, add received to local read buffer
		bytesRead[i] = readByte
	}
	return string(bytesRead), nil
}

// Event is made out of headers and body (if present)
func (self *FSock) readEvent() (header string, body string, err error) {
	var cl int

	if header, err = self.readHeaders(); err != nil {
		return "", "", err
	}
	if !strings.Contains(header, "Content-Length") { //No body
		return header, "", nil
	}
	if cl, err = strconv.Atoi(headerVal(header, "Content-Length")); err != nil {
		return "", "", errors.New("Cannot extract content length")
	}
	if body, err = self.readBody(cl); err != nil {
		return "", "", err
	}
	return
}

// Read events from network buffer, stop when exitChan is closed, report on errReadEvents on error and exit
// Receive exitChan and errReadEvents as parameters so we avoid concurrency on using self.
func (self *FSock) readEvents() {
	for {
		select {
		case <-self.stopReadEvents:
			return
		default: // Unlock waiting here
		}
		hdr, body, err := self.readEvent()
		if err != nil {
			self.errReadEvents <- err
			return
		}
		if strings.Contains(hdr, "api/response") {
			self.cmdChan <- body
		} else if strings.Contains(hdr, "command/reply") {
			self.cmdChan <- headerVal(hdr, "Reply-Text")
		} else if body != "" { // We got a body, could be event, try dispatching it
			self.dispatchEvent(body)
		}
	}
}

// Subscribe to events
func (self *FSock) eventsPlain(events []string) error {
	// if len(events) == 0 {
	// 	return nil
	// }
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
		eventsCmd += " BACKGROUND_JOB" // For bgapi
		if len(customEvents) != 0 {    // Add CUSTOM events subscribing in the end otherwise unexpected events are received
			eventsCmd += " " + "CUSTOM" + customEvents
		}
	}
	eventsCmd += "\n\n"
	self.send(eventsCmd)
	if rply, err := self.readHeaders(); err != nil {
		return err
	} else if !strings.Contains(rply, "Reply-Text: +OK") {
		self.Disconnect()
		return fmt.Errorf("Unexpected events-subscribe reply received: <%s>", rply)
	}
	return nil
}

// Enable filters
func (self *FSock) filterEvents(filters map[string][]string) error {
	if len(filters) == 0 {
		return nil
	}
	filters["Event-Name"] = append(filters["Event-Name"], "BACKGROUND_JOB") // for bgapi
	for hdr, vals := range filters {
		for _, val := range vals {
			cmd := "filter " + hdr + " " + val + "\n\n"
			self.send(cmd)
			if rply, err := self.readHeaders(); err != nil {
				return err
			} else if !strings.Contains(rply, "Reply-Text: +OK") {
				return fmt.Errorf("Unexpected filter-events reply received: <%s>", rply)
			}
		}
	}
	return nil
}

// Dispatch events to handlers in async mode
func (self *FSock) dispatchEvent(event string) {
	eventName := headerVal(event, "Event-Name")
	if eventName == "BACKGROUND_JOB" { // for bgapi BACKGROUND_JOB
		go self.doBackgroudJob(event)
		return
	}

	if eventName == "CUSTOM" {
		eventSubclass := headerVal(event, "Event-Subclass")
		if len(eventSubclass) != 0 {
			eventName += " " + urlDecode(eventSubclass)
		}
	}

	for _, handleName := range []string{eventName, "ALL"} {
		if _, hasHandlers := self.eventHandlers[handleName]; hasHandlers {
			// We have handlers, dispatch to all of them
			for _, handlerFunc := range self.eventHandlers[handleName] {
				go handlerFunc(event, self.connIdx)
			}
			return
		}
	}
	if self.logger != nil {
		fmt.Printf("No dispatcher, event name: %s, handlers: %+v\n", eventName, self.eventHandlers)
		self.logger.Warning(fmt.Sprintf("<FSock> No dispatcher for event: <%+v>", event))
	}
}

// bgapi event lisen fuction
func (self *FSock) doBackgroudJob(event string) { // add mutex protection
	evMap := EventToMap(event)
	jobUuid, has := evMap["Job-UUID"]
	if !has {
		self.logger.Err("<FSock> BACKGROUND_JOB with no Job-UUID")
		return
	}

	var out chan string
	self.fsMutex.RLock()
	out, has = self.backgroundChans[jobUuid]
	self.fsMutex.RUnlock()
	if !has {
		self.logger.Err(fmt.Sprintf("<FSock> BACKGROUND_JOB with UUID %s lost!", jobUuid))
		return // not a requested bgapi
	}

	self.fsMutex.Lock()
	delete(self.backgroundChans, jobUuid)
	self.fsMutex.Unlock()

	out <- evMap[EventBodyTag]
}

// Instantiates a new FSockPool
func NewFSockPool(maxFSocks int, fsaddr, fspasswd string, reconnects int, maxWaitConn time.Duration,
	eventHandlers map[string][]func(string, int), eventFilters map[string][]string, l *syslog.Writer, connIdx int) (*FSockPool, error) {
	pool := &FSockPool{
		connIdx:       connIdx,
		fsAddr:        fsaddr,
		fsPasswd:      fspasswd,
		reconnects:    reconnects,
		maxWaitConn:   maxWaitConn,
		eventHandlers: eventHandlers,
		eventFilters:  eventFilters,
		logger:        l,
		allowedConns:  make(chan struct{}, maxFSocks),
		fSocks:        make(chan *FSock, maxFSocks),
	}
	for i := 0; i < maxFSocks; i++ {
		pool.allowedConns <- struct{}{} // Empty initiate so we do not need to wait later when we pop
	}
	return pool, nil
}

// Connection handler for commands sent to FreeSWITCH
type FSockPool struct {
	connIdx       int
	fsAddr        string
	fsPasswd      string
	reconnects    int
	eventHandlers map[string][]func(string, int)
	eventFilters  map[string][]string
	logger        *syslog.Writer
	allowedConns  chan struct{} // Will be populated with members allowed
	fSocks        chan *FSock   // Keep here reference towards the list of opened sockets
	maxWaitConn   time.Duration // Maximum duration to wait for a connection to be returned by Pop
}

func (self *FSockPool) PopFSock() (fsock *FSock, err error) {
	if self == nil {
		return nil, errors.New("Unconfigured ConnectionPool")
	}
	if len(self.fSocks) != 0 { // Select directly if available, so we avoid randomness of selection
		fsock = <-self.fSocks
		return fsock, nil
	}
	select { // No fsock available in the pool, wait for first one showing up
	case fsock = <-self.fSocks:
	case <-self.allowedConns:
		fsock, err = NewFSock(self.fsAddr, self.fsPasswd, self.reconnects, self.eventHandlers, self.eventFilters, self.logger, self.connIdx)
		if err != nil {
			return nil, err
		}
		return fsock, nil
	case <-time.After(self.maxWaitConn):
		return nil, ErrConnectionPoolTimeout
	}
	return fsock, nil
}

func (self *FSockPool) PushFSock(fsk *FSock) {
	if self == nil { // Did not initialize the pool
		return
	}
	if fsk == nil || !fsk.Connected() {
		self.allowedConns <- struct{}{}
		return
	}
	self.fSocks <- fsk
}
