package server

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"text/template"
	"time"

	"github.com/Saku0512/specter/config"
	"github.com/gin-gonic/gin"
)

// forwardRequest proxies the incoming request to targetBase and writes the response.
// targetBase is a URL like "http://api.example.com"; the original path and query are preserved.
func forwardRequest(c *gin.Context, targetBase string) {
	target, err := url.Parse(targetBase)
	if err != nil {
		log.Printf("proxy: invalid target URL %q: %v", targetBase, err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "invalid proxy target"})
		return
	}

	outURL := *c.Request.URL
	outURL.Scheme = target.Scheme
	outURL.Host = target.Host

	var bodyReader io.Reader
	if c.Request.Body != nil {
		bodyReader = c.Request.Body
	}
	outReq, err := http.NewRequestWithContext(c.Request.Context(), c.Request.Method, outURL.String(), bodyReader)
	if err != nil {
		log.Printf("proxy: failed to create request: %v", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "proxy request failed"})
		return
	}
	for k, vs := range c.Request.Header {
		for _, v := range vs {
			outReq.Header.Add(k, v)
		}
	}
	outReq.Host = target.Host

	log.Printf("proxy → %s %s", outReq.Method, outReq.URL)
	resp, err := http.DefaultClient.Do(outReq)
	if err != nil {
		log.Printf("proxy: %s %s failed: %v", outReq.Method, outReq.URL, err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "proxy request failed"})
		return
	}
	defer resp.Body.Close()

	for k, vs := range resp.Header {
		for _, v := range vs {
			c.Header(k, v)
		}
	}
	c.Status(resp.StatusCode)
	io.Copy(c.Writer, resp.Body)
}

func fireWebhook(wh *config.Webhook, tmplData map[string]any, params gin.Params, store *DataStore) {
	if wh == nil || wh.URL == "" {
		return
	}
	go func() {
		if wh.Delay > 0 {
			time.Sleep(time.Duration(wh.Delay) * time.Millisecond)
		}

		method := strings.ToUpper(wh.Method)
		if method == "" {
			method = http.MethodPost
		}

		// Apply template to URL
		targetURL := wh.URL
		if strings.Contains(targetURL, "{{") {
			if tmpl, err := template.New("").Funcs(buildFuncMap(store)).Parse(targetURL); err == nil {
				var buf bytes.Buffer
				if err := tmpl.Execute(&buf, tmplData); err == nil {
					targetURL = buf.String()
				}
			}
		}

		var bodyReader io.Reader
		if wh.Body != nil {
			processed := applyParams(applyTemplate(wh.Body, tmplData, store), params)
			var bodyBytes []byte
			if s, ok := processed.(string); ok {
				bodyBytes = []byte(s)
			} else {
				bodyBytes, _ = json.Marshal(processed)
			}
			bodyReader = bytes.NewBuffer(bodyBytes)
		}

		req, err := http.NewRequest(method, targetURL, bodyReader)
		if err != nil {
			log.Printf("webhook: failed to build request to %s: %v", targetURL, err)
			return
		}
		if wh.Body != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		for k, v := range wh.Headers {
			req.Header.Set(k, v)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Printf("webhook: %s %s failed: %v", method, targetURL, err)
			return
		}
		resp.Body.Close()
		log.Printf("webhook: %s %s → %d", method, targetURL, resp.StatusCode)
	}()
}

