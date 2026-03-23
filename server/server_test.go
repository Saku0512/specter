package server

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

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
					{Query: map[string]string{"status": "^active$"}, Status: 200, Response: map[string]any{"active": true}},
					{Query: map[string]string{"status": "^inactive$"}, Status: 404, Response: map[string]any{"active": false}},
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

// --- Header matching ---

func doWithHeader(srv *Server, method, path, body, key, value string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set(key, value)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	return w
}

func TestHeaderMatch(t *testing.T) {
	srv := newSrv(&config.Config{
		Routes: []config.Route{
			{
				Path:   "/secure",
				Method: "GET",
				Match: []config.RouteMatch{
					{
						Headers:  map[string]string{"Authorization": "Bearer secret"},
						Response: map[string]any{"auth": true},
					},
				},
				Status:   401,
				Response: map[string]any{"auth": false},
			},
		},
	})

	// Correct token → matched response
	w := doWithHeader(srv, "GET", "/secure", "", "Authorization", "Bearer secret")
	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}
	body := jsonBody(t, w)
	if body["auth"] != true {
		t.Errorf("expected auth:true, got %v", body["auth"])
	}

	// No token → fallback
	w = do(srv, "GET", "/secure", "")
	if w.Code != 401 {
		t.Errorf("expected 401 fallback, got %d", w.Code)
	}
}

func TestHeaderMatch_caseInsensitiveName(t *testing.T) {
	srv := newSrv(&config.Config{
		Routes: []config.Route{
			{
				Path:   "/h",
				Method: "GET",
				Match: []config.RouteMatch{
					{
						Headers:  map[string]string{"x-api-key": "mykey"},
						Response: map[string]any{"ok": true},
					},
				},
				Response: map[string]any{"ok": false},
			},
		},
	})

	w := doWithHeader(srv, "GET", "/h", "", "X-Api-Key", "mykey")
	body := jsonBody(t, w)
	if body["ok"] != true {
		t.Errorf("expected ok:true for case-insensitive header match, got %v", body)
	}
}

