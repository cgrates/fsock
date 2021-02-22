/*
utils.go is released under the MIT License <http://www.opensource.org/licenses/mit-license.php
Copyright (C) ITsysCOM. All Rights Reserved.

Provides FreeSWITCH socket communication.

*/
package fsock

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/url"
	"regexp"
	"sort"
	"strings"
)

const EventBodyTag = "EvBody"

type logger interface {
	Alert(string) error
	Close() error
	Crit(string) error
	Debug(string) error
	Emerg(string) error
	Err(string) error
	Info(string) error
	Notice(string) error
	Warning(string) error
}
type nopLogger struct{}

func (nopLogger) Alert(string) error   { return nil }
func (nopLogger) Close() error         { return nil }
func (nopLogger) Crit(string) error    { return nil }
func (nopLogger) Debug(string) error   { return nil }
func (nopLogger) Emerg(string) error   { return nil }
func (nopLogger) Err(string) error     { return nil }
func (nopLogger) Info(string) error    { return nil }
func (nopLogger) Notice(string) error  { return nil }
func (nopLogger) Warning(string) error { return nil }

// Convert fseventStr into fseventMap
func FSEventStrToMap(fsevstr string, headers []string) map[string]string {
	fsevent := make(map[string]string)
	filtered := (len(headers) != 0)
	for _, strLn := range strings.Split(fsevstr, "\n") {
		if hdrVal := strings.SplitN(strLn, ": ", 2); len(hdrVal) == 2 {
			if filtered && isSliceMember(headers, hdrVal[0]) {
				continue // Loop again since we only work on filtered fields
			}
			fsevent[hdrVal[0]] = urlDecode(strings.TrimSpace(strings.TrimRight(hdrVal[1], "\n")))
		}
	}
	return fsevent
}

// Converts string received from fsock into a list of channel info, each represented in a map
func MapChanData(chanInfoStr string) (chansInfoMap []map[string]string) {
	chansInfoMap = make([]map[string]string, 0)
	spltChanInfo := strings.Split(chanInfoStr, "\n")
	if len(spltChanInfo) <= 4 {
		return
	}
	hdrs := strings.Split(spltChanInfo[0], ",")
	for _, chanInfoLn := range spltChanInfo[1 : len(spltChanInfo)-3] {
		chanInfo := splitIgnoreGroups(chanInfoLn, ",")
		if len(hdrs) != len(chanInfo) {
			continue
		}
		chnMp := make(map[string]string)
		for iHdr, hdr := range hdrs {
			chnMp[hdr] = chanInfo[iHdr]
		}
		chansInfoMap = append(chansInfoMap, chnMp)
	}
	return
}

func EventToMap(event string) (result map[string]string) {
	result = make(map[string]string)
	body := false
	spltevent := strings.Split(event, "\n")
	for i := 0; i < len(spltevent); i++ {
		if len(spltevent[i]) == 0 {
			body = true
			continue
		}
		if body {
			result[EventBodyTag] = strings.Join(spltevent[i:], "\n")
			return
		}
		if val := strings.SplitN(spltevent[i], ": ", 2); len(val) == 2 {
			result[val[0]] = urlDecode(strings.TrimSpace(val[1]))
		}
	}
	return
}

// helper function for uuid generation
func genUUID() string {
	b := make([]byte, 16)
	_, err := io.ReadFull(rand.Reader, b)
	if err != nil {
		log.Fatal(err)
	}
	b[6] = (b[6] & 0x0F) | 0x40
	b[8] = (b[8] &^ 0x40) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[:4], b[4:6], b[6:8], b[8:10],
		b[10:])
}

