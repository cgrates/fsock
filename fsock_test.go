/*
fsock_test.go is released under the MIT License <http://www.opensource.org/licenses/mit-license.php
Copyright (C) ITsysCOM. All Rights Reserved.

Provides FreeSWITCH socket communication.
*/
package fsock

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

/*

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
			fs := FSock{
				buffer:  bufio.NewReaderSize(buf, 8192),
				logger:  nopLogger{},
				fsMutex: new(sync.RWMutex),
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

func fillBuffer(buf *bytes.Buffer, content string) (int, error) {
	buf.Reset()
	return buf.Write([]byte(content))
}

func BenchmarkFSockReadBody(b *testing.B) {
	content := "Event-Name: CHANNEL_PARK\nCore-UUID: 44d90754-93de-4dd7-807a-9ad31e45d4de\nFreeSWITCH-Hostname: debian12\nFreeSWITCH-Switchname: debian12\nFreeSWITCH-IPv4: 10.0.2.15\nFreeSWITCH-IPv6: %3A%3A1\nEvent-Date-Local: 2023-12-22%2010%3A12%3A32\nEvent-Date-GMT: Fri,%2022%20Dec%202023%2015%3A12%3A32%20GMT\nEvent-Date-Timestamp: 1703257952506074\nEvent-Calling-File: switch_ivr.c\nEvent-Calling-Function: switch_ivr_park\nEvent-Calling-Line-Number: 1002\nEvent-Sequence: 498\nChannel-State: CS_EXECUTE\nChannel-Call-State: RINGING\nChannel-State-Number: 4\nChannel-Name: sofia/internal/1001%40192.168.56.120%3A5081\nUnique-ID: 4967ceb1-c6f9-4af9-9855-df323d6763ad\nCall-Direction: inbound\nPresence-Call-Direction: inbound\nChannel-HIT-Dialplan: true\nChannel-Presence-ID: 1001%40192.168.56.120\nChannel-Call-UUID: 4967ceb1-c6f9-4af9-9855-df323d6763ad\nAnswer-State: ringing\nCaller-Direction: inbound\nCaller-Logical-Direction: inbound\nCaller-Username: 1001\nCaller-Dialplan: XML\nCaller-Caller-ID-Name: 1001\nCaller-Caller-ID-Number: 1001\nCaller-Orig-Caller-ID-Name: 1001\nCaller-Orig-Caller-ID-Number: 1001\nCaller-Network-Addr: 192.168.56.120\nCaller-ANI: 1001\nCaller-Destination-Number: 1002\nCaller-Unique-ID: 4967ceb1-c6f9-4af9-9855-df323d6763ad\nCaller-Source: mod_sofia\nCaller-Context: default\nCaller-Channel-Name: sofia/internal/1001%40192.168.56.120%3A5081\nCaller-Profile-Index: 1\nCaller-Profile-Created-Time: 1703257952506074\nCaller-Channel-Created-Time: 1703257952506074\nCaller-Channel-Answered-Time: 0\nCaller-Channel-Progress-Time: 0\nCaller-Channel-Progress-Media-Time: 0\nCaller-Channel-Hangup-Time: 0\nCaller-Channel-Transfer-Time: 0\nCaller-Channel-Resurrect-Time: 0\nCaller-Channel-Bridged-Time: 0\nCaller-Channel-Last-Hold: 0\nCaller-Channel-Hold-Accum: 0\nCaller-Screen-Bit: true\nCaller-Privacy-Hide-Name: false\nCaller-Privacy-Hide-Number: false\nvariable_direction: inbound\nvariable_uuid: 4967ceb1-c6f9-4af9-9855-df323d6763ad\nvariable_session_id: 1\nvariable_sip_from_user: 1001\nvariable_sip_from_port: 5081\nvariable_sip_from_uri: 1001%40192.168.56.120%3A5081\nvariable_sip_from_host: 192.168.56.120\nvariable_video_media_flow: disabled\nvariable_audio_media_flow: disabled\nvariable_text_media_flow: disabled\nvariable_channel_name: sofia/internal/1001%40192.168.56.120%3A5081\nvariable_sip_call_id: 1-27764%40192.168.56.120\nvariable_sip_local_network_addr: 192.168.56.120\nvariable_sip_network_ip: 192.168.56.120\nvariable_sip_network_port: 5081\nvariable_sip_invite_stamp: 1703257952506074\nvariable_sip_received_ip: 192.168.56.120\nvariable_sip_received_port: 5081\nvariable_sip_via_protocol: udp\nvariable_sip_authorized: true\nvariable_sip_acl_authed_by: domains\nvariable_sip_from_user_stripped: 1001\nvariable_sip_from_tag: 27764SIPpTag001\nvariable_sofia_profile_name: internal\nvariable_sofia_profile_url: sip%3Amod_sofia%40192.168.56.120%3A5060\nvariable_recovery_profile_name: internal\nvariable_sip_full_via: SIP/2.0/UDP%20192.168.56.120%3A5081%3Bbranch%3Dz9hG4bK-27764-1-0\nvariable_sip_from_display: 1001\nvariable_sip_full_from: 1001%20%3Csip%3A1001%40192.168.56.120%3A5081%3E%3Btag%3D27764SIPpTag001\nvariable_sip_to_display: 1002\nvariable_sip_full_to: 1002%20%3Csip%3A1002%40192.168.56.120%3A5060%3E\nvariable_sip_req_user: 1002\nvariable_sip_req_port: 5060\nvariable_sip_req_uri: 1002%40192.168.56.120%3A5060\nvariable_sip_req_host: 192.168.56.120\nvariable_sip_to_user: 1002\nvariable_sip_to_port: 5060\nvariable_sip_to_uri: 1002%40192.168.56.120%3A5060\nvariable_sip_to_host: 192.168.56.120\nvariable_sip_contact_user: sipp\nvariable_sip_contact_port: 5081\nvariable_sip_contact_uri: sipp%40192.168.56.120%3A5081\nvariable_sip_contact_host: 192.168.56.120\nvariable_rtp_use_codec_string: G722,PCMU,PCMA,GSM\nvariable_sip_subject: Performance%20Test\nvariable_sip_via_host: 192.168.56.120\nvariable_sip_via_port: 5081\nvariable_max_forwards: 70\nvariable_presence_id: 1001%40192.168.56.120\nvariable_switch_r_sdp: v%3D0%0D%0Ao%3Duser1%2053655765%202353687637%20IN%20IP4%20192.168.56.120%0D%0As%3D-%0D%0Ac%3DIN%20IP4%20192.168.56.120%0D%0At%3D0%200%0D%0Am%3Daudio%206000%20RTP/AVP%200%0D%0Aa%3Drtpmap%3A0%20PCMU/8000%0D%0A\nvariable_ep_codec_string: CORE_PCM_MODULE.PCMU%408000h%4020i%4064000b\nvariable_endpoint_disposition: DELAYED%20NEGOTIATION\nvariable_call_uuid: 4967ceb1-c6f9-4af9-9855-df323d6763ad\nvariable_current_application: park\n\n"
	buf := &bytes.Buffer{}
	fs := &FSock{
		logger:  nopLogger{},
		buffer:  bufio.NewReaderSize(buf, 8092),
		fsMutex: new(sync.RWMutex),
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
*/
