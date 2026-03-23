package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Saku0512/specter/config"
)

func newSrv(cfg *config.Config) *Server {
	return New(cfg, false)
}

func do(srv *Server, method, path, body string) *httptest.ResponseRecorder {
	var b *bytes.Buffer
	if body != "" {
		b = bytes.NewBufferString(body)
	} else {
		b = &bytes.Buffer{}
	}
	req := httptest.NewRequest(method, path, b)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	return w
}

func jsonBody(t *testing.T, w *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &m); err != nil {
		t.Fatalf("failed to parse response body: %s", w.Body.String())
	}
	return m
}

// --- Basic ---

func TestBasicRoute(t *testing.T) {
	srv := newSrv(&config.Config{
		Routes: []config.Route{
			{Path: "/hello", Method: "GET", Status: 200, Response: map[string]any{"ok": true}},
		},
	})
	w := do(srv, "GET", "/hello", "")
	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}
	body := jsonBody(t, w)
	if body["ok"] != true {
		t.Errorf("expected ok:true, got %v", body)
	}
}

func TestDefaultStatus200(t *testing.T) {
	srv := newSrv(&config.Config{
		Routes: []config.Route{
			{Path: "/a", Method: "GET", Response: "ok"},
		},
	})
	w := do(srv, "GET", "/a", "")
	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestCustomStatus(t *testing.T) {
	srv := newSrv(&config.Config{
		Routes: []config.Route{
			{Path: "/created", Method: "POST", Status: 201, Response: map[string]any{"id": 1}},
		},
	})
	w := do(srv, "POST", "/created", "{}")
	if w.Code != 201 {
		t.Errorf("expected 201, got %d", w.Code)
	}
}

func TestNotFound(t *testing.T) {
	srv := newSrv(&config.Config{Routes: []config.Route{}})
	w := do(srv, "GET", "/missing", "")
	if w.Code != 404 {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

// --- Custom headers ---

func TestCustomHeaders(t *testing.T) {
	srv := newSrv(&config.Config{
		Routes: []config.Route{
			{Path: "/h", Method: "GET", Headers: map[string]string{"X-Foo": "bar"}, Response: nil},
		},
	})
	w := do(srv, "GET", "/h", "")
	if w.Header().Get("X-Foo") != "bar" {
		t.Errorf("expected X-Foo: bar, got %s", w.Header().Get("X-Foo"))
	}
}

// --- CORS ---

func TestCORSHeaders(t *testing.T) {
	srv := newSrv(&config.Config{
		CORS:   true,
		Routes: []config.Route{{Path: "/c", Method: "GET", Response: nil}},
	})
	w := do(srv, "GET", "/c", "")
	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("expected CORS header, got %s", w.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestCORSPreflight(t *testing.T) {
	srv := newSrv(&config.Config{
		CORS:   true,
		Routes: []config.Route{{Path: "/c", Method: "GET", Response: nil}},
	})
	w := do(srv, "OPTIONS", "/c", "")
	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", w.Code)
	}
}

// --- Path parameters ---

func TestPathParamString(t *testing.T) {
	srv := newSrv(&config.Config{
		Routes: []config.Route{
			{Path: "/users/:name", Method: "GET", Response: map[string]any{"name": ":name"}},
		},
	})
	w := do(srv, "GET", "/users/alice", "")
	body := jsonBody(t, w)
	if body["name"] != "alice" {
		t.Errorf("expected name:alice, got %v", body["name"])
	}
}

func TestPathParamNumeric(t *testing.T) {
	srv := newSrv(&config.Config{
		Routes: []config.Route{
			{Path: "/users/:id", Method: "GET", Response: map[string]any{"id": ":id"}},
		},
	})
	w := do(srv, "GET", "/users/42", "")
	body := jsonBody(t, w)
	// numeric params are converted to float64 by JSON unmarshaling
	if body["id"] != float64(42) {
		t.Errorf("expected id:42, got %v (%T)", body["id"], body["id"])
	}
}

// --- Multiple responses ---

func TestSequentialResponses(t *testing.T) {
	srv := newSrv(&config.Config{
		Routes: []config.Route{
			{
				Path:   "/seq",
				Method: "GET",
				Mode:   "sequential",
				Responses: []config.RouteResponse{
					{Status: 500, Response: map[string]any{"err": true}},
					{Status: 200, Response: map[string]any{"ok": true}},
				},
			},
		},
	})
	w1 := do(srv, "GET", "/seq", "")
	w2 := do(srv, "GET", "/seq", "")
	w3 := do(srv, "GET", "/seq", "") // loops back

	if w1.Code != 500 {
		t.Errorf("1st: expected 500, got %d", w1.Code)
	}
	if w2.Code != 200 {
		t.Errorf("2nd: expected 200, got %d", w2.Code)
	}
	if w3.Code != 500 {
		t.Errorf("3rd: expected 500 (loop), got %d", w3.Code)
	}
}

func TestRandomResponses(t *testing.T) {
	srv := newSrv(&config.Config{
		Routes: []config.Route{
			{
				Path:   "/rand",
				Method: "GET",
				Mode:   "random",
				Responses: []config.RouteResponse{
					{Status: 200, Response: map[string]any{"v": 1}},
					{Status: 201, Response: map[string]any{"v": 2}},
				},
			},
		},
	})
	seen := map[int]bool{}
	for i := 0; i < 50; i++ {
		w := do(srv, "GET", "/rand", "")
		seen[w.Code] = true
	}
	if !seen[200] || !seen[201] {
		t.Errorf("expected both 200 and 201 to appear in 50 requests, got %v", seen)
	}
}

// --- Query matching ---

func TestQueryMatch(t *testing.T) {
	srv := newSrv(&config.Config{
		Routes: []config.Route{
			{
				Path:   "/q",
				Method: "GET",
				Match: []config.RouteMatch{
					{Query: map[string]string{"status": "active"}, Status: 200, Response: map[string]any{"active": true}},
					{Query: map[string]string{"status": "inactive"}, Status: 404, Response: map[string]any{"active": false}},
				},
				Response: map[string]any{"default": true},
			},
		},
	})

	w := do(srv, "GET", "/q?status=active", "")
	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "true") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}

	w = do(srv, "GET", "/q?status=inactive", "")
	if w.Code != 404 {
		t.Errorf("expected 404, got %d", w.Code)
	}

	w = do(srv, "GET", "/q", "")
	if w.Code != 200 {
		t.Errorf("expected default 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "default") {
		t.Errorf("expected default response, got %s", w.Body.String())
	}
}

// --- Body matching ---

func TestBodyMatch(t *testing.T) {
	srv := newSrv(&config.Config{
		Routes: []config.Route{
			{
				Path:   "/login",
				Method: "POST",
				Match: []config.RouteMatch{
					{Body: map[string]any{"role": "admin"}, Status: 200, Response: map[string]any{"token": "admin-token"}},
					{Body: map[string]any{"role": "guest"}, Status: 403, Response: map[string]any{"error": "forbidden"}},
				},
				Response: map[string]any{"token": "default"},
			},
		},
	})

	w := do(srv, "POST", "/login", `{"role":"admin"}`)
	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "admin-token") {
		t.Errorf("expected admin-token, got %s", w.Body.String())
	}

	w = do(srv, "POST", "/login", `{"role":"guest"}`)
	if w.Code != 403 {
		t.Errorf("expected 403, got %d", w.Code)
	}

	w = do(srv, "POST", "/login", `{"role":"user"}`)
	if w.Code != 200 {
		t.Errorf("expected default 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "default") {
		t.Errorf("expected default token, got %s", w.Body.String())
	}
}

func TestBodyAndQueryMatch(t *testing.T) {
	srv := newSrv(&config.Config{
		Routes: []config.Route{
			{
				Path:   "/combined",
				Method: "POST",
				Match: []config.RouteMatch{
					{
						Query:    map[string]string{"v": "2"},
						Body:     map[string]any{"role": "admin"},
						Status:   200,
						Response: map[string]any{"match": "both"},
					},
				},
				Response: map[string]any{"match": "none"},
			},
		},
	})

	// both match
	w := do(srv, "POST", "/combined?v=2", `{"role":"admin"}`)
	if !strings.Contains(w.Body.String(), "both") {
		t.Errorf("expected both match, got %s", w.Body.String())
	}

	// only query matches
	w = do(srv, "POST", "/combined?v=2", `{"role":"guest"}`)
	if !strings.Contains(w.Body.String(), "none") {
		t.Errorf("expected none match, got %s", w.Body.String())
	}
}

// --- Content type ---

func TestContentTypePlainText(t *testing.T) {
	srv := newSrv(&config.Config{
		Routes: []config.Route{
			{Path: "/health", Method: "GET", ContentType: "text/plain", Response: "ok"},
		},
	})
	w := do(srv, "GET", "/health", "")
	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/plain") {
		t.Errorf("expected text/plain, got %s", ct)
	}
	if w.Body.String() != "ok" {
		t.Errorf("expected body 'ok', got %s", w.Body.String())
	}
}

func TestContentTypeHTML(t *testing.T) {
	srv := newSrv(&config.Config{
		Routes: []config.Route{
			{Path: "/page", Method: "GET", ContentType: "text/html", Response: "<h1>Hello</h1>"},
		},
	})
	w := do(srv, "GET", "/page", "")
	if ct := w.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Errorf("expected text/html, got %s", ct)
	}
	if w.Body.String() != "<h1>Hello</h1>" {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestContentTypeDefaultJSON(t *testing.T) {
	srv := newSrv(&config.Config{
		Routes: []config.Route{
			{Path: "/data", Method: "GET", Response: map[string]any{"ok": true}},
		},
	})
	w := do(srv, "GET", "/data", "")
	if ct := w.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Errorf("expected application/json, got %s", ct)
	}
}

func TestContentTypeMatchOverride(t *testing.T) {
	srv := newSrv(&config.Config{
		Routes: []config.Route{
			{
				Path:   "/mixed",
				Method: "GET",
				Match: []config.RouteMatch{
					{
						Query:       map[string]string{"fmt": "text"},
						ContentType: "text/plain",
						Response:    "plain",
					},
				},
				Response: map[string]any{"fmt": "json"},
			},
		},
	})

	w := do(srv, "GET", "/mixed?fmt=text", "")
	if ct := w.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/plain") {
		t.Errorf("expected text/plain, got %s", ct)
	}
	if w.Body.String() != "plain" {
		t.Errorf("expected body 'plain', got %s", w.Body.String())
	}

	w = do(srv, "GET", "/mixed", "")
	if ct := w.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Errorf("expected application/json for fallback, got %s", ct)
	}
}