func TestHeaderMatch_combined(t *testing.T) {
	srv := newSrv(&config.Config{
		Routes: []config.Route{
			{
				Path:   "/combo",
				Method: "POST",
				Match: []config.RouteMatch{
					{
						Headers:  map[string]string{"X-Role": "admin"},
						Body:     map[string]any{"action": "delete"},
						Status:   200,
						Response: map[string]any{"allowed": true},
					},
				},
				Status:   403,
				Response: map[string]any{"allowed": false},
			},
		},
	})

	// Both match
	req := httptest.NewRequest("POST", "/combo", strings.NewReader(`{"action":"delete"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Role", "admin")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// Header matches but body doesn't
	req = httptest.NewRequest("POST", "/combo", strings.NewReader(`{"action":"read"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Role", "admin")
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != 403 {
		t.Errorf("expected 403 fallback, got %d", w.Code)
	}
}

// --- Request verification ---

func assertAPI(srv *Server, body string) *httptest.ResponseRecorder {
	return do(srv, "POST", "/__specter/requests/assert", body)
}

func TestAssert_matchByPath(t *testing.T) {
	srv := newSrv(&config.Config{
		Routes: []config.Route{{Path: "/users", Method: "GET", Response: nil}},
	})
	do(srv, "GET", "/users", "")

	w := assertAPI(srv, `{"path":"/users"}`)
	if w.Code != 200 {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAssert_matchByMethodAndPath(t *testing.T) {
	srv := newSrv(&config.Config{
		Routes: []config.Route{{Path: "/users", Method: "POST", Response: nil}},
	})
	do(srv, "POST", "/users", `{"name":"Alice"}`)

	w := assertAPI(srv, `{"method":"POST","path":"/users"}`)
	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestAssert_matchByBody(t *testing.T) {
	srv := newSrv(&config.Config{
		Routes: []config.Route{{Path: "/users", Method: "POST", Response: nil}},
	})
	do(srv, "POST", "/users", `{"name":"Alice","role":"admin"}`)

	w := assertAPI(srv, `{"path":"/users","body":{"name":"Alice"}}`)
	if w.Code != 200 {
		t.Errorf("expected 200 for body match, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAssert_noMatch(t *testing.T) {
	srv := newSrv(&config.Config{
		Routes: []config.Route{{Path: "/users", Method: "GET", Response: nil}},
	})
	do(srv, "GET", "/users", "")

	w := assertAPI(srv, `{"path":"/orders"}`)
	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d", w.Code)
	}
	var body map[string]any
	json.Unmarshal(w.Body.Bytes(), &body)
	if body["ok"] != false {
		t.Errorf("expected ok:false")
	}
}

func TestAssert_exactCount(t *testing.T) {
	srv := newSrv(&config.Config{
		Routes: []config.Route{{Path: "/users", Method: "GET", Response: nil}},
	})
	do(srv, "GET", "/users", "")
	do(srv, "GET", "/users", "")

	// count: 2 → pass
	w := assertAPI(srv, `{"path":"/users","count":2}`)
	if w.Code != 200 {
		t.Errorf("expected 200 for count:2, got %d", w.Code)
	}

	// count: 1 → fail
	w = assertAPI(srv, `{"path":"/users","count":1}`)
	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422 for count:1 mismatch, got %d", w.Code)
	}
}

func TestAssert_countZero(t *testing.T) {
	srv := newSrv(&config.Config{
		Routes: []config.Route{{Path: "/users", Method: "GET", Response: nil}},
	})
	// No requests made yet
	w := assertAPI(srv, `{"path":"/admin","count":0}`)
	if w.Code != 200 {
		t.Errorf("expected 200 for count:0 (none recorded), got %d", w.Code)
	}
}

func TestAssert_matchByQuery(t *testing.T) {
	srv := newSrv(&config.Config{
		Routes: []config.Route{{Path: "/search", Method: "GET", Response: nil}},
	})
	do(srv, "GET", "/search?q=hello", "")

	w := assertAPI(srv, `{"path":"/search","query":{"q":"hello"}}`)
	if w.Code != 200 {
		t.Errorf("expected 200 for query match, got %d: %s", w.Code, w.Body.String())
	}
}

func TestRequestByIndex(t *testing.T) {
	srv := newSrv(&config.Config{
		Routes: []config.Route{{Path: "/a", Method: "GET", Response: nil}},
	})
	do(srv, "GET", "/a", "")
	do(srv, "GET", "/a", "")

	w := do(srv, "GET", "/__specter/requests/0", "")
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var entry map[string]any
	json.Unmarshal(w.Body.Bytes(), &entry)
	if entry["path"] != "/a" {
		t.Errorf("unexpected entry: %v", entry)
	}
}

func TestRequestByIndex_outOfRange(t *testing.T) {
	srv := newSrv(&config.Config{Routes: []config.Route{}})

	w := do(srv, "GET", "/__specter/requests/5", "")
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestRequestByIndex_invalid(t *testing.T) {
	srv := newSrv(&config.Config{Routes: []config.Route{}})

	w := do(srv, "GET", "/__specter/requests/abc", "")
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// --- Stateful mocking ---

func setServerState(srv *Server, state string) {
	srv.state.Set(state)
}

func getServerState(srv *Server) string {
	return srv.state.Get()
}

func TestState_matchesCurrentState(t *testing.T) {
	setStr := "logged_in"
	srv := newSrv(&config.Config{
		Routes: []config.Route{
			{Path: "/profile", Method: "GET", State: "logged_in", Response: map[string]any{"name": "Alice"}},
		},
	})

	// State is "" → 409
	w := do(srv, "GET", "/profile", "")
	if w.Code != http.StatusConflict {
		t.Errorf("expected 409 when state mismatch, got %d", w.Code)
	}

	// Set state → 200
	srv.state.Set(setStr)
	w = do(srv, "GET", "/profile", "")
	if w.Code != 200 {
		t.Errorf("expected 200 when state matches, got %d", w.Code)
	}
}

func TestState_fallbackToNoStateCondition(t *testing.T) {
	srv := newSrv(&config.Config{
		Routes: []config.Route{
			{Path: "/profile", Method: "GET", State: "logged_in", Status: 200, Response: map[string]any{"auth": true}},
			{Path: "/profile", Method: "GET", Status: 401, Response: map[string]any{"auth": false}},
		},
	})

	// No state → fallback route
	w := do(srv, "GET", "/profile", "")
	if w.Code != 401 {
		t.Errorf("expected 401 fallback, got %d", w.Code)
	}

	// Logged in → first route
	setServerState(srv, "logged_in")
	w = do(srv, "GET", "/profile", "")
	if w.Code != 200 {
		t.Errorf("expected 200 when logged_in, got %d", w.Code)
	}
}

func TestState_setStateTransitions(t *testing.T) {
	loggedIn := "logged_in"
	empty := ""
	srv := newSrv(&config.Config{
		Routes: []config.Route{
			{Path: "/login", Method: "POST", SetState: &loggedIn, Response: map[string]any{"ok": true}},
			{Path: "/logout", Method: "POST", State: "logged_in", SetState: &empty, Response: map[string]any{"ok": true}},
		},
	})

	if getServerState(srv) != "" {
		t.Errorf("initial state should be empty")
	}

	do(srv, "POST", "/login", "")
	if getServerState(srv) != "logged_in" {
		t.Errorf("expected logged_in after login, got %q", getServerState(srv))
	}

	do(srv, "POST", "/logout", "")
	if getServerState(srv) != "" {
		t.Errorf("expected empty after logout, got %q", getServerState(srv))
	}
}

func TestState_persistsAcrossReload(t *testing.T) {
	loggedIn := "logged_in"
	srv := newSrv(&config.Config{
		Routes: []config.Route{
			{Path: "/login", Method: "POST", SetState: &loggedIn, Response: map[string]any{"ok": true}},
		},
	})

	do(srv, "POST", "/login", "")
	if getServerState(srv) != "logged_in" {
		t.Fatalf("state not set before reload")
	}

	srv.Reload(&config.Config{
		Routes: []config.Route{
			{Path: "/profile", Method: "GET", State: "logged_in", Response: map[string]any{"name": "Alice"}},
		},
	})

	if getServerState(srv) != "logged_in" {
		t.Errorf("state lost after reload, got %q", getServerState(srv))
	}
	w := do(srv, "GET", "/profile", "")
	if w.Code != 200 {
		t.Errorf("expected 200 after reload with state, got %d", w.Code)
	}
}

func TestState_specterStateEndpoint(t *testing.T) {
	srv := newSrv(&config.Config{Routes: []config.Route{}})

	// GET initial state
	w := do(srv, "GET", "/__specter/state", "")
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var body map[string]any
	json.Unmarshal(w.Body.Bytes(), &body)
	if body["state"] != "" {
		t.Errorf("expected empty initial state, got %v", body["state"])
	}

	// PUT to set state
	w = do(srv, "PUT", "/__specter/state", `{"state":"testing"}`)
	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", w.Code)
	}
	if getServerState(srv) != "testing" {
		t.Errorf("expected state=testing, got %q", getServerState(srv))
	}
}

// --- Faker templates ---

func TestFakeTemplateName(t *testing.T) {
	srv := newSrv(&config.Config{
		Routes: []config.Route{
			{Path: "/fake", Method: "GET", Response: map[string]any{"name": `{{ fake "name" }}`}},
		},
	})
	w := do(srv, "GET", "/fake", "")
	body := jsonBody(t, w)
	name, ok := body["name"].(string)
	if !ok || name == "" || name == `{{ fake "name" }}` {
		t.Errorf("expected non-empty rendered name, got %v", body["name"])
	}
}

func TestFakeTemplateEmail(t *testing.T) {
	srv := newSrv(&config.Config{
		Routes: []config.Route{
			{Path: "/fake", Method: "GET", Response: map[string]any{"email": `{{ fake "email" }}`}},
		},
	})
	w := do(srv, "GET", "/fake", "")
	body := jsonBody(t, w)
	email, ok := body["email"].(string)
	if !ok || !strings.Contains(email, "@") {
		t.Errorf("expected email with @, got %v", body["email"])
	}
}

func TestFakeTemplateUUID(t *testing.T) {
	srv := newSrv(&config.Config{
		Routes: []config.Route{
			{Path: "/fake", Method: "GET", Response: map[string]any{"id": `{{ fake "uuid" }}`}},
		},
	})
	w := do(srv, "GET", "/fake", "")
	body := jsonBody(t, w)
	id, ok := body["id"].(string)
	if !ok || len(id) != 36 {
		t.Errorf("expected UUID (36 chars), got %v", body["id"])
	}
}

func TestFakeTemplateUnknownReturnsEmpty(t *testing.T) {
	srv := newSrv(&config.Config{
		Routes: []config.Route{
			{Path: "/fake", Method: "GET", Response: map[string]any{"x": `{{ fake "unknown" }}`}},
		},
	})
	w := do(srv, "GET", "/fake", "")
	body := jsonBody(t, w)
	if body["x"] != "" {
		t.Errorf("expected empty string for unknown fake type, got %v", body["x"])
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

// --- Webhook ---

func TestWebhook_fired(t *testing.T) {
	received := make(chan []byte, 1)
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		received <- b
		w.WriteHeader(http.StatusOK)
	}))
	defer target.Close()

	srv := newSrv(&config.Config{
		Routes: []config.Route{
			{
				Path:     "/pay",
				Method:   "POST",
				Status:   200,
				Response: map[string]any{"ok": true},
				Webhook: &config.Webhook{
					URL:  target.URL + "/cb",
					Body: map[string]any{"event": "payment"},
				},
			},
		},
	})

	w := do(srv, "POST", "/pay", `{"amount":100}`)
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	select {
	case body := <-received:
		var m map[string]any
		if err := json.Unmarshal(body, &m); err != nil {
			t.Fatalf("webhook body not valid JSON: %s", body)
		}
		if m["event"] != "payment" {
			t.Errorf("unexpected webhook body: %v", m)
		}
	case <-waitTimeout(500):
		t.Fatal("webhook not received within timeout")
	}
}

func TestWebhook_bodyTemplate(t *testing.T) {
	received := make(chan []byte, 1)
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		received <- b
		w.WriteHeader(http.StatusOK)
	}))
	defer target.Close()

	srv := newSrv(&config.Config{
		Routes: []config.Route{
			{
				Path:     "/orders",
				Method:   "POST",
				Status:   201,
				Response: map[string]any{"created": true},
				Webhook: &config.Webhook{
					URL:  target.URL + "/notify",
					Body: map[string]any{"user": "{{ .body.user }}"},
				},
			},
		},
	})

	do(srv, "POST", "/orders", `{"user":"alice"}`)

	select {
	case body := <-received:
		var m map[string]any
		if err := json.Unmarshal(body, &m); err != nil {
			t.Fatalf("webhook body not valid JSON: %s", body)
		}
		if m["user"] != "alice" {
			t.Errorf("expected user=alice in webhook body, got %v", m)
		}
	case <-waitTimeout(500):
		t.Fatal("webhook not received within timeout")
	}
}

func TestWebhook_customMethod(t *testing.T) {
	method := make(chan string, 1)
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method <- r.Method
		w.WriteHeader(http.StatusOK)
	}))
	defer target.Close()

	srv := newSrv(&config.Config{
		Routes: []config.Route{
			{
				Path:   "/evt",
				Method: "GET",
				Status: 200,
				Webhook: &config.Webhook{
					URL:    target.URL + "/hook",
					Method: "PUT",
				},
			},
		},
	})

	do(srv, "GET", "/evt", "")

	select {
	case m := <-method:
		if m != "PUT" {
			t.Errorf("expected PUT, got %s", m)
		}
	case <-waitTimeout(500):
		t.Fatal("webhook not received within timeout")
	}
}

