package server

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func historyMiddleware(h *RequestHistory) gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/__specter/") {
			c.Next()
			return
		}

		var bodyStr string
		if c.Request.Body != nil {
			b, _ := io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(b))
			bodyStr = string(b)
		}

		query := map[string]string{}
		for k, vs := range c.Request.URL.Query() {
			if len(vs) > 0 {
				query[k] = vs[0]
			}
		}

		headers := map[string]string{}
		for k, vs := range c.Request.Header {
			headers[k] = strings.Join(vs, ", ")
		}

		h.add(RequestEntry{
			Time:    time.Now().UTC(),
			Method:  c.Request.Method,
			Path:    c.Request.URL.Path,
			Query:   query,
			Headers: headers,
			Body:    bodyStr,
		})

		c.Next()
	}
}

func verboseLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Printf("→ %s %s", c.Request.Method, redactURL(c.Request.URL.RequestURI()))

		for k, v := range c.Request.Header {
			log.Printf("  %s: %s", k, redactHeaderValue(k, strings.Join(v, ", ")))
		}

		if c.Request.Body != nil && c.Request.ContentLength != 0 {
			body, err := io.ReadAll(c.Request.Body)
			if err == nil && len(body) > 0 {
				c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
				log.Printf("  Body: %s", redactBodyForLog(body))
			}
		}

		c.Next()
	}
}

func redactedGinLogger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		path := redactURL(param.Path)
		msg := fmt.Sprintf(
			"[GIN] %s | %3d | %13v | %15s | %-7s %q",
			param.TimeStamp.Format("2006/01/02 - 15:04:05"),
			param.StatusCode,
			param.Latency,
			param.ClientIP,
			param.Method,
			path,
		)
		if param.ErrorMessage != "" {
			msg += " | " + param.ErrorMessage
		}
		return msg + "\n"
	})
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
