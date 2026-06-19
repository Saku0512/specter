package server

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"math/rand/v2"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sort"
	"strconv"
	"sync/atomic"
	"time"

	validate_cmd "github.com/Saku0512/specter/cmd/validate"
	"github.com/Saku0512/specter/config"
	"github.com/gin-gonic/gin"
)

type routeEntry struct {
	route     config.Route
	key       string
	source    string
	counter   *atomic.Uint64 // sequential responses cycling index
	callCount *atomic.Uint64 // total matched calls (for on_call)
	limiter   *rateLimiter
}

func timelineDefinitions(cfg *config.Config, dynamic *DynamicRouteStore) []TimelineProgress {
	var out []TimelineProgress
	for i, route := range cfg.Routes {
		if len(route.Timeline) == 0 {
			continue
		}
		out = append(out, TimelineProgress{
			Key:    fmt.Sprintf("config-%d", i),
			Method: route.Method,
			Path:   route.Path,
			Source: "config",
			Total:  len(route.Timeline),
		})
	}
	for _, dr := range dynamic.All() {
		if len(dr.Route.Timeline) == 0 {
			continue
		}
		out = append(out, TimelineProgress{
			Key:    dr.ID,
			Method: dr.Route.Method,
			Path:   dr.Route.Path,
			Source: "dynamic",
			Total:  len(dr.Route.Timeline),
		})
	}
	return out
}