// --- Response templates ---

func TestTemplateBody(t *testing.T) {
	srv := newSrv(&config.Config{
		Routes: []config.Route{
			{
				Path:     "/echo",
				Method:   "POST",
				Response: map[string]any{"hello": "{{ .body.name }}"},
			},
		},
	})
	w := do(srv, "POST", "/echo", `{"name":"Alice"}`)
	body := jsonBody(t, w)
	if body["hello"] != "Alice" {
		t.Errorf("expected hello:Alice, got %v", body["hello"])
	}
}

func TestTemplateQuery(t *testing.T) {
	srv := newSrv(&config.Config{
		Routes: []config.Route{
			{
				Path:     "/greet",
				Method:   "GET",
				Response: map[string]any{"msg": "hello {{ .query.name }}"},
			},
		},
	})
	w := do(srv, "GET", "/greet?name=Bob", "")
	body := jsonBody(t, w)
	if body["msg"] != "hello Bob" {
		t.Errorf("expected 'hello Bob', got %v", body["msg"])
	}
}

func TestTemplateParams(t *testing.T) {
	srv := newSrv(&config.Config{
		Routes: []config.Route{
			{
				Path:     "/users/:id",
				Method:   "GET",
				Response: map[string]any{"msg": "user {{ .params.id }}"},
			},
		},
	})
	w := do(srv, "GET", "/users/42", "")
	body := jsonBody(t, w)
	if body["msg"] != "user 42" {
		t.Errorf("expected 'user 42', got %v", body["msg"])
	}
}

