/*
fsock.go is released under the MIT License <http://www.opensource.org/licenses/mit-license.php
Copyright (C) ITsysCOM. All Rights Reserved.

Provides FreeSWITCH socket communication.

*/

package fsock

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"log"
	"log/syslog"
	"net"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// helper function for uuid generation
func genUUID() string {
	b := make([]byte, 16)
	_, err := io.ReadFull(rand.Reader, b)
	if err != nil {
		log.Fatal(err)
	}
	b[6] = (b[6] & 0x0F) | 0x40
	b[8] = (b[8] &^ 0x40) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[:4], b[4:6], b[6:8], b[8:10],
		b[10:])
}

func indexStringAll(origStr, srchd string) []int {
	foundIdxs := make([]int, 0)
	lenSearched := len(srchd)
	startIdx := 0
	for {
		idxFound := strings.Index(origStr[startIdx:], srchd)
		if idxFound == -1 {
			break
		} else {
			idxFound += startIdx
			foundIdxs = append(foundIdxs, idxFound)
			startIdx = idxFound + lenSearched // Skip the characters found on next check
		}
	}
	return foundIdxs
}

// Split considering {}[] which cancel separator
// In the end we merge groups which are having consecutive [] or {} in beginning since this is how FS builts them
func splitIgnoreGroups(origStr, sep string) []string {
	if len(origStr) == 0 {
		return []string{}
	} else if len(sep) == 0 {
		return []string{origStr}
	}
	retSplit := make([]string, 0)
	cmIdxs := indexStringAll(origStr, ",") // Main indexes of separators
	if len(cmIdxs) == 0 {
		return []string{origStr}
	}
	oCrlyIdxs := indexStringAll(origStr, "{") // Index  { for exceptions
	cCrlyIdxs := indexStringAll(origStr, "}") // Index  } for exceptions closing
	oBrktIdxs := indexStringAll(origStr, "[") // Index [ for exceptions
	cBrktIdxs := indexStringAll(origStr, "]") // Index ] for exceptions closing
	lastNonexcludedIdx := 0
	for i, cmdIdx := range cmIdxs {
		if len(oCrlyIdxs) == len(cCrlyIdxs) && len(oBrktIdxs) == len(cBrktIdxs) { // We assume exceptions and closing them are symetrical, otherwise don't handle exceptions
			exceptFound := false
			for iCrlyIdx := range oCrlyIdxs {
				if oCrlyIdxs[iCrlyIdx] < cmdIdx && cCrlyIdxs[iCrlyIdx] > cmdIdx { // Parentheses canceling indexing found
					exceptFound = true
					break
				}
			}
			for oBrktIdx := range oBrktIdxs {
				if oBrktIdxs[oBrktIdx] < cmdIdx && cBrktIdxs[oBrktIdx] > cmdIdx { // Parentheses canceling indexing found
					exceptFound = true
					break
				}
			}
			if exceptFound {
				continue
			}
		}
		switch i {
		case 0: // First one
			retSplit = append(retSplit, origStr[:cmIdxs[i]])
		case len(cmIdxs) - 1: // Last one
			postpendStr := ""
			if len(origStr) > cmIdxs[i]+1 { // Our separator is not the last character in the string
				postpendStr = origStr[cmIdxs[i]+1:]
			}
			retSplit = append(retSplit, origStr[cmIdxs[lastNonexcludedIdx]+1:cmIdxs[i]], postpendStr)
		default:
			retSplit = append(retSplit, origStr[cmIdxs[lastNonexcludedIdx]+1:cmIdxs[i]]) // Discard the separator from end string
		}
		lastNonexcludedIdx = i
	}
	groupedSplt := make([]string, 0)
	// Merge more consecutive groups (this is how FS displays app data from dial strings)
	for idx, spltData := range retSplit {
		if idx == 0 {
			groupedSplt = append(groupedSplt, spltData)
			continue // Nothing to do for first data
		}
		isGroup, _ := regexp.MatchString("{.*}|[.*]", spltData)
		if !isGroup {
			groupedSplt = append(groupedSplt, spltData)
			continue
		}
		isPrevGroup, _ := regexp.MatchString("{.*}|[.*]", retSplit[idx-1])
		if !isPrevGroup {
			groupedSplt = append(groupedSplt, spltData)
			continue
		}
		groupedSplt[len(groupedSplt)-1] = groupedSplt[len(groupedSplt)-1] + sep + spltData // Merge it with the previous data
	}
	return groupedSplt
}

