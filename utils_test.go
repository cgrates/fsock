/*
utils_test.go is released under the MIT License <http://www.opensource.org/licenses/mit-license.php
Copyright (C) ITsysCOM. All Rights Reserved.

Provides FreeSWITCH socket communication.
*/
package fsock

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func TestSplitIgnoreGroups(t *testing.T) {
	strNoGroups := "d775e082-4309-4629-b08a-ae174271f2e1,outbound,2014-10-27 10:30:11,1414402211,sofia/ipbxas/dan@172.16.254.66,CS_EXCHANGE_MEDIA,dan,+4986517174963,172.16.254.66,dan,,,XML,ipbxas,PCMA,8000,64000,PCMA,8000,64000,,iPBXDev,dan@172.16.254.66,,ACTIVE,Outbound Call,dan,,ba23506f-e36b-4c12-9c17-9146077bb240,,"
	if !reflect.DeepEqual(strings.Split(strNoGroups, ","), splitIgnoreGroups(strNoGroups, ",", 0)) {
		t.Errorf("NormalSplit: \n%+v\n, resultWithGroups: \n%+v\n", strings.Split(strNoGroups, ","), splitIgnoreGroups(strNoGroups, ",", 0))
	}
	strNoGroups2 := "d775e082-4309-4629-b08a-ae174271f2e1,outbound,2014-10-27 10:30:11,1414402211,sofia/ipbxas/dan@172.16.254.66,CS_EXCHANGE_MEDIA,dan,+4986517174963,172.16.254.66,dan,,,XML,ipbxas,PCMA,8000,64000,PCMA,8000,64000,,iPBXDev,dan@172.16.254.66,,ACTIVE,Outbound Call,dan,,ba23506f-e36b-4c12-9c17-9146077bb240"
	if !reflect.DeepEqual(strings.Split(strNoGroups2, ","), splitIgnoreGroups(strNoGroups2, ",", 0)) {
		t.Errorf("NormalSplit: \n%+v\n, resultWithGroups: \n%+v\n", strings.Split(strNoGroups, ","), splitIgnoreGroups(strNoGroups, ",", 0))
	}
	strNoGroups3 := ",d775e082-4309-4629-b08a-ae174271f2e1,outbound,2014-10-27 10:30:11,1414402211,sofia/ipbxas/dan@172.16.254.66,CS_EXCHANGE_MEDIA,dan,+4986517174963,172.16.254.66,dan,,,XML,ipbxas,PCMA,8000,64000,PCMA,8000,64000,,iPBXDev,dan@172.16.254.66,,ACTIVE,Outbound Call,dan,,ba23506f-e36b-4c12-9c17-9146077bb240"
	if !reflect.DeepEqual(strings.Split(strNoGroups3, ","), splitIgnoreGroups(strNoGroups3, ",", 0)) {
		t.Errorf("NormalSplit: \n%+v\n, resultWithGroups: \n%+v\n", strings.Split(strNoGroups, ","), splitIgnoreGroups(strNoGroups, ",", 0))
	}
	strWithGroups := "ba23506f-e36b-4c12-9c17-9146077bb240,inbound,2014-10-27 10:30:11,1414402211,sofia/ipbxas/dan@172.16.254.66,CS_EXECUTE,dan,dan,172.16.254.66,+4986517174963,bridge,{sip_contact_user=iPBXSuite}[origination_caller_id_number=+4986517174963,to_domain_tag=172.16.254.66,sip_h_X-CalledEPType=SIP,sip_h_X-CalledEPTag=dan,sip_h_X-ForwardedCall=false,presence_id=dan@172.16.254.66,leg_progress_timeout=50,leg_timeout=100,to_ep_type=SIP,to_ep_tag=dan,sip_h_X-CalledDomainTag=172.16.254.66,sip_h_X-Billable=false,sip_h_X-LoopApp=LOOP_ROUTED]sofia/ipbxas/dan@172.16.254.66;fs_path=sip:127.0.0.1,XML,ipbxas,PCMA,8000,64000,PCMA,8000,64000,,iPBXDev,dan@172.16.254.66,,ACTIVE,,,,ba23506f-e36b-4c12-9c17-9146077bb240,,"
	eSplt := []string{"ba23506f-e36b-4c12-9c17-9146077bb240", "inbound", "2014-10-27 10:30:11", "1414402211", "sofia/ipbxas/dan@172.16.254.66", "CS_EXECUTE", "dan", "dan", "172.16.254.66", "+4986517174963", "bridge",
		"{sip_contact_user=iPBXSuite}[origination_caller_id_number=+4986517174963,to_domain_tag=172.16.254.66,sip_h_X-CalledEPType=SIP,sip_h_X-CalledEPTag=dan,sip_h_X-ForwardedCall=false,presence_id=dan@172.16.254.66,leg_progress_timeout=50,leg_timeout=100,to_ep_type=SIP,to_ep_tag=dan,sip_h_X-CalledDomainTag=172.16.254.66,sip_h_X-Billable=false,sip_h_X-LoopApp=LOOP_ROUTED]sofia/ipbxas/dan@172.16.254.66;fs_path=sip:127.0.0.1",
		"XML", "ipbxas", "PCMA", "8000", "64000", "PCMA", "8000", "64000", "", "iPBXDev", "dan@172.16.254.66", "", "ACTIVE", "", "", "", "ba23506f-e36b-4c12-9c17-9146077bb240", "", ""}
	if splt := splitIgnoreGroups(strWithGroups, ",", 0); !reflect.DeepEqual(eSplt, splt) {
		t.Errorf("Expecting : %+v, received: %v", eSplt, splt)
	}
	strWithGroups = "8009b347-fe46-4c99-9bb8-89e52e05d35f,inbound,2014-11-19 12:05:13,1416395113,sofia/ipbxas/+4986517174963@1.2.3.4,CS_EXECUTE,004986517174963,+4986517174963,2.3.4.5,0049850210795,bridge,{sip_contact_user=CloudIPBX.com,bridge_early_media=true}[to_domain_tag=sip.test.cloudipbx.com,sip_h_X-CalledEPType=SIP,sip_h_X-CalledEPTag=user3,sip_h_X-ForwardedCall=false,sip_h_X-LoopApp=LOOP_ROUTED,origination_caller_id_number=+4986517174963,to_ep_type=SIP,to_ep_tag=user3,sip_h_X-CalledDomainTag=sip.test.cloudipbx.com,sip_h_X-Billable=false,max_forwards=50]sofia/ipbxas/user3@sip.test.cloudipbx.com;fs_path=sip:127.0.0.1;transport=tcp,[to_domain_tag=sip.test.cloudipbx.com,sip_h_X-ForwardedCall=false,sip_h_X-Billable=false,sip_h_X-LoopApp=LOOP_ROUTED,origination_caller_id_number=+4986517174963,to_ep_type=SIP,to_ep_tag=user4,sip_h_X-CalledDomainTag=sip.test.cloudipbx.com,sip_h_X-CalledEPType=SIP,sip_h_X-CalledEPTag=user4,max_forwards=50]sofia/ipbxas/user4@sip.test.cloudipbx.com;fs_path=sip:127.0.0.1;transport=tcp,XML,ipbxas,PCMA,8000,64000,PCMA,8000,64000,,nl-asd-dev-sbc01,+4986517174963@1.2.3.4,,ACTIVE,,,,8009b347-fe46-4c99-9bb8-89e52e05d35f,,,004986517174963,+4986517174963,2.3.4.5,0049850210795,XML,ipbxas"
	eSplt = []string{"8009b347-fe46-4c99-9bb8-89e52e05d35f", "inbound", "2014-11-19 12:05:13", "1416395113", "sofia/ipbxas/+4986517174963@1.2.3.4", "CS_EXECUTE", "004986517174963", "+4986517174963", "2.3.4.5", "0049850210795", "bridge",
		"{sip_contact_user=CloudIPBX.com,bridge_early_media=true}[to_domain_tag=sip.test.cloudipbx.com,sip_h_X-CalledEPType=SIP,sip_h_X-CalledEPTag=user3,sip_h_X-ForwardedCall=false,sip_h_X-LoopApp=LOOP_ROUTED,origination_caller_id_number=+4986517174963,to_ep_type=SIP,to_ep_tag=user3,sip_h_X-CalledDomainTag=sip.test.cloudipbx.com,sip_h_X-Billable=false,max_forwards=50]sofia/ipbxas/user3@sip.test.cloudipbx.com;fs_path=sip:127.0.0.1;transport=tcp",
		"[to_domain_tag=sip.test.cloudipbx.com,sip_h_X-ForwardedCall=false,sip_h_X-Billable=false,sip_h_X-LoopApp=LOOP_ROUTED,origination_caller_id_number=+4986517174963,to_ep_type=SIP,to_ep_tag=user4,sip_h_X-CalledDomainTag=sip.test.cloudipbx.com,sip_h_X-CalledEPType=SIP,sip_h_X-CalledEPTag=user4,max_forwards=50]sofia/ipbxas/user4@sip.test.cloudipbx.com;fs_path=sip:127.0.0.1;transport=tcp",
		"XML", "ipbxas", "PCMA", "8000", "64000", "PCMA", "8000", "64000", "", "nl-asd-dev-sbc01", "+4986517174963@1.2.3.4", "", "ACTIVE", "", "", "", "8009b347-fe46-4c99-9bb8-89e52e05d35f", "", "", "004986517174963", "+4986517174963", "2.3.4.5", "0049850210795", "XML", "ipbxas"}
	if splt := splitIgnoreGroups(strWithGroups, ",", 0); !reflect.DeepEqual(eSplt, splt) {
		t.Errorf("Expecting : %+v, received: %v", eSplt, splt)
	}
	strWithGroups = "c2a2b753-8283-4347-94be-7df10a7710e3,inbound,2014-11-30 19:17:59,1417371479,sofia/loop_ipbxas/08765431@192.168.42.142,CS_EXECUTE,anonymous,08765431,192.168.42.142,eprou-eibhoog_g204_EPR,bridge,{sip_contact_user=Sipean,bridge_early_media=true}[origination_caller_id_name=DLH,to_ep_type=SIP,sip_h_X-CalledEPType=SIP,sip_h_X-CalledEPTag=user0008,sip_h_X-ForwardedCall=false,max_forwards=50,origination_caller_id_number=anonymous,to_domain_tag=sampledomain.com,to_ep_tag=user0008,sip_h_X-CalledDomainTag=sampledomain.com,sip_h_X-Billable=false,sip_h_X-LoopApp=LOOP_ROUTED]sofia/ipbxas/user0008@sampledomain.com;fs_path=sip:127.0.0.1;transport=tcp,[to_domain_tag=sampledomain.com,to_ep_type=SIP,to_ep_tag=user0010,sip_h_X-CalledDomainTag=sampledomain.com,sip_h_X-CalledEPTag=user0010,sip_h_X-LoopApp=LOOP_ROUTED,max_forwards=50,origination_caller_id_name=DLH,origination_caller_id_number=anonymous,sip_h_X-CalledEPType=SIP,sip_h_X-ForwardedCall=false,sip_h_X-Billable=false]sofia/ipbxas/user0010@sampledomain.com;fs_path=sip:127.0.0.1;transport=tcp,[to_domain_tag=sampledomain.com,to_ep_type=SIP,to_ep_tag=user0009,sip_h_X-CalledDomainTag=sampledomain.com,sip_h_X-ForwardedCall=false,sip_h_X-LoopApp=LOOP_ROUTED,origination_caller_id_name=DLH,origination_caller_id_number=anonymous,sip_h_X-CalledEPType=SIP,sip_h_X-CalledEPTag=user0009,sip_h_X-Billable=false,max_forwards=50]sofia/ipbxas/user0009@sampledomain.com;fs_path=sip:127.0.0.1;transport=tcp,[origination_caller_id_name=DLH,sip_h_X-CalledDomainTag=sampledomain.com,sip_h_X-ForwardedCall=false,sip_h_X-LoopApp=LOOP_ROUTED,origination_caller_id_number=anonymous,to_domain_tag=sampledomain.com,to_ep_type=SIP,to_ep_tag=user0001,sip_h_X-CalledEPType=SIP,sip_h_X-CalledEPTag=user0001,sip_h_X-Billable=false,max_forwards=50]sofia/ipbxas/user0001@sampledomain.com;fs_path=sip:127.0.0.1;transport=tcp,[origination_caller_id_name=DLH,to_ep_type=SIP,to_ep_tag=user0011,sip_h_X-CalledEPTag=user0011,origination_caller_id_number=anonymous,to_domain_tag=sampledomain.com,sip_h_X-CalledDomainTag=sampledomain.com,sip_h_X-CalledEPType=SIP,sip_h_X-ForwardedCall=false,sip_h_X-Billable=false,sip_h_X-LoopApp=LOOP_ROUTED,max_forwards=50]sofia/ipbxas/user0011@sampledomain.com;fs_path=sip:127.0.0.1;transport=tcp,[to_ep_type=SIP,sip_h_X-ForwardedCall=false,sip_h_X-Billable=false,sip_h_X-LoopApp=LOOP_ROUTED,origination_caller_id_name=DLH,origination_caller_id_number=anonymous,to_domain_tag=sampledomain.com,to_ep_tag=user0002,sip_h_X-CalledDomainTag=sampledomain.com,sip_h_X-CalledEPType=SIP,sip_h_X-CalledEPTag=user0002,max_forwards=50]sofia/ipbxas/user0002@sampledomain.com;fs_path=sip:127.0.0.1;transport=tcp,[origination_caller_id_name=DLH,origination_caller_id_number=anonymous,to_domain_tag=sampledomain.com,sip_h_X-CalledEPType=SIP,sip_h_X-Billable=false,sip_h_X-LoopApp=LOOP_ROUTED,max_forwards=50,to_ep_type=SIP,to_ep_tag=user0003,sip_h_X-CalledDomainTag=sampledomain.com,sip_h_X-CalledEPTag=user0003,sip_h_X-ForwardedCall=false]sofia/ipbxas/user0003@sampledomain.com;fs_path=sip:127.0.0.1;transport=tcp,[origination_caller_id_name=DLH,to_ep_type=SIP,sip_h_X-CalledDomainTag=sampledomain.com,sip_h_X-CalledEPTag=user0004,sip_h_X-ForwardedCall=false,max_forwards=50,origination_caller_id_number=anonymous,to_domain_tag=sampledomain.com,to_ep_tag=user0004,sip_h_X-CalledEPType=SIP,sip_h_X-Billable=false,sip_h_X-LoopApp=LOOP_ROUTED]sofia/ipbxas/user0004@sampledomain.com;fs_path=sip:127.0.0.1;transport=tcp,[to_domain_tag=sampledomain.com,to_ep_type=SIP,sip_h_X-CalledDomainTag=sampledomain.com,sip_h_X-CalledEPType=SIP,sip_h_X-ForwardedCall=false,sip_h_X-Billable=false,sip_h_X-LoopApp=LOOP_ROUTED,max_forwards=50,origination_caller_id_name=DLH,origination_caller_id_number=anonymous,to_ep_tag=user0012,sip_h_X-CalledEPTag=user0012]sofia/ipbxas/user0012@sampledomain.com;fs_path=sip:127.0.0.1;transport=tcp,XML,ipbxas_lo,PCMA,8000,64000,PCMA,8000,64000,,nl-asd-gls-sbc01,,,RINGING,,,,,,,anonymous,08765431,192.168.42.142,eprou-eibhoog_g204_EPR,XML,ipbxas_lo"
	eSplt = []string{"c2a2b753-8283-4347-94be-7df10a7710e3", "inbound", "2014-11-30 19:17:59", "1417371479", "sofia/loop_ipbxas/08765431@192.168.42.142", "CS_EXECUTE", "anonymous", "08765431", "192.168.42.142", "eprou-eibhoog_g204_EPR", "bridge",
		"{sip_contact_user=Sipean,bridge_early_media=true}[origination_caller_id_name=DLH,to_ep_type=SIP,sip_h_X-CalledEPType=SIP,sip_h_X-CalledEPTag=user0008,sip_h_X-ForwardedCall=false,max_forwards=50,origination_caller_id_number=anonymous,to_domain_tag=sampledomain.com,to_ep_tag=user0008,sip_h_X-CalledDomainTag=sampledomain.com,sip_h_X-Billable=false,sip_h_X-LoopApp=LOOP_ROUTED]sofia/ipbxas/user0008@sampledomain.com;fs_path=sip:127.0.0.1;transport=tcp",
		"[to_domain_tag=sampledomain.com,to_ep_type=SIP,to_ep_tag=user0010,sip_h_X-CalledDomainTag=sampledomain.com,sip_h_X-CalledEPTag=user0010,sip_h_X-LoopApp=LOOP_ROUTED,max_forwards=50,origination_caller_id_name=DLH,origination_caller_id_number=anonymous,sip_h_X-CalledEPType=SIP,sip_h_X-ForwardedCall=false,sip_h_X-Billable=false]sofia/ipbxas/user0010@sampledomain.com;fs_path=sip:127.0.0.1;transport=tcp",
		"[to_domain_tag=sampledomain.com,to_ep_type=SIP,to_ep_tag=user0009,sip_h_X-CalledDomainTag=sampledomain.com,sip_h_X-ForwardedCall=false,sip_h_X-LoopApp=LOOP_ROUTED,origination_caller_id_name=DLH,origination_caller_id_number=anonymous,sip_h_X-CalledEPType=SIP,sip_h_X-CalledEPTag=user0009,sip_h_X-Billable=false,max_forwards=50]sofia/ipbxas/user0009@sampledomain.com;fs_path=sip:127.0.0.1;transport=tcp",
		"[origination_caller_id_name=DLH,sip_h_X-CalledDomainTag=sampledomain.com,sip_h_X-ForwardedCall=false,sip_h_X-LoopApp=LOOP_ROUTED,origination_caller_id_number=anonymous,to_domain_tag=sampledomain.com,to_ep_type=SIP,to_ep_tag=user0001,sip_h_X-CalledEPType=SIP,sip_h_X-CalledEPTag=user0001,sip_h_X-Billable=false,max_forwards=50]sofia/ipbxas/user0001@sampledomain.com;fs_path=sip:127.0.0.1;transport=tcp",
		"[origination_caller_id_name=DLH,to_ep_type=SIP,to_ep_tag=user0011,sip_h_X-CalledEPTag=user0011,origination_caller_id_number=anonymous,to_domain_tag=sampledomain.com,sip_h_X-CalledDomainTag=sampledomain.com,sip_h_X-CalledEPType=SIP,sip_h_X-ForwardedCall=false,sip_h_X-Billable=false,sip_h_X-LoopApp=LOOP_ROUTED,max_forwards=50]sofia/ipbxas/user0011@sampledomain.com;fs_path=sip:127.0.0.1;transport=tcp",
		"[to_ep_type=SIP,sip_h_X-ForwardedCall=false,sip_h_X-Billable=false,sip_h_X-LoopApp=LOOP_ROUTED,origination_caller_id_name=DLH,origination_caller_id_number=anonymous,to_domain_tag=sampledomain.com,to_ep_tag=user0002,sip_h_X-CalledDomainTag=sampledomain.com,sip_h_X-CalledEPType=SIP,sip_h_X-CalledEPTag=user0002,max_forwards=50]sofia/ipbxas/user0002@sampledomain.com;fs_path=sip:127.0.0.1;transport=tcp",
		"[origination_caller_id_name=DLH,origination_caller_id_number=anonymous,to_domain_tag=sampledomain.com,sip_h_X-CalledEPType=SIP,sip_h_X-Billable=false,sip_h_X-LoopApp=LOOP_ROUTED,max_forwards=50,to_ep_type=SIP,to_ep_tag=user0003,sip_h_X-CalledDomainTag=sampledomain.com,sip_h_X-CalledEPTag=user0003,sip_h_X-ForwardedCall=false]sofia/ipbxas/user0003@sampledomain.com;fs_path=sip:127.0.0.1;transport=tcp", "[origination_caller_id_name=DLH,to_ep_type=SIP,sip_h_X-CalledDomainTag=sampledomain.com,sip_h_X-CalledEPTag=user0004,sip_h_X-ForwardedCall=false,max_forwards=50,origination_caller_id_number=anonymous,to_domain_tag=sampledomain.com,to_ep_tag=user0004,sip_h_X-CalledEPType=SIP,sip_h_X-Billable=false,sip_h_X-LoopApp=LOOP_ROUTED]sofia/ipbxas/user0004@sampledomain.com;fs_path=sip:127.0.0.1;transport=tcp", "[to_domain_tag=sampledomain.com,to_ep_type=SIP,sip_h_X-CalledDomainTag=sampledomain.com,sip_h_X-CalledEPType=SIP,sip_h_X-ForwardedCall=false,sip_h_X-Billable=false,sip_h_X-LoopApp=LOOP_ROUTED,max_forwards=50,origination_caller_id_name=DLH,origination_caller_id_number=anonymous,to_ep_tag=user0012,sip_h_X-CalledEPTag=user0012]sofia/ipbxas/user0012@sampledomain.com;fs_path=sip:127.0.0.1;transport=tcp",
		"XML", "ipbxas_lo", "PCMA", "8000", "64000", "PCMA", "8000", "64000", "", "nl-asd-gls-sbc01", "", "", "RINGING", "", "", "", "", "", "", "anonymous", "08765431", "192.168.42.142", "eprou-eibhoog_g204_EPR", "XML", "ipbxas_lo"}
	if splt := splitIgnoreGroups(strWithGroups, ",", 0); !reflect.DeepEqual(eSplt, splt) {
		t.Errorf("Expecting : %+v, received: %v", eSplt, splt)
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

func TestMapChanData(t *testing.T) {
	chanInfoStr := `uuid,direction,created,created_epoch,name,state,cid_name,cid_num,ip_addr,dest,application,application_data,dialplan,context,read_codec,read_rate,read_bit_rate,write_codec,write_rate,write_bit_rate,secure,hostname,presence_id,presence_data,callstate,callee_name,callee_num,callee_direction,call_uuid,sent_callee_name,sent_callee_num
fed464b3-a328-453f-9437-92b9b6a400fd,inbound,2014-10-26 18:08:32,1414343312,sofia/ipbxas/dan@172.16.254.66,CS_EXECUTE,dan,dan,172.16.254.66,+4986517174963,,,XML,ipbxas,PCMA,8000,64000,PCMA,8000,64000,,iPBXDev,dan@172.16.254.66,,HELD,,,,fed464b3-a328-453f-9437-92b9b6a400fd,,
c56125cc-024a-48a2-adbc-9612f6c02334,outbound,2014-10-26 18:08:32,1414343312,sofia/ipbxas/dan@172.16.254.66,CS_EXCHANGE_MEDIA,dan,+4986517174963,172.16.254.66,dan,playback,local_stream://moh,XML,ipbxas,PCMA,8000,64000,PCMA,8000,64000,,iPBXDev,dan@172.16.254.66,,ACTIVE,Outbound Call,dan,,fed464b3-a328-453f-9437-92b9b6a400fd,,
e604a792-172a-4e8f-8fc9-9198f0d15f15,inbound,2014-10-26 18:08:32,1414343312,sofia/loop_ipbxas/+4986517174963@172.16.254.66,CS_EXECUTE,dan,+4986517174963,127.0.0.1,dan,bridge,[sip_h_X-EpTransport=udp]sofia/ipbxas/dan@172.16.254.1:5060;registering_acc=172_16_254_66;fs_path=sip:172.16.254.66,XML,redirected,PCMA,8000,64000,PCMA,8000,64000,,iPBXDev,,,ACTIVE,Outbound Call,dan,SEND,e604a792-172a-4e8f-8fc9-9198f0d15f15,Outbound Call,dan
eacd0ae4-e1d5-447d-a7aa-e422a3a7abad,outbound,2014-10-26 18:08:32,1414343312,sofia/ipbxas/dan@172.16.254.1:5060,CS_EXCHANGE_MEDIA,dan,+4986517174963,127.0.0.1,dan,,,XML,redirected,PCMA,8000,64000,PCMA,8000,64000,,iPBXDev,,,ACTIVE,Outbound Call,dan,SEND,e604a792-172a-4e8f-8fc9-9198f0d15f15,dan,+4986517174963

4 total.
`
	eChanData := []map[string]string{
		{"application": "", "call_uuid": "fed464b3-a328-453f-9437-92b9b6a400fd", "direction": "inbound", "name": "sofia/ipbxas/dan@172.16.254.66", "application_data": "", "callstate": "HELD",
			"created": "2014-10-26 18:08:32", "cid_num": "dan", "dialplan": "XML", "read_bit_rate": "64000", "hostname": "iPBXDev", "callee_num": "", "created_epoch": "1414343312", "dest": "+4986517174963",
			"write_codec": "PCMA", "presence_data": "", "callee_direction": "", "sent_callee_num": "", "uuid": "fed464b3-a328-453f-9437-92b9b6a400fd", "state": "CS_EXECUTE", "ip_addr": "172.16.254.66",
			"cid_name": "dan", "write_rate": "8000", "write_bit_rate": "64000", "callee_name": "", "context": "ipbxas", "read_codec": "PCMA", "read_rate": "8000", "secure": "",
			"presence_id": "dan@172.16.254.66", "sent_callee_name": ""},
		{"application": "playback", "call_uuid": "fed464b3-a328-453f-9437-92b9b6a400fd", "direction": "outbound", "name": "sofia/ipbxas/dan@172.16.254.66", "application_data": "local_stream://moh",
			"callstate": "ACTIVE", "created": "2014-10-26 18:08:32", "cid_num": "+4986517174963", "dialplan": "XML", "read_bit_rate": "64000", "hostname": "iPBXDev", "callee_num": "dan", "created_epoch": "1414343312",
			"dest": "dan", "write_codec": "PCMA", "presence_data": "", "callee_direction": "", "sent_callee_num": "", "uuid": "c56125cc-024a-48a2-adbc-9612f6c02334", "state": "CS_EXCHANGE_MEDIA",
			"ip_addr": "172.16.254.66", "cid_name": "dan", "write_rate": "8000", "write_bit_rate": "64000", "callee_name": "Outbound Call", "context": "ipbxas", "read_codec": "PCMA", "read_rate": "8000",
			"secure": "", "presence_id": "dan@172.16.254.66", "sent_callee_name": ""},
		{"dialplan": "XML", "read_codec": "PCMA", "secure": "", "hostname": "iPBXDev", "callstate": "ACTIVE", "callee_num": "dan", "sent_callee_name": "Outbound Call", "created_epoch": "1414343312",
			"application": "bridge", "write_codec": "PCMA", "write_rate": "8000", "presence_data": "", "call_uuid": "e604a792-172a-4e8f-8fc9-9198f0d15f15", "uuid": "e604a792-172a-4e8f-8fc9-9198f0d15f15", "cid_num": "+4986517174963",
			"application_data": "[sip_h_X-EpTransport=udp]sofia/ipbxas/dan@172.16.254.1:5060;registering_acc=172_16_254_66;fs_path=sip:172.16.254.66", "created": "2014-10-26 18:08:32", "dest": "dan", "direction": "inbound",
			"state": "CS_EXECUTE", "ip_addr": "127.0.0.1", "cid_name": "dan", "write_bit_rate": "64000", "sent_callee_num": "dan", "name": "sofia/loop_ipbxas/+4986517174963@172.16.254.66", "context": "redirected",
			"read_rate": "8000", "read_bit_rate": "64000", "presence_id": "", "callee_name": "Outbound Call", "callee_direction": "SEND"},
		{"direction": "outbound", "state": "CS_EXCHANGE_MEDIA", "ip_addr": "127.0.0.1", "cid_name": "dan", "write_bit_rate": "64000", "sent_callee_num": "+4986517174963", "name": "sofia/ipbxas/dan@172.16.254.1:5060",
			"context": "redirected", "read_rate": "8000", "read_bit_rate": "64000", "presence_id": "", "callee_name": "Outbound Call", "callee_direction": "SEND", "dialplan": "XML", "read_codec": "PCMA", "secure": "",
			"hostname": "iPBXDev", "callstate": "ACTIVE", "callee_num": "dan", "sent_callee_name": "dan", "created_epoch": "1414343312", "application": "", "write_codec": "PCMA", "write_rate": "8000", "presence_data": "",
			"call_uuid": "e604a792-172a-4e8f-8fc9-9198f0d15f15", "uuid": "eacd0ae4-e1d5-447d-a7aa-e422a3a7abad", "cid_num": "+4986517174963", "application_data": "", "created": "2014-10-26 18:08:32", "dest": "dan"},
	}
	if rcvChanData := MapChanData(chanInfoStr, ","); !reflect.DeepEqual(eChanData, rcvChanData) {
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
		{"created_epoch": "1414402211", "application": "bridge", "write_codec": "PCMA", "write_rate": "8000", "presence_data": "", "call_uuid": "ba23506f-e36b-4c12-9c17-9146077bb240", "uuid": "ba23506f-e36b-4c12-9c17-9146077bb240", "cid_num": "dan",
			"application_data": "{sip_contact_user=iPBXSuite}[origination_caller_id_number=+4986517174963,to_domain_tag=172.16.254.66,sip_h_X-CalledEPType=SIP,sip_h_X-CalledEPTag=dan,sip_h_X-ForwardedCall=false,presence_id=dan@172.16.254.66,leg_progress_timeout=50,leg_timeout=100,to_ep_type=SIP,to_ep_tag=dan,sip_h_X-CalledDomainTag=172.16.254.66,sip_h_X-Billable=false,sip_h_X-LoopApp=LOOP_ROUTED]sofia/ipbxas/dan@172.16.254.66;fs_path=sip:127.0.0.1",
			"created":          "2014-10-27 10:30:11", "dest": "+4986517174963", "direction": "inbound", "state": "CS_EXECUTE", "ip_addr": "172.16.254.66", "cid_name": "dan", "write_bit_rate": "64000", "sent_callee_num": "", "name": "sofia/ipbxas/dan@172.16.254.66", "context": "ipbxas", "read_rate": "8000",
			"read_bit_rate": "64000", "presence_id": "dan@172.16.254.66", "callee_name": "", "callee_direction": "", "dialplan": "XML", "read_codec": "PCMA", "secure": "", "hostname": "iPBXDev", "callstate": "ACTIVE", "callee_num": "", "sent_callee_name": ""},
		{"state": "CS_EXCHANGE_MEDIA", "dialplan": "XML", "read_codec": "PCMA", "secure": "", "hostname": "iPBXDev", "callstate": "ACTIVE", "callee_num": "dan", "write_codec": "PCMA", "write_bit_rate": "64000",
			"call_uuid": "ba23506f-e36b-4c12-9c17-9146077bb240", "context": "ipbxas", "read_rate": "8000", "read_bit_rate": "64000", "presence_id": "dan@172.16.254.66", "created": "2014-10-27 10:30:11", "dest": "dan", "callee_name": "Outbound Call",
			"callee_direction": "", "direction": "outbound", "ip_addr": "172.16.254.66", "sent_callee_name": "", "created_epoch": "1414402211", "cid_name": "dan", "application": "", "write_rate": "8000", "presence_data": "", "sent_callee_num": "",
			"uuid": "d775e082-4309-4629-b08a-ae174271f2e1", "name": "sofia/ipbxas/dan@172.16.254.66", "cid_num": "+4986517174963", "application_data": ""},
		{"cid_name": "dan", "write_rate": "8000", "write_bit_rate": "64000", "callee_name": "Outbound Call", "context": "redirected", "read_codec": "PCMA", "read_rate": "8000", "secure": "", "presence_id": "", "sent_callee_name": "Outbound Call",
			"application": "playback", "call_uuid": "7c6a423e-7d2d-40c3-8f7f-06dc534d6576", "direction": "inbound", "name": "sofia/loop_ipbxas/+4986517174963@172.16.254.66", "application_data": "local_stream://moh", "callstate": "ACTIVE", "created": "2014-10-27 10:30:11",
			"cid_num": "+4986517174963", "dialplan": "XML", "read_bit_rate": "64000", "hostname": "iPBXDev", "callee_num": "dan", "created_epoch": "1414402211", "dest": "dan", "write_codec": "PCMA", "presence_data": "", "callee_direction": "SEND", "sent_callee_num": "dan",
			"uuid": "7c6a423e-7d2d-40c3-8f7f-06dc534d6576", "state": "CS_EXECUTE", "ip_addr": "127.0.0.1"},
		{"ip_addr": "127.0.0.1", "application_data": "", "cid_name": "dan", "cid_num": "+4986517174963", "read_codec": "PCMA", "read_bit_rate": "64000", "secure": "", "created_epoch": "1414402211", "presence_data": "", "callee_direction": "SEND",
			"call_uuid": "7c6a423e-7d2d-40c3-8f7f-06dc534d6576", "uuid": "81a05714-5a89-4a1c-848c-5e592527ae03", "direction": "outbound", "name": "sofia/ipbxas/dan@172.16.254.1:5060", "state": "CS_EXCHANGE_MEDIA", "callstate": "HELD", "created": "2014-10-27 10:30:11",
			"write_rate": "8000", "write_bit_rate": "64000", "callee_name": "Outbound Call", "dialplan": "XML", "context": "redirected", "read_rate": "8000", "hostname": "iPBXDev", "presence_id": "", "callee_num": "dan", "sent_callee_name": "dan", "dest": "dan",
			"application": "", "write_codec": "PCMA", "sent_callee_num": "+4986517174963"},
	}
	if rcvChanData := MapChanData(chanInfoStr, ","); !reflect.DeepEqual(eChanData, rcvChanData) {
		t.Errorf("Expected: %+v, received: %+v", eChanData, rcvChanData)
	}
}

func TestMapChanData3(t *testing.T) {
	chanInfoStr := `uuid,direction,created,created_epoch,name,state,cid_name,cid_num,ip_addr,dest,application,application_data,dialplan,context,read_codec,read_rate,read_bit_rate,write_codec,write_rate,write_bit_rate,secure,hostname,presence_id,presence_data,callstate,callee_name,callee_num,callee_direction,call_uuid,sent_callee_name,sent_callee_num,initial_cid_name,initial_cid_num,initial_ip_addr,initial_dest,initial_dialplan,initial_context
8009b347-fe46-4c99-9bb8-89e52e05d35f,inbound,2014-11-19 12:05:13,1416395113,sofia/ipbxas/+4986517174963@1.2.3.4,CS_EXECUTE,004986517174963,+4986517174963,2.3.4.5,0049850210795,bridge,{sip_contact_user=CloudIPBX.com,bridge_early_media=true}[to_domain_tag=sip.test.cloudipbx.com,sip_h_X-CalledEPType=SIP,sip_h_X-CalledEPTag=user3,sip_h_X-ForwardedCall=false,sip_h_X-LoopApp=LOOP_ROUTED,origination_caller_id_number=+4986517174963,to_ep_type=SIP,to_ep_tag=user3,sip_h_X-CalledDomainTag=sip.test.cloudipbx.com,sip_h_X-Billable=false,max_forwards=50]sofia/ipbxas/user3@sip.test.cloudipbx.com;fs_path=sip:127.0.0.1;transport=tcp,XML,ipbxas,PCMA,8000,64000,PCMA,8000,64000,,nl-asd-dev-sbc01,+4986517174963@1.2.3.4,,ACTIVE,,,,8009b347-fe46-4c99-9bb8-89e52e05d35f,,,004986517174963,+4986517174963,2.3.4.5,0049850210795,XML,ipbxas
91f198d3-3e4d-4885-b2f7-fd58865fa9a5,outbound,2014-11-19 12:05:13,1416395113,sofia/ipbxas/user3@sip.test.cloudipbx.com,CS_EXCHANGE_MEDIA,004986517174963,+4986517174963,2.3.4.5,user3,,,XML,ipbxas,PCMA,8000,64000,PCMA,8000,64000,,nl-asd-dev-sbc01,,,ACTIVE,Outbound Call,user3,,8009b347-fe46-4c99-9bb8-89e52e05d35f,,,004986517174963,+4986517174963,2.3.4.5,user3,XML,ipbxas
e657365d-c51b-4487-85f8-188c0771664e,inbound,2014-11-19 12:05:13,1416395113,sofia/loop_ipbxas/+4986517174963@2.3.4.5,CS_EXECUTE,004986517174963,+4986517174963,192.168.50.136,user3,bridge,[sip_h_X-EpTransport=tls]sofia/ipbxas/user3@10.10.10.142:40268;alias=87.139.12.167~40268~3;registering_acc=sip_test_deanconnect_nl;fs_path=sip:2.3.4.5;transport=tcp,XML,redirected,PCMA,8000,64000,PCMA,8000,64000,,nl-asd-dev-sbc01,,,ACTIVE,Outbound Call,user3,SEND,e657365d-c51b-4487-85f8-188c0771664e,Outbound Call,user3,004986517174963,+4986517174963,192.168.50.136,user3,XML,ipbxas_lo
2a7efd05-6f6f-400e-b319-4b8ff6a77a80,outbound,2014-11-19 12:05:13,1416395113,sofia/ipbxas/user3@10.10.10.142:40268,CS_EXCHANGE_MEDIA,004986517174963,+4986517174963,192.168.50.136,user3,,,XML,redirected,PCMA,8000,64000,PCMA,8000,64000,,nl-asd-dev-sbc01,,,ACTIVE,Outbound Call,user3,SEND,e657365d-c51b-4487-85f8-188c0771664e,004986517174963,+4986517174963,004986517174963,+4986517174963,192.168.50.136,user3,XML,redirected

4 total.
`
	eChanData := []map[string]string{
		{"uuid": "8009b347-fe46-4c99-9bb8-89e52e05d35f", "direction": "inbound", "created": "2014-11-19 12:05:13", "created_epoch": "1416395113", "name": "sofia/ipbxas/+4986517174963@1.2.3.4", "state": "CS_EXECUTE",
			"cid_name": "004986517174963", "cid_num": "+4986517174963", "ip_addr": "2.3.4.5", "dest": "0049850210795", "application": "bridge",
			"application_data": "{sip_contact_user=CloudIPBX.com,bridge_early_media=true}[to_domain_tag=sip.test.cloudipbx.com,sip_h_X-CalledEPType=SIP,sip_h_X-CalledEPTag=user3,sip_h_X-ForwardedCall=false,sip_h_X-LoopApp=LOOP_ROUTED,origination_caller_id_number=+4986517174963,to_ep_type=SIP,to_ep_tag=user3,sip_h_X-CalledDomainTag=sip.test.cloudipbx.com,sip_h_X-Billable=false,max_forwards=50]sofia/ipbxas/user3@sip.test.cloudipbx.com;fs_path=sip:127.0.0.1;transport=tcp",
			"dialplan":         "XML", "context": "ipbxas", "read_codec": "PCMA", "read_rate": "8000", "read_bit_rate": "64000", "write_codec": "PCMA", "write_rate": "8000", "write_bit_rate": "64000", "secure": "", "hostname": "nl-asd-dev-sbc01",
			"presence_id": "+4986517174963@1.2.3.4", "presence_data": "", "callstate": "ACTIVE", "callee_name": "", "callee_num": "", "callee_direction": "", "call_uuid": "8009b347-fe46-4c99-9bb8-89e52e05d35f", "sent_callee_name": "",
			"sent_callee_num": "", "initial_cid_name": "004986517174963", "initial_cid_num": "+4986517174963", "initial_ip_addr": "2.3.4.5", "initial_dest": "0049850210795", "initial_dialplan": "XML", "initial_context": "ipbxas"},
		{"direction": "outbound", "sent_callee_name": "", "application": "", "secure": "", "callstate": "ACTIVE", "call_uuid": "8009b347-fe46-4c99-9bb8-89e52e05d35f", "initial_dialplan": "XML", "name": "sofia/ipbxas/user3@sip.test.cloudipbx.com", "ip_addr": "2.3.4.5",
			"context": "ipbxas", "read_codec": "PCMA", "callee_num": "user3", "initial_cid_name": "004986517174963", "initial_dest": "user3", "uuid": "91f198d3-3e4d-4885-b2f7-fd58865fa9a5", "cid_num": "+4986517174963", "dest": "user3", "dialplan": "XML", "read_rate": "8000", "write_rate": "8000",
			"write_bit_rate": "64000", "presence_id": "", "created": "2014-11-19 12:05:13", "cid_name": "004986517174963", "presence_data": "", "callee_name": "Outbound Call", "initial_cid_num": "+4986517174963", "initial_context": "ipbxas", "state": "CS_EXCHANGE_MEDIA", "callee_direction": "",
			"created_epoch": "1416395113", "application_data": "", "read_bit_rate": "64000", "write_codec": "PCMA", "hostname": "nl-asd-dev-sbc01", "sent_callee_num": "", "initial_ip_addr": "2.3.4.5"},
		{"dialplan": "XML", "write_codec": "PCMA", "presence_id": "", "callee_direction": "SEND", "created": "2014-11-19 12:05:13", "read_rate": "8000", "secure": "", "sent_callee_name": "Outbound Call", "initial_ip_addr": "192.168.50.136", "name": "sofia/loop_ipbxas/+4986517174963@2.3.4.5",
			"context": "redirected", "write_bit_rate": "64000", "cid_name": "004986517174963", "presence_data": "", "callstate": "ACTIVE", "callee_name": "Outbound Call", "initial_dest": "user3", "direction": "inbound", "cid_num": "+4986517174963",
			"application_data": "[sip_h_X-EpTransport=tls]sofia/ipbxas/user3@10.10.10.142:40268;alias=87.139.12.167~40268~3;registering_acc=sip_test_deanconnect_nl;fs_path=sip:2.3.4.5;transport=tcp", "read_bit_rate": "64000", "hostname": "nl-asd-dev-sbc01", "callee_num": "user3", "call_uuid": "e657365d-c51b-4487-85f8-188c0771664e",
			"initial_context": "ipbxas_lo", "state": "CS_EXECUTE", "dest": "user3", "read_codec": "PCMA", "write_rate": "8000", "initial_cid_name": "004986517174963", "uuid": "e657365d-c51b-4487-85f8-188c0771664e", "created_epoch": "1416395113", "ip_addr": "192.168.50.136", "application": "bridge",
			"sent_callee_num": "user3", "initial_cid_num": "+4986517174963", "initial_dialplan": "XML"},
		{"created_epoch": "1416395113", "cid_num": "+4986517174963", "application": "", "read_bit_rate": "64000", "callee_num": "user3", "initial_ip_addr": "192.168.50.136", "initial_dest": "user3", "name": "sofia/ipbxas/user3@10.10.10.142:40268", "ip_addr": "192.168.50.136",
			"application_data": "", "write_codec": "PCMA", "write_bit_rate": "64000", "presence_data": "", "callstate": "ACTIVE", "call_uuid": "e657365d-c51b-4487-85f8-188c0771664e", "sent_callee_name": "004986517174963", "direction": "outbound", "cid_name": "004986517174963", "write_rate": "8000",
			"uuid": "2a7efd05-6f6f-400e-b319-4b8ff6a77a80", "context": "redirected", "callee_direction": "SEND", "state": "CS_EXCHANGE_MEDIA", "dest": "user3", "sent_callee_num": "+4986517174963", "created": "2014-11-19 12:05:13", "dialplan": "XML", "read_codec": "PCMA", "initial_cid_num": "+4986517174963",
			"read_rate": "8000", "secure": "", "hostname": "nl-asd-dev-sbc01", "presence_id": "", "initial_cid_name": "004986517174963", "callee_name": "Outbound Call", "initial_dialplan": "XML", "initial_context": "redirected"},
	}
	if rcvChanData := MapChanData(chanInfoStr, ","); !reflect.DeepEqual(eChanData, rcvChanData) {
		for _, mp := range eChanData {
			found := false
			for _, rcvMp := range rcvChanData {
				if reflect.DeepEqual(mp, rcvMp) {
					found = true
					break
				}
			}
			if !found {
				fmt.Printf("Not matching expected map: %+v\n", mp)
			}
		}
		t.Errorf("Expected: %+v, received: %+v", eChanData, rcvChanData)
	}
}

func TestMapChanData4(t *testing.T) {
	chanInfoStr := `uuid,direction,created,created_epoch,name,state,cid_name,cid_num,ip_addr,dest,application,application_data,dialplan,context,read_codec,read_rate,read_bit_rate,write_codec,write_rate,write_bit_rate,secure,hostname,presence_id,presence_data,accountcode,callstate,callee_name,callee_num,callee_direction,call_uuid,sent_callee_name,sent_callee_num,initial_cid_name,initial_cid_num,initial_ip_addr,initial_dest,initial_dialplan,initial_context
f66a1563-3d86-4a93-914d-3f9436f830d2,inbound,2018-06-29 04:37:18,1530261438,sofia/internal/1001@192.168.56.203,CS_EXECUTE,1001,1001,192.168.56.2,1002,playback,/home/teo/DesiJourney.wav,XML,default,G722,16000,64000,G722,16000,64000,,teo,1001@192.168.56.203,,1001,ACTIVE,,,,,,,1001,1001,192.168.56.2,1002,XML,default

1 total.
`
	eChanData := []map[string]string{
		{"uuid": "f66a1563-3d86-4a93-914d-3f9436f830d2",
			"direction": "inbound", "created": "2018-06-29 04:37:18",
			"created_epoch": "1530261438", "name": "sofia/internal/1001@192.168.56.203", "state": "CS_EXECUTE",
			"cid_name": "1001", "cid_num": "1001", "ip_addr": "192.168.56.2",
			"dest": "1002", "application": "playback",
			"application_data": "/home/teo/DesiJourney.wav",
			"dialplan":         "XML", "context": "default", "read_codec": "G722",
			"read_rate": "16000", "read_bit_rate": "64000",
			"write_codec": "G722", "write_rate": "16000",
			"write_bit_rate": "64000", "secure": "", "hostname": "teo",
			"presence_id": "1001@192.168.56.203", "presence_data": "",
			"accountcode": "1001",
			"callstate":   "ACTIVE", "callee_name": "", "callee_num": "",
			"callee_direction": "", "call_uuid": "",
			"sent_callee_name": "", "sent_callee_num": "", "initial_cid_name": "1001",
			"initial_cid_num": "1001", "initial_ip_addr": "192.168.56.2",
			"initial_dest": "1002", "initial_dialplan": "XML", "initial_context": "default"},
	}
	rcvChanData := MapChanData(chanInfoStr, ",")
	if !reflect.DeepEqual(eChanData, rcvChanData) {
		t.Errorf("Expected: %+v, received: %+v", eChanData, rcvChanData)
	}
}

func TestMapChanData5(t *testing.T) {
	chanInfoStr := `uuid,direction,created,created_epoch,name,state,cid_name,cid_num,ip_addr,dest,application,application_data,dialplan,context,read_codec,read_rate,read_bit_rate,write_codec,write_rate,write_bit_rate,secure,hostname,presence_id,presence_data,accountcode,callstate,callee_name,callee_num,callee_direction,call_uuid,sent_callee_name,sent_callee_num,initial_cid_name,initial_cid_num,initial_ip_addr,initial_dest,initial_dialplan,initial_context
f66a1563-3d86-4a93-914d-3f9436f830d2,inbound,2023-01-17 05:14:50,1673925290,sofia/internal/1001@192.168.56.203,CS_EXECUTE,1001,1001,192.168.56.2,1002,playback,1 1 1 5000 # media/mypbx_mainmenu.mp3 tone_stream://%(210,0,622.37,440)%(120,0,197,109.33) var_menu_dtmf [1,3],XML,default,G722,16000,64000,G722,16000,64000,,test,,,1001,ACTIVE,,,,,,,1001,1001,192.168.56.2,1002,XML,default

1 total. 
`
	eChanData := []map[string]string{
		{
			"uuid":             "f66a1563-3d86-4a93-914d-3f9436f830d2",
			"direction":        "inbound",
			"created":          "2023-01-17 05:14:50",
			"created_epoch":    "1673925290",
			"name":             "sofia/internal/1001@192.168.56.203",
			"state":            "CS_EXECUTE",
			"cid_name":         "1001",
			"cid_num":          "1001",
			"ip_addr":          "192.168.56.2",
			"dest":             "1002",
			"application":      "playback",
			"application_data": "1 1 1 5000 # media/mypbx_mainmenu.mp3 tone_stream://%(210,0,622.37,440)%(120,0,197,109.33) var_menu_dtmf [1,3]",
			"dialplan":         "XML",
			"context":          "default",
			"read_codec":       "G722",
			"read_rate":        "16000",
			"read_bit_rate":    "64000",
			"write_codec":      "G722",
			"write_rate":       "16000",
			"write_bit_rate":   "64000",
			"secure":           "",
			"hostname":         "test",
			"presence_id":      "",
			"presence_data":    "",
			"accountcode":      "1001",
			"callstate":        "ACTIVE",
			"callee_name":      "",
			"callee_num":       "",
			"callee_direction": "",
			"call_uuid":        "",
			"sent_callee_name": "",
			"sent_callee_num":  "",
			"initial_cid_name": "1001",
			"initial_cid_num":  "1001",
			"initial_ip_addr":  "192.168.56.2",
			"initial_dest":     "1002",
			"initial_dialplan": "XML",
			"initial_context":  "default"},
	}
	rcvChanData := MapChanData(chanInfoStr, ",")
	if !reflect.DeepEqual(eChanData, rcvChanData) {
		t.Errorf("expected: %+v,\nreceived: %+v", eChanData, rcvChanData)
	}
}

func TestEventToMap1(t *testing.T) {
	event := `Event-Name: BACKGROUND_JOB
Core-UUID: 32a090b2-7279-4d0f-b33d-1e42c87af186
FreeSWITCH-Hostname: debian
FreeSWITCH-Switchname: debian
FreeSWITCH-IPv4: 10.0.3.15
FreeSWITCH-IPv6: %3A%3A1
Event-Date-Local: 2019-01-11%2012%3A42%3A33
Event-Date-GMT: Fri,%2011%20Jan%202019%2010%3A42%3A33%20GMT
Event-Date-Timestamp: 1547203353175785
Event-Calling-File: mod_event_socket.c
Event-Calling-Function: api_exec
Event-Calling-Line-Number: 1557
Event-Sequence: 2177
Job-UUID: c2ee21ba-4cd0-413a-bcc3-7ac3f7888b61
Job-Command: status
Content-Length: 342

UP 0 years, 0 days, 3 hours, 34 minutes, 57 seconds, 300 milliseconds, 531 microseconds
FreeSWITCH (Version 1.8.2 -3-a98a958ac3 64bit) is ready
0 session(s) since startup
0 session(s) - peak 0, last 5min 0 
0 session(s) per Sec out of max 30, peak 0, last 5min 0 
1000 session(s) max
min idle cpu 0.00/99.50
Current Stack Size/Max 240K/8192K`
	expected := map[string]string{
		"Event-Name":                "BACKGROUND_JOB",
		"Core-UUID":                 "32a090b2-7279-4d0f-b33d-1e42c87af186",
		"FreeSWITCH-Hostname":       "debian",
		"FreeSWITCH-Switchname":     "debian",
		"FreeSWITCH-IPv4":           "10.0.3.15",
		"FreeSWITCH-IPv6":           "::1",
		"Event-Date-Local":          "2019-01-11 12:42:33",
		"Event-Date-GMT":            "Fri, 11 Jan 2019 10:42:33 GMT",
		"Event-Date-Timestamp":      "1547203353175785",
		"Event-Calling-File":        "mod_event_socket.c",
		"Event-Calling-Function":    "api_exec",
		"Event-Calling-Line-Number": "1557",
		"Event-Sequence":            "2177",
		"Job-UUID":                  "c2ee21ba-4cd0-413a-bcc3-7ac3f7888b61",
		"Job-Command":               "status",
		"Content-Length":            "342",
		EventBodyTag: `UP 0 years, 0 days, 3 hours, 34 minutes, 57 seconds, 300 milliseconds, 531 microseconds
FreeSWITCH (Version 1.8.2 -3-a98a958ac3 64bit) is ready
0 session(s) since startup
0 session(s) - peak 0, last 5min 0 
0 session(s) per Sec out of max 30, peak 0, last 5min 0 
1000 session(s) max
min idle cpu 0.00/99.50
Current Stack Size/Max 240K/8192K`}
	if rply := EventToMap(event); !reflect.DeepEqual(rply, expected) {
		t.Errorf("Expected: %s , recieved: %s", toJSON(expected), toJSON(rply))
	}
}

func TestEventToMap2(t *testing.T) {
	event := `Event-Name: BACKGROUND_JOB
Core-UUID: 32a090b2-7279-4d0f-b33d-1e42c87af186
FreeSWITCH-Hostname: debian
FreeSWITCH-Switchname: debian
FreeSWITCH-IPv4: 10.0.3.15
FreeSWITCH-IPv6: %3A%3A1
Event-Date-Local: 2019-01-11%2012%3A42%3A33
Event-Date-GMT: Fri,%2011%20Jan%202019%2010%3A42%3A33%20GMT
Event-Date-Timestamp: 1547203353175785
Event-Calling-File: mod_event_socket.c
Event-Calling-Function: api_exec
Event-Calling-Line-Number: 1557
Event-Sequence: 2177
Job-UUID: c2ee21ba-4cd0-413a-bcc3-7ac3f7888b61
Job-Command: status
Content-Length: 342
`
	expected := map[string]string{
		"Event-Name":                "BACKGROUND_JOB",
		"Core-UUID":                 "32a090b2-7279-4d0f-b33d-1e42c87af186",
		"FreeSWITCH-Hostname":       "debian",
		"FreeSWITCH-Switchname":     "debian",
		"FreeSWITCH-IPv4":           "10.0.3.15",
		"FreeSWITCH-IPv6":           "::1",
		"Event-Date-Local":          "2019-01-11 12:42:33",
		"Event-Date-GMT":            "Fri, 11 Jan 2019 10:42:33 GMT",
		"Event-Date-Timestamp":      "1547203353175785",
		"Event-Calling-File":        "mod_event_socket.c",
		"Event-Calling-Function":    "api_exec",
		"Event-Calling-Line-Number": "1557",
		"Event-Sequence":            "2177",
		"Job-UUID":                  "c2ee21ba-4cd0-413a-bcc3-7ac3f7888b61",
		"Job-Command":               "status",
		"Content-Length":            "342",
	}
	if rply := EventToMap(event); !reflect.DeepEqual(rply, expected) {
		t.Errorf("Expected: %s , recieved: %s", toJSON(expected), toJSON(rply))
	}
}

func TestGetMapKeys(t *testing.T) {
	fct := func(string, int) {}
	hMap := map[string][]func(string, int){
		"HEARTBEAT":                {fct},
		"RE_SCHEDULE":              {fct},
		"CHANNEL_STATE":            {fct},
		"CODEC":                    {fct},
		"CHANNEL_CREATE":           {fct},
		"CHANNEL_CALLSTATE":        {fct},
		"API":                      {fct},
		"CHANNEL_EXECUTE":          {fct},
		"CHANNEL_EXECUTE_COMPLETE": {fct},
		"CHANNEL_PARK":             {fct},
		"CHANNEL_HANGUP":           {fct},
		"CHANNEL_HANGUP_COMPLETE":  {fct},
		"CHANNEL_UNPARK":           {fct},
		"CHANNEL_DESTROY":          {fct},
	}
	expected := []string{"HEARTBEAT", "RE_SCHEDULE", "CHANNEL_STATE", "CODEC",
		"CHANNEL_CREATE", "CHANNEL_CALLSTATE", "API", "CHANNEL_EXECUTE",
		"CHANNEL_EXECUTE_COMPLETE", "CHANNEL_PARK", "CHANNEL_HANGUP",
		"CHANNEL_HANGUP_COMPLETE", "CHANNEL_UNPARK", "CHANNEL_DESTROY",
	}
	rply := getMapKeys(hMap)
	sort.Strings(expected)
	sort.Strings(rply)
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expected: %s , recieved: %s", toJSON(expected), toJSON(rply))
	}
}

func TestUtilsHeaderValNotFound(t *testing.T) {
	hdrs := "test: value"
	hdr := "fail"
	expected := ""
	received := headerVal(hdrs, hdr)
	if received != expected {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", expected, received)
	}
}

func TestUtilsToJSON(t *testing.T) {
	m := map[string]int{
		"testKey1": 1,
		"testKey2": 2,
	}
	expected := "{\"testKey1\":1,\"testKey2\":2}"
	received := toJSON(m)
	if expected != received {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", expected, received)
	}
}

func TestUtilsNopLoggerAlert(t *testing.T) {
	var l nopLogger
	err := l.Alert("alert")
	if err != nil {
		t.Errorf("\nExpected nil, received <%+v>", err)
	}
}

func TestUtilsNopLoggerClose(t *testing.T) {
	var l nopLogger
	err := l.Close()
	if err != nil {
		t.Errorf("\nExpected nil, received <%+v>", err)
	}
}

func TestUtilsNopLoggerCrit(t *testing.T) {
	var l nopLogger
	err := l.Crit("crit")
	if err != nil {
		t.Errorf("\nExpected nil, received <%+v>", err)
	}
}

func TestUtilsNopLoggerDebug(t *testing.T) {
	var l nopLogger
	err := l.Debug("debug")
	if err != nil {
		t.Errorf("\nExpected nil, received <%+v>", err)
	}
}

func TestUtilsNopLoggerEmerg(t *testing.T) {
	var l nopLogger
	err := l.Emerg("emerg")
	if err != nil {
		t.Errorf("\nExpected nil, received <%+v>", err)
	}
}

func TestUtilsNopLoggerErr(t *testing.T) {
	var l nopLogger
	err := l.Err("err")
	if err != nil {
		t.Errorf("\nExpected nil, received <%+v>", err)
	}
}

func TestUtilsNopLoggerInfo(t *testing.T) {
	var l nopLogger
	err := l.Info("info")
	if err != nil {
		t.Errorf("\nExpected nil, received <%+v>", err)
	}
}

func TestUtilsNopLoggerNotice(t *testing.T) {
	var l nopLogger
	err := l.Notice("notice")
	if err != nil {
		t.Errorf("\nExpected nil, received <%+v>", err)
	}
}

func TestUtilsNopLoggerWarning(t *testing.T) {
	var l nopLogger
	err := l.Warning("warning")
	if err != nil {
		t.Errorf("\nExpected nil, received <%+v>", err)
	}
}

func TestUtilsMapChanDataInsufficientStr(t *testing.T) {
	infoStr := "test1,\ntest2\ntest3\ntest4"

	expected := make([]map[string]string, 0)
	received := MapChanData(infoStr, ",")

	if !reflect.DeepEqual(expected, received) {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", expected, received)
	}
}

func TestUtilsMapChanDataContinue(t *testing.T) {
	infoStr := "test1,value\ntest2\ntest3\ntest4\ntest5\ntest6"

	expected := make([]map[string]string, 0)
	received := MapChanData(infoStr, ",")

	if !reflect.DeepEqual(expected, received) {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", expected, received)
	}
}

func TestUtilsgenUUID(t *testing.T) {
	uuid := genUUID()
	if len(uuid) == 0 {
		t.Fatalf("GenUUID error %s", uuid)
	}
	uuid2 := genUUID()
	if len(uuid2) == 0 {
		t.Fatalf("GenUUID error %s", uuid)
	}
	if uuid == uuid2 {
		t.Error("GenUUID error.")
	}
}

func TestUtilsSplitIgnoreGroups(t *testing.T) {
	testName := func(s string, i int) string {
		return fmt.Sprintf("case %d: %s", i, s)
	}
	testCases := []struct {
		desc     string
		params   []string
		expected []string
	}{
		{
			desc:     "EmptyInput",
			params:   []string{"", ","},
			expected: []string{},
		},
		{
			desc:     "NoGroups",
			params:   []string{"el1,el2,el3", ""},
			expected: []string{"el1,el2,el3"},
		},
		{
			desc:     "SingleElement",
			params:   []string{"el1", ","},
			expected: []string{"el1"},
		},
		{
			desc:     "LeadingAndTrailingSeparator",
			params:   []string{",el2,el3,el4,el5,el6,el7,el8,", ","},
			expected: []string{"", "el2", "el3", "el4", "el5", "el6", "el7", "el8", ""},
		},
		{
			desc:     "NoGroups",
			params:   []string{"el1,el2,el3,el4,el5,el6,el7,el8,el9", ","},
			expected: []string{"el1", "el2", "el3", "el4", "el5", "el6", "el7", "el8", "el9"},
		},
		{
			desc:     "SeparateGroups",
			params:   []string{"el1,[el2,el3,el4],el5,el6,{el7,el8},el9", ","},
			expected: []string{"el1", "[el2,el3,el4]", "el5", "el6", "{el7,el8}", "el9"},
		},
		{
			desc:     "SeparateGroups2",
			params:   []string{"el1,(el2,el3),el4,(el5,el6),{el7,el8},el9", ","},
			expected: []string{"el1", "(el2,el3)", "el4", "(el5,el6)", "{el7,el8}", "el9"},
		},
		{
			desc:     "NestedGroups1",
			params:   []string{"el1,el2,[el3,{el4,el5},el6],el7,el8,el9", ","},
			expected: []string{"el1", "el2", "[el3,{el4,el5},el6]", "el7", "el8", "el9"},
		},
		{
			desc:     "NestedGroups2",
			params:   []string{"el1,el2,el3,{el4,el5,[el6,el7]},el8,el9", ","},
			expected: []string{"el1", "el2", "el3", "{el4,el5,[el6,el7]}", "el8", "el9"},
		},
		{
			desc:     "NestedGroups3",
			params:   []string{"el1,[el2,[el3,el4]],el5,{{el6,el7},el8},el9", ","},
			expected: []string{"el1", "[el2,[el3,el4]]", "el5", "{{el6,el7},el8}", "el9"},
		},
		{
			desc:     "NestedGroups4",
			params:   []string{"el1,[el2,(el3,el4)],el5,({el6,el7},el8),el9", ","},
			expected: []string{"el1", "[el2,(el3,el4)]", "el5", "({el6,el7},el8)", "el9"},
		},
		{
			desc:     "EmptyInputPipeSep",
			params:   []string{"", "|"},
			expected: []string{},
		},
		{
			desc:     "SingleElementPipeSep",
			params:   []string{"el1", "|"},
			expected: []string{"el1"},
		},
		{
			desc:     "LeadingAndTrailingPipeSep",
			params:   []string{"|el2|el3|el4|el5|el6|el7|el8|", "|"},
			expected: []string{"", "el2", "el3", "el4", "el5", "el6", "el7", "el8", ""},
		},
		{
			desc:     "NoGroupsPipeSep",
			params:   []string{"el1|el2|el3|el4|el5|el6|el7|el8|el9", "|"},
			expected: []string{"el1", "el2", "el3", "el4", "el5", "el6", "el7", "el8", "el9"},
		},
		{
			desc:     "SeparateGroupsPipeSep",
			params:   []string{"el1|[el2|el3|el4]|el5|el6|{el7|el8}|el9", "|"},
			expected: []string{"el1", "[el2|el3|el4]", "el5", "el6", "{el7|el8}", "el9"},
		},
		{
			desc:     "SeparateGroupsPipeSep2",
			params:   []string{"el1|(el2|el3)|el4|(el5|el6)|{el7|el8}|el9", "|"},
			expected: []string{"el1", "(el2|el3)", "el4", "(el5|el6)", "{el7|el8}", "el9"},
		},
		{
			desc:     "NestedGroupsPipeSep1",
			params:   []string{"el1|el2|[el3|{el4|el5}|el6]|el7|el8|el9", "|"},
			expected: []string{"el1", "el2", "[el3|{el4|el5}|el6]", "el7", "el8", "el9"},
		},
		{
			desc:     "NestedGroupsPipeSep2",
			params:   []string{"el1|el2|el3|{el4|el5|[el6|el7]}|el8|el9", "|"},
			expected: []string{"el1", "el2", "el3", "{el4|el5|[el6|el7]}", "el8", "el9"},
		},
		{
			desc:     "NestedGroupsPipeSep3",
			params:   []string{"el1|[el2|[el3|el4]]|el5|{{el6|el7}|el8}|el9", "|"},
			expected: []string{"el1", "[el2|[el3|el4]]", "el5", "{{el6|el7}|el8}", "el9"},
		},
		{
			desc:     "NestedGroupsPipeSep4",
			params:   []string{"el1|[el2|(el3|el4)]|el5|({el6|el7}|el8)|el9", "|"},
			expected: []string{"el1", "[el2|(el3|el4)]", "el5", "({el6|el7}|el8)", "el9"},
		},
		{
			desc:     "EmptyInputTab",
			params:   []string{"", "\t"},
			expected: []string{},
		},
		{
			desc:     "SingleElementTab",
			params:   []string{"el1", "\t"},
			expected: []string{"el1"},
		},
		{
			desc:     "LeadingAndTrailingTab",
			params:   []string{"\tel2\tel3\tel4\tel5\tel6\tel7\tel8\t", "\t"},
			expected: []string{"", "el2", "el3", "el4", "el5", "el6", "el7", "el8", ""},
		},
		{
			desc:     "NoGroupsTab",
			params:   []string{"el1\tel2\tel3\tel4\tel5\tel6\tel7\tel8\tel9", "\t"},
			expected: []string{"el1", "el2", "el3", "el4", "el5", "el6", "el7", "el8", "el9"},
		},
		{
			desc:     "SeparateGroupsTab",
			params:   []string{"el1\t[el2\tel3\tel4]\tel5\tel6\t{el7\tel8}\tel9", "\t"},
			expected: []string{"el1", "[el2\tel3\tel4]", "el5", "el6", "{el7\tel8}", "el9"},
		},
		{
			desc:     "SeparateGroupsTab2",
			params:   []string{"el1\t(el2\tel3)\tel4\t(el5\tel6)\t{el7\tel8}\tel9", "\t"},
			expected: []string{"el1", "(el2\tel3)", "el4", "(el5\tel6)", "{el7\tel8}", "el9"},
		},
		{
			desc:     "NestedGroupsTab1",
			params:   []string{"el1\tel2\t[el3\t{el4\tel5}\tel6]\tel7\tel8\tel9", "\t"},
			expected: []string{"el1", "el2", "[el3\t{el4\tel5}\tel6]", "el7", "el8", "el9"},
		},
		{
			desc:     "NestedGroupsTab2",
			params:   []string{"el1\tel2\tel3\t{el4\tel5\t[el6\tel7]}\tel8\tel9", "\t"},
			expected: []string{"el1", "el2", "el3", "{el4\tel5\t[el6\tel7]}", "el8", "el9"},
		},
		{
			desc:     "NestedGroupsTab3",
			params:   []string{"el1\t[el2\t[el3\tel4]]\tel5\t{{el6\tel7}\tel8}\tel9", "\t"},
			expected: []string{"el1", "[el2\t[el3\tel4]]", "el5", "{{el6\tel7}\tel8}", "el9"},
		},
		{
			desc:     "NestedGroupsTab4",
			params:   []string{"el1\t[el2\t(el3\tel4)]\tel5\t({el6\tel7}\tel8)\tel9", "\t"},
			expected: []string{"el1", "[el2\t(el3\tel4)]", "el5", "({el6\tel7}\tel8)", "el9"},
		},
		{
			desc:     "EmptyInputMultipleCharSep",
			params:   []string{"", " sep "},
			expected: []string{},
		},
		{
			desc:     "SingleElementMultipleCharSep",
			params:   []string{"el1", " sep "},
			expected: []string{"el1"},
		},
		{
			desc:     "LeadingAndTrailingMultipleCharSep",
			params:   []string{" sep el2 sep el3 sep el4 sep el5 sep el6 sep el7 sep el8 sep ", " sep "},
			expected: []string{"", "el2", "el3", "el4", "el5", "el6", "el7", "el8", ""},
		},
		{
			desc:     "NoGroupsMultipleCharSep",
			params:   []string{"el1 sep el2 sep el3 sep el4 sep el5 sep el6 sep el7 sep el8 sep el9", " sep "},
			expected: []string{"el1", "el2", "el3", "el4", "el5", "el6", "el7", "el8", "el9"},
		},
		{
			desc:     "SeparateGroupsMultipleCharSep",
			params:   []string{"el1 sep [el2 sep el3 sep el4] sep el5 sep el6 sep {el7 sep el8} sep el9", " sep "},
			expected: []string{"el1", "[el2 sep el3 sep el4]", "el5", "el6", "{el7 sep el8}", "el9"},
		},
		{
			desc:     "SeparateGroupsMultipleCharSep2",
			params:   []string{"el1 sep (el2 sep el3) sep el4 sep (el5 sep el6) sep {el7 sep el8} sep el9", " sep "},
			expected: []string{"el1", "(el2 sep el3)", "el4", "(el5 sep el6)", "{el7 sep el8}", "el9"},
		},
		{
			desc:     "NestedGroupsMultipleCharSep1",
			params:   []string{"el1 sep el2 sep [el3 sep {el4 sep el5} sep el6] sep el7 sep el8 sep el9", " sep "},
			expected: []string{"el1", "el2", "[el3 sep {el4 sep el5} sep el6]", "el7", "el8", "el9"},
		},
		{
			desc:     "NestedGroupsMultipleCharSep2",
			params:   []string{"el1 sep el2 sep el3 sep {el4 sep el5 sep [el6 sep el7]} sep el8 sep el9", " sep "},
			expected: []string{"el1", "el2", "el3", "{el4 sep el5 sep [el6 sep el7]}", "el8", "el9"},
		},
		{
			desc:     "NestedGroupsMultipleCharSep3",
			params:   []string{"el1 sep [el2 sep [el3 sep el4]] sep el5 sep {{el6 sep el7} sep el8} sep el9", " sep "},
			expected: []string{"el1", "[el2 sep [el3 sep el4]]", "el5", "{{el6 sep el7} sep el8}", "el9"},
		},
		{
			desc:     "NestedGroupsMultipleCharSep4",
			params:   []string{"el1 sep [el2 sep (el3 sep el4)] sep el5 sep ({el6 sep el7} sep el8) sep el9", " sep "},
			expected: []string{"el1", "[el2 sep (el3 sep el4)]", "el5", "({el6 sep el7} sep el8)", "el9"},
		},
	}

	for testNr, testData := range testCases {
		t.Run(testName(testData.desc, testNr), func(t *testing.T) {
			if rcv := splitIgnoreGroups(testData.params[0], testData.params[1], 0); !reflect.DeepEqual(rcv, testData.expected) {
				t.Errorf("expected: <%+v>, \nreceived: <%+v>", testData.expected, rcv)
			}
		})
	}
}

/*********************** Benchmarks ************************/

func BenchmarkHeaderVal(b *testing.B) {
	for i := 0; i < b.N; i++ {
		headerVal(HEADER, "Content-Length")
		headerVal(BODY, "Event-Date-Loca")
	}
}

/*
$ go test -bench=.  -run=^$ -benchtime=1s -count=5 -benchmem
goos: linux
goarch: amd64
pkg: github.com/cgrates/fsock
cpu: Intel(R) Core(TM) i7-10510U CPU @ 1.80GHz
BenchmarkSplitIgnoreGroupsUnallocatedSlice-3   	  332491	      3188 ns/op	    1008 B/op	       6 allocs/op
BenchmarkSplitIgnoreGroupsUnallocatedSlice-3   	  330399	      3186 ns/op	    1008 B/op	       6 allocs/op
BenchmarkSplitIgnoreGroupsUnallocatedSlice-3   	  318609	      3172 ns/op	    1008 B/op	       6 allocs/op
BenchmarkSplitIgnoreGroupsUnallocatedSlice-3   	  340814	      3291 ns/op	    1008 B/op	       6 allocs/op
BenchmarkSplitIgnoreGroupsUnallocatedSlice-3   	  341724	      3167 ns/op	    1008 B/op	       6 allocs/op
BenchmarkSplitIgnoreGroupsPreallocatedSlice-3     	  445351	      2349 ns/op	     480 B/op	       1 allocs/op
BenchmarkSplitIgnoreGroupsPreallocatedSlice-3     	  439710	      2387 ns/op	     480 B/op	       1 allocs/op
BenchmarkSplitIgnoreGroupsPreallocatedSlice-3     	  443182	      2375 ns/op	     480 B/op	       1 allocs/op
BenchmarkSplitIgnoreGroupsPreallocatedSlice-3     	  508722	      2328 ns/op	     480 B/op	       1 allocs/op
BenchmarkSplitIgnoreGroupsPreallocatedSlice-3     	  510158	      2351 ns/op	     480 B/op	       1 allocs/op
*/
func BenchmarkSplitIgnoreGroupsUnallocatedSlice(b *testing.B) {
	input := "el1,el2,el3,el4,el5,el6,el7,el8,el9,el10,el11,el12,el13,el14,el15,el16,el17,el18,el19,el20,el21,el22,el23,el24,el25,el26,el27,el28,el29,el30"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = splitIgnoreGroups(input, ",", 0)
	}
}

func BenchmarkSplitIgnoreGroupsPreallocatedSlice(b *testing.B) {
	input := "el1,el2,el3,el4,el5,el6,el7,el8,el9,el10,el11,el12,el13,el14,el15,el16,el17,el18,el19,el20,el21,el22,el23,el24,el25,el26,el27,el28,el29,el30"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = splitIgnoreGroups(input, ",", 30)
	}
}