func TestTemplateNoTemplate(t *testing.T) {
	srv := newSrv(&config.Config{
		Routes: []config.Route{
			{Path: "/static", Method: "GET", Response: map[string]any{"ok": true}},
		},
	})
	w := do(srv, "GET", "/static", "")
	body := jsonBody(t, w)
	if body["ok"] != true {
		t.Errorf("static response broken: %v", body)
	}
}

// --- Rate limiting ---

func TestRateLimit(t *testing.T) {
	srv := newSrv(&config.Config{
		Routes: []config.Route{
			{Path: "/limited", Method: "GET", RateLimit: 2, Response: map[string]any{"ok": true}},
		},
	})

	w := do(srv, "GET", "/limited", "")
	if w.Code != 200 {
		t.Errorf("1st: expected 200, got %d", w.Code)
	}
	w = do(srv, "GET", "/limited", "")
	if w.Code != 200 {
		t.Errorf("2nd: expected 200, got %d", w.Code)
	}
	w = do(srv, "GET", "/limited", "")
	if w.Code != http.StatusTooManyRequests {
		t.Errorf("3rd: expected 429, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "rate limit exceeded") {
		t.Errorf("expected rate limit error body, got %s", w.Body.String())
	}
}

func TestRateLimitWithReset(t *testing.T) {
	srv := newSrv(&config.Config{
		Routes: []config.Route{
			{Path: "/window", Method: "GET", RateLimit: 1, RateReset: 1, Response: map[string]any{"ok": true}},
		},
	})

	w := do(srv, "GET", "/window", "")
	if w.Code != 200 {
		t.Errorf("1st: expected 200, got %d", w.Code)
	}
	w = do(srv, "GET", "/window", "")
	if w.Code != http.StatusTooManyRequests {
		t.Errorf("2nd: expected 429, got %d", w.Code)
	}
	if w.Header().Get("Retry-After") == "" {
		t.Errorf("expected Retry-After header")
	}
}

func TestRateLimitUnaffectedRoutes(t *testing.T) {
	srv := newSrv(&config.Config{
		Routes: []config.Route{
			{Path: "/limited", Method: "GET", RateLimit: 1, Response: nil},
			{Path: "/unlimited", Method: "GET", Response: map[string]any{"ok": true}},
		},
	})

	do(srv, "GET", "/limited", "")
	do(srv, "GET", "/limited", "") // triggers 429

	w := do(srv, "GET", "/unlimited", "")
	if w.Code != 200 {
		t.Errorf("unlimited route: expected 200, got %d", w.Code)
	}
}

// --- Request history ---

func TestHistoryRecordsRequests(t *testing.T) {
	srv := newSrv(&config.Config{
		Routes: []config.Route{
			{Path: "/hello", Method: "GET", Response: map[string]any{"ok": true}},
		},
	})

	do(srv, "GET", "/hello", "")
	do(srv, "GET", "/hello", "")

	w := do(srv, "GET", "/__specter/requests", "")
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var entries []map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &entries); err != nil {
		t.Fatalf("failed to parse history: %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(entries))
	}
	if entries[0]["method"] != "GET" || entries[0]["path"] != "/hello" {
		t.Errorf("unexpected entry: %v", entries[0])
	}
}

func TestHistoryExcludesSpecterRoutes(t *testing.T) {
	srv := newSrv(&config.Config{Routes: []config.Route{}})

	do(srv, "GET", "/__specter/requests", "")
	do(srv, "GET", "/__specter/requests", "")

	w := do(srv, "GET", "/__specter/requests", "")
	var entries []map[string]any
	json.Unmarshal(w.Body.Bytes(), &entries)
	if len(entries) != 0 {
		t.Errorf("expected 0 entries (specter routes excluded), got %d", len(entries))
	}
}

func TestHistoryClear(t *testing.T) {
	srv := newSrv(&config.Config{
		Routes: []config.Route{
			{Path: "/a", Method: "GET", Response: nil},
		},
	})

	do(srv, "GET", "/a", "")
	do(srv, "GET", "/a", "")

	w := do(srv, "DELETE", "/__specter/requests", "")
	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", w.Code)
	}

	w = do(srv, "GET", "/__specter/requests", "")
	var entries []map[string]any
	json.Unmarshal(w.Body.Bytes(), &entries)
	if len(entries) != 0 {
		t.Errorf("expected 0 after clear, got %d", len(entries))
	}
}

