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
		logger:  nopLogger{},
		fsMutex: &sync.RWMutex{},
		conn:    new(connMock),
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
