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
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"text/template"
	"time"

	"github.com/Saku0512/specter/config"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/gin-gonic/gin"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/legacy"
	"gopkg.in/yaml.v3"
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


type VarStore struct {
	mu   sync.Mutex
	vars map[string]string
}

func newVarStore() *VarStore { return &VarStore{vars: map[string]string{}} }

func (v *VarStore) Get(key string) string {
	v.mu.Lock()
	defer v.mu.Unlock()
	return v.vars[key]
}

func (v *VarStore) Set(key, val string) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.vars[key] = val
}

func (v *VarStore) Delete(key string) {
	v.mu.Lock()
	defer v.mu.Unlock()
	delete(v.vars, key)
}

func (v *VarStore) All() map[string]string {
	v.mu.Lock()
	defer v.mu.Unlock()
	out := make(map[string]string, len(v.vars))
	for k, val := range v.vars {
		out[k] = val
	}
	return out
}

func (v *VarStore) Clear() {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.vars = map[string]string{}
}


// DynamicRoute is a route added at runtime via the introspection API.
type DynamicRoute struct {
	ID    string       `json:"id"`
	Route config.Route `json:"route"`
}

type DynamicRouteStore struct {
	mu     sync.Mutex
	routes []DynamicRoute
}

func (d *DynamicRouteStore) Add(route config.Route) string {
	id := newID()
	d.mu.Lock()
	d.routes = append(d.routes, DynamicRoute{ID: id, Route: route})
	d.mu.Unlock()
	return id
}

func (d *DynamicRouteStore) Remove(id string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	for i, r := range d.routes {
		if r.ID == id {
			d.routes = append(d.routes[:i], d.routes[i+1:]...)
			return true
		}
	}
	return false
}

func (d *DynamicRouteStore) All() []DynamicRoute {
	d.mu.Lock()
	defer d.mu.Unlock()
	out := make([]DynamicRoute, len(d.routes))
	copy(out, d.routes)
	return out
}

func (d *DynamicRouteStore) Clear() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.routes = nil
}

func newID() string { return gofakeit.UUID() }

type Server struct {
	engine  atomic.Pointer[gin.Engine]
	verbose bool
	history *RequestHistory
	state   *StateStore
	vars    *VarStore
	cfg     atomic.Pointer[config.Config]
	dynamic *DynamicRouteStore
}

func New(cfg *config.Config, verbose bool) *Server {
	s := &Server{verbose: verbose, history: &RequestHistory{}, state: &StateStore{}, vars: newVarStore(), dynamic: &DynamicRouteStore{}}
	s.cfg.Store(cfg)
	s.engine.Store(newEngine(cfg, verbose, s.history, s.state, s.vars, s.dynamic, s.rebuild))
	return s
}

func (s *Server) rebuild() {
	cfg := s.cfg.Load()
	s.engine.Store(newEngine(cfg, s.verbose, s.history, s.state, s.vars, s.dynamic, s.rebuild))
}

func (s *Server) Reload(cfg *config.Config) {
	s.cfg.Store(cfg)
	s.rebuild()
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.engine.Load().ServeHTTP(w, r)
}

type routeEntry struct {
	route     config.Route
	counter   *atomic.Uint64 // sequential responses cycling index
	callCount *atomic.Uint64 // total matched calls (for on_call)
	limiter   *rateLimiter
}

