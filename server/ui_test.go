package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRenderUIInjectsAPIAddress(t *testing.T) {
	got := renderUI("http://localhost:8080")
	if !strings.Contains(got, "http://localhost:8080") {
		t.Fatalf("expected API address in UI, got %q", got)
	}
	if !strings.Contains(got, "Reset All") {
		t.Fatalf("expected Reset All control in UI, got %q", got)
	}
}

func TestUIHandlerServesHTML(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	uiHandler("http://localhost:8080").ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if got := rec.Header().Get("Content-Type"); got != "text/html; charset=utf-8" {
		t.Fatalf("expected HTML content type, got %q", got)
	}
	if !strings.Contains(rec.Body.String(), "Auto Refresh On") {
		t.Fatalf("expected auto refresh toggle in HTML, got %q", rec.Body.String())
	}
}