func TestWebhook_nilSkipped(t *testing.T) {
	// Route with no webhook must not panic
	srv := newSrv(&config.Config{
		Routes: []config.Route{
			{Path: "/ok", Method: "GET", Status: 200, Response: map[string]any{"ok": true}},
		},
	})
	w := do(srv, "GET", "/ok", "")
	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func waitTimeout(ms int) <-chan struct{} {
	ch := make(chan struct{})
	go func() {
		time.Sleep(time.Duration(ms) * time.Millisecond)
		close(ch)
	}()
	return ch
}

// --- File Response ---

func TestFileResponse_json(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "*.json")
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString(`{"id":1,"name":"Alice"}`)
	f.Close()

	srv := newSrv(&config.Config{
		Routes: []config.Route{
			{Path: "/u", Method: "GET", File: f.Name()},
		},
	})

	w := do(srv, "GET", "/u", "")
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	m := jsonBody(t, w)
	if m["name"] != "Alice" {
		t.Errorf("unexpected body: %v", m)
	}
}

func TestFileResponse_yaml(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString("id: 2\nname: Bob\n")
	f.Close()

	srv := newSrv(&config.Config{
		Routes: []config.Route{
			{Path: "/u", Method: "GET", File: f.Name()},
		},
	})

	w := do(srv, "GET", "/u", "")
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	m := jsonBody(t, w)
	if m["name"] != "Bob" {
		t.Errorf("unexpected body: %v", m)
	}
}

func TestFileResponse_yml(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "*.yml")
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString("status: ok\n")
	f.Close()

	srv := newSrv(&config.Config{
		Routes: []config.Route{
			{Path: "/s", Method: "GET", File: f.Name()},
		},
	})

	w := do(srv, "GET", "/s", "")
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	m := jsonBody(t, w)
	if m["status"] != "ok" {
		t.Errorf("unexpected body: %v", m)
	}
}

func TestFileResponse_plainText(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "*.txt")
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString("hello world")
	f.Close()

	srv := newSrv(&config.Config{
		Routes: []config.Route{
			{Path: "/t", Method: "GET", File: f.Name(), ContentType: "text/plain"},
		},
	})

	w := do(srv, "GET", "/t", "")
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "hello world") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestFileResponse_missingFile(t *testing.T) {
	srv := newSrv(&config.Config{
		Routes: []config.Route{
			{Path: "/x", Method: "GET", File: "/nonexistent/file.json"},
		},
	})

	w := do(srv, "GET", "/x", "")
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	m := jsonBody(t, w)
	if m["error"] == nil {
		t.Errorf("expected error key in body, got %v", m)
	}
}

func TestFileResponse_inResponses(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "*.json")
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString(`{"seq":1}`)
	f.Close()

	srv := newSrv(&config.Config{
		Routes: []config.Route{
			{
				Path:   "/r",
				Method: "GET",
				Responses: []config.RouteResponse{
					{File: f.Name()},
					{Status: 204},
				},
			},
		},
	})

	w := do(srv, "GET", "/r", "")
	m := jsonBody(t, w)
	if m["seq"] != float64(1) {
		t.Errorf("expected seq=1, got %v", m)
	}

	w = do(srv, "GET", "/r", "")
	if w.Code != 204 {
		t.Errorf("expected 204 on second call, got %d", w.Code)
	}
}

