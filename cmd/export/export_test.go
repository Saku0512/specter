package export

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestBuildConfig_deduplicates(t *testing.T) {
	entries := []requestEntry{
		{Method: "GET", Path: "/users"},
		{Method: "GET", Path: "/users"},
		{Method: "POST", Path: "/users"},
		{Method: "GET", Path: "/users/1"},
	}
	cfg := buildConfig(entries)
	if len(cfg.Routes) != 3 {
		t.Fatalf("expected 3 unique routes, got %d", len(cfg.Routes))
	}
}

func TestBuildConfig_sortsRoutes(t *testing.T) {
	entries := []requestEntry{
		{Method: "GET", Path: "/z"},
		{Method: "GET", Path: "/a"},
		{Method: "POST", Path: "/a"},
	}
	cfg := buildConfig(entries)
	if cfg.Routes[0].Path != "/a" || cfg.Routes[0].Method != "GET" {
		t.Errorf("expected /a GET first, got %s %s", cfg.Routes[0].Method, cfg.Routes[0].Path)
	}
	if cfg.Routes[1].Path != "/a" || cfg.Routes[1].Method != "POST" {
		t.Errorf("expected /a POST second, got %s %s", cfg.Routes[1].Method, cfg.Routes[1].Path)
	}
	if cfg.Routes[2].Path != "/z" {
		t.Errorf("expected /z last, got %s", cfg.Routes[2].Path)
	}
}

func TestBuildConfig_normalizesMethod(t *testing.T) {
	entries := []requestEntry{{Method: "get", Path: "/items"}}
	cfg := buildConfig(entries)
	if cfg.Routes[0].Method != "GET" {
		t.Errorf("expected method GET, got %s", cfg.Routes[0].Method)
	}
}

func TestFetchHistory_success(t *testing.T) {
	entries := []requestEntry{
		{Method: "GET", Path: "/foo"},
		{Method: "POST", Path: "/bar"},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/__specter/requests" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(entries)
	}))
	defer srv.Close()

	got, err := fetchHistory(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(got))
	}
}

func TestFetchHistory_serverError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal", 500)
	}))
	defer srv.Close()

	_, err := fetchHistory(srv.URL)
	if err == nil {
		t.Error("expected error for 500 response")
	}
}

func TestRun_writesFile(t *testing.T) {
	entries := []requestEntry{{Method: "GET", Path: "/hello"}}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(entries)
	}))
	defer srv.Close()

	out := filepath.Join(t.TempDir(), "out.yml")
	Run([]string{"--from", srv.URL, "-o", out})

	if _, err := os.Stat(out); err != nil {
		t.Fatalf("output file not created: %v", err)
	}
}
