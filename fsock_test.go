package fsock

import (
	"bufio"
	"io/ioutil"
	"os"
	"reflect"
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
	FS.buffer = bufio.NewReader(r)
	w.Write([]byte(HEADER + BODY))
	h, b, err := FS.readEvent()
	if err != nil || h != HEADER[:len(HEADER)-1] || len(b) != 564 {
		t.Error("Error parsing event: ", h, b, len(b))
	}
}

func TestHeaderValMiddle(t *testing.T) {
	h := headerVal(BODY, "Event-Date-GMT")
	if h != "Fri,%2005%20Oct%202012%2011%3A41%3A38%20GMT" {
		t.Error("Header val error: ", h)
	}
}

func TestHeaderValStart(t *testing.T) {
	h := headerVal(BODY, "Event-Name")
	if h != "RE_SCHEDULE" {
		t.Error("Header val error: ", h)
	}
}

func TestHeaderValEnd(t *testing.T) {
	h := headerVal(BODY, "Task-Runtime")
	if h != "1349437318" {
		t.Error("Header val error: ", h)
	}
}

func TestEventToMapUnfiltered(t *testing.T) {
	fields := FSEventStrToMap(BODY, nil)
	if fields["Event-Name"] != "RE_SCHEDULE" {
		t.Error("Event not parsed correctly: ", fields)
	}
	if len(fields) != 17 {
		t.Error("Incorrect number of event fields: ", len(fields))
	}
}

