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
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"text/template"
	"time"

	"github.com/Saku0512/specter/config"
	"gopkg.in/yaml.v3"
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

type StateStore struct {
	mu    sync.Mutex
	value string
}

func (s *StateStore) Get() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.value
}

func (s *StateStore) Set(v string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.value = v
}

type Server struct {
	engine  atomic.Pointer[gin.Engine]
	verbose bool
	history *RequestHistory
	state   *StateStore
}

func New(cfg *config.Config, verbose bool) *Server {
	s := &Server{verbose: verbose, history: &RequestHistory{}, state: &StateStore{}}
	s.engine.Store(newEngine(cfg, verbose, s.history, s.state))
	return s
}

func (s *Server) Reload(cfg *config.Config) {
	s.engine.Store(newEngine(cfg, s.verbose, s.history, s.state))
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.engine.Load().ServeHTTP(w, r)
}

type routeEntry struct {
	route   config.Route
	counter *atomic.Uint64
	limiter *rateLimiter
}

func newEngine(cfg *config.Config, verbose bool, history *RequestHistory, state *StateStore) *gin.Engine {
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
	r.GET("/__specter/requests/:index", func(c *gin.Context) {
		idx, err := strconv.Atoi(c.Param("index"))
		if err != nil || idx < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "index must be a non-negative integer"})
			return
		}
		entries := history.all()
		if idx >= len(entries) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": fmt.Sprintf("index %d out of range (%d recorded)", idx, len(entries)),
			})
			return
		}
		c.JSON(http.StatusOK, entries[idx])
	})
	r.POST("/__specter/requests/assert", func(c *gin.Context) {
		var req assertRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		matched := filterEntries(history.all(), req)
		wantAtLeastOne := req.Count == nil
		if wantAtLeastOne {
			if len(matched) >= 1 {
				c.JSON(http.StatusOK, gin.H{"ok": true, "matched": len(matched)})
			} else {
				c.JSON(http.StatusUnprocessableEntity, gin.H{
					"ok": false, "matched": 0, "error": "no matching requests found",
				})
			}
			return
		}
		if len(matched) == *req.Count {
			c.JSON(http.StatusOK, gin.H{"ok": true, "matched": len(matched)})
		} else {
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"ok":      false,
				"matched": len(matched),
				"error":   fmt.Sprintf("expected %d matching request(s), got %d", *req.Count, len(matched)),
			})
		}
	})
	r.GET("/__specter/state", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"state": state.Get()})
	})
	r.PUT("/__specter/state", func(c *gin.Context) {
		var body struct {
			State string `json:"state"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		state.Set(body.State)
		c.Status(http.StatusNoContent)
	})

	// Group routes by (method, path) to support multiple state-conditional
	// entries for the same endpoint.
	type routeKey struct{ method, path string }
	groups := map[routeKey][]*routeEntry{}
	var order []routeKey
	seen := map[routeKey]bool{}

	for _, route := range cfg.Routes {
		key := routeKey{route.Method, route.Path}
		e := &routeEntry{route: route, counter: &atomic.Uint64{}}
		if route.RateLimit > 0 {
			e.limiter = newRateLimiter(route.RateLimit, route.RateReset)
		}
		groups[key] = append(groups[key], e)
		if !seen[key] {
			seen[key] = true
			order = append(order, key)
		}
	}

	for _, key := range order {
		k := key
		entries := groups[k]

		r.Handle(k.method, k.path, func(c *gin.Context) {
			var bodyBytes []byte
			if c.Request.Body != nil {
				bodyBytes, _ = io.ReadAll(c.Request.Body)
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}
			tmplData := buildTemplateData(c, bodyBytes)
			currentState := state.Get()

			for _, e := range entries {
				rt := e.route

				// Skip entries whose state condition doesn't match
				if rt.State != "" && rt.State != currentState {
					continue
				}

				// Rate limit
				if e.limiter != nil {
					if ok, retryAfter := e.limiter.allow(); !ok {
						if retryAfter > 0 {
							c.Header("Retry-After", strconv.Itoa(int(retryAfter.Seconds())+1))
						}
						c.JSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
						return
					}
				}

				// Delay: random range takes precedence over fixed delay
				if rt.DelayMin > 0 || rt.DelayMax > 0 {
					d := rt.DelayMin
					if rt.DelayMax > rt.DelayMin {
						d = rt.DelayMin + rand.IntN(rt.DelayMax-rt.DelayMin+1)
					}
					time.Sleep(time.Duration(d) * time.Millisecond)
				} else if rt.Delay > 0 {
					time.Sleep(time.Duration(rt.Delay) * time.Millisecond)
				}
				// Fault injection
				if rt.ErrorRate > 0 && rand.Float64() < rt.ErrorRate {
					status := rt.ErrorStatus
					if status == 0 {
						status = http.StatusServiceUnavailable
					}
					c.JSON(status, gin.H{"error": "injected fault"})
					return
				}
				for hk, hv := range rt.Headers {
					c.Header(hk, hv)
				}

				// match conditions
				for _, m := range rt.Match {
					if matchesQuery(c, m.Query) && matchesBody(bodyBytes, m.Body) && matchesHeaders(c, m.Headers) {
						status := m.Status
						if status == 0 {
							status = http.StatusOK
						}
						body, fileCT := resolveBody(m.Response, m.File, tmplData, c.Params)
						ct := m.ContentType
						if ct == "" {
							ct = fileCT
						}
						if ct == "" {
							ct = rt.ContentType
						}
						respond(c, status, ct, body)
						if rt.SetState != nil {
							state.Set(*rt.SetState)
						}
						fireWebhook(rt.Webhook, tmplData, c.Params)
						return
					}
				}

				// multiple responses
				if len(rt.Responses) > 0 {
					var picked config.RouteResponse
					switch rt.Mode {
					case "random":
						picked = rt.Responses[rand.IntN(len(rt.Responses))]
					default:
						idx := e.counter.Add(1) - 1
						picked = rt.Responses[idx%uint64(len(rt.Responses))]
					}
					status := picked.Status
					if status == 0 {
						status = http.StatusOK
					}
					body, fileCT := resolveBody(picked.Response, picked.File, tmplData, c.Params)
					ct := picked.ContentType
					if ct == "" {
						ct = fileCT
					}
					if ct == "" {
						ct = rt.ContentType
					}
					respond(c, status, ct, body)
					if rt.SetState != nil {
						state.Set(*rt.SetState)
					}
					fireWebhook(rt.Webhook, tmplData, c.Params)
					return
				}

				// default response
				status := rt.Status
				if status == 0 {
					status = http.StatusOK
				}
				body, fileCT := resolveBody(rt.Response, rt.File, tmplData, c.Params)
				ct := rt.ContentType
				if ct == "" {
					ct = fileCT
				}
				respond(c, status, ct, body)
				if rt.SetState != nil {
					state.Set(*rt.SetState)
				}
				fireWebhook(rt.Webhook, tmplData, c.Params)
				return
			}

			// No entry matched the current state
			c.JSON(http.StatusConflict, gin.H{"error": "no route matches current state", "state": currentState})
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

func fireWebhook(wh *config.Webhook, tmplData map[string]any, params gin.Params) {
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
			if tmpl, err := template.New("").Funcs(templateFuncs).Parse(targetURL); err == nil {
				var buf bytes.Buffer
				if err := tmpl.Execute(&buf, tmplData); err == nil {
					targetURL = buf.String()
				}
			}
		}

		var bodyReader io.Reader
		if wh.Body != nil {
			processed := applyParams(applyTemplate(wh.Body, tmplData), params)
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

// loadFile reads a file and returns its parsed content and an inferred content type.
// .json files are parsed as JSON, .yaml/.yml as YAML (served as JSON-compatible data).
// All other files are returned as a raw string with an empty content type.
func loadFile(path string) (any, string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, "", err
	}
	switch strings.ToLower(filepath.Ext(path)) {
	case ".json":
		var v any
		if err := json.Unmarshal(data, &v); err != nil {
			return nil, "", fmt.Errorf("parse %s: %w", path, err)
		}
		return v, "application/json", nil
	case ".yaml", ".yml":
		var v any
		if err := yaml.Unmarshal(data, &v); err != nil {
			return nil, "", fmt.Errorf("parse %s: %w", path, err)
		}
		return v, "application/json", nil
	default:
		return string(data), "", nil
	}
}

// resolveBody returns the response body and an inferred content type.
// If file is non-empty it takes precedence over body.
func resolveBody(body any, file string, tmplData map[string]any, params gin.Params) (any, string) {
	if file != "" {
		data, ct, err := loadFile(file)
		if err != nil {
			log.Printf("file response: %v", err)
			return map[string]any{"error": "failed to load response file"}, ""
		}
		return applyParams(applyTemplate(data, tmplData), params), ct
	}
	return applyParams(applyTemplate(body, tmplData), params), ""
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

func matchesHeaders(c *gin.Context, headers map[string]string) bool {
	for k, v := range headers {
		if c.GetHeader(k) != v {
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
