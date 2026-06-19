package server

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

const redactedValue = "[REDACTED]"

var sensitiveNames = []string{
	"authorization",
	"cookie",
	"set-cookie",
	"proxy-authorization",
	"x-api-key",
	"x-auth-token",
	"x-access-token",
	"api-key",
	"apikey",
	"access-token",
	"id-token",
	"refresh-token",
	"password",
	"passwd",
	"secret",
	"token",
	"credential",
	"session",
	"jwt",
}

func isSensitiveName(name string) bool {
	normalized := strings.ToLower(strings.TrimSpace(name))
	normalized = strings.ReplaceAll(normalized, "_", "-")
	for _, sensitive := range sensitiveNames {
		if normalized == sensitive || strings.Contains(normalized, sensitive) {
			return true
		}
	}
	return false
}

func redactHeaderValue(name string, value string) string {
	if isSensitiveName(name) {
		return redactedValue
	}
	return value
}

func redactURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return redactedValue
	}

	u.User = nil

	query := u.Query()
	for key := range query {
		if isSensitiveName(key) {
			query.Set(key, redactedValue)
		}
	}
	u.RawQuery = query.Encode()

	return u.String()
}

func redactBodyForLog(body []byte) string {
	if len(body) == 0 {
		return ""
	}

	var value any
	if err := json.Unmarshal(body, &value); err != nil {
		if form, ok := redactFormBodyForLog(string(body)); ok {
			return form
		}
		return redactedValue
	}

	redacted := redactJSONValue(value)
	out, err := json.Marshal(redacted)
	if err != nil {
		return redactedValue
	}
	return string(out)
}

func redactFormBodyForLog(raw string) (string, bool) {
	if !strings.Contains(raw, "=") {
		return "", false
	}
	values, err := url.ParseQuery(raw)
	if err != nil || len(values) == 0 {
		return "", false
	}
	for key := range values {
		if isSensitiveName(key) {
			values.Set(key, redactedValue)
		}
	}
	return values.Encode(), true
}

func redactJSONValue(value any) any {
	switch v := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(v))
		for key, child := range v {
			if isSensitiveName(key) {
				out[key] = redactedValue
				continue
			}
			out[key] = redactJSONValue(child)
		}
		return out
	case []any:
		out := make([]any, len(v))
		for i, child := range v {
			out[i] = redactJSONValue(child)
		}
		return out
	default:
		return v
	}
}

func safeURLForLog(value any) string {
	return redactURL(fmt.Sprint(value))
}
