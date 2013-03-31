# FreeSWITCH socket client written in [Go](http://cgrates.org/ "Go Website")#

## Installation ##

`go get github.com/cgrates/fsock`

## Support ##
Join [CGRateS](http://www.cgrates.org/ "CGRateS Website") on Google Groups [here](https://groups.google.com/forum/#!forum/cgrates "CGRateS on GoogleGroups").

## License ##
fsock.go is released under the [MIT License](http://www.opensource.org/licenses/mit-license.php "MIT License").
Copyright (C) ITsysCOM. All Rights Reserved.

## Sample usage code ##
```
package main

import (
    "github.com/cgrates/fsock"
    "log/syslog"
    "fmt"
)

// Formats the event as map and prints it out
func printHeartbeat( eventStr string ) {
    // Format the event from string into Go's map type
    eventMap := fsock.FSEventStrToMap(eventStr, []string{})
    fmt.Printf("%v",eventMap)
}

func main() {
    // Init a syslog writter for our test
    l,errLog := syslog.New(syslog.LOG_INFO, "TestFSock")
    if errLog!=nil {
        l.Crit(fmt.Sprintf("Cannot connect to syslog:", errLog))
        return
    }
    // No filters
    evFilters := map[string]string{}
    // We are interested in heartbeats, define handler for them
    evHandlers := map[string][]func(string){"HEARTBEAT": []func(string){printHeartbeat}}
    fs, err := fsock.NewFSock("127.0.0.1:8021", "ClueCon", 10, evHandlers, evFilters, l)
    if err != nil {
        l.Crit(fmt.Sprintf("FreeSWITCH error:", err))
	return
    } else if !fs.Connected() {
	l.Warning("Cannot connect to FreeSWITCH")
        return
    }
    fs.ReadEvents()
}
```

Continous integration: [![Build Status](https://goci.herokuapp.com/project/image/github.com/cgrates/fsock "Continous integration")](http://goci.me/project/github.com/cgrates/fsock) [![Build Status](https://secure.travis-ci.org/cgrates/fsock.png)](http://travis-ci.org/cgrates/fsock)

