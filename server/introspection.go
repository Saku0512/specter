package server

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

type assertRequest struct {
	Request  string            `json:"request"`
	Method   string            `json:"method"`
	Path     string            `json:"path"`
	Query    map[string]string `json:"query"`
	Headers  map[string]string `json:"headers"`
	Body     map[string]any    `json:"body"`
	BodyPath map[string]string `json:"body_path"`
	BodyMode string            `json:"body_mode"` // subset (default) or exact
	Count    *int              `json:"count"`     // nil = at least 1
	Called   *int              `json:"called"`    // alias for count
}

func filterEntries(entries []RequestEntry, req assertRequest) []RequestEntry {
	req = normalizeAssertRequest(req)
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
		if !assertHeadersMatch(e.Headers, req.Headers) {
			continue
		}
		if !assertBodyMatches(e.Body, req.Body, req.BodyMode) {
			continue
		}
		if !assertBodyPathMatches(e.Body, req.BodyPath) {
			continue
		}
		matched = append(matched, e)
	}
	return matched
}

func normalizeAssertRequest(req assertRequest) assertRequest {
	if req.Count == nil && req.Called != nil {
		req.Count = req.Called
	}
	if req.Request != "" && (req.Method == "" || req.Path == "") {
		parts := strings.Fields(req.Request)
		if len(parts) >= 2 {
			if req.Method == "" {
				req.Method = parts[0]
			}
			if req.Path == "" {
				req.Path = parts[1]
			}
		}
	}
	req.BodyMode = strings.ToLower(strings.TrimSpace(req.BodyMode))
	return req
}

func assertQueryMatches(recorded, expected map[string]string) bool {
	for k, v := range expected {
		if recorded[k] != v {
			return false
		}
	}
	return true
}

func assertHeadersMatch(recorded, expected map[string]string) bool {
	for k, v := range expected {
		if recorded[k] != v {
			return false
		}
	}
	return true
}

func assertBodyMatches(recordedBody string, expected map[string]any, mode string) bool {
	if len(expected) == 0 {
		return true
	}
	var parsed map[string]any
	if err := json.Unmarshal([]byte(recordedBody), &parsed); err != nil {
		return false
	}
	if mode == "exact" {
		return reflect.DeepEqual(parsed, expected)
	}
	for k, v := range expected {
		if fmt.Sprint(parsed[k]) != fmt.Sprint(v) {
			return false
		}
	}
	return true
}

func assertBodyPathMatches(recordedBody string, expected map[string]string) bool {
	if len(expected) == 0 {
		return true
	}
	return matchesBodyPath([]byte(recordedBody), expected)
}
