package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Saku0512/specter/config"
	"github.com/gin-gonic/gin"
)

// handleSSE sends a Server-Sent Events stream to the client.
// Events are sent in order; each event may have an optional per-event delay (ms).
// When rt.StreamRepeat is true the event list cycles until the client disconnects.
func handleSSE(c *gin.Context, rt config.Route) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "streaming not supported"})
		return
	}

	ctx := c.Request.Context()

	sendAll := func() bool {
		for _, ev := range rt.Events {
			if ev.Delay > 0 {
				select {
				case <-time.After(time.Duration(ev.Delay) * time.Millisecond):
				case <-ctx.Done():
					return false
				}
			}
			select {
			case <-ctx.Done():
				return false
			default:
			}
			if ev.Event != "" {
				fmt.Fprintf(c.Writer, "event: %s\n", ev.Event)
			}
			if ev.ID != "" {
				fmt.Fprintf(c.Writer, "id: %s\n", ev.ID)
			}
			var dataStr string
			if s, ok2 := ev.Data.(string); ok2 {
				dataStr = s
			} else if ev.Data != nil {
				b, _ := json.Marshal(ev.Data)
				dataStr = string(b)
			}
			fmt.Fprintf(c.Writer, "data: %s\n\n", dataStr)
			flusher.Flush()
		}
		return true
	}

	if rt.StreamRepeat {
		for {
			if !sendAll() {
				return
			}
			select {
			case <-ctx.Done():
				return
			default:
			}
		}
	} else {
		sendAll()
	}
}