func TestEventToMapFiltered(t *testing.T) {
	fields := FSEventStrToMap(BODY, []string{"Event-Name", "Task-Group", "Event-Date-GMT"})
	if fields["Event-Date-Local"] != "2012-10-05 13:41:38" {
		t.Error("Event not parsed correctly: ", fields)
	}
	if len(fields) != 14 {
		t.Error("Incorrect number of event fields: ", len(fields))
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
	FS = &FSock{}
	FS.buffer = bufio.NewReader(r)
	var events int32
	FS.eventHandlers = map[string][]func(string){
		"HEARTBEAT":                []func(string){func(string) { events++ }},
		"RE_SCHEDULE":              []func(string){func(string) { events++ }},
		"CHANNEL_STATE":            []func(string){func(string) { events++ }},
		"CODEC":                    []func(string){func(string) { events++ }},
		"CHANNEL_CREATE":           []func(string){func(string) { events++ }},
		"CHANNEL_CALLSTATE":        []func(string){func(string) { events++ }},
		"API":                      []func(string){func(string) { events++ }},
		"CHANNEL_EXECUTE":          []func(string){func(string) { events++ }},
		"CHANNEL_EXECUTE_COMPLETE": []func(string){func(string) { events++ }},
		"CHANNEL_PARK":             []func(string){func(string) { events++ }},
		"CHANNEL_HANGUP":           []func(string){func(string) { events++ }},
		"CHANNEL_HANGUP_COMPLETE":  []func(string){func(string) { events++ }},
		"CHANNEL_UNPARK":           []func(string){func(string) { events++ }},
		"CHANNEL_DESTROY":          []func(string){func(string) { events++ }},
	}
	go FS.ReadEvents()
	w.Write(data)
	time.Sleep(50 * time.Millisecond)
	if events != 45 {
		t.Error("Error reading events: ", events)
	}
}

func TestMapChanData(t *testing.T) {
	chanInfoStr := `uuid,direction,created,created_epoch,name,state,cid_name,cid_num,ip_addr,dest,application,application_data,dialplan,context,read_codec,read_rate,read_bit_rate,write_codec,write_rate,write_bit_rate,secure,hostname,presence_id,presence_data,callstate,callee_name,callee_num,callee_direction,call_uuid,sent_callee_name,sent_callee_num
fed464b3-a328-453f-9437-92b9b6a400fd,inbound,2014-10-26 18:08:32,1414343312,sofia/ipbxas/dan@172.16.254.66,CS_EXECUTE,dan,dan,172.16.254.66,+4986517174963,,,XML,ipbxas,PCMA,8000,64000,PCMA,8000,64000,,iPBXDev,dan@172.16.254.66,,HELD,,,,fed464b3-a328-453f-9437-92b9b6a400fd,,
c56125cc-024a-48a2-adbc-9612f6c02334,outbound,2014-10-26 18:08:32,1414343312,sofia/ipbxas/dan@172.16.254.66,CS_EXCHANGE_MEDIA,dan,+4986517174963,172.16.254.66,dan,playback,local_stream://moh,XML,ipbxas,PCMA,8000,64000,PCMA,8000,64000,,iPBXDev,dan@172.16.254.66,,ACTIVE,Outbound Call,dan,,fed464b3-a328-453f-9437-92b9b6a400fd,,
e604a792-172a-4e8f-8fc9-9198f0d15f15,inbound,2014-10-26 18:08:32,1414343312,sofia/loop_ipbxas/+4986517174963@172.16.254.66,CS_EXECUTE,dan,+4986517174963,127.0.0.1,dan,bridge,[sip_h_X-EpTransport=udp]sofia/ipbxas/dan@172.16.254.1:5060;registering_acc=172_16_254_66;fs_path=sip:172.16.254.66,XML,redirected,PCMA,8000,64000,PCMA,8000,64000,,iPBXDev,,,ACTIVE,Outbound Call,dan,SEND,e604a792-172a-4e8f-8fc9-9198f0d15f15,Outbound Call,dan
eacd0ae4-e1d5-447d-a7aa-e422a3a7abad,outbound,2014-10-26 18:08:32,1414343312,sofia/ipbxas/dan@172.16.254.1:5060,CS_EXCHANGE_MEDIA,dan,+4986517174963,127.0.0.1,dan,,,XML,redirected,PCMA,8000,64000,PCMA,8000,64000,,iPBXDev,,,ACTIVE,Outbound Call,dan,SEND,e604a792-172a-4e8f-8fc9-9198f0d15f15,dan,+4986517174963

4 total.
`
	eChanData := []map[string]string{
		map[string]string{"application": "", "call_uuid": "fed464b3-a328-453f-9437-92b9b6a400fd", "direction": "inbound", "name": "sofia/ipbxas/dan@172.16.254.66", "application_data": "", "callstate": "HELD",
			"created": "2014-10-26 18:08:32", "cid_num": "dan", "dialplan": "XML", "read_bit_rate": "64000", "hostname": "iPBXDev", "callee_num": "", "created_epoch": "1414343312", "dest": "+4986517174963",
			"write_codec": "PCMA", "presence_data": "", "callee_direction": "", "sent_callee_num": "", "uuid": "fed464b3-a328-453f-9437-92b9b6a400fd", "state": "CS_EXECUTE", "ip_addr": "172.16.254.66",
			"cid_name": "dan", "write_rate": "8000", "write_bit_rate": "64000", "callee_name": "", "context": "ipbxas", "read_codec": "PCMA", "read_rate": "8000", "secure": "",
			"presence_id": "dan@172.16.254.66", "sent_callee_name": ""},
		map[string]string{"application": "playback", "call_uuid": "fed464b3-a328-453f-9437-92b9b6a400fd", "direction": "outbound", "name": "sofia/ipbxas/dan@172.16.254.66", "application_data": "local_stream://moh",
			"callstate": "ACTIVE", "created": "2014-10-26 18:08:32", "cid_num": "+4986517174963", "dialplan": "XML", "read_bit_rate": "64000", "hostname": "iPBXDev", "callee_num": "dan", "created_epoch": "1414343312",
			"dest": "dan", "write_codec": "PCMA", "presence_data": "", "callee_direction": "", "sent_callee_num": "", "uuid": "c56125cc-024a-48a2-adbc-9612f6c02334", "state": "CS_EXCHANGE_MEDIA",
			"ip_addr": "172.16.254.66", "cid_name": "dan", "write_rate": "8000", "write_bit_rate": "64000", "callee_name": "Outbound Call", "context": "ipbxas", "read_codec": "PCMA", "read_rate": "8000",
			"secure": "", "presence_id": "dan@172.16.254.66", "sent_callee_name": ""},
		map[string]string{"dialplan": "XML", "read_codec": "PCMA", "secure": "", "hostname": "iPBXDev", "callstate": "ACTIVE", "callee_num": "dan", "sent_callee_name": "Outbound Call", "created_epoch": "1414343312",
			"application": "bridge", "write_codec": "PCMA", "write_rate": "8000", "presence_data": "", "call_uuid": "e604a792-172a-4e8f-8fc9-9198f0d15f15", "uuid": "e604a792-172a-4e8f-8fc9-9198f0d15f15", "cid_num": "+4986517174963",
			"application_data": "[sip_h_X-EpTransport=udp]sofia/ipbxas/dan@172.16.254.1:5060;registering_acc=172_16_254_66;fs_path=sip:172.16.254.66", "created": "2014-10-26 18:08:32", "dest": "dan", "direction": "inbound",
			"state": "CS_EXECUTE", "ip_addr": "127.0.0.1", "cid_name": "dan", "write_bit_rate": "64000", "sent_callee_num": "dan", "name": "sofia/loop_ipbxas/+4986517174963@172.16.254.66", "context": "redirected",
			"read_rate": "8000", "read_bit_rate": "64000", "presence_id": "", "callee_name": "Outbound Call", "callee_direction": "SEND"},
		map[string]string{"direction": "outbound", "state": "CS_EXCHANGE_MEDIA", "ip_addr": "127.0.0.1", "cid_name": "dan", "write_bit_rate": "64000", "sent_callee_num": "+4986517174963", "name": "sofia/ipbxas/dan@172.16.254.1:5060",
			"context": "redirected", "read_rate": "8000", "read_bit_rate": "64000", "presence_id": "", "callee_name": "Outbound Call", "callee_direction": "SEND", "dialplan": "XML", "read_codec": "PCMA", "secure": "",
			"hostname": "iPBXDev", "callstate": "ACTIVE", "callee_num": "dan", "sent_callee_name": "dan", "created_epoch": "1414343312", "application": "", "write_codec": "PCMA", "write_rate": "8000", "presence_data": "",
			"call_uuid": "e604a792-172a-4e8f-8fc9-9198f0d15f15", "uuid": "eacd0ae4-e1d5-447d-a7aa-e422a3a7abad", "cid_num": "+4986517174963", "application_data": "", "created": "2014-10-26 18:08:32", "dest": "dan"},
	}
	if rcvChanData := MapChanData(chanInfoStr); !reflect.DeepEqual(eChanData, rcvChanData) {
		t.Errorf("Expected: %+v, received: %+v", eChanData, rcvChanData)
	}
}

/*********************** Benchmarks ************************/

func BenchmarkHeaderVal(b *testing.B) {
	for i := 0; i < b.N; i++ {
		headerVal(HEADER, "Content-Length")
		headerVal(BODY, "Event-Date-Loca")
	}
}