func TestHistoryRecordsBody(t *testing.T) {
	srv := newSrv(&config.Config{
		Routes: []config.Route{
			{Path: "/post", Method: "POST", Response: nil},
		},
	})

	do(srv, "POST", "/post", `{"name":"Alice"}`)

	w := do(srv, "GET", "/__specter/requests", "")
	var entries []map[string]any
	json.Unmarshal(w.Body.Bytes(), &entries)
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0]["body"] != `{"name":"Alice"}` {
		t.Errorf("unexpected body in history: %v", entries[0]["body"])
	}
}

func TestHistoryPersistsAcrossReload(t *testing.T) {
	cfg := &config.Config{
		Routes: []config.Route{
			{Path: "/a", Method: "GET", Response: nil},
		},
	}
	srv := newSrv(cfg)

	do(srv, "GET", "/a", "")

	srv.Reload(&config.Config{
		Routes: []config.Route{
			{Path: "/b", Method: "GET", Response: nil},
		},
	})

	do(srv, "GET", "/b", "")

	w := do(srv, "GET", "/__specter/requests", "")
	var entries []map[string]any
	json.Unmarshal(w.Body.Bytes(), &entries)
	if len(entries) != 2 {
		t.Errorf("expected 2 entries after reload, got %d", len(entries))
	}
}

// --- Reload ---

func TestReload(t *testing.T) {
	cfg1 := &config.Config{
		Routes: []config.Route{
			{Path: "/r", Method: "GET", Status: 200, Response: map[string]any{"v": 1}},
		},
	}
	srv := newSrv(cfg1)

	w := do(srv, "GET", "/r", "")
	if w.Code != 200 {
		t.Errorf("before reload: expected 200, got %d", w.Code)
	}

	cfg2 := &config.Config{
		Routes: []config.Route{
			{Path: "/r", Method: "GET", Status: 202, Response: map[string]any{"v": 2}},
		},
	}
	srv.Reload(cfg2)

	w = do(srv, "GET", "/r", "")
	if w.Code != 202 {
		t.Errorf("after reload: expected 202, got %d", w.Code)
	}
}