func newEngine(cfg *config.Config, verbose bool, history *RequestHistory, state *StateStore, vars *VarStore, dynamic *DynamicRouteStore, rebuild func()) *gin.Engine {
	r := gin.Default()

	if cfg.CORS {
		r.Use(corsMiddleware())
	}
	if verbose {
		r.Use(verboseLogger())
	}
	r.Use(historyMiddleware(history))
	r.Use(openAPIMiddleware(cfg.OpenAPI))

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

	// Vars endpoints
	r.GET("/__specter/vars", func(c *gin.Context) {
		c.JSON(http.StatusOK, vars.All())
	})
	r.PUT("/__specter/vars", func(c *gin.Context) {
		var body map[string]string
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		for k, v := range body {
			vars.Set(k, v)
		}
		c.Status(http.StatusNoContent)
	})
	r.DELETE("/__specter/vars", func(c *gin.Context) {
		vars.Clear()
		c.Status(http.StatusNoContent)
	})
	r.GET("/__specter/vars/:key", func(c *gin.Context) {
		key := c.Param("key")
		val := vars.Get(key)
		c.JSON(http.StatusOK, gin.H{"key": key, "value": val})
	})
	r.PUT("/__specter/vars/:key", func(c *gin.Context) {
		var body struct {
			Value string `json:"value"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		vars.Set(c.Param("key"), body.Value)
		c.Status(http.StatusNoContent)
	})
	r.DELETE("/__specter/vars/:key", func(c *gin.Context) {
		vars.Delete(c.Param("key"))
		c.Status(http.StatusNoContent)
	})

	// Dynamic routes endpoints
	r.GET("/__specter/routes", func(c *gin.Context) {
		all := dynamic.All()
		type routeInfo struct {
			ID     string       `json:"id,omitempty"`
			Source string       `json:"source"`
			Route  config.Route `json:"route"`
		}
		var out []routeInfo
		for _, r := range cfg.Routes {
			out = append(out, routeInfo{Source: "config", Route: r})
		}
		for _, dr := range all {
			out = append(out, routeInfo{ID: dr.ID, Source: "dynamic", Route: dr.Route})
		}
		c.JSON(http.StatusOK, out)
	})
	r.POST("/__specter/routes", func(c *gin.Context) {
		var route config.Route
		if err := c.ShouldBindJSON(&route); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if route.Path == "" || route.Method == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "path and method are required"})
			return
		}
		id := dynamic.Add(route)
		go rebuild()
		c.JSON(http.StatusCreated, gin.H{"id": id})
	})
	r.DELETE("/__specter/routes", func(c *gin.Context) {
		dynamic.Clear()
		go rebuild()
		c.Status(http.StatusNoContent)
	})
	r.DELETE("/__specter/routes/:id", func(c *gin.Context) {
		if !dynamic.Remove(c.Param("id")) {
			c.JSON(http.StatusNotFound, gin.H{"error": "route not found"})
			return
		}
		go rebuild()
		c.Status(http.StatusNoContent)
	})

	// Group routes by (method, path) to support multiple state-conditional
	// entries for the same endpoint.
	type routeKey struct{ method, path string }
	groups := map[routeKey][]*routeEntry{}
	var order []routeKey
	seen := map[routeKey]bool{}

	allRoutes := cfg.Routes
	for _, dr := range dynamic.All() {
		allRoutes = append(allRoutes, dr.Route)
	}
	for _, route := range allRoutes {
		key := routeKey{route.Method, route.Path}
		e := &routeEntry{route: route, counter: &atomic.Uint64{}, callCount: &atomic.Uint64{}}
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

				// Skip entries whose state or vars conditions don't match
				if rt.State != "" && rt.State != currentState {
					continue
				}
				if !matchesVars(vars, rt.Vars) {
					continue
				}

				// Increment call counter (used for on_call matching)
				callN := e.callCount.Add(1)

				// Skip if on_call is set and this call number doesn't match
				if rt.OnCall > 0 && int(callN) != rt.OnCall {
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

				// Per-route proxy: forward and return, skip mock logic
				if rt.Proxy != "" {
					forwardRequest(c, rt.Proxy)
					return
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
					if matchesQuery(c, m.Query) && matchesBody(bodyBytes, m.Body) && matchesHeaders(c, m.Headers) && matchesBodyPath(bodyBytes, m.BodyPath) {
						status := m.Status
						if status == 0 {
							status = http.StatusOK
						}
						body, fileCT, scriptStatus := resolveBody(m.Response, m.File, m.Script, tmplData, c.Params)
						if scriptStatus != 0 {
							status = scriptStatus
						}
						ct := m.ContentType
						if ct == "" {
							ct = fileCT
						}
						if ct == "" {
							ct = rt.ContentType
						}
						respond(c, status, ct, body)
						// match-level set_state / set_vars take priority over route-level
						if m.SetState != nil {
							state.Set(*m.SetState)
						} else if rt.SetState != nil {
							state.Set(*rt.SetState)
						}
						for k, v := range rt.SetVars {
							vars.Set(k, v)
						}
						for k, v := range m.SetVars {
							vars.Set(k, v)
						}
						fireWebhook(rt.Webhook, tmplData, c.Params)
						return
					}
				}

				// multiple responses
				if len(rt.Responses) > 0 {
					var picked config.RouteResponse
					// Check for on_call-pinned entry first (callN already incremented above)
					var found bool
					for _, resp := range rt.Responses {
						if resp.OnCall > 0 && int(callN) == resp.OnCall {
							picked = resp
							found = true
							break
						}
					}
					if !found {
						// Fall back to sequential/random among entries without on_call
						var pool []config.RouteResponse
						for _, resp := range rt.Responses {
							if resp.OnCall == 0 {
								pool = append(pool, resp)
							}
						}
						if len(pool) == 0 {
							pool = rt.Responses
						}
						switch rt.Mode {
						case "random":
							picked = pool[rand.IntN(len(pool))]
						default:
							idx := e.counter.Add(1) - 1
							picked = pool[idx%uint64(len(pool))]
						}
					}
					status := picked.Status
					if status == 0 {
						status = http.StatusOK
					}
					body, fileCT, scriptStatus2 := resolveBody(picked.Response, picked.File, picked.Script, tmplData, c.Params)
					if scriptStatus2 != 0 {
						status = scriptStatus2
					}
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
					for k, v := range rt.SetVars {
						vars.Set(k, v)
					}
					fireWebhook(rt.Webhook, tmplData, c.Params)
					return
				}

				// default response
				status := rt.Status
				if status == 0 {
					status = http.StatusOK
				}
				body, fileCT, scriptStatus3 := resolveBody(rt.Response, rt.File, rt.Script, tmplData, c.Params)
				if scriptStatus3 != 0 {
					status = scriptStatus3
				}
				ct := rt.ContentType
				if ct == "" {
					ct = fileCT
				}
				respond(c, status, ct, body)
				if rt.SetState != nil {
					state.Set(*rt.SetState)
				}
				for k, v := range rt.SetVars {
					vars.Set(k, v)
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
	headers := map[string]any{}
	for k, vs := range c.Request.Header {
		headers[k] = strings.Join(vs, ", ")
	}
	return map[string]any{
		"params":  params,
		"query":   query,
		"body":    body,
		"headers": headers,
		"method":  c.Request.Method,
		"path":    c.Request.URL.Path,
	}
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
	"default": func(def, val string) string {
		if val == "" {
			return def
		}
		return val
	},
	"upper":  strings.ToUpper,
	"lower":  strings.ToLower,
	"trim":   strings.TrimSpace,
	"now":    func() string { return time.Now().UTC().Format(time.RFC3339) },
	"add":    func(a, b int) int { return a + b },
	"sub":    func(a, b int) int { return a - b },
}

// execScript renders a Go template script and returns the result plus an optional
// HTTP status override. If the JSON output contains a "_status" key (integer 100-599)
// it is extracted and returned as the status override; the key is removed from the body.
// If the output is valid JSON it is decoded; otherwise the raw string is returned.
func execScript(script string, data map[string]any) (body any, statusOverride int) {
	if script == "" {
		return nil, 0
	}
	tmpl, err := template.New("").Funcs(templateFuncs).Parse(script)
	if err != nil {
		log.Printf("script: parse error: %v", err)
		return nil, 0
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		log.Printf("script: execute error: %v", err)
		return nil, 0
	}
	out := strings.TrimSpace(buf.String())
	var v any
	if err := json.Unmarshal([]byte(out), &v); err == nil {
		if m, ok := v.(map[string]any); ok {
			if s, ok := m["_status"].(float64); ok && s >= 100 && s < 600 {
				delete(m, "_status")
				return m, int(s)
			}
		}
		return v, 0
	}
	return out, 0
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

// resolveBody returns the response body, an inferred content type, and an optional
// HTTP status override (non-zero only when script uses the _status envelope).
// Priority: script > file > body.
func resolveBody(body any, file, script string, tmplData map[string]any, params gin.Params) (any, string, int) {
	if script != "" {
		b, statusOverride := execScript(script, tmplData)
		return b, "", statusOverride
	}
	if file != "" {
		data, ct, err := loadFile(file)
		if err != nil {
			log.Printf("file response: %v", err)
			return map[string]any{"error": "failed to load response file"}, "", 0
		}
		return applyParams(applyTemplate(data, tmplData), params), ct, 0
	}
	return applyParams(applyTemplate(body, tmplData), params), "", 0
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
	for k, pattern := range headers {
		if !matchRegexOrExact(c.GetHeader(k), pattern) {
			return false
		}
	}
	return true
}

func matchesVars(vs *VarStore, expected map[string]string) bool {
	for k, v := range expected {
		if vs.Get(k) != v {
			return false
		}
	}
	return true
}

func matchesQuery(c *gin.Context, query map[string]string) bool {
	for k, pattern := range query {
		if !matchRegexOrExact(c.Query(k), pattern) {
			return false
		}
	}
	return true
}

// matchRegexOrExact treats pattern as a Go regular expression.
// If the pattern fails to compile it falls back to exact string comparison.
func matchRegexOrExact(actual, pattern string) bool {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return actual == pattern
	}
	return re.MatchString(actual)
}

// getJSONPath traverses a nested map using a dot-separated path.
// e.g. "user.role" returns the value at data["user"]["role"].
func getJSONPath(data map[string]any, path string) (string, bool) {
	idx := strings.Index(path, ".")
	if idx < 0 {
		v, ok := data[path]
		if !ok {
			return "", false
		}
		return fmt.Sprint(v), true
	}
	nested, ok := data[path[:idx]].(map[string]any)
	if !ok {
		return "", false
	}
	return getJSONPath(nested, path[idx+1:])
}

// matchesBodyPath checks dot-notation path → regex pattern conditions against the request body.
func matchesBodyPath(body []byte, paths map[string]string) bool {
	if len(paths) == 0 {
		return true
	}
	var parsed map[string]any
	if err := json.Unmarshal(body, &parsed); err != nil {
		return false
	}
	for path, pattern := range paths {
		actual, ok := getJSONPath(parsed, path)
		if !ok {
			return false
		}
		re, err := regexp.Compile(pattern)
		if err != nil || !re.MatchString(actual) {
			return false
		}
	}
	return true
}

// openAPIMiddleware returns a gin middleware that validates incoming requests
// against an OpenAPI spec. Validation errors are non-blocking: the mock response
// is always served, but an X-Specter-Validation-Error header and a log line are added.
func openAPIMiddleware(specPath string) gin.HandlerFunc {
	if specPath == "" {
		return func(c *gin.Context) { c.Next() }
	}

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromFile(specPath)
	if err != nil {
		log.Printf("openapi: failed to load spec %q: %v", specPath, err)
		return func(c *gin.Context) { c.Next() }
	}
	if err := doc.Validate(loader.Context); err != nil {
		log.Printf("openapi: spec %q is invalid: %v", specPath, err)
		return func(c *gin.Context) { c.Next() }
	}

	router, err := legacy.NewRouter(doc)
	if err != nil {
		log.Printf("openapi: failed to build router from %q: %v", specPath, err)
		return func(c *gin.Context) { c.Next() }
	}

	return func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/__specter/") {
			c.Next()
			return
		}

		// Read body for validation without consuming it
		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		route, pathParams, err := router.FindRoute(c.Request)
		if err != nil {
			// Route not in spec — not an error, spec may be partial
			c.Next()
			return
		}

		input := &openapi3filter.RequestValidationInput{
			Request:    c.Request,
			PathParams: pathParams,
			Route:      route,
			Options: &openapi3filter.Options{
				AuthenticationFunc: openapi3filter.NoopAuthenticationFunc,
			},
		}
		// Restore body after FindRoute may have consumed it
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		if err := openapi3filter.ValidateRequest(loader.Context, input); err != nil {
			msg := err.Error()
			c.Header("X-Specter-Validation-Error", msg)
			log.Printf("openapi validation: %s %s — %s", c.Request.Method, c.Request.URL.Path, msg)
		}

		// Always restore body for the route handler
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
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
