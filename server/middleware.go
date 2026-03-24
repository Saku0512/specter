package server

import (
	"bytes"
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
		log.Printf("→ %s %s", c.Request.Method, c.Request.URL.RequestURI())

		for k, v := range c.Request.Header {
			log.Printf("  %s: %s", k, strings.Join(v, ", "))
		}

		if c.Request.Body != nil && c.Request.ContentLength != 0 {
			body, err := io.ReadAll(c.Request.Body)
			if err == nil && len(body) > 0 {
				c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
				log.Printf("  Body: %s", body)
			}
		}

		c.Next()
	}
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
