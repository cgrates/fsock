/*
fsock.go is released under the MIT License <http://www.opensource.org/licenses/mit-license.php
Copyright (C) ITsysCOM. All Rights Reserved.

Provides FreeSWITCH socket communication.

*/

package fsock

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

// NewFSConn constructs and connects a FSConn
func NewFSConn(addr, passwd string,
	connIdx int,
	replyTimeout time.Duration,
	connErr chan error,
	lgr logger,
	evFilters map[string][]string,
	eventHandlers map[string][]func(string, int),
	bgapi bool,
) (*FSConn, error) {

	fsConn := &FSConn{
		connIdx:       connIdx,
		replyTimeout:  replyTimeout,
		lgr:           lgr,
		err:           connErr,
		replies:       make(chan string),
		eventHandlers: eventHandlers,
		bgapiChan:     make(map[string]chan string),
		bgapiMux:      new(sync.RWMutex),
	}

	// Build the TCP connection and the buffer reading it
	var err error
	if fsConn.conn, err = net.Dial("tcp", addr); err != nil {
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

	if err = fsConn.auth(passwd); err != nil { // Auth did not succeed
		return nil, err
	}

	if err = fsConn.filterEvents(evFilters, bgapi); err != nil {
		return nil, err
	}

	if err = fsConn.eventsPlain(getMapKeys(eventHandlers),
		bgapi); err != nil {
		return nil, err
	}

	go fsConn.readEvents() // Fork read events in it's own goroutine

	return fsConn, nil
}

type FSConn struct {
	connIdx       int                            // Identifier for the component using this instance of FSConn, optional
	replyTimeout  time.Duration                  // Timeout for awaiting replies
	conn          net.Conn                       // TCP connection to FreeSWITCH
	rdr           *bufio.Reader                  // Reader for the TCP connection
	lgr           logger                         // Logger for logging messages
	err           chan error                     // Channel for reporting errors
	replies       chan string                    // Channel for receiving replies
	eventHandlers map[string][]func(string, int) // eventStr, connId, handles events
	bgapiChan     map[string]chan string         // Channels used by bgapi
	bgapiMux      *sync.RWMutex                  // Protects the bgapiChan map
}

// readHeaders reads and parses the headers from a FreeSWITCH response.
func (fsConn *FSConn) readHeaders() (header string, err error) {
	bytesRead := make([]byte, 0) // buffer to accumulate header bytes
	var readLine []byte          // temporary slice to hold each line read

	for {
		if readLine, err = fsConn.rdr.ReadBytes('\n'); err != nil {
			fsConn.lgr.Err(fmt.Sprintf(
				"<FSock> Error reading headers: <%v>", err))
			fsConn.conn.Close() // close the connection regardless

			// Distinguish between different types of network errors to handle reconnection:
			// Return io.EOF (triggering a reconnect) if either:
			// - The error is not a network operation error (net.OpError)
			// - The connection timed out (opErr.Timeout())
			// - The connection was reset by peer (syscall.ECONNRESET)
			// For all other network operation errors, return the original error.
			var opErr *net.OpError
			if !errors.As(err, &opErr) || opErr.Timeout() ||
				errors.Is(opErr.Err, syscall.ECONNRESET) {
				return "", io.EOF
			}
			return "", err
		}

		// Check if the line is empty.
		if len(bytes.TrimSpace(readLine)) == 0 {
			// Empty line indicates the end of the headers, exit loop.
			break
		}
		bytesRead = append(bytesRead, readLine...)
	}
	return string(bytesRead), nil
}

// auth authenticates the connection with FreeSWITCH using the provided password.
func (fsConn *FSConn) auth(passwd string) (err error) {
	if err = fsConn.send("auth " + passwd + "\n\n"); err != nil {
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

// filterEvents will filter the Events coming from FreeSWITCH.
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

// send will send the content over the connection.
func (fsConn *FSConn) send(sendContent string) (err error) {
	if _, err = fsConn.conn.Write([]byte(sendContent)); err != nil {
		fsConn.lgr.Err(fmt.Sprintf("<FSock> Cannot write command to socket <%s>", err.Error()))
	}
	return
}

// eventsPlain will subscribe for events in plain mode.
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

// readEvent will read one Event from FreeSWITCH, made out of headers and body (if present).
func (fsConn *FSConn) readEvent() (header string, body string, err error) {
	if header, err = fsConn.readHeaders(); err != nil {
		return "", "", err
	}
	if !strings.Contains(header, "Content-Length") { //No body
		return header, "", nil
	}
	cl, err := strconv.Atoi(headerVal(header, "Content-Length"))
	if err != nil {
		return "", "", fmt.Errorf("invalid Content-Length header: %v", err)
	}
	if body, err = fsConn.readBody(cl); err != nil {
		return "", "", err
	}
	return header, body, nil
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

// readEvents continuously reads and processes events from the network buffer. It stops
// and exits the loop if an error is encountered, after sending it to fsConn.err.
func (fsConn *FSConn) readEvents() {
	for {
		hdr, body, err := fsConn.readEvent()

		// If an error occurs during the read operation, report
		// it on the error channel and exit the loop.
		if err != nil {
			fsConn.err <- err
			return
		}
		switch {
		case strings.Contains(hdr, "api/response"):
			// For API responses, send the body
			// directly to the replies channel.
			fsConn.replies <- body

		case strings.Contains(hdr, "command/reply"):
			// For command replies, extract the "Reply-Text" from
			// the header and send it to the replies channel.
			fsConn.replies <- headerVal(hdr, "Reply-Text")

		case body != "":
			// Could be an event, try dispatching it.
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
func (fsConn *FSConn) Send(payload string) (string, error) {
	if err := fsConn.send(payload); err != nil {
		return "", err
	}

	// Prepare a context based on fsConn.replyTimeout
	var ctx context.Context
	var cancel context.CancelFunc
	if fsConn.replyTimeout > 0 {
		ctx, cancel = context.WithTimeout(context.Background(), fsConn.replyTimeout)
	} else {
		ctx, cancel = context.WithCancel(context.Background())
	}
	defer cancel()

	replies := make(chan string)
	replyErrors := make(chan error)

	go func() {
		select {
		case reply := <-fsConn.replies:
			if strings.Contains(reply, "-ERR") {
				replyErrors <- errors.New(strings.TrimSpace(reply))
				return
			}
			replies <- reply
		case <-ctx.Done():
			replyErrors <- ctx.Err()
		}
	}()

	select {
	case reply := <-replies:
		return reply, nil
	case err := <-replyErrors:
		return "", err
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
	return fsConn.conn.Close()
}

// LocalAddr returns the local address of the connection
func (fsConn *FSConn) LocalAddr() net.Addr {
	return fsConn.conn.LocalAddr()
}
