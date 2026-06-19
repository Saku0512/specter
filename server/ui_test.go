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

func TestRenderUIIncludesDynamicRouteEditor(t *testing.T) {
	got := renderUI("http://localhost:8080")
	for _, want := range []string{
		"Dynamic Route Editor",
		`id="route-json"`,
		"onclick=\"newRoute()\"",
		"onclick=\"saveRoute()\"",
		"onclick=\"editSelectedRoute()\"",
		"Route JSON must be valid JSON.",
		"Route JSON must include a string path.",
		"Route JSON must include a string method.",
		"Config routes cannot be edited in memory; saving creates a dynamic copy.",
		"POST",
		"PUT",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected UI to contain %q", want)
		}
	}
}

func TestRenderUIIncludesConfigPlayground(t *testing.T) {
	got := renderUI("http://localhost:8080")
	for _, want := range []string{
		"Config Playground",
		`id="config-yaml"`,
		"onclick=\"validateConfigPlayground()\"",
		"/__specter/config/validate",
		"Registered Routes",
		"Seeded Stores",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected UI to contain %q", want)
		}
	}
}

func TestRenderUIIncludesTimelinePanel(t *testing.T) {
	got := renderUI("http://localhost:8080")
	for _, want := range []string{
		"Timelines",
		`id="timelines-body"`,
		"/__specter/timelines",
		"resetTimeline",
		"Reset Timelines",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected UI to contain %q", want)
		}
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
