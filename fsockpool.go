/*
fsock.go is released under the MIT License <http://www.opensource.org/licenses/mit-license.php
Copyright (C) ITsysCOM. All Rights Reserved.

Provides FreeSWITCH socket communication.

*/

package fsock

import (
	"errors"
	"time"
)

// Instantiates a new FSockPool
func NewFSockPool(maxFSocks int, fsaddr, fspasswd string, reconnects int, maxWaitConn time.Duration,
	maxReconnectInterval time.Duration, delayFuncConstructor func(time.Duration, time.Duration) func() time.Duration,
	eventHandlers map[string][]func(string, int), eventFilters map[string][]string,
	l logger, connIdx int, bgapi bool, stopError chan error) *FSockPool {
	if l == nil {
		l = nopLogger{}
	}
	pool := &FSockPool{
		connIdx:              connIdx,
		fsAddr:               fsaddr,
		fsPasswd:             fspasswd,
		reconnects:           reconnects,
		maxReconnectInterval: maxReconnectInterval,
		delayFuncConstructor: delayFuncConstructor,
		maxWaitConn:          maxWaitConn,
		eventHandlers:        eventHandlers,
		eventFilters:         eventFilters,
		logger:               l,
		allowedConns:         make(chan struct{}, maxFSocks),
		fSocks:               make(chan *FSock, maxFSocks),
		bgapi:                bgapi,
		stopError:            stopError,
	}
	for i := 0; i < maxFSocks; i++ {
		pool.allowedConns <- struct{}{} // Empty initiate so we do not need to wait later when we pop
	}
	return pool
}

// Connection handler for commands sent to FreeSWITCH
type FSockPool struct {
	connIdx              int
	fsAddr               string
	fsPasswd             string
	reconnects           int
	maxReconnectInterval time.Duration
	delayFuncConstructor func(time.Duration, time.Duration) func() time.Duration
	eventHandlers        map[string][]func(string, int)
	eventFilters         map[string][]string
	logger               logger
	allowedConns         chan struct{} // Will be populated with members allowed
	fSocks               chan *FSock   // Keep here reference towards the list of opened sockets
	maxWaitConn          time.Duration // Maximum duration to wait for a connection to be returned by Pop
	bgapi                bool
	stopError            chan error
}

func (fs *FSockPool) PopFSock() (fsock *FSock, err error) {
	if fs == nil {
		return nil, errors.New("Unconfigured ConnectionPool")
	}
	if len(fs.fSocks) != 0 { // Select directly if available, so we avoid randomness of selection
		fsock = <-fs.fSocks
		return
	}
	tm := time.NewTimer(fs.maxWaitConn)
	select { // No fsock available in the pool, wait for first one showing up
	case fsock = <-fs.fSocks:
		tm.Stop()
		return
	case <-fs.allowedConns:
		tm.Stop()
		return NewFSock(fs.fsAddr, fs.fsPasswd, fs.reconnects, fs.maxReconnectInterval, fs.delayFuncConstructor,
			fs.eventHandlers, fs.eventFilters, fs.logger, fs.connIdx, fs.bgapi, fs.stopError)
	case <-tm.C:
		return nil, ErrConnectionPoolTimeout
	}
}

func (fs *FSockPool) PushFSock(fsk *FSock) {
	if fs == nil { // Did not initialize the pool
		return
	}
	if fsk == nil || !fsk.Connected() {
		fs.allowedConns <- struct{}{}
		return
	}
	fs.fSocks <- fsk
}
