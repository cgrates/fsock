/*
fsock.go is released under the MIT License <http://www.opensource.org/licenses/mit-license.php
Copyright (C) ITsysCOM. All Rights Reserved.

Provides FreeSWITCH socket communication.

*/

package fsock

import (
	"errors"
	"io"
	"net"
	"reflect"
	"sync"
	"time"
)

var (
	ErrConnectionPoolTimeout = errors.New("ConnectionPool timeout")
	ErrDisconnected          = errors.New("DISCONNECTED")
)

// NewFSock connects to FS and starts buffering input
func NewFSock(fsaddr, fsPasswd string, reconnects int, maxReconnectInterval time.Duration,
	delayFunc func(time.Duration, time.Duration) func() time.Duration,
	eventHandlers map[string][]func(string, int), eventFilters map[string][]string,
	l logger, connIdx int, bgapi bool, stopError chan error) (fsock *FSock, err error) {
	if l == nil ||
		(reflect.ValueOf(l).Kind() == reflect.Ptr && reflect.ValueOf(l).IsNil()) {
		l = nopLogger{}
	}
	fsock = &FSock{
		fsMux:                new(sync.RWMutex),
		connIdx:              connIdx,
		fsAddr:               fsaddr,
		fsPasswd:             fsPasswd,
		eventFilters:         eventFilters,
		eventHandlers:        eventHandlers,
		reconnects:           reconnects,
		maxReconnectInterval: maxReconnectInterval,
		delayFunc:            delayFunc,
		logger:               l,
		bgapi:                bgapi,
		stopError:            stopError,
	}
	if err = fsock.Connect(); err != nil {
		return nil, err
	}
	return
}

// FSock reperesents the connection to FreeSWITCH Socket
type FSock struct {
	fsMux                *sync.RWMutex
	fsConn               *FSConn
	connIdx              int // Indetifier for the component using this instance of FSock, optional
	fsAddr               string
	fsPasswd             string
	eventFilters         map[string][]string
	eventHandlers        map[string][]func(string, int) // eventStr, connId
	reconnects           int
	maxReconnectInterval time.Duration
	delayFunc            func(time.Duration, time.Duration) func() time.Duration // used to create/reset the delay function
	stopReadEvents       chan struct{}                                           // Keep a reference towards forkedReadEvents so we can stop them whenever necessary
	readEventsFStopped   chan struct{}                                           // FSConn will signal the abormal disconnect
	logger               logger
	bgapi                bool
	stopError            chan error // will communicate on final disconnect
}

// Connect adds loking to connect method
func (fs *FSock) Connect() (err error) {
	fs.fsMux.Lock()
	defer fs.fsMux.Unlock()
	return fs.connect()
}

// Connect is a non-thread safe connect method
func (fs *FSock) connect() (err error) {
	// Reinit readEvents channels so we avoid concurrency issues between goroutines
	fs.stopReadEvents = make(chan struct{})
	fs.readEventsFStopped = make(chan struct{})
	fs.fsConn, err = NewFSConn(fs.fsAddr, fs.fsPasswd,
		fs.connIdx, fs.logger,
		fs.eventFilters, fs.eventHandlers,
		fs.bgapi, fs.stopReadEvents, fs.readEventsFStopped)
	if err != nil {
		return
	}
	go func() { // automatic restart if the connection was dropped due to read errors
		if forced := <-fs.readEventsFStopped; forced == struct{}{} {
			fs.fsMux.RLock()
			defer fs.fsMux.RUnlock()
			fs.disconnect()
			fs.reconnectIfNeeded()
			if !fs.connected() {
				fs.stopError <- io.EOF // signal to external sources that we will not attempt reconnect
			}
		}
	}()
	return
}

// Connected adds up locking on top of normal connected method.
func (fs *FSock) Connected() (ok bool) {
	fs.fsMux.RLock()
	defer fs.fsMux.RUnlock()
	return fs.connected()
}

// connected checks if socket connected. Not thread safe.
func (fs *FSock) connected() (ok bool) {
	return fs.fsConn != nil
}

