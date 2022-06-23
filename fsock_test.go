/*
fsock_test.go is released under the MIT License <http://www.opensource.org/licenses/mit-license.php
Copyright (C) ITsysCOM. All Rights Reserved.

Provides FreeSWITCH socket communication.

*/
package fsock

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
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
	fs := &FSock{}
	fs.fsMutex = new(sync.RWMutex)
	fs.buffer = bufio.NewReader(r)
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
	fs := &FSock{}
	fs.fsMutex = new(sync.RWMutex)
	fs.buffer = bufio.NewReader(r)
	w.Write([]byte(HEADER + BODY))
	h, b, err := fs.readEvent()
	if err != nil || h != HEADER[:len(HEADER)-1] || len(b) != 564 {
		t.Error("Error parsing event: ", h, b, len(b))
	}
}

func TestReadEvents(t *testing.T) {
	data, err := ioutil.ReadFile("test_data.txt")
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

	fs := &FSock{logger: nopLogger{}}
	fs.fsMutex = new(sync.RWMutex)
	fs.buffer = bufio.NewReader(r)
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
		fsMutex:        new(sync.RWMutex),
		eventHandlers:  make(map[string][]func(string, int)),
		eventFilters:   make(map[string][]string),
		stopReadEvents: make(chan struct{}),
		logger:         nopLogger{},
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
	fs := &FSock{
		logger:    nopLogger{},
		delayFunc: fibDuration,
		fsMutex:   &sync.RWMutex{},
		conn:      new(connMock),
	}

	expected := ErrConnectionPoolTimeout
	err := fs.send("testString")

	if err == nil || err != expected {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", expected, err)
	}
}

func TestFSockAuthFailSend(t *testing.T) {
	fs := &FSock{
		logger:  nopLogger{},
		fsMutex: &sync.RWMutex{},
		conn:    new(connMock),
	}

	err := fs.auth()

	if err == nil || err != ErrConnectionPoolTimeout {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", ErrConnectionPoolTimeout, err)
	}
}

func TestFSockAuthFailReply(t *testing.T) {
	buf := new(bytes.Buffer)
	fs := &FSock{
		fspaswd: "test",
		conn:    &connMock2{buf: buf},
		buffer:  bufio.NewReader(bytes.NewBuffer([]byte("Reply-Text: +OK accepted\n\n"))),
		fsMutex: new(sync.RWMutex),
		logger:  new(nopLogger),
	}

	expected := fmt.Sprintf("Unexpected auth reply received: <%s>", strings.TrimSuffix(HEADER, "\n"))
	err := fs.auth()
	if err != nil {
		t.Fatal(err)
	}

	expectedbuf := "auth test\n\n"
	if rcv := buf.String(); rcv != expectedbuf {
		t.Errorf("\nReceived: %q, \nExpected: %q", rcv, expectedbuf)
	}

	buf.Reset()
	fs.buffer = bufio.NewReader(bytes.NewBuffer([]byte(HEADER)))
	err = fs.auth()

	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", expected, err.Error())
	}

	if rcv := buf.String(); rcv != expectedbuf {
		t.Errorf("\nExpected: %q, \nReceived: %q", expectedbuf, rcv)
	}
}

func TestFSockAuthFailRead(t *testing.T) {
	fs := &FSock{
		fspaswd: "test",
		fsMutex: &sync.RWMutex{},
		buffer:  bufio.NewReader(bytes.NewBuffer([]byte("Reply-Text: +OK accepted"))),
		logger:  new(nopLogger),
		conn:    new(connMock3),
	}
	expected := io.EOF
	err := fs.auth()

	if err == nil || err != expected {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", expected, err)
	}
}

func TestFSockSendBgapiCmdNonNilErr(t *testing.T) {
	fs := &FSock{
		fsMutex:         &sync.RWMutex{},
		delayFunc:       fibDuration,
		backgroundChans: make(map[string]chan string),
	}

	expected := "Not connected to FreeSWITCH"
	_, err := fs.SendBgapiCmd("test")

	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", expected, err)
	}
}

func TestFSockSendMsgCmdWithBodyEmptyArguments(t *testing.T) {
	fs := &FSock{}
	uuid := ""
	cmdargs := make(map[string]string)
	body := ""

	expected := "Need command arguments"
	err := fs.SendMsgCmdWithBody(uuid, cmdargs, body)

	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", expected, err)
	}
}

