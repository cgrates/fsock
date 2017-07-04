# FreeSWITCH socket client written in [Go](http://cgrates.org/ "Go Website")

## Installation ##

`go get github.com/cgrates/fsock`

## Support ##
Join [CGRateS](http://www.cgrates.org/ "CGRateS Website") on Google Groups [here](https://groups.google.com/forum/#!forum/cgrates "CGRateS on GoogleGroups").

## License ##
fsock.go is released under the [MIT License](http://www.opensource.org/licenses/mit-license.php "MIT License").
Copyright (C) ITsysCOM GmbH. All Rights Reserved.

## Sample usage code ##
```
package main

import (
    "github.com/cgrates/fsock"
    "log/syslog"
    "fmt"
)

// Formats the event as map and prints it out
func printHeartbeat( eventStr, connId string) {
    // Format the event from string into Go's map type
    eventMap := fsock.FSEventStrToMap(eventStr, []string{})
    fmt.Printf("%v, connId: %s\n",eventMap, connId)
}

// Formats the event as map and prints it out
func printChannelAnswer( eventStr, connId string) {
    // Format the event from string into Go's map type
    eventMap := fsock.FSEventStrToMap(eventStr, []string{})
    fmt.Printf("%v, connId: %s\n",eventMap, connId)
}

// Formats the event as map and prints it out
func printChannelHangup( eventStr, connId string) {
    // Format the event from string into Go's map type
    eventMap := fsock.FSEventStrToMap(eventStr, []string{})
    fmt.Printf("%v, connId: %s\n",eventMap, connId)
}

func main() {
    // Init a syslog writter for our test
    l,errLog := syslog.New(syslog.LOG_INFO, "TestFSock")
    if errLog!=nil {
        l.Crit(fmt.Sprintf("Cannot connect to syslog:", errLog))
        return
    }

    // Filters
    evFilters := make(map[string][]string)
    evFilters["Event-Name"] = append(evFilters["Event-Name"], "CHANNEL_ANSWER")
    evFilters["Event-Name"] = append(evFilters["Event-Name"], "CHANNEL_HANGUP_COMPLETE")

    // We are interested in heartbeats, channel_answer, channel_hangup define handler for them
    evHandlers := map[string][]func(string, string){
	    "HEARTBEAT":               {printHeartbeat},
	    "CHANNEL_ANSWER":          {printChannelAnswer},
	    "CHANNEL_HANGUP_COMPLETE": {printChannelHangup},
    }

    fs, err := fsock.NewFSock("127.0.0.1:8021", "ClueCon", 10, evHandlers, evFilters, l, "wetsfnmretiewrtpj")
    if err != nil {
        l.Crit(fmt.Sprintf("FreeSWITCH error:", err))
        return
    }
    fs.ReadEvents()
}
```

[![Build Status](https://secure.travis-ci.org/cgrates/fsock.png)](http://travis-ci.org/cgrates/fsock)

