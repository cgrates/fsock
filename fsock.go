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
	"log/syslog"
	"net"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

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

// successive Fibonacci numbers.
func fib() func() int {
	a, b := 0, 1
	return func() int {
		a, b = b, a+b
		return a
	}
}

var FS *fSock // Used to share FS connection via package globals

// Connection to FreeSWITCH Socket
type fSock struct {
	conn               net.Conn
	buffer             *bufio.Reader
	fsaddress, fspaswd string
	eventHandlers      map[string][]func(string)
	eventFilters       map[string]string
	apiChan, cmdChan   chan string
	reconnects         int
	delayFunc          func() int
	logger             *syslog.Writer
}

// Reads headers until delimiter reached
func (self *fSock) readHeaders() (s string, err error) {
	bytesRead := make([]byte, 0)
	var readLine []byte
	for {
		readLine, err = self.buffer.ReadBytes('\n')
		if err != nil {
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
func (self *fSock) readBody(ln int) (string, error) {
	bytesRead := make([]byte, ln)
	for i := 0; i < ln; i++ {
		if readByte, err := self.buffer.ReadByte(); err != nil {
			return "", err
		} else { // No Error, add received to localread buffer
			bytesRead[i] = readByte // Add received line to the local read buffer
		}
	}
	return string(bytesRead), nil
}

// Event is made out of headers and body (if present)
func (self *fSock) readEvent() (string, string, error) {
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

// Checks if socket connected. Can be extended with pings
func (self *fSock) Connected() bool {
	if self.conn == nil {
		return false
	}
	return true
}

// Disconnects from socket
func (self *fSock) Disconnect() (err error) {
	if self.conn != nil {
		err = self.conn.Close()
	}
	return
}

// Auth to FS
func (self *fSock) auth() error {
	authCmd := fmt.Sprintf("auth %s\n\n", self.fspaswd)
	fmt.Fprint(self.conn, authCmd)
	if rply, err := self.readHeaders(); err != nil || !strings.Contains(rply, "Reply-Text: +OK accepted") {
		fmt.Println("Got reply to auth:", rply)
		return errors.New("auth error")
	}
	return nil
}

// Subscribe to events
func (self *fSock) eventsPlain(events []string) error {
	if len(events) == 0 {
		return nil
	}
	eventsCmd := "event plain"
	for _, ev := range events {
		if ev == "ALL" {
			eventsCmd = "event plain all"
			break
		}
		eventsCmd += " " + ev
	}
	eventsCmd += "\n\n"
	fmt.Fprint(self.conn, eventsCmd) // Send command here
	if rply, err := self.readHeaders(); err != nil || !strings.Contains(rply, "Reply-Text: +OK") {
		return errors.New("event error")
	}
	return nil
}

// Enable filters
func (self *fSock) filterEvents(filters map[string]string) error {
	if len(filters) == 0 { //Nothing to filter
		return nil
	}
	cmd := "filter"
	for hdr, val := range filters {
		cmd += " " + hdr + " " + val
	}
	cmd += "\n\n"
	fmt.Fprint(self.conn, cmd)
	if rply, err := self.readHeaders(); err != nil || !strings.Contains(rply, "Reply-Text: +OK") {
		return errors.New("filter error")
	}
	return nil
}

// Connect or reconnect
func (self *fSock) Connect() error {
	if self.Connected() {
		self.Disconnect()
	}
	var conErr error
	for i := 0; i < self.reconnects; i++ {
		self.conn, conErr = net.Dial("tcp", self.fsaddress)
		if conErr == nil {
			self.logger.Info("Successfully connected to FreeSWITCH")
			// Connected, init buffer, auth and subscribe to desired events and filters
			self.buffer = bufio.NewReaderSize(self.conn, 8192) // reinit buffer
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
			return nil
		}
		time.Sleep(time.Duration(self.delayFunc()) * time.Second)
	}
	return conErr
}

// Send API command
func (self *fSock) SendApiCmd(cmdStr string) error {
	if !self.Connected() {
		return errors.New("Not connected to FS")
	}
	cmd := fmt.Sprintf("api %s\n\n", cmdStr)
	fmt.Fprint(self.conn, cmd)
	resEvent := <-self.apiChan
	if strings.Contains(resEvent, "-ERR") {
		return errors.New("Command failed")
	}
	return nil
}

// SendMessage command
func (self *fSock) SendMsgCmd(uuid string, cmdargs map[string]string) error {
	if len(cmdargs) == 0 {
		return errors.New("Need command arguments")
	}
	if !self.Connected() {
		return errors.New("Not connected to FS")
	}
	argStr := ""
	for k, v := range cmdargs {
		argStr += fmt.Sprintf("%s:%s\n", k, v)
	}
	fmt.Fprint(self.conn, fmt.Sprintf("sendmsg %s\n%s\n", uuid, argStr))
	replyTxt := <-self.cmdChan
	if strings.HasPrefix(replyTxt, "-ERR") {
		return fmt.Errorf("SendMessage: %s", replyTxt)
	}
	return nil
}

// Reads events from socket
func (self *fSock) ReadEvents() {
	// Read events from buffer, firing them up further
	for {
		hdr, body, err := self.readEvent()
		if err != nil {
			self.logger.Warning("FreeSWITCH connection broken: attemting reconnect")
			connErr := self.Connect()
			if connErr != nil {
				return
			}
			continue // Connection reset
		}
		if strings.Contains(hdr, "api/response") {
			self.apiChan <- hdr + body
		} else if strings.Contains(hdr, "command/reply") {
			self.cmdChan <- headerVal(hdr, "Reply-Text")
		}
		if body != "" { // We got a body, could be event, try dispatching it
			self.dispatchEvent(body)
		}
	}

	return
}

// Dispatch events to handlers in async mode
func (self *fSock) dispatchEvent(event string) {
	eventName := headerVal(event, "Event-Name")
	if _, hasHandlers := self.eventHandlers[eventName]; hasHandlers {
		// We have handlers, dispatch to all of them
		for _, handlerFunc := range self.eventHandlers[eventName] {
			go handlerFunc(event)
		}
	}
}



// Connects to FS and starts buffering input
func NewFSock(fsaddr, fspaswd string, reconnects int, eventHandlers map[string][]func(string), eventFilters map[string]string, l *syslog.Writer) (*fSock, error) {
	fsock := fSock{fsaddress: fsaddr, fspaswd: fspaswd, eventHandlers: eventHandlers, eventFilters: eventFilters, reconnects: reconnects, logger: l}
	fsock.apiChan = make(chan string) // Init apichan so we can use it to pass api replies
	fsock.cmdChan = make(chan string)
	fsock.delayFunc = fib()
	errConn := fsock.Connect()
	if errConn != nil {
		return nil, errConn
	}
	return &fsock, nil
}