func TestFSockSendMsgCmd(t *testing.T) {
	fs := &FSock{}
	uuid := "testID"
	cmdargs := make(map[string]string)

	expected := "Need command arguments"
	err := fs.SendMsgCmd(uuid, cmdargs)

	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", expected, err)
	}
}

func TestFSockLocalAddrNotConnected(t *testing.T) {
	fs := &FSock{
		fsMutex: &sync.RWMutex{},
	}
	addr := fs.LocalAddr()
	if addr != nil {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", nil, addr)
	}
}

func TestFSockReadEvents(t *testing.T) {
	fs := &FSock{
		fsMutex:        &sync.RWMutex{},
		delayFunc:      fibDuration,
		stopReadEvents: make(chan struct{}),
		errReadEvents:  make(chan error, 1),
	}

	fs.errReadEvents <- io.EOF

	expected := "Not connected to FreeSWITCH"
	err := fs.ReadEvents()

	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", expected, err)
	}
}

func TestFSockReadBody(t *testing.T) {
	fs := &FSock{
		fsMutex: &sync.RWMutex{},
		logger:  nopLogger{},
		buffer:  bufio.NewReader(bytes.NewBuffer([]byte(""))),
	}
	rply, err := fs.readBody(2)

	if err == nil || err != io.EOF {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", io.EOF, err)
	}

	if rply != "" {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", "", rply)
	}
}

func TestFSockSendCmdErrSend(t *testing.T) {
	fs := &FSock{
		fsMutex:    &sync.RWMutex{},
		logger:     nopLogger{},
		reconnects: 5,
		conn:       &connMock{},
	}
	rply, err := fs.sendCmd("test")

	if err == nil || err != ErrConnectionPoolTimeout {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", ErrConnectionPoolTimeout, err)
	}

	if rply != "" {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", "", rply)
	}
}

func TestFSockSendCmdErrContains(t *testing.T) {
	fs := &FSock{
		fsMutex:    &sync.RWMutex{},
		logger:     nopLogger{},
		reconnects: 2,
		conn:       &connMock3{},
		cmdChan:    make(chan string, 1),
	}

	fs.cmdChan <- "test-ERR"

	expected := "test-ERR"
	rply, err := fs.sendCmd("test")
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", expected, err)
	}

	if rply != "" {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", "", rply)
	}

}

func TestFSockReconnectIfNeeded(t *testing.T) {
	fs := &FSock{
		fsMutex:    &sync.RWMutex{},
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
		fsMutex:   &sync.RWMutex{},
		delayFunc: fibDuration,
	}
	uuid := "testID"
	cmdargs := map[string]string{
		"testKey": "testValue",
	}
	body := "testBody"

	expected := "Not connected to FreeSWITCH"
	err := fs.SendMsgCmdWithBody(uuid, cmdargs, body)

	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", expected, err)
	}
}

func TestFSockLocalAddr(t *testing.T) {
	fs := &FSock{
		conn:    &connMock{},
		fsMutex: &sync.RWMutex{},
	}
	addr := fs.LocalAddr()
	if addr != nil {
		t.Errorf("\nExpected nil, got %v", addr)
	}
}

func TestFSockreadEvent(t *testing.T) {
	fs := &FSock{
		buffer:  bufio.NewReader(bytes.NewBuffer([]byte("Content-Length\n\n"))),
		logger:  nopLogger{},
		fsMutex: &sync.RWMutex{},
	}

	expected := fmt.Sprintf("Cannot extract content length because<%s>", "strconv.Atoi: parsing \"\": invalid syntax")
	exphead := "Content-Length\n"
	expbody := ""
	head, body, err := fs.readEvent()
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", expected, err)
	}

	if head != exphead {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", exphead, head)
	}

	if body != expbody {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", expbody, body)
	}
}

func TestFSockreadEventsStopRead(t *testing.T) {
	// nothing to check only for coverage
	fs := &FSock{
		stopReadEvents: make(chan struct{}, 1),
	}

	close(fs.stopReadEvents)
	fs.readEvents()
}

func TestFSockeventsPlainErrSend(t *testing.T) {
	fs := &FSock{
		fsMutex: &sync.RWMutex{},
		conn:    &connMock{},
		logger:  nopLogger{},
	}
	events := []string{""}

	expected := ErrConnectionPoolTimeout
	err := fs.eventsPlain(events, true)

	if err == nil || err != expected {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", expected, err)
	}
}