// Extracts value of a header from anywhere in content string
func headerVal(hdrs, hdr string) string {
	var hdrSIdx, hdrEIdx int
	if hdrSIdx = strings.Index(hdrs, hdr); hdrSIdx == -1 {
		return ""
	} else if hdrEIdx = strings.Index(hdrs[hdrSIdx:], "\n"); hdrEIdx == -1 {
		hdrEIdx = len(hdrs[hdrSIdx:])
	}
	splt := strings.SplitN(hdrs[hdrSIdx:hdrSIdx+hdrEIdx], ": ", 2)
	if len(splt) != 2 {
		return ""
	}
	return strings.TrimSpace(strings.TrimRight(splt[1], "\n"))
}

// FS event header values are urlencoded. Use this to decode them. On error, use original value
func urlDecode(hdrVal string) string {
	if valUnescaped, errUnescaping := url.QueryUnescape(hdrVal); errUnescaping == nil {
		hdrVal = valUnescaped
	}
	return hdrVal
}

// Binary string search in slice
func isSliceMember(ss []string, s string) bool {
	sort.Strings(ss)
	i := sort.SearchStrings(ss, s)
	return (i < len(ss) && ss[i] == s)
}

// Convert fseventStr into fseventMap
func FSEventStrToMap(fsevstr string, headers []string) map[string]string {
	fsevent := make(map[string]string)
	filtered := (len(headers) != 0)
	for _, strLn := range strings.Split(fsevstr, "\n") {
		if hdrVal := strings.SplitN(strLn, ": ", 2); len(hdrVal) == 2 {
			if filtered && isSliceMember(headers, hdrVal[0]) {
				continue // Loop again since we only work on filtered fields
			}
			fsevent[hdrVal[0]] = urlDecode(strings.TrimSpace(strings.TrimRight(hdrVal[1], "\n")))
		}
	}
	return fsevent
}

// Converts string received from fsock into a list of channel info, each represented in a map
func MapChanData(chanInfoStr string) (chansInfoMap []map[string]string) {
	chansInfoMap = make([]map[string]string, 0)
	spltChanInfo := strings.Split(chanInfoStr, "\n")
	if len(spltChanInfo) <= 4 {
		return
	}
	hdrs := strings.Split(spltChanInfo[0], ",")
	for _, chanInfoLn := range spltChanInfo[1 : len(spltChanInfo)-3] {
		chanInfo := splitIgnoreGroups(chanInfoLn, ",")
		if len(hdrs) != len(chanInfo) {
			continue
		}
		chnMp := make(map[string]string, 0)
		for iHdr, hdr := range hdrs {
			chnMp[hdr] = chanInfo[iHdr]
		}
		chansInfoMap = append(chansInfoMap, chnMp)
	}
	return
}

func EventToMap(event string) (result map[string]string) {
	result = make(map[string]string, 0)
	body := false
	for _, line := range strings.Split(event, "\n") {
		if len(line) == 0 {
			body = true
		}
		if body {
			result["EvBody"] = result["EvBody"] + "\n" + line
			continue
		}
		if val := strings.SplitN(line, ": ", 2); len(val) == 2 {
			result[val[0]] = val[1]
		}

	}
	return
}

// successive Fibonacci numbers.
func fib() func() int {
	a, b := 0, 1
	return func() int {
		a, b = b, a+b
		return a
	}
}

var FS *FSock // Used to share FS connection via package globals

// Connection to FreeSWITCH Socket
type FSock struct {
	conn               net.Conn
	connMutex          *sync.RWMutex
	connId             string // Indetifier for the component using this instance of FSock, optional
	buffer             *bufio.Reader
	fsaddress, fspaswd string
	eventHandlers      map[string][]func(string, string) // eventStr, connId
	eventFilters       map[string][]string
	apiChan, cmdChan   chan string
	reconnects         int
	delayFunc          func() int
	stopReadEvents     chan struct{} //Keep a reference towards forkedReadEvents so we can stop them whenever necessary
	errReadEvents      chan error
	backgroundChans    map[string]chan string
	logger             *syslog.Writer
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
	if self.stopReadEvents == nil {
		self.stopReadEvents = make(chan struct{})
	}
	if self.errReadEvents == nil {
		self.errReadEvents = make(chan error)
	}
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
			self.apiChan <- body
		} else if strings.Contains(hdr, "command/reply") {
			self.cmdChan <- headerVal(hdr, "Reply-Text")
		} else if body != "" { // We got a body, could be event, try dispatching it
			self.dispatchEvent(body)
		}
	}
}

