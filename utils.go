/*
utils.go is released under the MIT License <http://www.opensource.org/licenses/mit-license.php
Copyright (C) ITsysCOM. All Rights Reserved.

Provides FreeSWITCH socket communication.
*/
package fsock

import (
	"crypto/rand"
	"fmt"
	"io"
	"net/url"
	"slices"
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

// FSEventStrToMap transforms an FreeSWITCH event string into a map, optionally filtering headers.
func FSEventStrToMap(fsevstr string, headers []string) map[string]string {
	fsevent := make(map[string]string)
	filtered := (len(headers) != 0)
	for _, strLn := range strings.Split(fsevstr, "\n") {
		if hdrVal := strings.SplitN(strLn, ": ", 2); len(hdrVal) == 2 {
			if filtered && slices.Contains(headers, hdrVal[0]) {
				continue // Loop again since we only work on filtered fields
			}
			fsevent[hdrVal[0]] = urlDecode(strings.TrimSpace(strings.TrimRight(hdrVal[1], "\n")))
		}
	}
	return fsevent
}

// MapChanData parses channel information from a given string coming from fsock
// into a slice of maps, where each map contains individual channel data.
func MapChanData(chanInfoStr string, chanDelim string) (chansInfoMap []map[string]string) {
	chansInfoMap = make([]map[string]string, 0)
	spltChanInfo := strings.Split(chanInfoStr, "\n")
	if len(spltChanInfo) <= 4 {
		return
	}
	hdrs := strings.Split(spltChanInfo[0], chanDelim)
	for _, chanInfoLn := range spltChanInfo[1 : len(spltChanInfo)-3] {
		chanInfo := splitIgnoreGroups(chanInfoLn, chanDelim, len(hdrs))
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
	io.ReadFull(rand.Reader, b)
	b[6] = (b[6] & 0x0F) | 0x40
	b[8] = (b[8] &^ 0x40) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[:4], b[4:6], b[6:8], b[8:10],
		b[10:])
}

// splitIgnoreGroups splits a string by a separator, excluding elements
// enclosed in {}, [], or ().
func splitIgnoreGroups(s, sep string, expectedLength int) []string {
	if s == "" {
		return []string{}
	}
	if sep == "" {
		return []string{s}
	}
	sl := make([]string, 0, expectedLength)
	var idx, sqBrackets, crlBrackets, parantheses int
	sepLen := len(sep)
	for i := 0; i < len(s); {
		if strings.HasPrefix(s[i:], sep) && sqBrackets == 0 &&
			crlBrackets == 0 && parantheses == 0 {
			sl = append(sl, s[idx:i])
			idx = i + sepLen
			i = idx
			continue
		}
		switch s[i] {
		case '[':
			sqBrackets++
		case ']':
			if sqBrackets > 0 {
				sqBrackets--
			}
		case '{':
			crlBrackets++
		case '}':
			if crlBrackets > 0 {
				crlBrackets--
			}
		case '(':
			parantheses++
		case ')':
			if parantheses > 0 {
				parantheses--
			}
		}
		i++
	}
	sl = append(sl, s[idx:])
	return sl
}

// headerVal extracts a header's value from a content string.
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

// urlDecode decodes URL-encoded FS event header values, reverting to the original on error.
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