func TestFSockeventsPlainErrRead(t *testing.T) {
	fs := &FSock{
		fsMutex: &sync.RWMutex{},
		conn:    &connMock3{},
		logger:  nopLogger{},
		buffer:  bufio.NewReader(bytes.NewBuffer([]byte("test\n"))),
	}
	events := []string{"ALL"}

	expected := io.EOF
	err := fs.eventsPlain(events, true)

	if err == nil || err != expected {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", expected, err)
	}
}

func TestFSockeventsPlainUnexpectedReply(t *testing.T) {
	fs := &FSock{
		fsMutex: &sync.RWMutex{},
		conn:    &connMock3{},
		logger:  nopLogger{},
		buffer:  bufio.NewReader(bytes.NewBuffer([]byte("test\n\n"))),
	}
	events := []string{"CUSTOMtest"}

	expected := fmt.Sprintf("Unexpected events-subscribe reply received: <%s>", "test\n")
	err := fs.eventsPlain(events, true)

	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", expected, err)
	}
}

func TestFSockfilterEventsUnexpectedReply(t *testing.T) {
	fs := &FSock{
		fsMutex: &sync.RWMutex{},
		conn:    &connMock3{},
		buffer:  bufio.NewReader(bytes.NewBuffer([]byte("test\n\n"))),
		logger:  nopLogger{},
	}
	filters := map[string][]string{
		"Event-Name": nil,
	}

	expected := fmt.Sprintf("Unexpected filter-events reply received: <%s>", "test\n")
	err := fs.filterEvents(filters, true)

	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", expected, err)
	}
}

func TestFSockfilterEventsErrRead(t *testing.T) {
	fs := &FSock{
		fsMutex: &sync.RWMutex{},
		conn:    &connMock3{},
		buffer:  bufio.NewReader(bytes.NewBuffer([]byte("test\n"))),
		logger:  nopLogger{},
	}
	filters := map[string][]string{
		"Event-Name": nil,
	}

	expected := io.EOF
	err := fs.filterEvents(filters, true)

	if err == nil || err != expected {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", expected, err)
	}
}

func TestFSockfilterEventsErrSend(t *testing.T) {
	fs := &FSock{
		fsMutex: &sync.RWMutex{},
		conn:    &connMock{},
		buffer:  bufio.NewReader(bytes.NewBuffer([]byte("test\n\n"))),
		logger:  nopLogger{},
	}
	filters := map[string][]string{
		"Event-Name": nil,
	}

	expected := ErrConnectionPoolTimeout
	err := fs.filterEvents(filters, true)

	if err == nil || err != expected {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", expected, err)
	}
}

func TestFSockfilterEventsErrNil(t *testing.T) {
	fs := &FSock{
		fsMutex: &sync.RWMutex{},
		conn:    &connMock3{},
		buffer:  bufio.NewReader(bytes.NewBuffer([]byte("testReply-Text: +OK\n\n"))),
		logger:  nopLogger{},
	}
	filters := map[string][]string{
		"Event-Name": nil,
	}

	err := fs.filterEvents(filters, true)

	if err != nil {
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
	fs := &FSock{
		logger: l,
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
	fs := &FSock{
		logger: l,
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
	fs := &FSock{
		logger:  l,
		fsMutex: &sync.RWMutex{},
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

	evHandlers := make(map[string][]func(string, int))
	evFilters := make(map[string][]string)

	fspool := &FSockPool{
		connIdx:       connIdx,
		fsAddr:        fsaddr,
		fsPasswd:      fspw,
		reconnects:    reconns,
		maxWaitConn:   maxWait,
		eventHandlers: evHandlers,
		eventFilters:  evFilters,
		logger:        nopLogger{},
		allowedConns:  nil,
		fSocks:        nil,
		bgapiSup:      true,
	}
	fsnew := NewFSockPool(maxFSocks, fsaddr, fspw, reconns, maxWait, 0, fibDuration, evHandlers, evFilters, nil, connIdx, true)
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
	fsk := &FSock{
		fsMutex: &sync.RWMutex{},
		conn:    &connMock{},
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

	expected := "Unconfigured ConnectionPool"
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
	fsock, err := fs.PopFSock()

	if err != nil {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", nil, err)
	} else if fsock != expected { // the pointer should be the same
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", expected, fsock)
	}
}

func TestFSockPopFSockTimeout(t *testing.T) {
	fs := &FSockPool{}

	expected := ErrConnectionPoolTimeout
	fsk, err := fs.PopFSock()

	if err == nil || err != expected {
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
		fsAddr:               "testAddr",
		fsPasswd:             "testPw",
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