func toJSON(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func indexStringAll(origStr, srchd string) []int {
	foundIdxs := make([]int, 0)
	lenSearched := len(srchd)
	startIdx := 0
	for {
		idxFound := strings.Index(origStr[startIdx:], srchd)
		if idxFound == -1 {
			break
		}
		idxFound += startIdx
		foundIdxs = append(foundIdxs, idxFound)
		startIdx = idxFound + lenSearched // Skip the characters found on next check
	}
	return foundIdxs
}

// Split considering {}[] which cancel separator
// In the end we merge groups which are having consecutive [] or {} in beginning since this is how FS builts them
func splitIgnoreGroups(origStr, sep string) []string {
	if len(origStr) == 0 {
		return []string{}
	} else if len(sep) == 0 {
		return []string{origStr}
	}
	retSplit := make([]string, 0)
	cmIdxs := indexStringAll(origStr, ",") // Main indexes of separators
	if len(cmIdxs) == 0 {
		return []string{origStr}
	}
	oCrlyIdxs := indexStringAll(origStr, "{") // Index  { for exceptions
	cCrlyIdxs := indexStringAll(origStr, "}") // Index  } for exceptions closing
	oBrktIdxs := indexStringAll(origStr, "[") // Index [ for exceptions
	cBrktIdxs := indexStringAll(origStr, "]") // Index ] for exceptions closing
	lastNonexcludedIdx := 0
	for i, cmdIdx := range cmIdxs {
		if len(oCrlyIdxs) == len(cCrlyIdxs) && len(oBrktIdxs) == len(cBrktIdxs) { // We assume exceptions and closing them are symetrical, otherwise don't handle exceptions
			exceptFound := false
			for iCrlyIdx := range oCrlyIdxs {
				if oCrlyIdxs[iCrlyIdx] < cmdIdx && cCrlyIdxs[iCrlyIdx] > cmdIdx { // Parentheses canceling indexing found
					exceptFound = true
					break
				}
			}
			for oBrktIdx := range oBrktIdxs {
				if oBrktIdxs[oBrktIdx] < cmdIdx && cBrktIdxs[oBrktIdx] > cmdIdx { // Parentheses canceling indexing found
					exceptFound = true
					break
				}
			}
			if exceptFound {
				continue
			}
		}
		switch i {
		case 0: // First one
			retSplit = append(retSplit, origStr[:cmIdxs[i]])
		case len(cmIdxs) - 1: // Last one
			postpendStr := ""
			if len(origStr) > cmIdxs[i]+1 { // Our separator is not the last character in the string
				postpendStr = origStr[cmIdxs[i]+1:]
			}
			retSplit = append(retSplit, origStr[cmIdxs[lastNonexcludedIdx]+1:cmIdxs[i]], postpendStr)
		default:
			retSplit = append(retSplit, origStr[cmIdxs[lastNonexcludedIdx]+1:cmIdxs[i]]) // Discard the separator from end string
		}
		lastNonexcludedIdx = i
	}
	groupedSplt := make([]string, 0)
	// Merge more consecutive groups (this is how FS displays app data from dial strings)
	for idx, spltData := range retSplit {
		if idx == 0 {
			groupedSplt = append(groupedSplt, spltData)
			continue // Nothing to do for first data
		}
		isGroup, _ := regexp.MatchString("{.*}|[.*]", spltData)
		if !isGroup {
			groupedSplt = append(groupedSplt, spltData)
			continue
		}
		isPrevGroup, _ := regexp.MatchString("{.*}|[.*]", retSplit[idx-1])
		if !isPrevGroup {
			groupedSplt = append(groupedSplt, spltData)
			continue
		}
		groupedSplt[len(groupedSplt)-1] = groupedSplt[len(groupedSplt)-1] + sep + spltData // Merge it with the previous data
	}
	return groupedSplt
}

// Extracts value of a header from anywhere in content string
func headerVal(hdrs, hdr string) string {
	var hdrSIdx, hdrEIdx int
	if hdrSIdx = strings.Index(hdrs, hdr); hdrSIdx == -1 {
		return ""
	} else if hdrEIdx = strings.Index(hdrs[hdrSIdx:], "\n"); hdrEIdx == -1 {
		hdrEIdx = len(hdrs[hdrSIdx:])
	}
	splt := strings.SplitN(hdrs[hdrSIdx:hdrSIdx+hdrEIdx], ": ", 2)
	if len(splt) != 2 {
		return ""
	}
	return strings.TrimSpace(strings.TrimRight(splt[1], "\n"))
}

// FS event header values are urlencoded. Use this to decode them. On error, use original value
func urlDecode(hdrVal string) string {
	if valUnescaped, errUnescaping := url.QueryUnescape(hdrVal); errUnescaping == nil {
		hdrVal = valUnescaped
	}
	return hdrVal
}

func getMapKeys(m map[string][]func(string, int)) (keys []string) {
	keys = make([]string, len(m))
	indx := 0
	for key := range m {
		keys[indx] = key
		indx++
	}
	return
}

// Binary string search in slice
func isSliceMember(ss []string, s string) bool {
	sort.Strings(ss)
	i := sort.SearchStrings(ss, s)
	return (i < len(ss) && ss[i] == s)
}

// successive Fibonacci numbers.
func fib() func() int {
	a, b := 0, 1
	return func() int {
		a, b = b, a+b
		return a
	}
}
