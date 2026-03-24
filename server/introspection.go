package server

import (
	"encoding/json"
	"fmt"
	"strings"
)

type assertRequest struct {
	Method string            `json:"method"`
	Path   string            `json:"path"`
	Query  map[string]string `json:"query"`
	Body   map[string]any    `json:"body"`
	Count  *int              `json:"count"` // nil = at least 1
}

func filterEntries(entries []RequestEntry, req assertRequest) []RequestEntry {
	var matched []RequestEntry
	for _, e := range entries {
		if req.Method != "" && !strings.EqualFold(e.Method, req.Method) {
			continue
		}
		if req.Path != "" && e.Path != req.Path {
			continue
		}
		if !assertQueryMatches(e.Query, req.Query) {
			continue
		}
		if !assertBodyMatches(e.Body, req.Body) {
			continue
		}
		matched = append(matched, e)
	}
	return matched
}

func assertQueryMatches(recorded, expected map[string]string) bool {
	for k, v := range expected {
		if recorded[k] != v {
			return false
		}
	}
	return true
}

func assertBodyMatches(recordedBody string, expected map[string]any) bool {
	if len(expected) == 0 {
		return true
	}
	var parsed map[string]any
	if err := json.Unmarshal([]byte(recordedBody), &parsed); err != nil {
		return false
	}
	for k, v := range expected {
		if fmt.Sprint(parsed[k]) != fmt.Sprint(v) {
			return false
		}
	}
	return true
}
