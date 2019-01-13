package fsock

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log/syslog"
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

var FSTests = []func(*FSock, *testing.T){
	testSendCmd,
	testSendApiCmd,
	testSendBgapiCmd1,
	testReconect,
	testSendCmd,
	testSendApiCmd,
	testSendBgapiCmd1,
}

func TestFSock(t *testing.T) {
	faddr := "127.0.0.1:8021"
	fpass := "ClueCon"
	noreconects := 10
	conID := "wetsfnmretiewrtpj"
	l, errLog := syslog.New(syslog.LOG_INFO, "TestFSock")
	if errLog != nil {
		t.Fatal(errLog)
	}
	evFilters := make(map[string][]string)
	evHandlers := make(map[string][]func(string, string))

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
	// fs.CloseChans()
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

func testSendBgapiCmd1(fs *FSock, t *testing.T) {
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
