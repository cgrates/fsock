/*
fsock_test.go is released under the MIT License <http://www.opensource.org/licenses/mit-license.php
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
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"
)

const (
	HEADER = `Content-Length: 564
Content-Type: text/event-plain

`
	BODY = `Event-Name: RE_SCHEDULE
Core-UUID: 792e181c-b6e6-499c-82a1-52a778e7d82d
FreeSWITCH-Hostname: h1.cgrates.org
FreeSWITCH-Switchname: h1.cgrates.org
FreeSWITCH-IPv4: 172.16.16.16
FreeSWITCH-IPv6: %3A%3A1
Event-Date-Local: 2012-10-05%2013%3A41%3A38
Event-Date-GMT: Fri,%2005%20Oct%202012%2011%3A41%3A38%20GMT
Event-Date-Timestamp: 1349437298012866
Event-Calling-File: switch_scheduler.c
Event-Calling-Function: switch_scheduler_execute
Event-Calling-Line-Number: 65
Event-Sequence: 34263
Task-ID: 2
Task-Desc: heartbeat
Task-Group: core
Task-Runtime: 1349437318

extra data
`
)

func TestHeaders(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Error("Error creating pipe!")
	}
	fs := &FSConn{}
	fs.rdr = bufio.NewReader(r)
	w.Write([]byte(HEADER))
	h, err := fs.readHeaders()
	if err != nil || h != "Content-Length: 564\nContent-Type: text/event-plain\n" {
		t.Error("Error parsing headers: ", h, err)
	}
}

func TestEvent(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Error("Error creating pype!")
	}
	fs := &FSConn{}
	fs.rdr = bufio.NewReader(r)
	w.Write([]byte(HEADER + BODY))
	h, b, err := fs.readEvent()
	if err != nil || h != HEADER[:len(HEADER)-1] || len(b) != 564 {
		t.Error("Error parsing event: ", h, b, len(b))
	}
}

func TestReadEvents(t *testing.T) {
	data, err := os.ReadFile("test_data.txt")
	if err != nil {
		t.Error("Error reading test data file!")
	}
	r, w, err := os.Pipe()
	if err != nil {
		t.Error("Error creating pipe!")
	}
	funcMutex := new(sync.RWMutex)
	var events int32
	evfunc := func(string, int) {
		funcMutex.Lock()
		events++
		funcMutex.Unlock()
	}

	fs := &FSConn{}
	fs.lgr = nopLogger{}
	fs.rdr = bufio.NewReader(r)
	fs.eventHandlers = map[string][]func(string, int){
		"HEARTBEAT":                {evfunc},
		"RE_SCHEDULE":              {evfunc},
		"CHANNEL_STATE":            {evfunc},
		"CODEC":                    {evfunc},
		"CHANNEL_CREATE":           {evfunc},
		"CHANNEL_CALLSTATE":        {evfunc},
		"API":                      {evfunc},
		"CHANNEL_EXECUTE":          {evfunc},
		"CHANNEL_EXECUTE_COMPLETE": {evfunc},
		"CHANNEL_PARK":             {evfunc},
		"CHANNEL_HANGUP":           {evfunc},
		"CHANNEL_HANGUP_COMPLETE":  {evfunc},
		"CHANNEL_UNPARK":           {evfunc},
		"CHANNEL_DESTROY":          {evfunc},
	}
	go fs.readEvents()
	w.Write(data)
	time.Sleep(50 * time.Millisecond)
	funcMutex.RLock()
	if events != 45 {
		t.Error("Error reading events: ", events)
	}
	funcMutex.RUnlock()
}

func TestFSockConnect(t *testing.T) {
	fs := &FSock{
		mu:            new(sync.RWMutex),
		eventHandlers: make(map[string][]func(string, int)),
		eventFilters:  make(map[string][]string),
		logger:        nopLogger{},
	}

	err := fs.Connect()
	if err == nil {
		t.Fatal("Expected non-nil error")
	}

}

type connMock struct{}

func (cM *connMock) Close() error {
	return nil
}

func (cM *connMock) LocalAddr() net.Addr {
	return nil
}

func (cM *connMock) RemoteAddr() net.Addr {
	return nil
}

func (cM *connMock) Read(b []byte) (n int, err error) {
	return 0, nil
}

func (cM *connMock) Write(b []byte) (n int, err error) {
	return 0, ErrConnectionPoolTimeout
}

func (cM *connMock) SetDeadline(t time.Time) error {
	return nil
}

func (cM *connMock) SetReadDeadline(t time.Time) error {
	return nil
}

func (cM *connMock) SetWriteDeadline(t time.Time) error {
	return nil
}

type connMock2 struct {
	buf *bytes.Buffer
}

func (cM *connMock2) Close() error {
	return nil
}

func (cM *connMock2) LocalAddr() net.Addr {
	return nil
}

func (cM *connMock2) RemoteAddr() net.Addr {
	return nil
}

func (cM *connMock2) Read(b []byte) (n int, err error) {
	return 0, nil
}

func (cM *connMock2) Write(b []byte) (n int, err error) {
	return cM.buf.Write(b)
}

func (cM *connMock2) SetDeadline(t time.Time) error {
	return nil
}

func (cM *connMock2) SetReadDeadline(t time.Time) error {
	return nil
}

func (cM *connMock2) SetWriteDeadline(t time.Time) error {
	return nil
}

type connMock3 struct{}

func (cM *connMock3) Close() error {
	return nil
}

func (cM *connMock3) LocalAddr() net.Addr {
	return nil
}

func (cM *connMock3) RemoteAddr() net.Addr {
	return nil
}

func (cM *connMock3) Read(b []byte) (n int, err error) {
	return 0, nil
}

func (cM *connMock3) Write(b []byte) (n int, err error) {
	return 0, nil
}

func (cM *connMock3) SetDeadline(t time.Time) error {
	return nil
}

func (cM *connMock3) SetReadDeadline(t time.Time) error {
	return nil
}

func (cM *connMock3) SetWriteDeadline(t time.Time) error {
	return nil
}
func TestFSockSend(t *testing.T) {
	fs := &FSConn{
		lgr:  nopLogger{},
		conn: new(connMock),
	}

	expected := ErrConnectionPoolTimeout
	err := fs.send("testString")

	if err == nil || err != expected {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", expected, err)
	}
}

func TestFSockAuthFailSend(t *testing.T) {

	fs := FSConn{
		lgr:  nopLogger{},
		conn: new(connMock),
	}
	err := fs.auth("")

	if err == nil || err != ErrConnectionPoolTimeout {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", ErrConnectionPoolTimeout, err)
	}
}

func TestFSockAuthFailReply(t *testing.T) {
	buf := new(bytes.Buffer)
	fs := &FSConn{
		conn: &connMock2{buf: buf},
		rdr:  bufio.NewReader(bytes.NewBuffer([]byte("Reply-Text: +OK accepted\n\n"))),
		lgr:  new(nopLogger),
	}

	expected := fmt.Sprintf("unexpected auth reply received: <%s>", strings.TrimSuffix(HEADER, "\n"))
	err := fs.auth("test")
	if err != nil {
		t.Fatal(err)
	}

	expectedbuf := "auth test\n\n"
	if rcv := buf.String(); rcv != expectedbuf {
		t.Errorf("\nReceived: %q, \nExpected: %q", rcv, expectedbuf)
	}

	buf.Reset()
	fs.rdr = bufio.NewReader(bytes.NewBuffer([]byte(HEADER)))
	err = fs.auth("test")

	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", expected, err.Error())
	}

	if rcv := buf.String(); rcv != expectedbuf {
		t.Errorf("\nExpected: %q, \nReceived: %q", expectedbuf, rcv)
	}
}

func TestFSockAuthFailRead(t *testing.T) {
	fs := &FSConn{
		rdr:  bufio.NewReader(bytes.NewBuffer([]byte("Reply-Text: +OK accepted"))),
		lgr:  new(nopLogger),
		conn: new(connMock3),
	}
	expected := io.EOF
	err := fs.auth("test")

	if err == nil || err != expected {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", expected, err)
	}
}

func TestFSockSendBgapiCmdNonNilErr(t *testing.T) {
	fs := &FSock{
		mu:        &sync.RWMutex{},
		delayFunc: fibDuration,
	}

	expected := "not connected to FreeSWITCH"
	_, err := fs.SendBgapiCmd("test")

	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", expected, err)
	}
}

// fibDuration returns successive Fibonacci numbers converted to time.Duration.
func fibDuration(durationUnit, maxDuration time.Duration) func() time.Duration {
	a, b := 0, 1
	return func() time.Duration {
		a, b = b, a+b
		fibNrAsDuration := time.Duration(a) * durationUnit
		if maxDuration > 0 && maxDuration < fibNrAsDuration {
			return maxDuration
		}
		return fibNrAsDuration
	}
}

func TestFSockSendMsgCmdWithBodyEmptyArguments(t *testing.T) {
	fs := &FSock{}
	uuid := ""
	cmdargs := make(map[string]string)
	body := ""

	expected := "need command arguments"
	err := fs.SendMsgCmdWithBody(uuid, cmdargs, body)

	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", expected, err)
	}
}

func TestFSockSendMsgCmd(t *testing.T) {
	fs := &FSock{}
	uuid := "testID"
	cmdargs := make(map[string]string)

	expected := "need command arguments"
	err := fs.SendMsgCmd(uuid, cmdargs)

	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", expected, err)
	}
}

func TestFSockLocalAddrNotConnected(t *testing.T) {
	fs := &FSock{
		mu: &sync.RWMutex{},
	}
	addr := fs.LocalAddr()
	if addr != nil {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", nil, addr)
	}
}

func TestFSockReadEvents(t *testing.T) {
	fs := &FSock{
		mu:        &sync.RWMutex{},
		delayFunc: fibDuration,
	}

	expected := "not connected to FreeSWITCH"
	err := fs.reconnectIfNeeded()
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", expected, err)
	}
}

func TestFSockReadBody(t *testing.T) {
	fs := &FSConn{
		conn: new(connMock),
		lgr:  nopLogger{},
		rdr:  bufio.NewReader(bytes.NewBuffer([]byte(""))),
	}
	if rply, err := fs.readBody(2); err == nil || err != io.EOF {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", io.EOF, err)
	} else if rply != "" {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", "", rply)
	}
}

func TestFSockSendCmdErrSend(t *testing.T) {

	fs := &FSConn{
		lgr:  nopLogger{},
		conn: &connMock{},
	}
	rply, err := fs.Send("test")

	if err == nil || err != ErrConnectionPoolTimeout {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", ErrConnectionPoolTimeout, err)
	}

	if rply != "" {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", "", rply)
	}
}

func TestFSockSendCmdErrContains(t *testing.T) {
	fs := &FSConn{
		lgr:     nopLogger{},
		conn:    &connMock3{},
		replies: make(chan string, 1),
	}

	fs.replies <- "test-ERR"

	expected := "test-ERR"
	if rply, err := fs.Send("test"); err == nil || err.Error() != expected {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", expected, err)
	} else if rply != "" {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", "", rply)
	}

}

func TestFSockReconnectIfNeeded(t *testing.T) {
	fs := &FSock{
		mu:         &sync.RWMutex{},
		logger:     nopLogger{},
		reconnects: 2,
		delayFunc:  fibDuration,
	}

	expected := "dial tcp: missing address"
	err := fs.ReconnectIfNeeded()

	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", expected, err)
	}
}

func TestFSockSendMsgCmdWithBody(t *testing.T) {
	fs := &FSock{
		mu:        &sync.RWMutex{},
		delayFunc: fibDuration,
	}
	uuid := "testID"
	cmdargs := map[string]string{
		"testKey": "testValue",
	}
	body := "testBody"

	expected := "not connected to FreeSWITCH"
	err := fs.SendMsgCmdWithBody(uuid, cmdargs, body)

	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", expected, err)
	}
}

func TestFSockLocalAddr(t *testing.T) {
	fs := &FSock{
		mu: &sync.RWMutex{},
	}
	addr := fs.LocalAddr()
	if addr != nil {
		t.Errorf("\nExpected nil, got %v", addr)
	}
}

func TestFSockreadEvent(t *testing.T) {
	fs := &FSConn{
		rdr: bufio.NewReader(bytes.NewBuffer([]byte("Content-Length\n\n"))),
		lgr: nopLogger{},
	}

	expected := fmt.Sprintf("cannot extract content length, err: <%s>", "strconv.Atoi: parsing \"\": invalid syntax")
	exphead := "Content-Length\n"
	expbody := ""
	if head, body, err := fs.readEvent(); err == nil || err.Error() != expected {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", expected, err)
	} else if head != exphead {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", exphead, head)
	} else if body != expbody {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", expbody, body)
	}
}

func TestFSockeventsPlainErrSend(t *testing.T) {
	fs := &FSConn{
		conn: &connMock{},
		lgr:  nopLogger{},
	}
	events := []string{""}

	expected := ErrConnectionPoolTimeout
	err := fs.eventsPlain(events, true)

	if err == nil || err != expected {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", expected, err)
	}
}

func TestFSockeventsPlainErrRead(t *testing.T) {
	fs := &FSConn{

		conn: &connMock3{},
		lgr:  nopLogger{},
		rdr:  bufio.NewReader(bytes.NewBuffer([]byte("test\n"))),
	}
	events := []string{"ALL"}

	expected := io.EOF
	if err := fs.eventsPlain(events, true); err == nil || err != expected {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", expected, err)
	}
}

func TestFSockeventsPlainUnexpectedReply(t *testing.T) {
	fs := &FSConn{
		conn: &connMock3{},
		lgr:  nopLogger{},
		rdr:  bufio.NewReader(bytes.NewBuffer([]byte("test\n\n"))),
	}
	events := []string{"CUSTOMtest"}

	expected := fmt.Sprintf("unexpected events-subscribe reply received: <%s>", "test\n")
	if err := fs.eventsPlain(events, true); err == nil || err.Error() != expected {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", expected, err)
	}
}

func TestFSockfilterEventsUnexpectedReply(t *testing.T) {
	fs := &FSConn{
		conn: &connMock3{},
		rdr:  bufio.NewReader(bytes.NewBuffer([]byte("test\n\n"))),
		lgr:  nopLogger{},
	}
	filters := map[string][]string{
		"Event-Name": nil,
	}

	expected := fmt.Sprintf("unexpected filter-events reply received: <%s>", "test\n")
	err := fs.filterEvents(filters, true)

	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", expected, err)
	}
}

func TestFSockfilterEventsErrRead(t *testing.T) {
	fs := &FSConn{
		conn: &connMock3{},
		rdr:  bufio.NewReader(bytes.NewBuffer([]byte("test\n"))),
		lgr:  nopLogger{},
	}
	filters := map[string][]string{
		"Event-Name": nil,
	}

	expected := io.EOF
	if err := fs.filterEvents(filters, true); err == nil || err != expected {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", expected, err)
	}
}

func TestFSockfilterEventsErrSend(t *testing.T) {
	fs := &FSConn{

		conn: &connMock{},
		rdr:  bufio.NewReader(bytes.NewBuffer([]byte("test\n\n"))),
		lgr:  nopLogger{},
	}
	filters := map[string][]string{
		"Event-Name": nil,
	}

	expected := ErrConnectionPoolTimeout
	if err := fs.filterEvents(filters, true); err == nil || err != expected {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", expected, err)
	}
}

func TestFSockfilterEventsErrNil(t *testing.T) {
	fs := &FSConn{
		conn: &connMock3{},
		rdr:  bufio.NewReader(bytes.NewBuffer([]byte("testReply-Text: +OK\n\n"))),
		lgr:  nopLogger{},
	}
	filters := map[string][]string{
		"Event-Name": nil,
	}

	if err := fs.filterEvents(filters, true); err != nil {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", nil, err)
	}
}

type loggerMock struct {
	msgType, msg string
}

func (lM *loggerMock) Alert(string) error {
	return nil
}

func (lM *loggerMock) Close() error {
	return nil
}

func (lM *loggerMock) Crit(string) error {
	return nil
}

func (lM *loggerMock) Debug(string) error {
	return nil
}

func (lM *loggerMock) Emerg(string) error {
	return nil
}

func (lM *loggerMock) Err(s string) error {
	lM.msgType = "error"
	lM.msg = s
	return nil
}

func (lM *loggerMock) Info(string) error {
	return nil
}

func (lM *loggerMock) Notice(string) error {
	return nil
}

func (lM *loggerMock) Warning(event string) error {
	lM.msgType = "warning"
	lM.msg = event
	return nil
}

func TestFSockdispatchEvent(t *testing.T) {
	l := &loggerMock{}
	fs := &FSConn{
		lgr: l,
	}
	event := "Event-Name: CUSTOM\n"
	event += "Event-Subclass: test"

	expected := fmt.Sprintf("<FSock> No dispatcher for event: <%+v> with event name: %s", event, "CUSTOM test")
	fs.dispatchEvent(event)

	if l.msgType != "warning" {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", "warning", l.msgType)
	} else if l.msg != expected {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", expected, l.msg)
	}
}

func TestFSockdoBackgroundJobLogErr1(t *testing.T) {
	l := &loggerMock{}
	fs := &FSConn{
		lgr: l,
	}
	event := "test"
	expected := "<FSock> BACKGROUND_JOB with no Job-UUID"
	fs.doBackgroundJob(event)

	if l.msgType != "error" {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", "error", l.msgType)
	} else if l.msg != expected {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", expected, l.msg)
	}
}

func TestFSockdoBackgroundJobLogErr2(t *testing.T) {
	l := &loggerMock{}
	fs := &FSConn{
		bgapiMux: &sync.RWMutex{},
		lgr:      l,
	}
	event := "Event-Name: CUSTOM\n"
	event += "Event-Subclass: test\n"
	event += "Job-UUID: testID"

	expected := fmt.Sprintf("<FSock> BACKGROUND_JOB with UUID %s lost!", "testID")
	fs.doBackgroundJob(event)

	if l.msgType != "error" {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", "error", l.msgType)
	} else if l.msg != expected {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", expected, l.msg)
	}
}

func TestFSockNewFSockPool(t *testing.T) {
	fsaddr := "testAddr"
	fspw := "testPw"
	reconns := 2
	connIdx := 0
	maxFSocks := 1

	var maxWait time.Duration
	chanErr := make(chan error, 1)
	evHandlers := make(map[string][]func(string, int))
	evFilters := make(map[string][]string)

	fspool := &FSockPool{
		connIdx:       connIdx,
		addr:          fsaddr,
		passwd:        fspw,
		reconnects:    reconns,
		maxWaitConn:   maxWait,
		replyTimeout:  5 * time.Second,
		eventHandlers: evHandlers,
		eventFilters:  evFilters,
		bgapi:         true,
		logger:        nopLogger{},
		allowedConns:  nil,
		fSocks:        nil,
		stopError:     chanErr,
	}
	fsnew := NewFSockPool(maxFSocks, fsaddr, fspw, reconns, maxWait, 0, 5*time.Second, fibDuration, evHandlers, evFilters, nil, connIdx, true, chanErr)
	fsnew.allowedConns = nil
	fsnew.fSocks = nil
	fsnew.delayFuncConstructor = nil

	if !reflect.DeepEqual(fspool, fsnew) {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", fspool, fsnew)
	}
}

func TestFSockPushFSockAllowedConns(t *testing.T) {
	var fs *FSockPool
	var fsk *FSock
	fs.PushFSock(fsk)

	fs = &FSockPool{
		allowedConns: make(chan struct{}, 3),
	}

	fs.PushFSock(fsk)
	if len(fs.allowedConns) != 1 {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", 1, len(fs.allowedConns))
	}
}

func TestFSockPushFSock(t *testing.T) {
	fs := &FSockPool{
		allowedConns: make(chan struct{}, 1),
		fSocks:       make(chan *FSock, 1),
	}
	fsConn := &FSConn{
		conn: &connMock{},
	}
	fsk := &FSock{
		fsConn: fsConn,
		mu:     &sync.RWMutex{},
	}
	fs.PushFSock(fsk)
	if len(fs.fSocks) != 1 {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", 1, len(fs.fSocks))
	} else if rcv := <-fs.fSocks; !reflect.DeepEqual(rcv, fsk) {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", fsk, rcv)
	}
}

func TestFSockPopFSockEmpty(t *testing.T) {
	var fs *FSockPool

	expected := "unconfigured connection pool"
	fsk, err := fs.PopFSock()

	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", expected, err)
	} else if fs != nil {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", nil, fsk)
	}
}

func TestFSockPopFSock2(t *testing.T) {
	fs := &FSockPool{
		fSocks: make(chan *FSock, 1),
	}

	expected := &FSock{}
	fs.fSocks <- expected
	if fsock, err := fs.PopFSock(); err != nil {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", nil, err)
	} else if fsock != expected { // the pointer should be the same
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", expected, fsock)
	}
}

func TestFSockPopFSockTimeout(t *testing.T) {
	fs := &FSockPool{}

	expected := ErrConnectionPoolTimeout
	if fsk, err := fs.PopFSock(); err == nil || err != expected {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", expected, err)
	} else if fsk != nil {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", nil, fsk)
	}
}

func TestFSockPopFSock4(t *testing.T) {
	fs := &FSockPool{
		fSocks:      make(chan *FSock, 1),
		maxWaitConn: 20 * time.Millisecond,
	}

	expected := &FSock{}
	go func() {
		time.Sleep(5 * time.Millisecond)
		fs.fSocks <- expected
	}()
	fsock, err := fs.PopFSock()

	if err != nil {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", nil, err)
	} else if fsock != expected { // the pointer should be the same
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", expected, fsock)
	}
}

func TestFSockPopFSock5(t *testing.T) {
	fs := &FSockPool{
		addr:                 "testAddr",
		passwd:               "testPw",
		reconnects:           2,
		maxReconnectInterval: 0,
		delayFuncConstructor: fibDuration,
		eventHandlers:        make(map[string][]func(string, int)),
		eventFilters:         make(map[string][]string),
		logger:               nopLogger{},
		connIdx:              0,
		fSocks:               make(chan *FSock, 1),
		allowedConns:         make(chan struct{}),
		maxWaitConn:          20 * time.Millisecond,
	}

	expected := "dial tcp: address testAddr: missing port in address"
	close(fs.allowedConns)
	fsock, err := fs.PopFSock()

	if err.Error() != expected {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", expected, err)
	} else if fsock != nil {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", nil, fsock)
	}
}

func TestFSockReadBodyTT(t *testing.T) {
	testCases := []struct {
		name        string
		input       string
		expected    string
		bytesToRead int
		expectedErr error
	}{
		{
			name:     "simple string",
			input:    "Hello, World!",
			expected: "Hello, World!",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "multiple-line string",
			input:    "Line 1\nLine 2\nLine 3",
			expected: "Line 1\nLine 2\nLine 3",
		},
		{
			name:     "fs event",
			input:    "Event-Name: CHANNEL_PARK\nCore-UUID: 44d90754-93de-4dd7-807a-9ad31e45d4de\nFreeSWITCH-Hostname: debian12\nFreeSWITCH-Switchname: debian12\nFreeSWITCH-IPv4: 10.0.2.15\nFreeSWITCH-IPv6: %3A%3A1\nEvent-Date-Local: 2023-12-22%2010%3A12%3A32\nEvent-Date-GMT: Fri,%2022%20Dec%202023%2015%3A12%3A32%20GMT\nEvent-Date-Timestamp: 1703257952506074\nEvent-Calling-File: switch_ivr.c\nEvent-Calling-Function: switch_ivr_park\nEvent-Calling-Line-Number: 1002\nEvent-Sequence: 498\nChannel-State: CS_EXECUTE\nChannel-Call-State: RINGING\nChannel-State-Number: 4\nChannel-Name: sofia/internal/1001%40192.168.56.120%3A5081\nUnique-ID: 4967ceb1-c6f9-4af9-9855-df323d6763ad\nCall-Direction: inbound\nPresence-Call-Direction: inbound\nChannel-HIT-Dialplan: true\nChannel-Presence-ID: 1001%40192.168.56.120\nChannel-Call-UUID: 4967ceb1-c6f9-4af9-9855-df323d6763ad\nAnswer-State: ringing\nCaller-Direction: inbound\nCaller-Logical-Direction: inbound\nCaller-Username: 1001\nCaller-Dialplan: XML\nCaller-Caller-ID-Name: 1001\nCaller-Caller-ID-Number: 1001\nCaller-Orig-Caller-ID-Name: 1001\nCaller-Orig-Caller-ID-Number: 1001\nCaller-Network-Addr: 192.168.56.120\nCaller-ANI: 1001\nCaller-Destination-Number: 1002\nCaller-Unique-ID: 4967ceb1-c6f9-4af9-9855-df323d6763ad\nCaller-Source: mod_sofia\nCaller-Context: default\nCaller-Channel-Name: sofia/internal/1001%40192.168.56.120%3A5081\nCaller-Profile-Index: 1\nCaller-Profile-Created-Time: 1703257952506074\nCaller-Channel-Created-Time: 1703257952506074\nCaller-Channel-Answered-Time: 0\nCaller-Channel-Progress-Time: 0\nCaller-Channel-Progress-Media-Time: 0\nCaller-Channel-Hangup-Time: 0\nCaller-Channel-Transfer-Time: 0\nCaller-Channel-Resurrect-Time: 0\nCaller-Channel-Bridged-Time: 0\nCaller-Channel-Last-Hold: 0\nCaller-Channel-Hold-Accum: 0\nCaller-Screen-Bit: true\nCaller-Privacy-Hide-Name: false\nCaller-Privacy-Hide-Number: false\nvariable_direction: inbound\nvariable_uuid: 4967ceb1-c6f9-4af9-9855-df323d6763ad\nvariable_session_id: 1\nvariable_sip_from_user: 1001\nvariable_sip_from_port: 5081\nvariable_sip_from_uri: 1001%40192.168.56.120%3A5081\nvariable_sip_from_host: 192.168.56.120\nvariable_video_media_flow: disabled\nvariable_audio_media_flow: disabled\nvariable_text_media_flow: disabled\nvariable_channel_name: sofia/internal/1001%40192.168.56.120%3A5081\nvariable_sip_call_id: 1-27764%40192.168.56.120\nvariable_sip_local_network_addr: 192.168.56.120\nvariable_sip_network_ip: 192.168.56.120\nvariable_sip_network_port: 5081\nvariable_sip_invite_stamp: 1703257952506074\nvariable_sip_received_ip: 192.168.56.120\nvariable_sip_received_port: 5081\nvariable_sip_via_protocol: udp\nvariable_sip_authorized: true\nvariable_sip_acl_authed_by: domains\nvariable_sip_from_user_stripped: 1001\nvariable_sip_from_tag: 27764SIPpTag001\nvariable_sofia_profile_name: internal\nvariable_sofia_profile_url: sip%3Amod_sofia%40192.168.56.120%3A5060\nvariable_recovery_profile_name: internal\nvariable_sip_full_via: SIP/2.0/UDP%20192.168.56.120%3A5081%3Bbranch%3Dz9hG4bK-27764-1-0\nvariable_sip_from_display: 1001\nvariable_sip_full_from: 1001%20%3Csip%3A1001%40192.168.56.120%3A5081%3E%3Btag%3D27764SIPpTag001\nvariable_sip_to_display: 1002\nvariable_sip_full_to: 1002%20%3Csip%3A1002%40192.168.56.120%3A5060%3E\nvariable_sip_req_user: 1002\nvariable_sip_req_port: 5060\nvariable_sip_req_uri: 1002%40192.168.56.120%3A5060\nvariable_sip_req_host: 192.168.56.120\nvariable_sip_to_user: 1002\nvariable_sip_to_port: 5060\nvariable_sip_to_uri: 1002%40192.168.56.120%3A5060\nvariable_sip_to_host: 192.168.56.120\nvariable_sip_contact_user: sipp\nvariable_sip_contact_port: 5081\nvariable_sip_contact_uri: sipp%40192.168.56.120%3A5081\nvariable_sip_contact_host: 192.168.56.120\nvariable_rtp_use_codec_string: G722,PCMU,PCMA,GSM\nvariable_sip_subject: Performance%20Test\nvariable_sip_via_host: 192.168.56.120\nvariable_sip_via_port: 5081\nvariable_max_forwards: 70\nvariable_presence_id: 1001%40192.168.56.120\nvariable_switch_r_sdp: v%3D0%0D%0Ao%3Duser1%2053655765%202353687637%20IN%20IP4%20192.168.56.120%0D%0As%3D-%0D%0Ac%3DIN%20IP4%20192.168.56.120%0D%0At%3D0%200%0D%0Am%3Daudio%206000%20RTP/AVP%200%0D%0Aa%3Drtpmap%3A0%20PCMU/8000%0D%0A\nvariable_ep_codec_string: CORE_PCM_MODULE.PCMU%408000h%4020i%4064000b\nvariable_endpoint_disposition: DELAYED%20NEGOTIATION\nvariable_call_uuid: 4967ceb1-c6f9-4af9-9855-df323d6763ad\nvariable_current_application: park\n\n",
			expected: "Event-Name: CHANNEL_PARK\nCore-UUID: 44d90754-93de-4dd7-807a-9ad31e45d4de\nFreeSWITCH-Hostname: debian12\nFreeSWITCH-Switchname: debian12\nFreeSWITCH-IPv4: 10.0.2.15\nFreeSWITCH-IPv6: %3A%3A1\nEvent-Date-Local: 2023-12-22%2010%3A12%3A32\nEvent-Date-GMT: Fri,%2022%20Dec%202023%2015%3A12%3A32%20GMT\nEvent-Date-Timestamp: 1703257952506074\nEvent-Calling-File: switch_ivr.c\nEvent-Calling-Function: switch_ivr_park\nEvent-Calling-Line-Number: 1002\nEvent-Sequence: 498\nChannel-State: CS_EXECUTE\nChannel-Call-State: RINGING\nChannel-State-Number: 4\nChannel-Name: sofia/internal/1001%40192.168.56.120%3A5081\nUnique-ID: 4967ceb1-c6f9-4af9-9855-df323d6763ad\nCall-Direction: inbound\nPresence-Call-Direction: inbound\nChannel-HIT-Dialplan: true\nChannel-Presence-ID: 1001%40192.168.56.120\nChannel-Call-UUID: 4967ceb1-c6f9-4af9-9855-df323d6763ad\nAnswer-State: ringing\nCaller-Direction: inbound\nCaller-Logical-Direction: inbound\nCaller-Username: 1001\nCaller-Dialplan: XML\nCaller-Caller-ID-Name: 1001\nCaller-Caller-ID-Number: 1001\nCaller-Orig-Caller-ID-Name: 1001\nCaller-Orig-Caller-ID-Number: 1001\nCaller-Network-Addr: 192.168.56.120\nCaller-ANI: 1001\nCaller-Destination-Number: 1002\nCaller-Unique-ID: 4967ceb1-c6f9-4af9-9855-df323d6763ad\nCaller-Source: mod_sofia\nCaller-Context: default\nCaller-Channel-Name: sofia/internal/1001%40192.168.56.120%3A5081\nCaller-Profile-Index: 1\nCaller-Profile-Created-Time: 1703257952506074\nCaller-Channel-Created-Time: 1703257952506074\nCaller-Channel-Answered-Time: 0\nCaller-Channel-Progress-Time: 0\nCaller-Channel-Progress-Media-Time: 0\nCaller-Channel-Hangup-Time: 0\nCaller-Channel-Transfer-Time: 0\nCaller-Channel-Resurrect-Time: 0\nCaller-Channel-Bridged-Time: 0\nCaller-Channel-Last-Hold: 0\nCaller-Channel-Hold-Accum: 0\nCaller-Screen-Bit: true\nCaller-Privacy-Hide-Name: false\nCaller-Privacy-Hide-Number: false\nvariable_direction: inbound\nvariable_uuid: 4967ceb1-c6f9-4af9-9855-df323d6763ad\nvariable_session_id: 1\nvariable_sip_from_user: 1001\nvariable_sip_from_port: 5081\nvariable_sip_from_uri: 1001%40192.168.56.120%3A5081\nvariable_sip_from_host: 192.168.56.120\nvariable_video_media_flow: disabled\nvariable_audio_media_flow: disabled\nvariable_text_media_flow: disabled\nvariable_channel_name: sofia/internal/1001%40192.168.56.120%3A5081\nvariable_sip_call_id: 1-27764%40192.168.56.120\nvariable_sip_local_network_addr: 192.168.56.120\nvariable_sip_network_ip: 192.168.56.120\nvariable_sip_network_port: 5081\nvariable_sip_invite_stamp: 1703257952506074\nvariable_sip_received_ip: 192.168.56.120\nvariable_sip_received_port: 5081\nvariable_sip_via_protocol: udp\nvariable_sip_authorized: true\nvariable_sip_acl_authed_by: domains\nvariable_sip_from_user_stripped: 1001\nvariable_sip_from_tag: 27764SIPpTag001\nvariable_sofia_profile_name: internal\nvariable_sofia_profile_url: sip%3Amod_sofia%40192.168.56.120%3A5060\nvariable_recovery_profile_name: internal\nvariable_sip_full_via: SIP/2.0/UDP%20192.168.56.120%3A5081%3Bbranch%3Dz9hG4bK-27764-1-0\nvariable_sip_from_display: 1001\nvariable_sip_full_from: 1001%20%3Csip%3A1001%40192.168.56.120%3A5081%3E%3Btag%3D27764SIPpTag001\nvariable_sip_to_display: 1002\nvariable_sip_full_to: 1002%20%3Csip%3A1002%40192.168.56.120%3A5060%3E\nvariable_sip_req_user: 1002\nvariable_sip_req_port: 5060\nvariable_sip_req_uri: 1002%40192.168.56.120%3A5060\nvariable_sip_req_host: 192.168.56.120\nvariable_sip_to_user: 1002\nvariable_sip_to_port: 5060\nvariable_sip_to_uri: 1002%40192.168.56.120%3A5060\nvariable_sip_to_host: 192.168.56.120\nvariable_sip_contact_user: sipp\nvariable_sip_contact_port: 5081\nvariable_sip_contact_uri: sipp%40192.168.56.120%3A5081\nvariable_sip_contact_host: 192.168.56.120\nvariable_rtp_use_codec_string: G722,PCMU,PCMA,GSM\nvariable_sip_subject: Performance%20Test\nvariable_sip_via_host: 192.168.56.120\nvariable_sip_via_port: 5081\nvariable_max_forwards: 70\nvariable_presence_id: 1001%40192.168.56.120\nvariable_switch_r_sdp: v%3D0%0D%0Ao%3Duser1%2053655765%202353687637%20IN%20IP4%20192.168.56.120%0D%0As%3D-%0D%0Ac%3DIN%20IP4%20192.168.56.120%0D%0At%3D0%200%0D%0Am%3Daudio%206000%20RTP/AVP%200%0D%0Aa%3Drtpmap%3A0%20PCMU/8000%0D%0A\nvariable_ep_codec_string: CORE_PCM_MODULE.PCMU%408000h%4020i%4064000b\nvariable_endpoint_disposition: DELAYED%20NEGOTIATION\nvariable_call_uuid: 4967ceb1-c6f9-4af9-9855-df323d6763ad\nvariable_current_application: park\n\n",
		},
		{
			name:        "less characters",
			input:       "test_input",
			bytesToRead: 11,
			expected:    "",
			expectedErr: io.EOF,
		},
		{
			name:        "more characters",
			input:       "test_input",
			bytesToRead: 7,
			expected:    "test_in",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			fs := FSConn{
				rdr:  bufio.NewReaderSize(buf, 8192),
				lgr:  nopLogger{},
				conn: &net.TCPConn{},
			}
			_, err := fillBuffer(buf, tc.input)
			if err != nil {
				t.Fatalf("failed to fill buffer: %v", err)
			}
			noBytes := len(tc.input)
			if tc.bytesToRead != 0 {
				noBytes = tc.bytesToRead
			}
			received, err := fs.readBody(noBytes)
			if !errors.Is(err, tc.expectedErr) {
				t.Fatalf("expected error %v, received %v", tc.expectedErr, err)
			}

			if received != tc.expected {
				t.Errorf("expected %q,\nreceived %q", tc.expected, received)
			}
		})
	}
}

func TestFsConnReadEventErr(t *testing.T) {
	buf := new(bytes.Buffer)
	fs := FSConn{
		rdr:  bufio.NewReaderSize(buf, 8192),
		lgr:  nopLogger{},
		conn: &net.TCPConn{},
		err:  make(chan error, 1),
	}

	_, err := fillBuffer(buf, "Content-Length: error,	Content-Type: text/event-plain \n Event-Name: RE_SCHEDULE \n\n")
	if err != nil {
		t.Error(err)
	}
	fs.readEvents()
	select {
	case err = <-fs.err:
		if err == nil {
			t.Errorf("expected err")
		}
	case <-time.After(time.Millisecond * 1):
		t.Errorf("din't receive error from errorsChan")
	}
}

func fillBuffer(buf *bytes.Buffer, content string) (int, error) {
	buf.Reset()
	return buf.Write([]byte(content))
}

func BenchmarkFSockReadBody(b *testing.B) {
	content := "Event-Name: CHANNEL_PARK\nCore-UUID: 44d90754-93de-4dd7-807a-9ad31e45d4de\nFreeSWITCH-Hostname: debian12\nFreeSWITCH-Switchname: debian12\nFreeSWITCH-IPv4: 10.0.2.15\nFreeSWITCH-IPv6: %3A%3A1\nEvent-Date-Local: 2023-12-22%2010%3A12%3A32\nEvent-Date-GMT: Fri,%2022%20Dec%202023%2015%3A12%3A32%20GMT\nEvent-Date-Timestamp: 1703257952506074\nEvent-Calling-File: switch_ivr.c\nEvent-Calling-Function: switch_ivr_park\nEvent-Calling-Line-Number: 1002\nEvent-Sequence: 498\nChannel-State: CS_EXECUTE\nChannel-Call-State: RINGING\nChannel-State-Number: 4\nChannel-Name: sofia/internal/1001%40192.168.56.120%3A5081\nUnique-ID: 4967ceb1-c6f9-4af9-9855-df323d6763ad\nCall-Direction: inbound\nPresence-Call-Direction: inbound\nChannel-HIT-Dialplan: true\nChannel-Presence-ID: 1001%40192.168.56.120\nChannel-Call-UUID: 4967ceb1-c6f9-4af9-9855-df323d6763ad\nAnswer-State: ringing\nCaller-Direction: inbound\nCaller-Logical-Direction: inbound\nCaller-Username: 1001\nCaller-Dialplan: XML\nCaller-Caller-ID-Name: 1001\nCaller-Caller-ID-Number: 1001\nCaller-Orig-Caller-ID-Name: 1001\nCaller-Orig-Caller-ID-Number: 1001\nCaller-Network-Addr: 192.168.56.120\nCaller-ANI: 1001\nCaller-Destination-Number: 1002\nCaller-Unique-ID: 4967ceb1-c6f9-4af9-9855-df323d6763ad\nCaller-Source: mod_sofia\nCaller-Context: default\nCaller-Channel-Name: sofia/internal/1001%40192.168.56.120%3A5081\nCaller-Profile-Index: 1\nCaller-Profile-Created-Time: 1703257952506074\nCaller-Channel-Created-Time: 1703257952506074\nCaller-Channel-Answered-Time: 0\nCaller-Channel-Progress-Time: 0\nCaller-Channel-Progress-Media-Time: 0\nCaller-Channel-Hangup-Time: 0\nCaller-Channel-Transfer-Time: 0\nCaller-Channel-Resurrect-Time: 0\nCaller-Channel-Bridged-Time: 0\nCaller-Channel-Last-Hold: 0\nCaller-Channel-Hold-Accum: 0\nCaller-Screen-Bit: true\nCaller-Privacy-Hide-Name: false\nCaller-Privacy-Hide-Number: false\nvariable_direction: inbound\nvariable_uuid: 4967ceb1-c6f9-4af9-9855-df323d6763ad\nvariable_session_id: 1\nvariable_sip_from_user: 1001\nvariable_sip_from_port: 5081\nvariable_sip_from_uri: 1001%40192.168.56.120%3A5081\nvariable_sip_from_host: 192.168.56.120\nvariable_video_media_flow: disabled\nvariable_audio_media_flow: disabled\nvariable_text_media_flow: disabled\nvariable_channel_name: sofia/internal/1001%40192.168.56.120%3A5081\nvariable_sip_call_id: 1-27764%40192.168.56.120\nvariable_sip_local_network_addr: 192.168.56.120\nvariable_sip_network_ip: 192.168.56.120\nvariable_sip_network_port: 5081\nvariable_sip_invite_stamp: 1703257952506074\nvariable_sip_received_ip: 192.168.56.120\nvariable_sip_received_port: 5081\nvariable_sip_via_protocol: udp\nvariable_sip_authorized: true\nvariable_sip_acl_authed_by: domains\nvariable_sip_from_user_stripped: 1001\nvariable_sip_from_tag: 27764SIPpTag001\nvariable_sofia_profile_name: internal\nvariable_sofia_profile_url: sip%3Amod_sofia%40192.168.56.120%3A5060\nvariable_recovery_profile_name: internal\nvariable_sip_full_via: SIP/2.0/UDP%20192.168.56.120%3A5081%3Bbranch%3Dz9hG4bK-27764-1-0\nvariable_sip_from_display: 1001\nvariable_sip_full_from: 1001%20%3Csip%3A1001%40192.168.56.120%3A5081%3E%3Btag%3D27764SIPpTag001\nvariable_sip_to_display: 1002\nvariable_sip_full_to: 1002%20%3Csip%3A1002%40192.168.56.120%3A5060%3E\nvariable_sip_req_user: 1002\nvariable_sip_req_port: 5060\nvariable_sip_req_uri: 1002%40192.168.56.120%3A5060\nvariable_sip_req_host: 192.168.56.120\nvariable_sip_to_user: 1002\nvariable_sip_to_port: 5060\nvariable_sip_to_uri: 1002%40192.168.56.120%3A5060\nvariable_sip_to_host: 192.168.56.120\nvariable_sip_contact_user: sipp\nvariable_sip_contact_port: 5081\nvariable_sip_contact_uri: sipp%40192.168.56.120%3A5081\nvariable_sip_contact_host: 192.168.56.120\nvariable_rtp_use_codec_string: G722,PCMU,PCMA,GSM\nvariable_sip_subject: Performance%20Test\nvariable_sip_via_host: 192.168.56.120\nvariable_sip_via_port: 5081\nvariable_max_forwards: 70\nvariable_presence_id: 1001%40192.168.56.120\nvariable_switch_r_sdp: v%3D0%0D%0Ao%3Duser1%2053655765%202353687637%20IN%20IP4%20192.168.56.120%0D%0As%3D-%0D%0Ac%3DIN%20IP4%20192.168.56.120%0D%0At%3D0%200%0D%0Am%3Daudio%206000%20RTP/AVP%200%0D%0Aa%3Drtpmap%3A0%20PCMU/8000%0D%0A\nvariable_ep_codec_string: CORE_PCM_MODULE.PCMU%408000h%4020i%4064000b\nvariable_endpoint_disposition: DELAYED%20NEGOTIATION\nvariable_call_uuid: 4967ceb1-c6f9-4af9-9855-df323d6763ad\nvariable_current_application: park\n\n"
	buf := &bytes.Buffer{}
	fs := &FSConn{
		lgr: nopLogger{},
		rdr: bufio.NewReaderSize(buf, 8092),
	}
	noBytes := len(content)
	var err error
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err = fillBuffer(buf, content)
		if err != nil {
			b.Fatal(err)
		}
		_, err = fs.readBody(noBytes)
		if err != nil {
			b.Fatal(err)
		}
		// if body != content {
		// 	b.Fatalf("expected: %v, received: %v", content, body)
		// }
	}
}

// mockFreeSWITCH acts as a FreeSWITCH server. It goes through auth and then executes fn.
// The fn parameter can be customized based on the needs of the test.
// Returns the address of the listener.
func mockFreeSWITCH(t *testing.T, fn func(net.Conn)) string {
	t.Helper()

	// Start a ln on a random open port.
	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		defer ln.Close()
		conn, err := ln.Accept()
		if err != nil {
			t.Error(err)
			return
		}
		defer conn.Close()

		// Send auth challenge to the client.
		if _, err := conn.Write([]byte("auth/request\n\n")); err != nil {
			t.Error(err)
			return
		}

		rdr := bufio.NewReader(conn)
		auth := true
		for auth {
			// Read bytes until a newline.
			bytesRead, err := rdr.ReadBytes('\n')
			if err != nil {
				t.Error(err)
				return
			}

			// Ignore empty lines.
			if len(bytes.TrimSpace(bytesRead)) == 0 {
				continue
			}

			// Process auth/event plain requests.
			request := string(bytesRead)
			switch {
			case strings.Contains(request, "auth"):
				_, err = conn.Write([]byte("Reply-Text: +OK accepted\n\n"))
			case strings.Contains(request, "event plain"):
				_, err = conn.Write([]byte("Reply-Text: +OK\n\n"))

				// Final step during auth. End the loop.
				auth = false
			default:
				t.Error("unexpected request")
				return
			}
			if err != nil {
				t.Error(err)
				return
			}
		}

		// Execute the test-specific function after authentication.
		fn(conn)
	}()
	return ln.Addr().String()
}

func TestFSockHandleConnReset(t *testing.T) {
	addr := mockFreeSWITCH(t, func(c net.Conn) {
		// Simulate a syscall.ECONNRESET error by abruptly closing the connection after setting linger to 0.
		c.(*net.TCPConn).SetLinger(0)
		c.Close()
		// Closing the connection after setting linger to 0 causes an immediate reset, simulating a connection reset by peer.
	})

	stop := make(chan error)
	fs := &FSock{
		mu:         &sync.RWMutex{},
		connIdx:    0,
		addr:       addr,
		passwd:     "ClueCon",
		reconnects: 0, // no need to attempt reconnect
		logger:     nopLogger{},
		stopError:  stop,
		delayFunc:  fibDuration,
	}

	if err := fs.connect(); err != nil {
		t.Fatal("failed to connect to FreeSWITCH:", err)
	}

	// Encountering syscall.ECONNRESET while reading headers should trigger
	// reconnect attempts. With reconnects set to 0, expect a "not connected" error
	// on the stopError channel. A nil error means fsock mistakenly considered
	// the encountered error a signal for intentional shutdown.
	want := "not connected to FreeSWITCH"
	err := <-stop
	if err == nil || err.Error() != want {
		t.Errorf("conn error: got %v, want %s", err, want)
	}
}