func (self *FSock) send(cmd string) {
	self.connMutex.RLock()
	// fmt.Fprint(self.conn, cmd)

	w, err := self.conn.Write([]byte(cmd))
	if err != nil {
		if self.logger != nil {
			self.logger.Err(fmt.Sprintf("<FSock> Cannot write to socket <%s>", err.Error()))
		}
		return
	}
	/* safety check, since we don't know what may come out of the "net" lib*/
	if w == 0 {
		if self.logger != nil {
			self.logger.Debug("<FSock> Cannot write to socket: " + cmd)
		}
		return
	}
	self.connMutex.RUnlock()
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

// Subscribe to events
func (self *FSock) eventsPlain(events []string) error {
	if len(events) == 0 {
		return nil
	}
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
	eventsCmd += " BACKGROUND_JOB"                                // For bgapi
	if len(customEvents) != 0 && eventsCmd != "event plain all" { // Add CUSTOM events subscribing in the end otherwise unexpected events are received
		eventsCmd += " " + "CUSTOM" + customEvents
	}
	eventsCmd += "\n"
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
		go self.backgroudJob(event)
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
				go handlerFunc(event, self.connId)
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
func (self *FSock) backgroudJob(event string) { // add mutex protection
	evMap := EventToMap(event)
	jobUuid, has := evMap["Job-UUID"]
	if !has {
		self.logger.Err("<FSock> BACKGROUND_JOB with no Job-UUID")
		return
	}

	var out chan string
	out, has = self.backgroundChans[jobUuid]
	if !has {
		self.logger.Err(fmt.Sprintf("<FSock> BACKGROUND_JOB with UUID %s lost!", jobUuid))
		return // not a requested bgapi
	}

	delete(self.backgroundChans, jobUuid)
	out <- evMap["EvBody"]
}

// Checks if socket connected. Can be extended with pings
func (self *FSock) Connected() (ok bool) {
	self.connMutex.RLock()
	ok = (self.conn != nil)
	self.connMutex.RUnlock()
	return
}

// Disconnects from socket
func (self *FSock) Disconnect() (err error) {
	self.connMutex.Lock()
	if self.conn != nil {
		if self.logger != nil {
			self.logger.Info("<FSock> Disconnecting from FreeSWITCH!")
		}
		err = self.conn.Close()
		self.conn = nil
	}
	self.connMutex.Unlock()
	return
}

// Connect or reconnect
func (self *FSock) Connect() error {
	if self.Connected() {
		self.Disconnect()
	}
	if self.stopReadEvents != nil { // ToDo: Check if the channel is not already closed
		close(self.stopReadEvents) // we have read events already processing, request stop
		self.stopReadEvents = nil
	}
	conn, err := net.Dial("tcp", self.fsaddress)
	if err != nil {
		if self.logger != nil {
			self.logger.Err(fmt.Sprintf("<FSock> Attempt to connect to FreeSWITCH, received: %s", err.Error()))
		}
		return err
	}
	self.connMutex.Lock()
	self.conn = conn
	self.connMutex.Unlock()
	if self.logger != nil {
		self.logger.Info("<FSock> Successfully connected to FreeSWITCH!")
	}
	// Connected, init buffer, auth and subscribe to desired events and filters
	self.connMutex.RLock()
	self.buffer = bufio.NewReaderSize(self.conn, 8192) // reinit buffer
	self.connMutex.RUnlock()

	if authChg, err := self.readHeaders(); err != nil || !strings.Contains(authChg, "auth/request") {
		return errors.New("No auth challenge received")
	} else if errAuth := self.auth(); errAuth != nil { // Auth did not succeed
		return errAuth
	}
	// Subscribe to events handled by event handlers
	handledEvs := make([]string, len(self.eventHandlers))
	j := 0
	for k := range self.eventHandlers {
		handledEvs[j] = k
		j++
	}
	if err := self.eventsPlain(handledEvs); err != nil {
		return err
	}
	if err := self.filterEvents(self.eventFilters); err != nil {
		return err
	}
	// Reinit readEvents channels so we avoid concurrency issues between goroutines
	self.stopReadEvents = make(chan struct{})
	self.errReadEvents = make(chan error)
	go self.readEvents() // Fork read events in it's own goroutine
	return nil
}

// If not connected, attempt reconnect if allowed
func (self *FSock) ReconnectIfNeeded() (err error) {
	if self.Connected() { // No need to reconnect
		return nil
	}
	for i := 0; self.reconnects != -1 && i >= self.reconnects; i++ { // Maximum reconnects reached, -1 for infinite reconnects
		if err = self.Connect(); err == nil || self.Connected() {
			self.delayFunc = fib() // Reset the reconnect delay
			break                  // No error or unrelated to connection
		}
		time.Sleep(time.Duration(self.delayFunc()) * time.Second)
	}
	if err == nil && !self.Connected() {
		return errors.New("Not connected to FreeSWITCH")
	}
	return err // nil or last error in the loop
}

func (self *FSock) sendCmd(cmd string, cmdChan chan string) (rply string, err error) {
	if err = self.ReconnectIfNeeded(); err != nil {
		return "", err
	}
	cmd = fmt.Sprintf("%s\n", cmd)
	self.send(cmd)
	rply = <-cmdChan
	if strings.Contains(rply, "-ERR") {
		return "", errors.New(strings.TrimSpace(rply))
	}
	return rply, nil
}

// Generic proxy for commands
func (self *FSock) SendCmd(cmdStr string) (string, error) {
	return self.sendCmd(cmdStr+"\n", self.cmdChan)
}

// Send API command
func (self *FSock) SendApiCmd(cmdStr string) (string, error) {
	return self.sendCmd("api "+cmdStr+"\n", self.apiChan)
}

// Send BGAPI command
func (self *FSock) SendBgapiCmd(cmdStr string) (out chan string, err error) {
	jobUuid := genUUID()
	out = make(chan string)
	_, err = self.sendCmd(fmt.Sprintf("bgapi %s\nJob-UUID:%s", cmdStr, jobUuid), self.cmdChan)
	if err != nil {
		return nil, err
	}
	self.backgroundChans[jobUuid] = out
	return
}

// SendMessage command
func (self *FSock) SendMsgCmd(uuid string, cmdargs map[string]string) error {
	if len(cmdargs) == 0 {
		return errors.New("Need command arguments")
	}
	argStr := ""
	for k, v := range cmdargs {
		argStr += fmt.Sprintf("%s:%s\n", k, v)
	}
	_, err := self.sendCmd(fmt.Sprintf("sendmsg %s\n%s", uuid, argStr), self.cmdChan)
	return err
}

// SendEvent command
func (self *FSock) SendEvent(eventSubclass string, eventParams map[string]string) (string, error) {
	// Event-Name is overrided to CUSTOM by FreeSWITCH,
	// so we use Event-Subclass instead
	eventParams["Event-Subclass"] = eventSubclass
	cmd := fmt.Sprintf("sendevent %s\n", eventSubclass)
	for k, v := range eventParams {
		cmd += fmt.Sprintf("%s: %s\n", k, v)
	}
	return self.sendCmd(cmd, self.cmdChan)
}

// Reads events from socket, attempt reconnect if disconnected
func (self *FSock) ReadEvents() (err error) {
	for {
		if err = <-self.errReadEvents; err == io.EOF { // Disconnected, try reconnect
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

// Connects to FS and starts buffering input
func NewFSock(fsaddr, fspaswd string, reconnects int, eventHandlers map[string][]func(string, string), eventFilters map[string][]string, l *syslog.Writer, connId string) (fsock *FSock, err error) {

	fsock = &FSock{
		connMutex:     new(sync.RWMutex),
		connId:        connId,
		fsaddress:     fsaddr,
		fspaswd:       fspaswd,
		eventHandlers: eventHandlers,
		eventFilters:  eventFilters,
		reconnects:    reconnects,
		logger:        l,
		apiChan:       make(chan string),
		cmdChan:       make(chan string),
		delayFunc:     fib(),
	}
	if err = fsock.Connect(); err != nil {
		return nil, err
	}
	return
}

var ErrConnectionPoolTimeout = errors.New("ConnectionPool timeout")

// Connection handler for commands sent to FreeSWITCH
type FSockPool struct {
	connId        string
	fsAddr        string
	fsPasswd      string
	reconnects    int
	eventHandlers map[string][]func(string, string)
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
		fsock, err = NewFSock(self.fsAddr, self.fsPasswd, self.reconnects, self.eventHandlers, self.eventFilters, self.logger, self.connId)
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

// Instantiates a new FSockPool
func NewFSockPool(maxFSocks int, fsaddr, fspasswd string, reconnects int, maxWaitConn time.Duration,
	eventHandlers map[string][]func(string, string), eventFilters map[string][]string, l *syslog.Writer, connId string) (*FSockPool, error) {
	pool := &FSockPool{
		connId:        connId,
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
