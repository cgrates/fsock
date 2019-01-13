/*
fsock_test.go is released under the MIT License <http://www.opensource.org/licenses/mit-license.php
Copyright (C) ITsysCOM. All Rights Reserved.

Provides FreeSWITCH socket communication.

*/
package fsock

import (
	"bufio"
	"io/ioutil"
	"os"
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
		t.Error("Error creating pype!")
	}
	FS = &FSock{}
	FS.fsMutex = new(sync.RWMutex)
	FS.buffer = bufio.NewReader(r)
	w.Write([]byte(HEADER))
	h, err := FS.readHeaders()
	if err != nil || h != "Content-Length: 564\nContent-Type: text/event-plain\n" {
		t.Error("Error parsing headers: ", h, err)
	}
}

func TestEvent(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Error("Error creating pype!")
	}
	FS = &FSock{}
	FS.fsMutex = new(sync.RWMutex)
	FS.buffer = bufio.NewReader(r)
	w.Write([]byte(HEADER + BODY))
	h, b, err := FS.readEvent()
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
	evfunc := func(string, string) {
		funcMutex.Lock()
		events++
		funcMutex.Unlock()
	}

	FS = &FSock{}
	FS.fsMutex = new(sync.RWMutex)
	FS.buffer = bufio.NewReader(r)
	FS.eventHandlers = map[string][]func(string, string){
		"HEARTBEAT":                []func(string, string){evfunc},
		"RE_SCHEDULE":              []func(string, string){evfunc},
		"CHANNEL_STATE":            []func(string, string){evfunc},
		"CODEC":                    []func(string, string){evfunc},
		"CHANNEL_CREATE":           []func(string, string){evfunc},
		"CHANNEL_CALLSTATE":        []func(string, string){evfunc},
		"API":                      []func(string, string){evfunc},
		"CHANNEL_EXECUTE":          []func(string, string){evfunc},
		"CHANNEL_EXECUTE_COMPLETE": []func(string, string){evfunc},
		"CHANNEL_PARK":             []func(string, string){evfunc},
		"CHANNEL_HANGUP":           []func(string, string){evfunc},
		"CHANNEL_HANGUP_COMPLETE":  []func(string, string){evfunc},
		"CHANNEL_UNPARK":           []func(string, string){evfunc},
		"CHANNEL_DESTROY":          []func(string, string){evfunc},
	}
	go FS.readEvents()
	w.Write(data)
	time.Sleep(50 * time.Millisecond)
	funcMutex.RLock()
	if events != 45 {
		t.Error("Error reading events: ", events)
	}
	funcMutex.RUnlock()
}
