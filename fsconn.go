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
	"strconv"
	"strings"
	"sync"
	"time"
)

// NewFSConn constructs and connects a FSConn.
func NewFSConn(fsAddr, fsPasswd string, connIdx int,
	lgr logger, replyTimeout time.Duration,
	evFilters map[string][]string,
	eventHandlers map[string][]func(string, int),
	bgapi bool, fsClosedChan chan struct{},
) (fsConn *FSConn, err error) {

	fsConn = &FSConn{
		connIdx:        connIdx,
		lgr:            lgr,
		replyTimeout:   replyTimeout,
		eventHandlers:  eventHandlers,
		repliesChan:    make(chan string),
		errorsChan:     make(chan error),
		stopReadEvents: make(chan struct{}),
		fsClosedChan:   fsClosedChan,
		bgapiChan:      make(map[string]chan string),
		bgapiMux:       new(sync.RWMutex),
	}

	// Build the TCP connection and the buffer reading it
	if fsConn.conn, err = net.Dial("tcp", fsAddr); err != nil {
		fsConn.lgr.Err(fmt.Sprintf("<FSock> Attempt to connect to FreeSWITCH, received: %s", err.Error()))
		return nil, err
	}
	fsConn.rdr = bufio.NewReaderSize(fsConn.conn, 8192) // reinit buffer
	fsConn.lgr.Info("<FSock> Successfully connected to FreeSWITCH!")

	// Connected, auth and subscribe to desired events and filters
	var authChlng string
	if authChlng, err = fsConn.readHeaders(); err != nil {
		return nil, err
	}

	if !strings.Contains(authChlng, "auth/request") {
		fsConn.conn.Close()
		return nil, errors.New("no auth challenge received")
	}

	if err = fsConn.auth(fsPasswd); err != nil { // Auth did not succeed
		return
	}

	if err = fsConn.filterEvents(evFilters, bgapi); err != nil {
		return
	}

	if err = fsConn.eventsPlain(getMapKeys(eventHandlers),
		bgapi); err != nil {
		return
	}

	go fsConn.readEvents() // Fork read events in its own goroutine

	return

}

// FSConn represents one connection to FreeSWITCH
type FSConn struct {
	connIdx        int // Identifier for the component using this instance of FSock, optional
	conn           net.Conn
	rdr            *bufio.Reader
	lgr            logger
	replyTimeout   time.Duration
	repliesChan    chan string
	errorsChan     chan error
	stopReadEvents chan struct{}
	fsClosedChan   chan struct{}
	eventHandlers  map[string][]func(string, int) // eventStr, connId
	bgapiChan      map[string]chan string         // used by bgapi
	bgapiMux       *sync.RWMutex                  // protects the bgapiChan map
}

// readHeaders reads the headers part out of FreeSWITCH answer
func (fsConn *FSConn) readHeaders() (header string, err error) {
	bytesRead := make([]byte, 0)
	var readLine []byte

	for {
		if readLine, err = fsConn.rdr.ReadBytes('\n'); err != nil {
			fsConn.lgr.Err(fmt.Sprintf("<FSock> Error reading headers: <%s>", err.Error()))
			fsConn.conn.Close() // any read error leads to disconnection
			return "", io.EOF
		}
		// No Error, add received to localread buffer
		if len(bytes.TrimSpace(readLine)) == 0 {
			break
		}
		bytesRead = append(bytesRead, readLine...)
	}
	return string(bytesRead), nil
}

// Auth to FS
func (fsConn *FSConn) auth(fsPasswd string) (err error) {
	if err = fsConn.send("auth " + fsPasswd + "\n\n"); err != nil {
		fsConn.conn.Close()
		return
	}
	var rply string
	if rply, err = fsConn.readHeaders(); err != nil {
		return
	}
	if !strings.Contains(rply, "Reply-Text: +OK accepted") {
		fsConn.conn.Close()
		return fmt.Errorf("unexpected auth reply received: <%s>", rply)
	}
	return
}

// filterEvents will filter the Events coming from FreeSWITCH
func (fsConn *FSConn) filterEvents(filters map[string][]string, bgapi bool) (err error) {
	if len(filters) == 0 {
		return nil
	}
	if bgapi {
		filters["Event-Name"] = append(filters["Event-Name"], "BACKGROUND_JOB") // for bgapi
	}
	for hdr, vals := range filters {
		for _, val := range vals {
			if err = fsConn.send("filter " + hdr + " " + val + "\n\n"); err != nil {
				fsConn.lgr.Err(fmt.Sprintf("<FSock> Error filtering events: <%s>", err.Error()))
				fsConn.conn.Close()
				return
			}
			var rply string
			if rply, err = fsConn.readHeaders(); err != nil {
				return
			}
			if !strings.Contains(rply, "Reply-Text: +OK") {
				fsConn.conn.Close()
				return fmt.Errorf(`unexpected filter-events reply received: <%s>`, rply)
			}
		}
	}
	return nil
}

// send will send the content over the connection
func (fsConn *FSConn) send(payload string) (err error) {
	if _, err = fsConn.conn.Write([]byte(payload)); err != nil {
		fsConn.lgr.Err(fmt.Sprintf("<FSock> Cannot write command to socket <%s>", err.Error()))
	}
	return
}

// eventsPlain will subscribe for events in plain mode
func (fsConn *FSConn) eventsPlain(events []string, bgapi bool) (err error) {
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
		if bgapi {
			eventsCmd += " BACKGROUND_JOB" // For bgapi
		}
		if len(customEvents) != 0 { // Add CUSTOM events subscribing in the end otherwise unexpected events are received
			eventsCmd += " " + "CUSTOM" + customEvents
		}
	}

	if err = fsConn.send(eventsCmd + "\n\n"); err != nil {
		fsConn.conn.Close()
		return
	}
	var rply string
	if rply, err = fsConn.readHeaders(); err != nil {
		return
	}
	if !strings.Contains(rply, "Reply-Text: +OK") {
		fsConn.conn.Close()
		return fmt.Errorf("unexpected events-subscribe reply received: <%s>", rply)
	}
	return
}

