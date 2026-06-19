package server

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	faultHTTPError       = "http-error"
	faultTimeout         = "timeout"
	faultConnectionReset = "connection-reset"
	faultMalformedJSON   = "malformed-json"
	faultEmptyBody       = "empty-body"
)

func normalizeFault(name string) string {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "", "http-error", "error", "status":
		return faultHTTPError
	case "timeout":
		return faultTimeout
	case "connection-reset", "connection_reset", "reset":
		return faultConnectionReset
	case "malformed-json", "malformed_json", "bad-json", "bad_json":
		return faultMalformedJSON
	case "empty-body", "empty_body", "empty":
		return faultEmptyBody
	default:
		return strings.ToLower(strings.TrimSpace(name))
	}
}

func defaultStatus(status, fallback int) int {
	if status != 0 {
		return status
	}
	if fallback != 0 {
		return fallback
	}
	return http.StatusOK
}

func timeoutDuration(delayMS int) time.Duration {
	if delayMS > 0 {
		return time.Duration(delayMS) * time.Millisecond
	}
	return 30 * time.Second
}

func writeFault(c *gin.Context, fault string, status int, timeoutMS int) {
	c.Header("X-Specter-Fault", fault)
	switch fault {
	case faultTimeout:
		select {
		case <-time.After(timeoutDuration(timeoutMS)):
			c.Status(http.StatusGatewayTimeout)
		case <-c.Request.Context().Done():
		}
	case faultConnectionReset:
		hijacker, ok := c.Writer.(http.Hijacker)
		if !ok {
			c.Status(499)
			return
		}
		conn, _, err := hijacker.Hijack()
		if err != nil {
			c.Status(499)
			return
		}
		_ = conn.Close()
	case faultMalformedJSON:
		c.Data(defaultStatus(status, http.StatusOK), "application/json", []byte(`{"error":`))
	case faultEmptyBody:
		c.Status(defaultStatus(status, http.StatusOK))
	default:
		c.JSON(defaultStatus(status, http.StatusServiceUnavailable), gin.H{"error": "injected fault"})
	}
}
