package fsock

import (
	"bufio"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
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

func TestIndexStringAll(t *testing.T) {
	testStr := "a,b,c"
	if indxAll := IndexStringAll(testStr, ","); !reflect.DeepEqual([]int{1, 3}, indxAll) {
		t.Errorf("Expected %+v, received: %+v", []int{1, 3}, indxAll)
	}
	testStr = "a,,b,c,,"
	if indxAll := IndexStringAll(testStr, ",,"); !reflect.DeepEqual([]int{1, 6}, indxAll) {
		t.Errorf("Expected %+v, received: %+v", []int{1, 6}, indxAll)
	}
}

func TestSplitIgnoreGroups(t *testing.T) {
	strNoGroups := "d775e082-4309-4629-b08a-ae174271f2e1,outbound,2014-10-27 10:30:11,1414402211,sofia/ipbxas/dan@172.16.254.66,CS_EXCHANGE_MEDIA,dan,+4986517174963,172.16.254.66,dan,,,XML,ipbxas,PCMA,8000,64000,PCMA,8000,64000,,iPBXDev,dan@172.16.254.66,,ACTIVE,Outbound Call,dan,,ba23506f-e36b-4c12-9c17-9146077bb240,,"
	if !reflect.DeepEqual(strings.Split(strNoGroups, ","), SplitIgnoreGroups(strNoGroups, ",")) {
		t.Errorf("NormalSplit: \n%+v\n, resultWithGroups: \n%+v\n", strings.Split(strNoGroups, ","), SplitIgnoreGroups(strNoGroups, ","))
	}
	strNoGroups2 := "d775e082-4309-4629-b08a-ae174271f2e1,outbound,2014-10-27 10:30:11,1414402211,sofia/ipbxas/dan@172.16.254.66,CS_EXCHANGE_MEDIA,dan,+4986517174963,172.16.254.66,dan,,,XML,ipbxas,PCMA,8000,64000,PCMA,8000,64000,,iPBXDev,dan@172.16.254.66,,ACTIVE,Outbound Call,dan,,ba23506f-e36b-4c12-9c17-9146077bb240"
	if !reflect.DeepEqual(strings.Split(strNoGroups2, ","), SplitIgnoreGroups(strNoGroups2, ",")) {
		t.Errorf("NormalSplit: \n%+v\n, resultWithGroups: \n%+v\n", strings.Split(strNoGroups, ","), SplitIgnoreGroups(strNoGroups, ","))
	}
	strNoGroups3 := ",d775e082-4309-4629-b08a-ae174271f2e1,outbound,2014-10-27 10:30:11,1414402211,sofia/ipbxas/dan@172.16.254.66,CS_EXCHANGE_MEDIA,dan,+4986517174963,172.16.254.66,dan,,,XML,ipbxas,PCMA,8000,64000,PCMA,8000,64000,,iPBXDev,dan@172.16.254.66,,ACTIVE,Outbound Call,dan,,ba23506f-e36b-4c12-9c17-9146077bb240"
	if !reflect.DeepEqual(strings.Split(strNoGroups3, ","), SplitIgnoreGroups(strNoGroups3, ",")) {
		t.Errorf("NormalSplit: \n%+v\n, resultWithGroups: \n%+v\n", strings.Split(strNoGroups, ","), SplitIgnoreGroups(strNoGroups, ","))
	}
	strWithGroups := "ba23506f-e36b-4c12-9c17-9146077bb240,inbound,2014-10-27 10:30:11,1414402211,sofia/ipbxas/dan@172.16.254.66,CS_EXECUTE,dan,dan,172.16.254.66,+4986517174963,bridge,{sip_contact_user=iPBXSuite}[origination_caller_id_number=+4986517174963,to_domain_tag=172.16.254.66,sip_h_X-CalledEPType=SIP,sip_h_X-CalledEPTag=dan,sip_h_X-ForwardedCall=false,presence_id=dan@172.16.254.66,leg_progress_timeout=50,leg_timeout=100,to_ep_type=SIP,to_ep_tag=dan,sip_h_X-CalledDomainTag=172.16.254.66,sip_h_X-Billable=false,sip_h_X-LoopApp=LOOP_ROUTED]sofia/ipbxas/dan@172.16.254.66;fs_path=sip:127.0.0.1,XML,ipbxas,PCMA,8000,64000,PCMA,8000,64000,,iPBXDev,dan@172.16.254.66,,ACTIVE,,,,ba23506f-e36b-4c12-9c17-9146077bb240,,"
	eSplt := []string{"ba23506f-e36b-4c12-9c17-9146077bb240", "inbound", "2014-10-27 10:30:11", "1414402211", "sofia/ipbxas/dan@172.16.254.66", "CS_EXECUTE", "dan", "dan", "172.16.254.66", "+4986517174963", "bridge",
		"{sip_contact_user=iPBXSuite}[origination_caller_id_number=+4986517174963,to_domain_tag=172.16.254.66,sip_h_X-CalledEPType=SIP,sip_h_X-CalledEPTag=dan,sip_h_X-ForwardedCall=false,presence_id=dan@172.16.254.66,leg_progress_timeout=50,leg_timeout=100,to_ep_type=SIP,to_ep_tag=dan,sip_h_X-CalledDomainTag=172.16.254.66,sip_h_X-Billable=false,sip_h_X-LoopApp=LOOP_ROUTED]sofia/ipbxas/dan@172.16.254.66;fs_path=sip:127.0.0.1",
		"XML", "ipbxas", "PCMA", "8000", "64000", "PCMA", "8000", "64000", "", "iPBXDev", "dan@172.16.254.66", "", "ACTIVE", "", "", "", "ba23506f-e36b-4c12-9c17-9146077bb240", "", ""}
	if splt := SplitIgnoreGroups(strWithGroups, ","); !reflect.DeepEqual(eSplt, splt) {
		t.Errorf("Expecting : %+v, received: %v", eSplt, splt)
	}
}

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

func TestMapChanData2(t *testing.T) {
	chanInfoStr := `uuid,direction,created,created_epoch,name,state,cid_name,cid_num,ip_addr,dest,application,application_data,dialplan,context,read_codec,read_rate,read_bit_rate,write_codec,write_rate,write_bit_rate,secure,hostname,presence_id,presence_data,callstate,callee_name,callee_num,callee_direction,call_uuid,sent_callee_name,sent_callee_num
ba23506f-e36b-4c12-9c17-9146077bb240,inbound,2014-10-27 10:30:11,1414402211,sofia/ipbxas/dan@172.16.254.66,CS_EXECUTE,dan,dan,172.16.254.66,+4986517174963,bridge,{sip_contact_user=iPBXSuite}[origination_caller_id_number=+4986517174963,to_domain_tag=172.16.254.66,sip_h_X-CalledEPType=SIP,sip_h_X-CalledEPTag=dan,sip_h_X-ForwardedCall=false,presence_id=dan@172.16.254.66,leg_progress_timeout=50,leg_timeout=100,to_ep_type=SIP,to_ep_tag=dan,sip_h_X-CalledDomainTag=172.16.254.66,sip_h_X-Billable=false,sip_h_X-LoopApp=LOOP_ROUTED]sofia/ipbxas/dan@172.16.254.66;fs_path=sip:127.0.0.1,XML,ipbxas,PCMA,8000,64000,PCMA,8000,64000,,iPBXDev,dan@172.16.254.66,,ACTIVE,,,,ba23506f-e36b-4c12-9c17-9146077bb240,,
d775e082-4309-4629-b08a-ae174271f2e1,outbound,2014-10-27 10:30:11,1414402211,sofia/ipbxas/dan@172.16.254.66,CS_EXCHANGE_MEDIA,dan,+4986517174963,172.16.254.66,dan,,,XML,ipbxas,PCMA,8000,64000,PCMA,8000,64000,,iPBXDev,dan@172.16.254.66,,ACTIVE,Outbound Call,dan,,ba23506f-e36b-4c12-9c17-9146077bb240,,
7c6a423e-7d2d-40c3-8f7f-06dc534d6576,inbound,2014-10-27 10:30:11,1414402211,sofia/loop_ipbxas/+4986517174963@172.16.254.66,CS_EXECUTE,dan,+4986517174963,127.0.0.1,dan,playback,local_stream://moh,XML,redirected,PCMA,8000,64000,PCMA,8000,64000,,iPBXDev,,,ACTIVE,Outbound Call,dan,SEND,7c6a423e-7d2d-40c3-8f7f-06dc534d6576,Outbound Call,dan
81a05714-5a89-4a1c-848c-5e592527ae03,outbound,2014-10-27 10:30:11,1414402211,sofia/ipbxas/dan@172.16.254.1:5060,CS_EXCHANGE_MEDIA,dan,+4986517174963,127.0.0.1,dan,,,XML,redirected,PCMA,8000,64000,PCMA,8000,64000,,iPBXDev,,,HELD,Outbound Call,dan,SEND,7c6a423e-7d2d-40c3-8f7f-06dc534d6576,dan,+4986517174963

4 total.
`
	eChanData := []map[string]string{
		map[string]string{"created_epoch": "1414402211", "application": "bridge", "write_codec": "PCMA", "write_rate": "8000", "presence_data": "", "call_uuid": "ba23506f-e36b-4c12-9c17-9146077bb240", "uuid": "ba23506f-e36b-4c12-9c17-9146077bb240", "cid_num": "dan",
			"application_data": "{sip_contact_user=iPBXSuite}[origination_caller_id_number=+4986517174963,to_domain_tag=172.16.254.66,sip_h_X-CalledEPType=SIP,sip_h_X-CalledEPTag=dan,sip_h_X-ForwardedCall=false,presence_id=dan@172.16.254.66,leg_progress_timeout=50,leg_timeout=100,to_ep_type=SIP,to_ep_tag=dan,sip_h_X-CalledDomainTag=172.16.254.66,sip_h_X-Billable=false,sip_h_X-LoopApp=LOOP_ROUTED]sofia/ipbxas/dan@172.16.254.66;fs_path=sip:127.0.0.1",
			"created":          "2014-10-27 10:30:11", "dest": "+4986517174963", "direction": "inbound", "state": "CS_EXECUTE", "ip_addr": "172.16.254.66", "cid_name": "dan", "write_bit_rate": "64000", "sent_callee_num": "", "name": "sofia/ipbxas/dan@172.16.254.66", "context": "ipbxas", "read_rate": "8000",
			"read_bit_rate": "64000", "presence_id": "dan@172.16.254.66", "callee_name": "", "callee_direction": "", "dialplan": "XML", "read_codec": "PCMA", "secure": "", "hostname": "iPBXDev", "callstate": "ACTIVE", "callee_num": "", "sent_callee_name": ""},
		map[string]string{"state": "CS_EXCHANGE_MEDIA", "dialplan": "XML", "read_codec": "PCMA", "secure": "", "hostname": "iPBXDev", "callstate": "ACTIVE", "callee_num": "dan", "write_codec": "PCMA", "write_bit_rate": "64000",
			"call_uuid": "ba23506f-e36b-4c12-9c17-9146077bb240", "context": "ipbxas", "read_rate": "8000", "read_bit_rate": "64000", "presence_id": "dan@172.16.254.66", "created": "2014-10-27 10:30:11", "dest": "dan", "callee_name": "Outbound Call",
			"callee_direction": "", "direction": "outbound", "ip_addr": "172.16.254.66", "sent_callee_name": "", "created_epoch": "1414402211", "cid_name": "dan", "application": "", "write_rate": "8000", "presence_data": "", "sent_callee_num": "",
			"uuid": "d775e082-4309-4629-b08a-ae174271f2e1", "name": "sofia/ipbxas/dan@172.16.254.66", "cid_num": "+4986517174963", "application_data": ""},
		map[string]string{"cid_name": "dan", "write_rate": "8000", "write_bit_rate": "64000", "callee_name": "Outbound Call", "context": "redirected", "read_codec": "PCMA", "read_rate": "8000", "secure": "", "presence_id": "", "sent_callee_name": "Outbound Call",
			"application": "playback", "call_uuid": "7c6a423e-7d2d-40c3-8f7f-06dc534d6576", "direction": "inbound", "name": "sofia/loop_ipbxas/+4986517174963@172.16.254.66", "application_data": "local_stream://moh", "callstate": "ACTIVE", "created": "2014-10-27 10:30:11",
			"cid_num": "+4986517174963", "dialplan": "XML", "read_bit_rate": "64000", "hostname": "iPBXDev", "callee_num": "dan", "created_epoch": "1414402211", "dest": "dan", "write_codec": "PCMA", "presence_data": "", "callee_direction": "SEND", "sent_callee_num": "dan",
			"uuid": "7c6a423e-7d2d-40c3-8f7f-06dc534d6576", "state": "CS_EXECUTE", "ip_addr": "127.0.0.1"},
		map[string]string{"ip_addr": "127.0.0.1", "application_data": "", "cid_name": "dan", "cid_num": "+4986517174963", "read_codec": "PCMA", "read_bit_rate": "64000", "secure": "", "created_epoch": "1414402211", "presence_data": "", "callee_direction": "SEND",
			"call_uuid": "7c6a423e-7d2d-40c3-8f7f-06dc534d6576", "uuid": "81a05714-5a89-4a1c-848c-5e592527ae03", "direction": "outbound", "name": "sofia/ipbxas/dan@172.16.254.1:5060", "state": "CS_EXCHANGE_MEDIA", "callstate": "HELD", "created": "2014-10-27 10:30:11",
			"write_rate": "8000", "write_bit_rate": "64000", "callee_name": "Outbound Call", "dialplan": "XML", "context": "redirected", "read_rate": "8000", "hostname": "iPBXDev", "presence_id": "", "callee_num": "dan", "sent_callee_name": "dan", "dest": "dan",
			"application": "", "write_codec": "PCMA", "sent_callee_num": "+4986517174963"},
	}
	if rcvChanData := MapChanData(chanInfoStr); !reflect.DeepEqual(eChanData, rcvChanData) {
		t.Errorf("Expected: %+v, received: %+v", eChanData, rcvChanData)
	}
	/*if rcvChanData := MapChanData(chanInfoStr); len(rcvChanData) != 4 {
		t.Errorf("Unexpected channel map received: %+v", rcvChanData)
	}
	*/
}

/*********************** Benchmarks ************************/

func BenchmarkHeaderVal(b *testing.B) {
	for i := 0; i < b.N; i++ {
		headerVal(HEADER, "Content-Length")
		headerVal(BODY, "Event-Date-Loca")
	}
}
