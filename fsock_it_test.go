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
	"strings"
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