func newEngine(cfg *config.Config, verbose bool, random bool, history *RequestHistory, state *StateStore, vars *VarStore, scenario *ScenarioStore, dynamic *DynamicRouteStore, store *DataStore, timeline *TimelineStore, rebuild func()) *gin.Engine {
	r := gin.New()
	r.Use(redactedGinLogger(), gin.Recovery())

	if cfg.CORS {
		r.Use(corsMiddleware())
	}
	if verbose {
		r.Use(verboseLogger())
	}
	r.Use(historyMiddleware(history))
	_, oaRouter := buildOpenAPIRouter(cfg.OpenAPI)
	if oaRouter != nil {
		r.Use(openAPIRequestMiddleware(oaRouter, cfg.OpenAPIStrict))
	}

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
		req = normalizeAssertRequest(req)
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
	r.POST("/__specter/config/validate", func(c *gin.Context) {
		var body struct {
			YAML string `json:"yaml"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		preview, err := config.LoadBytes([]byte(body.YAML))
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"valid":     false,
				"errors":    []string{err.Error()},
				"routes":    []any{},
				"scenarios": []string{},
				"stores":    []string{},
			})
			return
		}
		errs := validate_cmd.CheckNoFilesystem(preview)
		if len(preview.Include) > 0 {
			errs = append(errs, "include is not supported in the config playground")
		}
		type routeSummary struct {
			Method string `json:"method"`
			Path   string `json:"path"`
			Status int    `json:"status,omitempty"`
			State  string `json:"state,omitempty"`
			Source string `json:"source"`
		}
		routes := make([]routeSummary, 0, len(preview.Routes))
		for _, route := range preview.Routes {
			routes = append(routes, routeSummary{
				Method: route.Method,
				Path:   route.Path,
				Status: route.Status,
				State:  route.State,
				Source: "config",
			})
		}
		scenarioNames := make([]string, 0, len(preview.Scenarios))
		storeNames := map[string]bool{}
		for name := range preview.Stores {
			storeNames[name] = true
		}
		for name, preset := range preview.Scenarios {
			scenarioNames = append(scenarioNames, name)
			for storeName := range preset.Stores {
				storeNames[storeName] = true
			}
		}
		sort.Strings(scenarioNames)
		stores := make([]string, 0, len(storeNames))
		for name := range storeNames {
			stores = append(stores, name)
		}
		sort.Strings(stores)
		c.JSON(http.StatusOK, gin.H{
			"valid":     len(errs) == 0,
			"errors":    errs,
			"routes":    routes,
			"scenarios": scenarioNames,
			"stores":    stores,
		})
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

	// Scenario preset endpoints
	r.GET("/__specter/scenarios", func(c *gin.Context) {
		names := make([]string, 0, len(cfg.Scenarios))
		for name := range cfg.Scenarios {
			names = append(names, name)
		}
		sort.Strings(names)
		c.JSON(http.StatusOK, gin.H{"active": scenario.Get(), "scenarios": names})
	})
	r.GET("/__specter/scenario", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"active": scenario.Get()})
	})
	r.POST("/__specter/scenarios/:name", func(c *gin.Context) {
		name := c.Param("name")
		preset, ok := cfg.Scenarios[name]
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": "scenario not found"})
			return
		}
		applyScenarioPreset(name, preset, state, vars, scenario, store)
		if storePersistFailed(c, store) {
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true, "active": name})
	})
	r.GET("/__specter/timelines", func(c *gin.Context) {
		c.JSON(http.StatusOK, timeline.Snapshot(timelineDefinitions(cfg, dynamic)))
	})
	r.POST("/__specter/timelines/:key/reset", func(c *gin.Context) {
		key := c.Param("key")
		found := false
		for _, def := range timelineDefinitions(cfg, dynamic) {
			if def.Key == key {
				found = true
				break
			}
		}
		if !found {
			c.JSON(http.StatusNotFound, gin.H{"error": "timeline not found"})
			return
		}
		timeline.Reset(key)
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	// Reset endpoint clears runtime state in one call.
	r.POST("/__specter/reset", func(c *gin.Context) {
		var body struct {
			Targets []string `json:"targets"` // optional: ["state","vars","history","stores","scenario","timelines"] — defaults to all
		}
		_ = c.ShouldBindJSON(&body)
		all := len(body.Targets) == 0
		reset := func(t string) bool {
			if all {
				return true
			}
			for _, v := range body.Targets {
				if v == t {
					return true
				}
			}
			return false
		}
		if reset("state") {
			state.Set("")
		}
		if reset("vars") {
			vars.Clear()
		}
		if reset("history") {
			history.clear()
		}
		if reset("stores") {
			store.ResetAll()
			if storePersistFailed(c, store) {
				return
			}
		}
		if reset("scenario") {
			scenario.Clear()
		}
		if reset("timelines") {
			timeline.Clear()
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
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
	r.PUT("/__specter/routes/:id", func(c *gin.Context) {
		var route config.Route
		if err := c.ShouldBindJSON(&route); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if route.Path == "" || route.Method == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "path and method are required"})
			return
		}
		if !dynamic.Update(c.Param("id"), route) {
			c.JSON(http.StatusNotFound, gin.H{"error": "route not found"})
			return
		}
		go rebuild()
		c.Status(http.StatusNoContent)
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

	// Store introspection endpoints
	r.GET("/__specter/stores", func(c *gin.Context) {
		names := store.Names()
		type storeInfo struct {
			Name  string `json:"name"`
			Count int    `json:"count"`
		}
		out := make([]storeInfo, 0, len(names))
		for _, name := range names {
			out = append(out, storeInfo{Name: name, Count: len(store.List(name))})
		}
		c.JSON(http.StatusOK, out)
	})
	r.GET("/__specter/stores/:name", func(c *gin.Context) {
		c.JSON(http.StatusOK, store.List(c.Param("name")))
	})
	r.PUT("/__specter/stores/:name", func(c *gin.Context) {
		var items []map[string]any
		if err := c.ShouldBindJSON(&items); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		store.SetCollection(c.Param("name"), items)
		if storePersistFailed(c, store) {
			return
		}
		c.Status(http.StatusNoContent)
	})
	r.DELETE("/__specter/stores/:name", func(c *gin.Context) {
		store.Clear(c.Param("name"))
		if storePersistFailed(c, store) {
			return
		}
		c.Status(http.StatusNoContent)
	})

	// Group routes by (method, path) to support multiple state-conditional
	// entries for the same endpoint.
	type routeKey struct{ method, path string }
	groups := map[routeKey][]*routeEntry{}
	var order []routeKey
	seen := map[routeKey]bool{}

	type routeSource struct {
		key    string
		source string
		route  config.Route
	}
	allRoutes := make([]routeSource, 0, len(cfg.Routes)+len(dynamic.All()))
	for i, route := range cfg.Routes {
		allRoutes = append(allRoutes, routeSource{key: fmt.Sprintf("config-%d", i), source: "config", route: route})
	}
	for _, dr := range dynamic.All() {
		allRoutes = append(allRoutes, routeSource{key: dr.ID, source: "dynamic", route: dr.Route})
	}
	for _, item := range allRoutes {
		route := item.route
		key := routeKey{route.Method, route.Path}
		e := &routeEntry{route: route, key: item.key, source: item.source, counter: &atomic.Uint64{}, callCount: &atomic.Uint64{}}
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
		sort.SliceStable(entries, func(i, j int) bool {
			return entries[i].route.Priority > entries[j].route.Priority
		})

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

				// Increment call counter (used for on_call and times matching)
				callN := e.callCount.Add(1)

				// Skip if times is set and this entry is exhausted
				if rt.Times > 0 && int(callN) > rt.Times {
					continue
				}
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

				// SSE streaming
				if rt.Stream {
					handleSSE(c, rt)
					return
				}

				// Redirect shorthand
				if rt.Redirect != "" {
					status := rt.RedirectStatus
					if status == 0 {
						status = http.StatusFound
					}
					c.Redirect(status, rt.Redirect)
					return
				}

				// Store CRUD operation
				if hasStoreOp(rt) {
					handleStoreOp(c, rt, bodyBytes, store)
					if rt.SetState != nil {
						state.Set(*rt.SetState)
					}
					for k, v := range rt.SetVars {
						vars.Set(k, v)
					}
					return
				}

				fault := ""
				shouldFault := false
				if rt.Fault != "" {
					fault = normalizeFault(rt.Fault)
					shouldFault = true
					if rt.ErrorRate > 0 {
						shouldFault = rand.Float64() < rt.ErrorRate
					}
				} else if rt.ErrorRate > 0 && rand.Float64() < rt.ErrorRate {
					fault = faultHTTPError
					shouldFault = true
				}
				if shouldFault && fault == faultTimeout {
					writeFault(c, fault, rt.ErrorStatus, rt.Delay)
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
				if shouldFault {
					status := rt.ErrorStatus
					if status == 0 && fault != faultHTTPError {
						status = rt.Status
					}
					writeFault(c, fault, status, rt.Delay)
					return
				}
				for hk, hv := range rt.Headers {
					c.Header(hk, hv)
				}
				applySetCookies(c, rt.SetCookies)

				// match conditions
				for _, m := range rt.Match {
					if matchesQuery(c, m.Query) &&
						matchesBody(bodyBytes, m.Body) &&
						matchesHeaders(c, m.Headers) &&
						matchesBodyPath(bodyBytes, m.BodyPath) &&
						matchesForm(c, bodyBytes, m.Form) &&
						matchesGraphQL(bodyBytes, m.GraphQL) &&
						matchesCookies(c, m.Cookies) &&
						matchesBodySchema(bodyBytes, m.BodySchema) {
						status := m.Status
						if status == 0 {
							status = http.StatusOK
						}
						body, fileCT, scriptStatus := resolveBody(m.Response, m.File, m.Script, tmplData, c.Params, store)
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
						// match-level delay (additive on top of route-level delay)
						if m.Delay > 0 {
							time.Sleep(time.Duration(m.Delay) * time.Millisecond)
						}
						if m.Fault != "" {
							writeFault(c, normalizeFault(m.Fault), m.Status, m.Delay)
							return
						}
						// match-level response headers override route-level headers
						for hk, hv := range m.ResponseHeaders {
							c.Header(hk, hv)
						}
						respondValidated(c, oaRouter, cfg.OpenAPIStrictResponse, status, ct, body)
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
						fireWebhook(rt.Webhook, tmplData, c.Params, store)
						return
					}
				}

				// timeline responses
				if len(rt.Timeline) > 0 {
					progress := timeline.Advance(e.key, rt.Method, rt.Path, e.source, len(rt.Timeline))
					picked := rt.Timeline[progress.Step-1]
					status := picked.Status
					if status == 0 {
						status = http.StatusOK
					}
					body, fileCT, scriptStatus2 := resolveBody(picked.Response, picked.File, picked.Script, tmplData, c.Params, store)
					if scriptStatus2 != 0 {
						status = scriptStatus2
					}
					if picked.Fault != "" {
						writeFault(c, normalizeFault(picked.Fault), status, 0)
						return
					}
					ct := picked.ContentType
					if ct == "" {
						ct = fileCT
					}
					if ct == "" {
						ct = rt.ContentType
					}
					c.Header("X-Specter-Timeline-Step", strconv.Itoa(progress.Step))
					c.Header("X-Specter-Timeline-Calls", strconv.FormatUint(progress.Calls, 10))
					respondValidated(c, oaRouter, cfg.OpenAPIStrictResponse, status, ct, body)
					if rt.SetState != nil {
						state.Set(*rt.SetState)
					}
					for k, v := range rt.SetVars {
						vars.Set(k, v)
					}
					fireWebhook(rt.Webhook, tmplData, c.Params, store)
					return
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
					body, fileCT, scriptStatus2 := resolveBody(picked.Response, picked.File, picked.Script, tmplData, c.Params, store)
					if scriptStatus2 != 0 {
						status = scriptStatus2
					}
					if picked.Fault != "" {
						writeFault(c, normalizeFault(picked.Fault), status, 0)
						return
					}
					ct := picked.ContentType
					if ct == "" {
						ct = fileCT
					}
					if ct == "" {
						ct = rt.ContentType
					}
					respondValidated(c, oaRouter, cfg.OpenAPIStrictResponse, status, ct, body)
					if rt.SetState != nil {
						state.Set(*rt.SetState)
					}
					for k, v := range rt.SetVars {
						vars.Set(k, v)
					}
					fireWebhook(rt.Webhook, tmplData, c.Params, store)
					return
				}

				// default response
				status := rt.Status
				if status == 0 {
					status = http.StatusOK
				}
				var body any
				var fileCT string
				var scriptStatus3 int
				if random && !routeHasBody(rt) {
					if randBody, randCT, ok := randomResponseBody(c.Request, oaRouter); ok {
						body = randBody
						if randCT != "" {
							fileCT = randCT
						}
					} else {
						body, fileCT, scriptStatus3 = resolveBody(rt.Response, rt.File, rt.Script, tmplData, c.Params, store)
					}
				} else {
					body, fileCT, scriptStatus3 = resolveBody(rt.Response, rt.File, rt.Script, tmplData, c.Params, store)
				}
				if scriptStatus3 != 0 {
					status = scriptStatus3
				}
				ct := rt.ContentType
				if ct == "" {
					ct = fileCT
				}
				respondValidated(c, oaRouter, cfg.OpenAPIStrictResponse, status, ct, body)
				if rt.SetState != nil {
					state.Set(*rt.SetState)
				}
				for k, v := range rt.SetVars {
					vars.Set(k, v)
				}
				fireWebhook(rt.Webhook, tmplData, c.Params, store)
				return
			}

			// No entry matched the current state
			c.JSON(http.StatusConflict, gin.H{"error": "no route matches current state", "state": currentState})
		})
	}

	if cfg.Proxy != "" {
		target, err := url.Parse(cfg.Proxy)
		if err != nil {
			log.Printf("invalid proxy URL %q: %v", safeURLForLog(cfg.Proxy), err)
		} else {
			proxy := httputil.NewSingleHostReverseProxy(target)
			r.NoRoute(func(c *gin.Context) {
				c.Request.Host = target.Host
				log.Printf("proxy → %s %s", c.Request.Method, redactURL(c.Request.URL.RequestURI()))
				proxy.ServeHTTP(c.Writer, c.Request)
			})
		}
	}

	return r
}