// --- Chaos / Fault Injection ---

func TestErrorRate_alwaysFault(t *testing.T) {
	srv := newSrv(&config.Config{
		Routes: []config.Route{
			{Path: "/api", Method: "GET", ErrorRate: 1.0, Response: map[string]any{"ok": true}},
		},
	})
	w := do(srv, "GET", "/api", "")
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", w.Code)
	}
}

func TestErrorRate_customStatus(t *testing.T) {
	srv := newSrv(&config.Config{
		Routes: []config.Route{
			{Path: "/api", Method: "GET", ErrorRate: 1.0, ErrorStatus: 500, Response: map[string]any{"ok": true}},
		},
	})
	w := do(srv, "GET", "/api", "")
	if w.Code != 500 {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestErrorRate_neverFault(t *testing.T) {
	srv := newSrv(&config.Config{
		Routes: []config.Route{
			{Path: "/api", Method: "GET", ErrorRate: 0.0, Response: map[string]any{"ok": true}},
		},
	})
	for i := 0; i < 10; i++ {
		w := do(srv, "GET", "/api", "")
		if w.Code != 200 {
			t.Errorf("expected 200, got %d", w.Code)
		}
	}
}

func TestDelayRange_valid(t *testing.T) {
	srv := newSrv(&config.Config{
		Routes: []config.Route{
			{Path: "/slow", Method: "GET", DelayMin: 0, DelayMax: 5, Response: map[string]any{"ok": true}},
		},
	})
	w := do(srv, "GET", "/slow", "")
	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

// --- VarStore / Multi-variable state ---

func putVars(srv *Server, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPut, "/__specter/vars", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	return w
}

func TestVars_setAndGet(t *testing.T) {
	srv := newSrv(&config.Config{Routes: []config.Route{}})

	w := putVars(srv, `{"role":"admin","tier":"gold"}`)
	if w.Code != http.StatusNoContent {
		t.Fatalf("PUT vars: expected 204, got %d", w.Code)
	}

	req := httptest.NewRequest(http.MethodGet, "/__specter/vars", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	var m map[string]string
	json.Unmarshal(rec.Body.Bytes(), &m)
	if m["role"] != "admin" || m["tier"] != "gold" {
		t.Errorf("unexpected vars: %v", m)
	}
}

func TestVars_perKeyPutAndDelete(t *testing.T) {
	srv := newSrv(&config.Config{Routes: []config.Route{}})

	req := httptest.NewRequest(http.MethodPut, "/__specter/vars/color",
		bytes.NewBufferString(`{"value":"blue"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/__specter/vars/color", nil)
	rec = httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	var m map[string]any
	json.Unmarshal(rec.Body.Bytes(), &m)
	if m["value"] != "blue" {
		t.Errorf("expected blue, got %v", m)
	}

	req = httptest.NewRequest(http.MethodDelete, "/__specter/vars/color", nil)
	rec = httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/__specter/vars/color", nil)
	rec = httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	json.Unmarshal(rec.Body.Bytes(), &m)
	if m["value"] != "" {
		t.Errorf("expected empty after delete, got %v", m)
	}
}

func TestVars_clear(t *testing.T) {
	srv := newSrv(&config.Config{Routes: []config.Route{}})
	putVars(srv, `{"a":"1","b":"2"}`)

	req := httptest.NewRequest(http.MethodDelete, "/__specter/vars", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/__specter/vars", nil)
	rec = httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	var m map[string]string
	json.Unmarshal(rec.Body.Bytes(), &m)
	if len(m) != 0 {
		t.Errorf("expected empty vars after clear, got %v", m)
	}
}

func TestVars_routeMatchCondition(t *testing.T) {
	srv := newSrv(&config.Config{
		Routes: []config.Route{
			{
				Path:   "/data",
				Method: "GET",
				Vars:   map[string]string{"role": "admin"},
				Status: 200,
				Response: map[string]any{"access": "granted"},
			},
			{
				Path:     "/data",
				Method:   "GET",
				Status:   403,
				Response: map[string]any{"access": "denied"},
			},
		},
	})

	// Without vars set → 403
	w := do(srv, "GET", "/data", "")
	if w.Code != 403 {
		t.Errorf("expected 403 without vars, got %d", w.Code)
	}

	// Set role=admin → 200
	putVars(srv, `{"role":"admin"}`)
	w = do(srv, "GET", "/data", "")
	if w.Code != 200 {
		t.Errorf("expected 200 with role=admin, got %d", w.Code)
	}
}

func TestVars_setVarsOnResponse(t *testing.T) {
	srv := newSrv(&config.Config{
		Routes: []config.Route{
			{
				Path:     "/login",
				Method:   "POST",
				Status:   200,
				Response: map[string]any{"ok": true},
				SetVars:  map[string]string{"logged_in": "true", "user": "alice"},
			},
			{
				Path:   "/profile",
				Method: "GET",
				Vars:   map[string]string{"logged_in": "true"},
				Status: 200,
				Response: map[string]any{"name": "alice"},
			},
			{
				Path:     "/profile",
				Method:   "GET",
				Status:   401,
				Response: map[string]any{"error": "unauthorized"},
			},
		},
	})

	w := do(srv, "GET", "/profile", "")
	if w.Code != 401 {
		t.Errorf("expected 401 before login, got %d", w.Code)
	}

	do(srv, "POST", "/login", `{}`)

	w = do(srv, "GET", "/profile", "")
	if w.Code != 200 {
		t.Errorf("expected 200 after login, got %d", w.Code)
	}
}

// --- Dynamic Route Management ---

func postRoute(srv *Server, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/__specter/routes",
		bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	return w
}

func TestDynamicRoute_addAndUse(t *testing.T) {
	srv := newSrv(&config.Config{})

	w := postRoute(srv, `{"path":"/dynamic","method":"GET","status":200,"response":{"ok":true}}`)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body)
	}

	// Give rebuild goroutine a moment
	time.Sleep(20 * time.Millisecond)

	w = do(srv, "GET", "/dynamic", "")
	if w.Code != 200 {
		t.Errorf("expected 200 after adding route, got %d", w.Code)
	}
}

func TestDynamicRoute_deleteByID(t *testing.T) {
	srv := newSrv(&config.Config{})

	w := postRoute(srv, `{"path":"/tmp","method":"GET","status":200,"response":{"ok":true}}`)
	var m map[string]any
	json.Unmarshal(w.Body.Bytes(), &m)
	id := m["id"].(string)

	time.Sleep(20 * time.Millisecond)

	req := httptest.NewRequest(http.MethodDelete, "/__specter/routes/"+id, nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}

	time.Sleep(20 * time.Millisecond)

	w = do(srv, "GET", "/tmp", "")
	if w.Code != 404 {
		t.Errorf("expected 404 after deleting route, got %d", w.Code)
	}
}

func TestDynamicRoute_clearAll(t *testing.T) {
	srv := newSrv(&config.Config{})

	postRoute(srv, `{"path":"/a","method":"GET","status":200}`)
	postRoute(srv, `{"path":"/b","method":"GET","status":200}`)
	time.Sleep(20 * time.Millisecond)

	req := httptest.NewRequest(http.MethodDelete, "/__specter/routes", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}

	time.Sleep(20 * time.Millisecond)

	if do(srv, "GET", "/a", "").Code != 404 {
		t.Errorf("expected 404 for /a after clear")
	}
	if do(srv, "GET", "/b", "").Code != 404 {
		t.Errorf("expected 404 for /b after clear")
	}
}

func TestDynamicRoute_listAll(t *testing.T) {
	srv := newSrv(&config.Config{
		Routes: []config.Route{{Path: "/static", Method: "GET"}},
	})
	postRoute(srv, `{"path":"/dyn","method":"POST","status":201}`)
	time.Sleep(20 * time.Millisecond)

	req := httptest.NewRequest(http.MethodGet, "/__specter/routes", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	var list []map[string]any
	json.Unmarshal(rec.Body.Bytes(), &list)
	if len(list) != 2 {
		t.Errorf("expected 2 routes, got %d", len(list))
	}
}

func TestDynamicRoute_missingFields(t *testing.T) {
	srv := newSrv(&config.Config{})
	w := postRoute(srv, `{"path":"/no-method"}`)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// --- OpenAPI Validation ---

const minimalOpenAPISpec = `
openapi: "3.0.0"
info:
  title: Test
  version: "1.0"
paths:
  /items:
    post:
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required: [name]
              properties:
                name:
                  type: string
      responses:
        "201":
          description: created
`

func TestOpenAPIValidation_valid(t *testing.T) {
	f, _ := os.CreateTemp(t.TempDir(), "*.yaml")
	f.WriteString(minimalOpenAPISpec)
	f.Close()

	srv := newSrv(&config.Config{
		OpenAPI: f.Name(),
		Routes:  []config.Route{{Path: "/items", Method: "POST", Status: 201}},
	})

	w := do(srv, "POST", "/items", `{"name":"widget"}`)
	if w.Code != 201 {
		t.Fatalf("expected 201, got %d", w.Code)
	}
	if w.Header().Get("X-Specter-Validation-Error") != "" {
		t.Errorf("expected no validation error, got: %s", w.Header().Get("X-Specter-Validation-Error"))
	}
}

func TestOpenAPIValidation_missingRequiredField(t *testing.T) {
	f, _ := os.CreateTemp(t.TempDir(), "*.yaml")
	f.WriteString(minimalOpenAPISpec)
	f.Close()

	srv := newSrv(&config.Config{
		OpenAPI: f.Name(),
		Routes:  []config.Route{{Path: "/items", Method: "POST", Status: 201}},
	})

	// Missing required "name" field
	w := do(srv, "POST", "/items", `{"price":10}`)
	if w.Code != 201 {
		t.Fatalf("mock should still respond 201, got %d", w.Code)
	}
	if w.Header().Get("X-Specter-Validation-Error") == "" {
		t.Errorf("expected X-Specter-Validation-Error header to be set")
	}
}

func TestOpenAPIValidation_noSpec(t *testing.T) {
	srv := newSrv(&config.Config{
		Routes: []config.Route{{Path: "/items", Method: "POST", Status: 201}},
	})
	w := do(srv, "POST", "/items", `{}`)
	if w.Code != 201 {
		t.Fatalf("expected 201, got %d", w.Code)
	}
	if w.Header().Get("X-Specter-Validation-Error") != "" {
		t.Errorf("no spec set, should have no validation header")
	}
}

func TestOpenAPIValidation_routeNotInSpec(t *testing.T) {
	f, _ := os.CreateTemp(t.TempDir(), "*.yaml")
	f.WriteString(minimalOpenAPISpec)
	f.Close()

	srv := newSrv(&config.Config{
		OpenAPI: f.Name(),
		Routes:  []config.Route{{Path: "/other", Method: "GET", Status: 200}},
	})

	// /other is not in the spec, should pass through silently
	w := do(srv, "GET", "/other", "")
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if w.Header().Get("X-Specter-Validation-Error") != "" {
		t.Errorf("route not in spec should not produce validation error")
	}
}

// --- BodyPath / Regex matching ---

func TestBodyPath_exactMatch(t *testing.T) {
	srv := newSrv(&config.Config{
		Routes: []config.Route{
			{
				Path:   "/users",
				Method: "POST",
				Match: []config.RouteMatch{
					{
						BodyPath: map[string]string{"role": "admin"},
						Status:   201,
						Response: map[string]any{"ok": true},
					},
				},
				Status:   400,
				Response: map[string]any{"error": "bad"},
			},
		},
	})

	w := do(srv, "POST", "/users", `{"role":"admin"}`)
	if w.Code != 201 {
		t.Fatalf("expected 201, got %d", w.Code)
	}
	w2 := do(srv, "POST", "/users", `{"role":"user"}`)
	if w2.Code != 400 {
		t.Fatalf("expected 400, got %d", w2.Code)
	}
}

func TestBodyPath_regexMatch(t *testing.T) {
	srv := newSrv(&config.Config{
		Routes: []config.Route{
			{
				Path:   "/items",
				Method: "POST",
				Match: []config.RouteMatch{
					{
						BodyPath: map[string]string{"name": "^pro_"},
						Status:   200,
						Response: map[string]any{"tier": "pro"},
					},
				},
				Status:   200,
				Response: map[string]any{"tier": "free"},
			},
		},
	})

	w := do(srv, "POST", "/items", `{"name":"pro_plan"}`)
	body := jsonBody(t, w)
	if body["tier"] != "pro" {
		t.Fatalf("expected tier=pro, got %v", body["tier"])
	}

	w2 := do(srv, "POST", "/items", `{"name":"free_plan"}`)
	body2 := jsonBody(t, w2)
	if body2["tier"] != "free" {
		t.Fatalf("expected tier=free, got %v", body2["tier"])
	}
}

func TestBodyPath_nestedPath(t *testing.T) {
	srv := newSrv(&config.Config{
		Routes: []config.Route{
			{
				Path:   "/data",
				Method: "POST",
				Match: []config.RouteMatch{
					{
						BodyPath: map[string]string{"user.role": "^admin$"},
						Status:   200,
						Response: map[string]any{"access": "granted"},
					},
				},
				Status:   403,
				Response: map[string]any{"access": "denied"},
			},
		},
	})

	w := do(srv, "POST", "/data", `{"user":{"role":"admin"}}`)
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	w2 := do(srv, "POST", "/data", `{"user":{"role":"guest"}}`)
	if w2.Code != 403 {
		t.Fatalf("expected 403, got %d", w2.Code)
	}
}

func TestBodyPath_missingField(t *testing.T) {
	srv := newSrv(&config.Config{
		Routes: []config.Route{
			{
				Path:   "/x",
				Method: "POST",
				Match: []config.RouteMatch{
					{
						BodyPath: map[string]string{"missing": ".*"},
						Status:   200,
						Response: map[string]any{"hit": true},
					},
				},
				Status:   404,
				Response: map[string]any{"hit": false},
			},
		},
	})

	w := do(srv, "POST", "/x", `{"other":"value"}`)
	if w.Code != 404 {
		t.Fatalf("expected 404 (field missing), got %d", w.Code)
	}
}

// --- on_call (N回目マッチ) ---

func TestOnCall_routeLevel(t *testing.T) {
	srv := newSrv(&config.Config{
		Routes: []config.Route{
			{Path: "/retry", Method: "GET", OnCall: 2, Status: 200, Response: map[string]any{"ok": true}},
			{Path: "/retry", Method: "GET", Status: 503, Response: map[string]any{"error": "unavailable"}},
		},
	})

	// Call 1: on_call=2 entry skips (callN=1 ≠ 2), fallback to 503
	w1 := do(srv, "GET", "/retry", "")
	if w1.Code != 503 {
		t.Fatalf("call 1: expected 503, got %d", w1.Code)
	}
	// Call 2: on_call=2 entry matches (callN=2 == 2) → 200
	w2 := do(srv, "GET", "/retry", "")
	if w2.Code != 200 {
		t.Fatalf("call 2: expected 200, got %d", w2.Code)
	}
	// Call 3: on_call=2 entry skips again → 503
	w3 := do(srv, "GET", "/retry", "")
	if w3.Code != 503 {
		t.Fatalf("call 3: expected 503, got %d", w3.Code)
	}
}

func TestOnCall_inResponses(t *testing.T) {
	srv := newSrv(&config.Config{
		Routes: []config.Route{
			{
				Path:   "/seq",
				Method: "GET",
				Responses: []config.RouteResponse{
					{OnCall: 2, Status: 201, Response: map[string]any{"special": true}},
					{Status: 200, Response: map[string]any{"normal": true}},
				},
			},
		},
	})

	// Call 1: on_call=2 doesn't match, picks from pool (normal)
	w1 := do(srv, "GET", "/seq", "")
	if w1.Code != 200 {
		t.Fatalf("call 1: expected 200, got %d", w1.Code)
	}
	// Call 2: on_call=2 matches → 201
	w2 := do(srv, "GET", "/seq", "")
	if w2.Code != 201 {
		t.Fatalf("call 2: expected 201, got %d", w2.Code)
	}
	// Call 3: on_call=2 doesn't match, picks normal again
	w3 := do(srv, "GET", "/seq", "")
	if w3.Code != 200 {
		t.Fatalf("call 3: expected 200, got %d", w3.Code)
	}
}

// --- Response Scripting ---

func TestScript_basicJSON(t *testing.T) {
	srv := newSrv(&config.Config{
		Routes: []config.Route{{
			Path:   "/greet",
			Method: "POST",
			Script: `{"msg": "hello {{ .body.name }}"}`,
		}},
	})

	w := do(srv, "POST", "/greet", `{"name":"Alice"}`)
	body := jsonBody(t, w)
	if body["msg"] != "hello Alice" {
		t.Fatalf("expected 'hello Alice', got %v", body["msg"])
	}
}

func TestScript_helpers(t *testing.T) {
	srv := newSrv(&config.Config{
		Routes: []config.Route{{
			Path:   "/up",
			Method: "GET",
			Script: `{"val": "{{ upper "hello" }}"}`,
		}},
	})

	w := do(srv, "GET", "/up", "")
	body := jsonBody(t, w)
	if body["val"] != "HELLO" {
		t.Fatalf("expected HELLO, got %v", body["val"])
	}
}

func TestScript_methodAndPath(t *testing.T) {
	srv := newSrv(&config.Config{
		Routes: []config.Route{{
			Path:   "/info",
			Method: "GET",
			Script: `{"method":"{{ .method }}","path":"{{ .path }}"}`,
		}},
	})

	w := do(srv, "GET", "/info", "")
	body := jsonBody(t, w)
	if body["method"] != "GET" || body["path"] != "/info" {
		t.Fatalf("unexpected body: %v", body)
	}
}

func TestScript_inMatch(t *testing.T) {
	srv := newSrv(&config.Config{
		Routes: []config.Route{{
			Path:   "/echo",
			Method: "POST",
			Match: []config.RouteMatch{{
				Body:   map[string]any{"echo": true},
				Script: `{"echoed":"{{ .body.text }}"}`,
			}},
			Response: map[string]any{"echoed": "none"},
		}},
	})

	w := do(srv, "POST", "/echo", `{"echo":true,"text":"hi"}`)
	body := jsonBody(t, w)
	if body["echoed"] != "hi" {
		t.Fatalf("expected hi, got %v", body["echoed"])
	}

	w2 := do(srv, "POST", "/echo", `{"echo":false,"text":"hi"}`)
	body2 := jsonBody(t, w2)
	if body2["echoed"] != "none" {
		t.Fatalf("expected none, got %v", body2["echoed"])
	}
}

func TestScript_inResponses(t *testing.T) {
	srv := newSrv(&config.Config{
		Routes: []config.Route{{
			Path:   "/seq",
			Method: "GET",
			Responses: []config.RouteResponse{
				{Script: `{"n":1}`},
				{Script: `{"n":2}`},
			},
		}},
	})

	w1 := do(srv, "GET", "/seq", "")
	b1 := jsonBody(t, w1)
	w2 := do(srv, "GET", "/seq", "")
	b2 := jsonBody(t, w2)

	if b1["n"] != float64(1) || b2["n"] != float64(2) {
		t.Fatalf("expected n=1 then n=2, got %v %v", b1["n"], b2["n"])
	}
}

// --- match-level set_state / set_vars ---

func TestMatch_setStateInMatch(t *testing.T) {
	loggedIn := "logged_in"
	srv := newSrv(&config.Config{
		Routes: []config.Route{
			{
				Path:   "/login",
				Method: "POST",
				Match: []config.RouteMatch{
					{
						Body:     map[string]any{"user": "alice"},
						SetState: &loggedIn,
						Status:   200,
						Response: map[string]any{"ok": true},
					},
				},
				Status:   401,
				Response: map[string]any{"error": "bad credentials"},
			},
			{
				Path:   "/profile",
				Method: "GET",
				State:  "logged_in",
				Status: 200,
				Response: map[string]any{"name": "alice"},
			},
			{
				Path:   "/profile",
				Method: "GET",
				Status: 401,
				Response: map[string]any{"error": "unauthorized"},
			},
		},
	})

	// wrong credentials → 401, state unchanged
	w := do(srv, "POST", "/login", `{"user":"bob"}`)
	if w.Code != 401 {
		t.Fatalf("expected 401, got %d", w.Code)
	}
	w2 := do(srv, "GET", "/profile", "")
	if w2.Code != 401 {
		t.Fatalf("state should not be set, expected 401 got %d", w2.Code)
	}

	// correct credentials → 200, state set to logged_in
	w3 := do(srv, "POST", "/login", `{"user":"alice"}`)
	if w3.Code != 200 {
		t.Fatalf("expected 200, got %d", w3.Code)
	}
	w4 := do(srv, "GET", "/profile", "")
	if w4.Code != 200 {
		t.Fatalf("expected 200 after login, got %d", w4.Code)
	}
}

func TestMatch_setVarsInMatch(t *testing.T) {
	srv := newSrv(&config.Config{
		Routes: []config.Route{
			{
				Path:   "/role",
				Method: "POST",
				Match: []config.RouteMatch{
					{
						Body:    map[string]any{"role": "admin"},
						SetVars: map[string]string{"role": "admin", "level": "5"},
						Status:  200,
						Response: map[string]any{"ok": true},
					},
				},
				Status:   200,
				Response: map[string]any{"ok": false},
			},
		},
	})

	do(srv, "POST", "/role", `{"role":"admin"}`)

	w := do(srv, "GET", "/__specter/vars", "")
	var vars map[string]string
	json.NewDecoder(w.Body).Decode(&vars)
	if vars["role"] != "admin" || vars["level"] != "5" {
		t.Fatalf("expected role=admin level=5, got %v", vars)
	}
}

func TestMatch_setVarsMatchLevelOverridesRoute(t *testing.T) {
	route := "route_state"
	match := "match_state"
	srv := newSrv(&config.Config{
		Routes: []config.Route{
			{
				Path:   "/action",
				Method: "POST",
				Match: []config.RouteMatch{
					{
						Body:     map[string]any{"x": "1"},
						SetState: &match,
						Response: map[string]any{"from": "match"},
					},
				},
				SetState: &route,
				Response: map[string]any{"from": "default"},
			},
		},
	})

	// match hits → match-level set_state wins
	do(srv, "POST", "/action", `{"x":"1"}`)
	w := do(srv, "GET", "/__specter/state", "")
	var s map[string]string
	json.NewDecoder(w.Body).Decode(&s)
	if s["state"] != "match_state" {
		t.Fatalf("expected match_state, got %v", s["state"])
	}
}

// --- query / headers regex matching ---

func TestQueryRegex_match(t *testing.T) {
	srv := newSrv(&config.Config{
		Routes: []config.Route{
			{
				Path:   "/items",
				Method: "GET",
				Match: []config.RouteMatch{
					{
						Query:    map[string]string{"sort": "^(asc|desc)$"},
						Status:   200,
						Response: map[string]any{"sorted": true},
					},
				},
				Status:   400,
				Response: map[string]any{"sorted": false},
			},
		},
	})

	for _, q := range []string{"asc", "desc"} {
		w := do(srv, "GET", "/items?sort="+q, "")
		if w.Code != 200 {
			t.Fatalf("sort=%s: expected 200, got %d", q, w.Code)
		}
	}
	w := do(srv, "GET", "/items?sort=random", "")
	if w.Code != 400 {
		t.Fatalf("sort=random: expected 400, got %d", w.Code)
	}
}

func TestQueryRegex_exactFallback(t *testing.T) {
	srv := newSrv(&config.Config{
		Routes: []config.Route{
			{
				Path:   "/search",
				Method: "GET",
				Match: []config.RouteMatch{
					{
						Query:    map[string]string{"q": "hello"},
						Status:   200,
						Response: map[string]any{"hit": true},
					},
				},
				Status:   404,
				Response: map[string]any{"hit": false},
			},
		},
	})

	// "hello" as regex matches substring — "hello world" should match
	w := do(srv, "GET", "/search?q=hello+world", "")
	if w.Code != 200 {
		t.Fatalf("expected 200 (substring match), got %d", w.Code)
	}
}

func TestHeadersRegex_match(t *testing.T) {
	srv := newSrv(&config.Config{
		Routes: []config.Route{
			{
				Path:   "/api",
				Method: "GET",
				Match: []config.RouteMatch{
					{
						Headers:  map[string]string{"Authorization": "^Bearer .+"},
						Status:   200,
						Response: map[string]any{"auth": true},
					},
				},
				Status:   401,
				Response: map[string]any{"auth": false},
			},
		},
	})

	req := httptest.NewRequest("GET", "/api", nil)
	req.Header.Set("Authorization", "Bearer my-token-123")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	req2 := httptest.NewRequest("GET", "/api", nil)
	req2.Header.Set("Authorization", "Basic dXNlcjpwYXNz")
	w2 := httptest.NewRecorder()
	srv.ServeHTTP(w2, req2)
	if w2.Code != 401 {
		t.Fatalf("expected 401, got %d", w2.Code)
	}
}

// --- script _status envelope ---

func TestScript_statusEnvelope(t *testing.T) {
	srv := newSrv(&config.Config{
		Routes: []config.Route{{
			Path:   "/order",
			Method: "POST",
			Script: `
{{- if eq .body.type "premium" -}}
{"_status": 201, "tier": "premium"}
{{- else -}}
{"_status": 400, "error": "invalid type"}
{{- end -}}`,
		}},
	})

	w := do(srv, "POST", "/order", `{"type":"premium"}`)
	if w.Code != 201 {
		t.Fatalf("expected 201, got %d", w.Code)
	}
	b := jsonBody(t, w)
	if b["tier"] != "premium" {
		t.Fatalf("expected tier=premium, got %v", b)
	}
	if _, hasStatus := b["_status"]; hasStatus {
		t.Error("_status should be removed from response body")
	}

	w2 := do(srv, "POST", "/order", `{"type":"free"}`)
	if w2.Code != 400 {
		t.Fatalf("expected 400, got %d", w2.Code)
	}
}

func TestScript_statusEnvelopeInMatch(t *testing.T) {
	srv := newSrv(&config.Config{
		Routes: []config.Route{{
			Path:   "/items",
			Method: "POST",
			Match: []config.RouteMatch{{
				Body:   map[string]any{"dynamic": true},
				Script: `{"_status": 202, "queued": true}`,
			}},
			Status:   200,
			Response: map[string]any{"queued": false},
		}},
	})

	w := do(srv, "POST", "/items", `{"dynamic":true}`)
	if w.Code != 202 {
		t.Fatalf("expected 202, got %d", w.Code)
	}
	b := jsonBody(t, w)
	if b["queued"] != true {
		t.Fatalf("expected queued=true, got %v", b)
	}
}

func TestScript_statusEnvelopeInResponses(t *testing.T) {
	srv := newSrv(&config.Config{
		Routes: []config.Route{{
			Path:   "/seq",
			Method: "GET",
			Responses: []config.RouteResponse{
				{Script: `{"_status": 503, "error": "down"}`},
				{Script: `{"_status": 200, "ok": true}`},
			},
		}},
	})

	w1 := do(srv, "GET", "/seq", "")
	if w1.Code != 503 {
		t.Fatalf("call 1: expected 503, got %d", w1.Code)
	}
	w2 := do(srv, "GET", "/seq", "")
	if w2.Code != 200 {
		t.Fatalf("call 2: expected 200, got %d", w2.Code)
	}
}

// --- per-route conditional proxy ---

func TestRouteProxy_forwardsRequest(t *testing.T) {
	// Start a real backend HTTP server
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"from":"backend"}`))
	}))
	defer backend.Close()

	srv := newSrv(&config.Config{
		Routes: []config.Route{
			{Path: "/real", Method: "GET", Proxy: backend.URL},
			{Path: "/mock", Method: "GET", Status: 200, Response: map[string]any{"from": "mock"}},
		},
	})

	// proxied route → backend response
	w := do(srv, "GET", "/real", "")
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	b := jsonBody(t, w)
	if b["from"] != "backend" {
		t.Fatalf("expected from=backend, got %v", b)
	}

	// normal mock route → mock response
	w2 := do(srv, "GET", "/mock", "")
	b2 := jsonBody(t, w2)
	if b2["from"] != "mock" {
		t.Fatalf("expected from=mock, got %v", b2)
	}
}

func TestRouteProxy_statefulSwitch(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"real":true}`))
	}))
	defer backend.Close()

	realMode := "real"
	srv := newSrv(&config.Config{
		Routes: []config.Route{
			// When state=real, forward to backend
			{Path: "/data", Method: "GET", State: "real", Proxy: backend.URL},
			// Default: mock
			{Path: "/data", Method: "GET", Status: 200, Response: map[string]any{"real": false}},
		},
	})

	// default state → mock
	w := do(srv, "GET", "/data", "")
	b := jsonBody(t, w)
	if b["real"] != false {
		t.Fatalf("expected mock response, got %v", b)
	}

	// set state to real
	srv.state.Set(realMode)

	// state=real → backend
	w2 := do(srv, "GET", "/data", "")
	b2 := jsonBody(t, w2)
	if b2["real"] != true {
		t.Fatalf("expected real=true, got %v", b2)
	}
}

func TestOpenAPIValidation_strictMode_blocks(t *testing.T) {
	f, _ := os.CreateTemp(t.TempDir(), "*.yaml")
	f.WriteString(minimalOpenAPISpec)
	f.Close()

	srv := newSrv(&config.Config{
		OpenAPI:       f.Name(),
		OpenAPIStrict: true,
		Routes:        []config.Route{{Path: "/items", Method: "POST", Status: 201}},
	})

	// missing required "name" field → strict mode returns 400
	w := do(srv, "POST", "/items", `{}`)
	if w.Code != 400 {
		t.Fatalf("strict mode: expected 400, got %d", w.Code)
	}
	// no mock response served
	if w.Code == 201 {
		t.Error("strict mode should not serve mock on invalid request")
	}
}

func TestOpenAPIValidation_strictMode_valid(t *testing.T) {
	f, _ := os.CreateTemp(t.TempDir(), "*.yaml")
	f.WriteString(minimalOpenAPISpec)
	f.Close()

	srv := newSrv(&config.Config{
		OpenAPI:       f.Name(),
		OpenAPIStrict: true,
		Routes:        []config.Route{{Path: "/items", Method: "POST", Status: 201}},
	})

	// valid request → 201 mock served normally
	w := do(srv, "POST", "/items", `{"name":"widget"}`)
	if w.Code != 201 {
		t.Fatalf("strict mode: expected 201 for valid request, got %d", w.Code)
	}
}

func TestOpenAPIValidation_nonStrict_stillServesMock(t *testing.T) {
	f, _ := os.CreateTemp(t.TempDir(), "*.yaml")
	f.WriteString(minimalOpenAPISpec)
	f.Close()

	srv := newSrv(&config.Config{
		OpenAPI: f.Name(),
		Routes:  []config.Route{{Path: "/items", Method: "POST", Status: 201}},
	})

	// invalid request in non-strict mode → still returns 201 with warning header
	w := do(srv, "POST", "/items", `{}`)
	if w.Code != 201 {
		t.Fatalf("non-strict: expected 201, got %d", w.Code)
	}
	if w.Header().Get("X-Specter-Validation-Error") == "" {
		t.Error("non-strict: expected X-Specter-Validation-Error header")
	}
}
