/*
fsock.go is released under the MIT License <http://www.opensource.org/licenses/mit-license.php
Copyright (C) ITsysCOM. All Rights Reserved.

Provides FreeSWITCH socket communication.

*/

package fsock

import (
	"errors"
	"fmt"
	"io"
	"net"
	"reflect"
	"sync"
	"time"
)

var (
	ErrConnectionPoolTimeout = errors.New("ConnectionPool timeout")
)

// NewFSock connects to FS and starts buffering input.
func NewFSock(addr, passwd string, reconnects int,
	maxReconnectInterval, replyTimeout time.Duration,
	delayFunc func(time.Duration, time.Duration) func() time.Duration,
	eventHandlers map[string][]func(string, int),
	eventFilters map[string][]string,
	logger logger, connIdx int, bgapi bool, stopError chan error,
) (fsock *FSock, err error) {
	if logger == nil ||
		(reflect.ValueOf(logger).Kind() == reflect.Ptr && reflect.ValueOf(logger).IsNil()) {
		logger = nopLogger{}
	}
	fsock = &FSock{
		mu:                   new(sync.RWMutex),
		connIdx:              connIdx,
		addr:                 addr,
		passwd:               passwd,
		eventFilters:         eventFilters,
		eventHandlers:        eventHandlers,
		reconnects:           reconnects,
		maxReconnectInterval: maxReconnectInterval,
		replyTimeout:         replyTimeout,
		delayFunc:            delayFunc,
		logger:               logger,
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
	mu      *sync.RWMutex
	connIdx int // identifier for the component using this instance of FSock, optional
	addr    string
	passwd  string
	fsConn  *FSConn

	reconnects           int
	maxReconnectInterval time.Duration
	replyTimeout         time.Duration
	delayFunc            func(time.Duration, time.Duration) func() time.Duration // used to create/reset the delay function

	eventFilters  map[string][]string
	eventHandlers map[string][]func(string, int) // eventStr, connId

	logger    logger
	bgapi     bool
	stopError chan error // will communicate on final disconnect
}

// Connect adds locking to connect method.
func (fs *FSock) Connect() (err error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	return fs.connect()
}

// connect establishes a connection to FreeSWITCH using the provided configuration details.
// This method is not thread-safe and should be called with external synchronization if used
// from multiple goroutines. Upon encountering read errors, it automatically attempts to
// restart the connection unless the error is intentionally triggered for stopping.
func (fs *FSock) connect() (err error) {

	// Create an error channel to listen for connection errors.
	connErr := make(chan error)

	// Initialize a new FSConn connection instance. Pass configuration and the error channel.
	fs.fsConn, err = NewFSConn(fs.addr, fs.passwd, fs.connIdx, fs.replyTimeout, connErr,
		fs.logger, fs.eventFilters, fs.eventHandlers, fs.bgapi)
	if err != nil {
		return err
	}

	// Start a goroutine to handle automatic reconnects in case the connection drops.
	go fs.handleConnectionError(connErr)

	return
}

// handleConnectionError listens for connection errors and decides whether to attempt a
// reconnection. It logs errors and manages the stopError channel signaling based on the
// encountered error.
func (fs *FSock) handleConnectionError(connErr chan error) {
	err := <-connErr // Wait for an error signal from readEvents.
	fs.logger.Err(fmt.Sprintf("<FSock> readEvents error (connection index: %d): %v", fs.connIdx, err))
	if err != io.EOF {
		// Signal nil error for intentional shutdowns.
		fs.signalError(nil)
		return // don't attempt reconnect
	}

	// Attempt to reconnect if the error indicates a dropped connection (io.EOF).
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	if err := fs.disconnect(); err != nil {
		fs.logger.Warning(fmt.Sprintf(
			"<FSock> Failed to disconnect from FreeSWITCH (connection index: %d): %v",
			fs.connIdx, err))
	}
	if err = fs.reconnectIfNeeded(); err != nil {
		fs.logger.Err(fmt.Sprintf(
			"<FSock> Failed to reconnect to FreeSWITCH (connection index: %d): %v",
			fs.connIdx, err))
		fs.signalError(err)
	}
}

// signalError handles logging or sending the error to the stopError channel.
func (fs *FSock) signalError(err error) {
	if fs.stopError == nil {
		// No stopError channel designated. Log the error if not nil.
		if err != nil {
			fs.logger.Err(fmt.Sprintf(
				"<FSock> Error encountered while reading events (connection index: %d): %v",
				fs.connIdx, err))
		}
		return
	}
	// Otherwise, signal on the stopError channel.
	fs.stopError <- err
}

// Connected adds up locking on top of normal connected method.
func (fs *FSock) Connected() (ok bool) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	return fs.connected()
}

// connected checks if socket connected. Not thread safe.
func (fs *FSock) connected() (ok bool) {
	return fs.fsConn != nil
}

// Disconnect adds up locking for disconnect
func (fs *FSock) Disconnect() (err error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
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
	fs.mu.Lock()
	defer fs.mu.Unlock()
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
	fs.mu.Lock() // make sure the fsConn does not get nil-ed after the reconnect
	defer fs.mu.Unlock()
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
	fs.mu.Lock()
	defer fs.mu.Unlock()
	if err := fs.reconnectIfNeeded(); err != nil {
		return out, err
	}
	return fs.fsConn.SendBgapiCmd(cmdStr)
}

func (fs *FSock) LocalAddr() net.Addr {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	if !fs.connected() {
		return nil
	}
	return fs.fsConn.LocalAddr()
}
