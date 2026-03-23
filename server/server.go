package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand/v2"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"text/template"
	"time"

	"github.com/Saku0512/specter/config"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/gin-gonic/gin"
)

const maxHistory = 200

type RequestEntry struct {
	Time    time.Time         `json:"time"`
	Method  string            `json:"method"`
	Path    string            `json:"path"`
	Query   map[string]string `json:"query,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    string            `json:"body,omitempty"`
}

type RequestHistory struct {
	mu      sync.Mutex
	entries []RequestEntry
}

func (h *RequestHistory) add(e RequestEntry) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.entries = append(h.entries, e)
	if len(h.entries) > maxHistory {
		h.entries = h.entries[len(h.entries)-maxHistory:]
	}
}

func (h *RequestHistory) all() []RequestEntry {
	h.mu.Lock()
	defer h.mu.Unlock()
	out := make([]RequestEntry, len(h.entries))
	copy(out, h.entries)
	return out
}

func (h *RequestHistory) clear() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.entries = nil
}

type Server struct {
	engine  atomic.Pointer[gin.Engine]
	verbose bool
	history *RequestHistory
}

func New(cfg *config.Config, verbose bool) *Server {
	s := &Server{verbose: verbose, history: &RequestHistory{}}
	s.engine.Store(newEngine(cfg, verbose, s.history))
	return s
}

func (s *Server) Reload(cfg *config.Config) {
	s.engine.Store(newEngine(cfg, s.verbose, s.history))
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.engine.Load().ServeHTTP(w, r)
}

func newEngine(cfg *config.Config, verbose bool, history *RequestHistory) *gin.Engine {
	r := gin.Default()

	if cfg.CORS {
		r.Use(corsMiddleware())
	}

	if verbose {
		r.Use(verboseLogger())
	}

	r.Use(historyMiddleware(history))

	r.GET("/__specter/requests", func(c *gin.Context) {
		c.JSON(http.StatusOK, history.all())
	})
	r.DELETE("/__specter/requests", func(c *gin.Context) {
		history.clear()
		c.Status(http.StatusNoContent)
	})

	for _, route := range cfg.Routes {
		rt := route

		var counter atomic.Uint64

		var rl *rateLimiter
		if rt.RateLimit > 0 {
			rl = newRateLimiter(rt.RateLimit, rt.RateReset)
		}

		r.Handle(rt.Method, rt.Path, func(c *gin.Context) {
			if rl != nil {
				if ok, retryAfter := rl.allow(); !ok {
					if retryAfter > 0 {
						c.Header("Retry-After", strconv.Itoa(int(retryAfter.Seconds())+1))
					}
					c.JSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
					return
				}
			}

			if rt.Delay > 0 {
				time.Sleep(time.Duration(rt.Delay) * time.Millisecond)
			}
			for k, v := range rt.Headers {
				c.Header(k, v)
			}

			// Pre-read body for match conditions and response templates
			var bodyBytes []byte
			if c.Request.Body != nil {
				bodyBytes, _ = io.ReadAll(c.Request.Body)
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}

			tmplData := buildTemplateData(c, bodyBytes)

			for _, m := range rt.Match {
				if matchesQuery(c, m.Query) && matchesBody(bodyBytes, m.Body) {
					status := m.Status
					if status == 0 {
						status = http.StatusOK
					}
					ct := m.ContentType
					if ct == "" {
						ct = rt.ContentType
					}
					respond(c, status, ct, applyParams(applyTemplate(m.Response, tmplData), c.Params))
					return
				}
			}

			if len(rt.Responses) > 0 {
				var picked config.RouteResponse
				switch rt.Mode {
				case "random":
					picked = rt.Responses[rand.IntN(len(rt.Responses))]
				default: // sequential
					idx := counter.Add(1) - 1
					picked = rt.Responses[idx%uint64(len(rt.Responses))]
				}
				status := picked.Status
				if status == 0 {
					status = http.StatusOK
				}
				ct := picked.ContentType
				if ct == "" {
					ct = rt.ContentType
				}
				respond(c, status, ct, applyParams(applyTemplate(picked.Response, tmplData), c.Params))
				return
			}

			status := rt.Status
			if status == 0 {
				status = http.StatusOK
			}
			respond(c, status, rt.ContentType, applyParams(applyTemplate(rt.Response, tmplData), c.Params))
		})
	}

	if cfg.Proxy != "" {
		target, err := url.Parse(cfg.Proxy)
		if err != nil {
			log.Printf("invalid proxy URL %q: %v", cfg.Proxy, err)
		} else {
			proxy := httputil.NewSingleHostReverseProxy(target)
			r.NoRoute(func(c *gin.Context) {
				c.Request.Host = target.Host
				log.Printf("proxy → %s %s", c.Request.Method, c.Request.URL.RequestURI())
				proxy.ServeHTTP(c.Writer, c.Request)
			})
		}
	}

	return r
}

type rateLimiter struct {
	mu          sync.Mutex
	count       int
	limit       int
	reset       time.Duration
	windowStart time.Time
}

func newRateLimiter(limit, resetSecs int) *rateLimiter {
	return &rateLimiter{
		limit:       limit,
		reset:       time.Duration(resetSecs) * time.Second,
		windowStart: time.Now(),
	}
}

// allow returns true if the request is within the rate limit.
// retryAfter is non-zero when a reset window is configured.
func (rl *rateLimiter) allow() (ok bool, retryAfter time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	if rl.reset > 0 && time.Since(rl.windowStart) >= rl.reset {
		rl.count = 0
		rl.windowStart = time.Now()
	}
	rl.count++
	if rl.count > rl.limit {
		if rl.reset > 0 {
			retryAfter = max(rl.reset-time.Since(rl.windowStart), 0)
		}
		return false, retryAfter
	}
	return true, 0
}

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

// applyParams recursively replaces ":paramName" strings in the response with
// actual path parameter values. Numeric params are converted to int or float64.
func applyParams(v any, params gin.Params) any {
	switch val := v.(type) {
	case string:
		if strings.HasPrefix(val, ":") {
			if p := params.ByName(val[1:]); p != "" {
				if n, err := strconv.Atoi(p); err == nil {
					return n
				}
				if f, err := strconv.ParseFloat(p, 64); err == nil {
					return f
				}
				return p
			}
		}
		return val
	case map[string]any:
		out := make(map[string]any, len(val))
		for k, v2 := range val {
			out[k] = applyParams(v2, params)
		}
		return out
	case []any:
		out := make([]any, len(val))
		for i, v2 := range val {
			out[i] = applyParams(v2, params)
		}
		return out
	default:
		return v
	}
}

func buildTemplateData(c *gin.Context, bodyBytes []byte) map[string]any {
	params := map[string]any{}
	for _, p := range c.Params {
		params[p.Key] = p.Value
	}
	query := map[string]any{}
	for k, vs := range c.Request.URL.Query() {
		if len(vs) == 1 {
			query[k] = vs[0]
		} else {
			query[k] = vs
		}
	}
	body := map[string]any{}
	if len(bodyBytes) > 0 {
		_ = json.Unmarshal(bodyBytes, &body)
	}
	return map[string]any{"params": params, "query": query, "body": body}
}

var templateFuncs = template.FuncMap{
	"fake": func(kind string) string {
		switch kind {
		case "name":
			return gofakeit.Name()
		case "first_name":
			return gofakeit.FirstName()
		case "last_name":
			return gofakeit.LastName()
		case "email":
			return gofakeit.Email()
		case "uuid":
			return gofakeit.UUID()
		case "phone":
			return gofakeit.Phone()
		case "url":
			return gofakeit.URL()
		case "ip":
			return gofakeit.IPv4Address()
		case "username":
			return gofakeit.Username()
		case "password":
			return gofakeit.Password(true, true, true, false, false, 12)
		case "word":
			return gofakeit.Word()
		case "sentence":
			return gofakeit.Sentence(6)
		case "paragraph":
			return gofakeit.Paragraph(1, 3, 10, " ")
		case "color":
			return gofakeit.Color()
		case "country":
			return gofakeit.Country()
		case "city":
			return gofakeit.City()
		case "zip":
			return gofakeit.Zip()
		case "street":
			return gofakeit.Street()
		case "company":
			return gofakeit.Company()
		case "job":
			return gofakeit.JobTitle()
		case "int":
			return strconv.Itoa(gofakeit.IntRange(1, 10000))
		case "float":
			return strconv.FormatFloat(float64(gofakeit.Float32Range(0, 1000)), 'f', 2, 64)
		case "bool":
			return strconv.FormatBool(gofakeit.Bool())
		case "date":
			return gofakeit.Date().Format("2006-01-02")
		case "datetime":
			return gofakeit.Date().Format(time.RFC3339)
		default:
			return ""
		}
	},
}

func applyTemplate(v any, data map[string]any) any {
	switch val := v.(type) {
	case string:
		if !strings.Contains(val, "{{") {
			return val
		}
		tmpl, err := template.New("").Funcs(templateFuncs).Parse(val)
		if err != nil {
			return val
		}
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			return val
		}
		return buf.String()
	case map[string]any:
		out := make(map[string]any, len(val))
		for k, v2 := range val {
			out[k] = applyTemplate(v2, data)
		}
		return out
	case []any:
		out := make([]any, len(val))
		for i, v2 := range val {
			out[i] = applyTemplate(v2, data)
		}
		return out
	default:
		return v
	}
}

func respond(c *gin.Context, status int, contentType string, body any) {
	if contentType == "" || contentType == "application/json" {
		c.JSON(status, body)
		return
	}
	s, _ := body.(string)
	c.Data(status, contentType, []byte(s))
}

func matchesBody(body []byte, expected map[string]any) bool {
	if len(expected) == 0 {
		return true
	}
	var parsed map[string]any
	if err := json.Unmarshal(body, &parsed); err != nil {
		return false
	}
	for k, v := range expected {
		if fmt.Sprint(parsed[k]) != fmt.Sprint(v) {
			return false
		}
	}
	return true
}

func matchesQuery(c *gin.Context, query map[string]string) bool {
	for k, v := range query {
		if c.Query(k) != v {
			return false
		}
	}
	return true
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
