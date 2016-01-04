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
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

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
	if i := sort.SearchStrings(ss, s); i < len(ss) && ss[i] == s {
		return true
	}
	return false
}

// Convert fseventStr into fseventMap
func FSEventStrToMap(fsevstr string, headers []string) map[string]string {
	fsevent := make(map[string]string)
	filtered := false
	if len(headers) != 0 {
		filtered = true
	}
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
func MapChanData(chanInfoStr string) []map[string]string {
	chansInfoMap := make([]map[string]string, 0)
	spltChanInfo := strings.Split(chanInfoStr, "\n")
	if len(spltChanInfo) < 5 {
		return chansInfoMap
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
	return chansInfoMap
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
	eventFilters       map[string]string
	apiChan, cmdChan   chan string
	reconnects         int
	delayFunc          func() int
	stopReadEvents     chan struct{} //Keep a reference towards forkedReadEvents so we can stop them whenever necessary
	errReadEvents      chan error
	logger             *syslog.Writer
}

// Reads headers until delimiter reached
func (self *FSock) readHeaders() (string, error) {
	bytesRead := make([]byte, 0)
	var readLine []byte
	var err error
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
func (self *FSock) readBody(ln int) (string, error) {
	bytesRead := make([]byte, ln)
	for i := 0; i < ln; i++ {
		if readByte, err := self.buffer.ReadByte(); err != nil {
			if self.logger != nil {
				self.logger.Err(fmt.Sprintf("<FSock> Error reading message body: <%s>", err.Error()))
			}
			self.Disconnect()
			return "", err
		} else { // No Error, add received to localread buffer
			bytesRead[i] = readByte // Add received line to the local read buffer
		}
	}
	return string(bytesRead), nil
}

// Event is made out of headers and body (if present)
func (self *FSock) readEvent() (string, string, error) {
	var hdrs, body string
	var cl int
	var err error

	if hdrs, err = self.readHeaders(); err != nil {
		return "", "", err
	}
	if !strings.Contains(hdrs, "Content-Length") { //No body
		return hdrs, "", nil
	}
	clStr := headerVal(hdrs, "Content-Length")
	if cl, err = strconv.Atoi(clStr); err != nil {
		return "", "", errors.New("Cannot extract content length")
	}
	if body, err = self.readBody(cl); err != nil {
		return "", "", err
	}
	return hdrs, body, nil
}

// Read events from network buffer, stop when exitChan is closed, report on errReadEvents on error and exit
// Receive exitChan and errReadEvents as parameters so we avoid concurrency on using self.
func (self *FSock) readEvents(exitChan chan struct{}, errReadEvents chan error) {
	for {
		select {
		case <-exitChan:
			return
		default: // Unlock waiting here
		}
		hdr, body, err := self.readEvent()
		if err != nil {
			errReadEvents <- err
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

// Auth to FS
func (self *FSock) auth() error {
	authCmd := fmt.Sprintf("auth %s\n\n", self.fspaswd)
	self.connMutex.RLock()
	fmt.Fprint(self.conn, authCmd)
	self.connMutex.RUnlock()
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
	if len(customEvents) != 0 && eventsCmd != "event plain all" { // Add CUSTOM events subscribing in the end otherwise unexpected events are received
		eventsCmd += " " + "CUSTOM" + customEvents
	}
	eventsCmd += "\n\n"
	self.connMutex.RLock()
	fmt.Fprint(self.conn, eventsCmd)
	self.connMutex.RUnlock()
	if rply, err := self.readHeaders(); err != nil {
		return err
	} else if !strings.Contains(rply, "Reply-Text: +OK") {
		self.Disconnect()
		return fmt.Errorf("Unexpected events-subscribe reply received: <%s>", rply)
	}
	return nil
}

// Enable filters
func (self *FSock) filterEvents(filters map[string]string) error {
	if len(filters) == 0 { //Nothing to filter
		return nil
	}
	for hdr, val := range filters {
		cmd := "filter " + hdr + " " + val + "\n\n"
		self.connMutex.RLock()
		fmt.Fprint(self.conn, cmd)
		self.connMutex.RUnlock()
		if rply, err := self.readHeaders(); err != nil {
			return err
		} else if !strings.Contains(rply, "Reply-Text: +OK") {
			return fmt.Errorf("Unexpected filter-events reply received: <%s>", rply)
		}
	}

	return nil
}

// Dispatch events to handlers in async mode
func (self *FSock) dispatchEvent(event string) {
	eventName := headerVal(event, "Event-Name")
	if eventName == "CUSTOM" {
		eventSubclass := headerVal(event, "Event-Subclass")
		if len(eventSubclass) != 0 {
			eventName += " " + urlDecode(eventSubclass)
		}
	}
	handleNames := []string{eventName, "ALL"}
	dispatched := false
	for _, handleName := range handleNames {
		if _, hasHandlers := self.eventHandlers[handleName]; hasHandlers {
			// We have handlers, dispatch to all of them
			for _, handlerFunc := range self.eventHandlers[handleName] {
				go handlerFunc(event, self.connId)
			}
			dispatched = true
		}
	}
	if !dispatched && self.logger != nil {
		fmt.Printf("No dispatcher, event name: %s, handlers: %+v\n", eventName, self.eventHandlers)
		self.logger.Warning(fmt.Sprintf("<FSock> No dispatcher for event: <%+v>", event))
	}
}

// Checks if socket connected. Can be extended with pings
func (self *FSock) Connected() bool {
	self.connMutex.RLock()
	defer self.connMutex.RUnlock()
	if self.conn == nil {
		return false
	}
	return true
}

// Disconnects from socket
func (self *FSock) Disconnect() (err error) {
	self.connMutex.Lock()
	defer self.connMutex.Unlock()
	if self.conn != nil {
		if self.logger != nil {
			self.logger.Info("<FSock> Disconnecting from FreeSWITCH!")
		}
		err = self.conn.Close()
		self.conn = nil
	}
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
	if subscribeErr := self.eventsPlain(handledEvs); subscribeErr != nil {
		return subscribeErr
	}
	if filterErr := self.filterEvents(self.eventFilters); filterErr != nil {
		return filterErr
	}
	// Reinit readEvents channels so we avoid concurrency issues between goroutines
	stopReadEvents := make(chan struct{})
	self.stopReadEvents = stopReadEvents
	self.errReadEvents = make(chan error)
	go self.readEvents(stopReadEvents, self.errReadEvents) // Fork read events in it's own goroutine
	return nil
}

// If not connected, attempt reconnect if allowed
func (self *FSock) ReconnectIfNeeded() error {
	if self.Connected() { // No need to reconnect
		return nil
	}
	var err error
	i := 0
	for {
		if self.reconnects != -1 && i >= self.reconnects { // Maximum reconnects reached, -1 for infinite reconnects
			break
		}
		if err = self.Connect(); err == nil || self.Connected() {
			self.delayFunc = fib() // Reset the reconnect delay
			break                  // No error or unrelated to connection
		}
		time.Sleep(time.Duration(self.delayFunc()) * time.Second)
		i++
	}
	if err == nil && !self.Connected() {
		return errors.New("Not connected to FreeSWITCH")
	}
	return err // nil or last error in the loop
}

// Generic proxy for commands
func (self *FSock) SendCmd(cmdStr string) (string, error) {
	if err := self.ReconnectIfNeeded(); err != nil {
		return "", err
	}
	cmd := fmt.Sprintf(cmdStr)
	self.connMutex.RLock()
	fmt.Fprint(self.conn, cmd)
	self.connMutex.RUnlock()
	resEvent := <-self.cmdChan
	if strings.Contains(resEvent, "-ERR") {
		return "", errors.New(strings.TrimSpace(resEvent))
	}
	return resEvent, nil
}

// Send API command
func (self *FSock) SendApiCmd(cmdStr string) (string, error) {
	if err := self.ReconnectIfNeeded(); err != nil {
		return "", err
	}
	cmd := fmt.Sprintf("api %s\n\n", cmdStr)
	self.connMutex.RLock()
	fmt.Fprint(self.conn, cmd)
	self.connMutex.RUnlock()
	resEvent := <-self.apiChan
	if strings.Contains(resEvent, "-ERR") {
		return "", errors.New(strings.TrimSpace(resEvent))
	}
	return resEvent, nil
}

// SendMessage command
func (self *FSock) SendMsgCmd(uuid string, cmdargs map[string]string) error {
	if err := self.ReconnectIfNeeded(); err != nil {
		return err
	}
	if len(cmdargs) == 0 {
		return errors.New("Need command arguments")
	}
	argStr := ""
	for k, v := range cmdargs {
		argStr += fmt.Sprintf("%s:%s\n", k, v)
	}
	self.connMutex.RLock()
	fmt.Fprint(self.conn, fmt.Sprintf("sendmsg %s\n%s\n", uuid, argStr))
	self.connMutex.RUnlock()
	replyTxt := <-self.cmdChan
	if strings.HasPrefix(replyTxt, "-ERR") {
		return errors.New(strings.TrimSpace(replyTxt))
	}
	return nil
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
func NewFSock(fsaddr, fspaswd string, reconnects int, eventHandlers map[string][]func(string, string), eventFilters map[string]string, l *syslog.Writer, connId string) (*FSock, error) {
	fsock := FSock{connMutex: new(sync.RWMutex), connId: connId, fsaddress: fsaddr, fspaswd: fspaswd, eventHandlers: eventHandlers, eventFilters: eventFilters, reconnects: reconnects,
		logger: l, apiChan: make(chan string), cmdChan: make(chan string), delayFunc: fib()}
	errConn := fsock.Connect()
	if errConn != nil {
		return nil, errConn
	}
	return &fsock, nil
}

var ErrConnectionTimeout = errors.New("ConnectionPool timeout")

// Connection handler for commands sent to FreeSWITCH
type FSockPool struct {
	connId, fsAddr, fsPasswd string
	reconnects               int
	eventHandlers            map[string][]func(string, string)
	eventFilters             map[string]string
	logger                   *syslog.Writer
	allowedConns             chan struct{} // Will be populated with members allowed
	fSocks                   chan *FSock   // Keep here reference towards the list of opened sockets
	maxWaitConn              time.Duration // Maximum duration to wait for a connection to be returned by Pop
}

func (self *FSockPool) PopFSock() (*FSock, error) {
	if self == nil {
		return nil, errors.New("Unconfigured ConnectionPool")
	}
	if len(self.fSocks) != 0 { // Select directly if available, so we avoid randomness of selection
		fsock := <-self.fSocks
		return fsock, nil
	}
	var fsock *FSock
	var err error
	select { // No fsock available in the pool, wait for first one showing up
	case fsock = <-self.fSocks:
	case <-self.allowedConns:
		fsock, err = NewFSock(self.fsAddr, self.fsPasswd, self.reconnects, self.eventHandlers, self.eventFilters, self.logger, self.connId)
		if err != nil {
			return nil, err
		}
		return fsock, nil
	case <-time.After(self.maxWaitConn):
		return nil, ErrConnectionTimeout
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
	eventHandlers map[string][]func(string, string), eventFilters map[string]string, l *syslog.Writer, connId string) (*FSockPool, error) {
	pool := &FSockPool{connId: connId, fsAddr: fsaddr, fsPasswd: fspasswd, reconnects: reconnects, maxWaitConn: maxWaitConn,
		eventHandlers: eventHandlers, eventFilters: eventFilters, logger: l}
	pool.allowedConns = make(chan struct{}, maxFSocks)
	var emptyConn struct{}
	for i := 0; i < maxFSocks; i++ {
		pool.allowedConns <- emptyConn // Empty initiate so we do not need to wait later when we pop
	}
	pool.fSocks = make(chan *FSock, maxFSocks)
	return pool, nil
}
