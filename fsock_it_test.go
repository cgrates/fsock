// +build integration

/*
fsock_it_test.go is released under the MIT License <http://www.opensource.org/licenses/mit-license.php
Copyright (C) ITsysCOM. All Rights Reserved.

Provides FreeSWITCH socket communication.

*/
package fsock

import (
	"fmt"
	"log/syslog"
	"net"
	"strings"
	"sync"
	"testing"
	"time"
)

var FSTests = []func(*FSock, *testing.T){
	testSendCmd,
	testSendApiCmd,
	testSendBgapiCmd,
	testReconect,
	testSendCmd,
	testSendApiCmd,
	testSendBgapiCmd,
	testSendEventWithBody,
	testSendEvent,
}

func TestFSock(t *testing.T) {
	faddr := "127.0.0.1:8021"
	fpass := "ClueCon"
	noreconects := 10
	conID := 0
	l, errLog := syslog.New(syslog.LOG_INFO, "TestFSock")
	if errLog != nil {
		t.Fatal(errLog)
	}
	evFilters := make(map[string][]string)
	evHandlers := make(map[string][]func(string, int))

	fs, err := NewFSock(faddr, fpass, noreconects, evHandlers, evFilters, l, conID)
	if err != nil {
		t.Fatal(err)
	}
	if !fs.Connected() {
		t.Errorf("Coudn't connect to freeswitch!")
	}

	for _, ft := range FSTests {
		t.Run("FSock", func(t *testing.T) { ft(fs, t) })
	}

	if err = fs.Disconnect(); err != nil {
		t.Error(err)
	}
}

func testReconect(fs *FSock, t *testing.T) {
	if err := fs.Disconnect(); err != nil {
		t.Error(err)
	}

	if err := fs.ReconnectIfNeeded(); err != nil {
		t.Error(err)
	}

	if !fs.Connected() {
		t.Errorf("Coudn't connect to freeswitch!")
	}
}

func testSendCmd(fs *FSock, t *testing.T) {
	expected := "Command recived!"
	cmd := fmt.Sprintf("api eval %s\n\n", expected)
	if rply, err := fs.SendCmd(cmd); err != nil {
		t.Error(err)
	} else if rply != expected {
		t.Errorf("Expected: %s , recieved: %s", expected, rply)
	}
}

func testSendApiCmd(fs *FSock, t *testing.T) {
	expected := "Command recived!"
	cmd := fmt.Sprintf("eval %s", expected)
	if rply, err := fs.SendApiCmd(cmd); err != nil {
		t.Error(err)
	} else if rply != expected {
		t.Errorf("Expected: %s , recieved: %s", expected, rply)
	}
}

func testSendBgapiCmd(fs *FSock, t *testing.T) {
	expected := "Command recived!"
	cmd := fmt.Sprintf("eval %s", expected)
	if ch, err := fs.SendBgapiCmd(cmd); err != nil {
		t.Error(err)
	} else {
		var rply string
		select {
		case rply = <-ch:
			if rply != expected {
				t.Errorf("Expected: %q , recieved: %q", expected, rply)
			}
		case <-time.After(5 * time.Second):
			t.Errorf("Timeout")
		}
	}
}

func testSendEventWithBody(fs *FSock, t *testing.T) {
	event := "NOTIFY"
	args := map[string]string{
		"profile":        "internal",
		"content-type":   "application/simple-message-summary",
		"event-string":   "check-sync",
		"user":           "1006",
		"host":           "99.157.44.194",
		"content-length": "2",
	}
	body := "OK"

	if rply, err := fs.SendEventWithBody(event, args, body); err != nil {
		t.Error(err)
	} else if !strings.HasPrefix(rply, "+OK") {
		t.Errorf("Event resonse wrong %s", rply)
	}
}

func testSendEvent(fs *FSock, t *testing.T) {
	event := "NOTIFY"
	args := map[string]string{
		"profile":      "internal",
		"content-type": "application/simple-message-summary",
		"event-string": "check-sync;reboot=false",
		"user":         "1005",
		"host":         "99.157.44.194",
	}
	if rply, err := fs.SendEvent(event, args); err != nil {
		t.Error(err)
	} else if !strings.HasPrefix(rply, "+OK") {
		t.Errorf("Event resonse wrong %s", rply)
	}
}