// Disconnect adds up locking for disconnect
func (fs *FSock) Disconnect() (err error) {
	fs.fsMux.Lock()
	defer fs.fsMux.Unlock()
	return fs.disconnect()
}

// Disconnect disconnects from socket
func (fs *FSock) disconnect() (err error) {
	if fs.fsConn != nil {
		fs.logger.Info("<FSock> Disconnecting from FreeSWITCH!")
		err = fs.fsConn.Disconnect()
		fs.fsConn = nil
	}
	return
}

// ReconnectIfNeeded adds up locking to reconnectIfNeeded
func (fs *FSock) ReconnectIfNeeded() (err error) {
	fs.fsMux.Lock()
	defer fs.fsMux.Unlock()
	return fs.reconnectIfNeeded()
}

// reconnectIfNeeded if not connected, attempt reconnect if allowed
func (fs *FSock) reconnectIfNeeded() (err error) {
	if fs.connected() { // No need to reconnect
		return
	}
	delay := fs.delayFunc(time.Second, fs.maxReconnectInterval)
	for i := 0; fs.reconnects == -1 || i < fs.reconnects; i++ { // Maximum reconnects reached, -1 for infinite reconnects
		if err = fs.connect(); err == nil && fs.connected() {
			break // No error or unrelated to connection
		}
		time.Sleep(delay())
	}
	if err == nil && !fs.connected() {
		return errors.New("not connected to FreeSWITCH")
	}
	return // nil or last error in the loop
}

// Generic proxy for commands
func (fs *FSock) SendCmd(cmdStr string) (rply string, err error) {
	fs.fsMux.Lock() // make sure the fsConn does not get nil-ed after the reconnect
	defer fs.fsMux.Unlock()
	if err = fs.reconnectIfNeeded(); err != nil {
		return
	}
	return fs.fsConn.Send(cmdStr + "\n") // ToDo: check if we have to send a secondary new line
}

func (fs *FSock) SendCmdWithArgs(cmd string, args map[string]string, body string) (string, error) {
	for k, v := range args {
		cmd += k + ": " + v + "\n"
	}
	if len(body) != 0 {
		cmd += "\n" + body + "\n"
	}
	return fs.SendCmd(cmd)
}

// Send API command
func (fs *FSock) SendApiCmd(cmdStr string) (string, error) {
	return fs.SendCmd("api " + cmdStr + "\n")
}

// SendMsgCmdWithBody command
func (fs *FSock) SendMsgCmdWithBody(uuid string, cmdargs map[string]string, body string) (err error) {
	if len(cmdargs) == 0 {
		return errors.New("need command arguments")
	}
	_, err = fs.SendCmdWithArgs("sendmsg "+uuid+"\n", cmdargs, body)
	return
}

// SendMsgCmd command
func (fs *FSock) SendMsgCmd(uuid string, cmdargs map[string]string) error {
	return fs.SendMsgCmdWithBody(uuid, cmdargs, "")
}

// SendEventWithBody command
func (fs *FSock) SendEventWithBody(eventSubclass string, eventParams map[string]string, body string) (string, error) {
	// Event-Name is overrided to CUSTOM by FreeSWITCH,
	// so we use Event-Subclass instead
	eventParams["Event-Subclass"] = eventSubclass
	return fs.SendCmdWithArgs("sendevent "+eventSubclass+"\n", eventParams, body)
}

// SendEvent command
func (fs *FSock) SendEvent(eventSubclass string, eventParams map[string]string) (string, error) {
	return fs.SendEventWithBody(eventSubclass, eventParams, "")
}

// Send BGAPI command
func (fs *FSock) SendBgapiCmd(cmdStr string) (out chan string, err error) {
	fs.fsMux.Lock()
	defer fs.fsMux.Unlock()
	fs.reconnectIfNeeded()
	return fs.fsConn.SendBgapiCmd(cmdStr)
}

func (fs *FSock) LocalAddr() net.Addr {
	fs.fsMux.RLock()
	defer fs.fsMux.RUnlock()
	if !fs.connected() {
		return nil
	}
	return fs.fsConn.LocalAddr()
}