// readEvent will read one Event from FreeSWITCH, made out of headers and body (if present)
func (fsConn *FSConn) readEvent() (header string, body string, err error) {
	if header, err = fsConn.readHeaders(); err != nil {
		return
	}
	if !strings.Contains(header, "Content-Length") { //No body
		return
	}
	var cl int
	if cl, err = strconv.Atoi(headerVal(header, "Content-Length")); err != nil {
		err = fmt.Errorf("cannot extract content length, err: <%s>", err)
		return
	}
	body, err = fsConn.readBody(cl)
	return
}

// readBody reads the specified number of bytes from the buffer.
// The number of bytes to read is given by 'noBytes', which is determined from the content-length header.
func (fsConn *FSConn) readBody(noBytes int) (string, error) {
	bytesRead := make([]byte, noBytes)
	_, err := io.ReadFull(fsConn.rdr, bytesRead)
	if err != nil {
		fsConn.lgr.Err(fmt.Sprintf("<FSock> Error reading message body: <%v>", err))
		fsConn.conn.Close()
		return "", io.EOF // Return io.EOF to trigger ReconnectIfNeeded.
	}
	return string(bytesRead), nil
}

// Read events from network buffer, stop when exitChan is closed, report on errorsChan on error and exit
// Receive exitChan and errorsChan as parameters so we avoid concurrency on using fs.
func (fsConn *FSConn) readEvents() {
	for {
		select {
		case <-fsConn.stopReadEvents:
			return
		default: // Unlock waiting here
		}
		hdr, body, err := fsConn.readEvent()
		if err != nil {
			if err == io.EOF {
				close(fsConn.fsClosedChan)
				return // not pushing the error since most probably nobody listens and the channel can be blocked
			}
			fsConn.errorsChan <- err
			return
		}
		if strings.Contains(hdr, "api/response") {
			fsConn.repliesChan <- body
		} else if strings.Contains(hdr, "command/reply") {
			fsConn.repliesChan <- headerVal(hdr, "Reply-Text")
		} else if body != "" { // We got a body, could be event, try dispatching it
			fsConn.dispatchEvent(body)
		}
	}
}

// Dispatch events to handlers in async mode
func (fsConn *FSConn) dispatchEvent(event string) {
	eventName := headerVal(event, "Event-Name")
	if eventName == "BACKGROUND_JOB" { // for bgapi BACKGROUND_JOB
		go fsConn.doBackgroundJob(event)
		return
	}

	if eventName == "CUSTOM" {
		eventSubclass := headerVal(event, "Event-Subclass")
		if len(eventSubclass) != 0 {
			eventName += " " + urlDecode(eventSubclass)
		}
	}

	for _, handleName := range []string{eventName, "ALL"} {
		if _, hasHandlers := fsConn.eventHandlers[handleName]; hasHandlers {
			// We have handlers, dispatch to all of them
			for _, handlerFunc := range fsConn.eventHandlers[handleName] {
				go handlerFunc(event, fsConn.connIdx)
			}
			return
		}
	}
	fsConn.lgr.Warning(fmt.Sprintf("<FSock> No dispatcher for event: <%+v> with event name: %s", event, eventName))
}

// bgapi event lisen fuction
func (fsConn *FSConn) doBackgroundJob(event string) { // add mutex protection
	evMap := EventToMap(event)
	jobUUID, has := evMap["Job-UUID"]
	if !has {
		fsConn.lgr.Err("<FSock> BACKGROUND_JOB with no Job-UUID")
		return
	}

	fsConn.bgapiMux.Lock()
	defer fsConn.bgapiMux.Unlock()
	var out chan string
	out, has = fsConn.bgapiChan[jobUUID]
	if !has {
		fsConn.lgr.Err(fmt.Sprintf("<FSock> BACKGROUND_JOB with UUID %s lost!", jobUUID))
		return // not a requested bgapi
	}

	delete(fsConn.bgapiChan, jobUUID)
	out <- evMap[EventBodyTag]
}

// Send will send the content over the connection, exposing synchronous interface outside
func (fsConn *FSConn) Send(payload string) (reply string, err error) {
	if err = fsConn.send(payload); err != nil {
		return "", err
	}
	select {
	case reply = <-fsConn.repliesChan:
		if strings.Contains(reply, "-ERR") {
			return "", errors.New(strings.TrimSpace(reply))
		}
		return reply, nil
	case err := <-fsConn.errorsChan:
		return "", err
	case <-time.After(fsConn.replyTimeout):
		return "", errors.New("timeout waiting for FreeSWITCH reply")
	}
}

// Send BGAPI command
func (fsConn *FSConn) SendBgapiCmd(cmdStr string) (out chan string, err error) {
	jobUUID := genUUID()
	out = make(chan string)

	fsConn.bgapiMux.Lock()
	fsConn.bgapiChan[jobUUID] = out
	fsConn.bgapiMux.Unlock()

	if _, err = fsConn.Send("bgapi " + cmdStr + "\nJob-UUID:" + jobUUID + "\n\n"); err != nil {
		return nil, err
	}
	return
}

// Disconnect will disconnect the fsConn from FreeSWITCH
func (fsConn *FSConn) Disconnect() error {
	close(fsConn.stopReadEvents)
	return fsConn.conn.Close()
}

// LocalAddr returns the local address of the connection
func (fsConn *FSConn) LocalAddr() net.Addr {
	return fsConn.conn.LocalAddr()
}