func TestFSockNewFSockNilLogger(t *testing.T) {
	fsaddr := "127.0.0.1:1234"
	fpaswd := "pw"
	noreconnects := 5
	conID := 0
	var l logger
	evFilters := make(map[string][]string)
	evHandlers := make(map[string][]func(string, int))

	fs, err := NewFSock(fsaddr, fpaswd, noreconnects, evHandlers, evFilters, l, conID)
	errexp := "dial tcp 127.0.0.1:1234: connect: connection refused"

	if err.Error() != errexp {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", errexp, err)
	}

	if fs != nil {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", nil, fs)
	}
}

func TestFSockconnect(t *testing.T) {
	const fsaddr = "127.0.0.1:8989"
	fs := &FSock{
		fsMutex:         &sync.RWMutex{},
		fsaddress:       fsaddr,
		fspaswd:         "pass",
		eventHandlers:   make(map[string][]func(string, int)),
		eventFilters:    make(map[string][]string),
		backgroundChans: make(map[string]chan string),
		cmdChan:         make(chan string),
		reconnects:      -1,
		delayFunc:       fib(),
		logger:          nopLogger{},
	}
	l, err := net.Listen("tcp", fsaddr)
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		conn, err := l.Accept()
		if err != nil {
			t.Error(err)
			return
		}
		conn.Write([]byte("not valid"))
		conn.Close()
	}()
	experr1 := "Received error<EOF> when receiving the auth challenge"
	if err := fs.connect(); err.Error() != experr1 {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", experr1, err)
	}
	go func() {
		conn, err := l.Accept()
		if err != nil {
			t.Error(err)
			return
		}
		conn.Write([]byte("not valid\n\n"))
		conn.Close()
	}()
	experr2 := "No auth challenge received"
	if err := fs.connect(); err.Error() != experr2 {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", experr2, err)
	}
	go func() {
		conn, err := l.Accept()
		if err != nil {
			t.Error(err)
			return
		}
		conn.Write([]byte("Content-Type: auth/request\n\n"))
		c := make([]byte, 512)
		conn.Read(c)
		conn.Write([]byte("Content-Type: command/reply\nReply-Text:  accepted\n\n"))
		// conn.Write([]byte("Content-Type: command/reply\nReply-Text: +OK accepted\n\n"))
		conn.Close()
	}()
	experr3 := "Unexpected auth reply received: <Content-Type: command/reply\nReply-Text:  accepted\n>"
	if err := fs.connect(); err.Error() != experr3 {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", experr3, err)
	}
	fs.eventFilters["Event-Name"] = []string{"CUSTOM"}
	go func() {
		conn, err := l.Accept()
		if err != nil {
			t.Error(err)
			return
		}
		conn.Write([]byte("Content-Type: auth/request\n\n"))
		c := make([]byte, 512)
		conn.Read(c)
		conn.Write([]byte("Content-Type: command/reply\nReply-Text: +OK accepted\n\n"))
		c = make([]byte, 512)
		conn.Read(c)
		conn.Write([]byte("not valid"))
		conn.Close()
	}()
	experr4 := "EOF"
	if err := fs.connect(); err.Error() != experr4 {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", experr4, err)
	}
	fs.eventHandlers["ALL"] = nil
	fs.eventFilters = make(map[string][]string)
	go func() {
		conn, err := l.Accept()
		if err != nil {
			t.Error(err)
			return
		}
		conn.Write([]byte("Content-Type: auth/request\n\n"))
		c := make([]byte, 512)
		conn.Read(c)
		conn.Write([]byte("Content-Type: command/reply\nReply-Text: +OK accepted\n\n"))
		c = make([]byte, 512)
		conn.Read(c)
		conn.Write([]byte("not valid"))
		conn.Close()
	}()
	experr5 := "EOF"
	if err := fs.connect(); err.Error() != experr5 {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", experr5, err)
	}
}
