package record

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBuildConfig_JSON(t *testing.T) {
	body, _ := json.Marshal(map[string]any{"id": 1, "name": "Alice"})
	routes := []recordedRoute{
		{method: "GET", path: "/users", status: 200, contentType: "application/json", body: body},
	}
	cfg := buildConfig(routes)
	if len(cfg.Routes) != 1 {
		t.Fatalf("expected 1 route, got %d", len(cfg.Routes))
	}
	r := cfg.Routes[0]
	if r.Method != "GET" || r.Path != "/users" || r.Status != 200 {
		t.Errorf("unexpected route: %+v", r)
	}
	if r.ContentType != "" {
		t.Errorf("expected no content_type for JSON, got %s", r.ContentType)
	}
	m, ok := r.Response.(map[string]any)
	if !ok {
		t.Fatalf("expected map response, got %T", r.Response)
	}
	if m["name"] != "Alice" {
		t.Errorf("expected name:Alice, got %v", m["name"])
	}
}

func TestBuildConfig_PlainText(t *testing.T) {
	routes := []recordedRoute{
		{method: "GET", path: "/health", status: 200, contentType: "text/plain; charset=utf-8", body: []byte("ok")},
	}
	cfg := buildConfig(routes)
	r := cfg.Routes[0]
	if r.ContentType != "text/plain" {
		t.Errorf("expected text/plain, got %s", r.ContentType)
	}
	if r.Response != "ok" {
		t.Errorf("expected response 'ok', got %v", r.Response)
	}
}

func TestBuildConfig_EmptyBody(t *testing.T) {
	routes := []recordedRoute{
		{method: "DELETE", path: "/users/1", status: 204, body: nil},
	}
	cfg := buildConfig(routes)
	r := cfg.Routes[0]
	if r.Response != nil {
		t.Errorf("expected nil response for empty body, got %v", r.Response)
	}
}

func TestBuildConfig_InvalidJSON(t *testing.T) {
	routes := []recordedRoute{
		{method: "GET", path: "/bad", status: 200, contentType: "application/json", body: []byte("not json")},
	}
	cfg := buildConfig(routes)
	r := cfg.Routes[0]
	if r.Response != "not json" {
		t.Errorf("expected raw string fallback, got %v", r.Response)
	}
}

func TestResponseRecorder_CapturesStatusAndBody(t *testing.T) {
	w := httptest.NewRecorder()
	rec := &responseRecorder{ResponseWriter: w, status: http.StatusOK}

	rec.WriteHeader(http.StatusCreated)
	rec.Write([]byte("hello"))

	if rec.status != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.status)
	}
	if rec.buf.String() != "hello" {
		t.Errorf("expected body 'hello', got %s", rec.buf.String())
	}
	if w.Code != http.StatusCreated {
		t.Errorf("underlying writer: expected 201, got %d", w.Code)
	}
}
