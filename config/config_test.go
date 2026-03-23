package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTemp(t *testing.T, name, content string) string {
	t.Helper()
	f := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(f, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return f
}

func TestLoad_yaml(t *testing.T) {
	path := writeTemp(t, "config.yaml", `
routes:
  - path: /users
    method: GET
    status: 200
    response:
      id: 1
`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Routes) != 1 {
		t.Fatalf("expected 1 route, got %d", len(cfg.Routes))
	}
	if cfg.Routes[0].Path != "/users" {
		t.Errorf("expected path /users, got %s", cfg.Routes[0].Path)
	}
}

func TestLoad_yml_extension(t *testing.T) {
	path := writeTemp(t, "config.yml", `
routes:
  - path: /ping
    method: GET
    status: 200
    response: ok
`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Routes) != 1 {
		t.Fatalf("expected 1 route, got %d", len(cfg.Routes))
	}
}

func TestLoad_fallback_to_yml(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig)
	os.Chdir(dir)

	os.WriteFile("config.yml", []byte("routes:\n  - path: /a\n    method: GET\n"), 0644)

	// request "config.yaml" (no such file) → should fall back to "config.yml"
	cfg, err := Load("config.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Routes) != 1 {
		t.Fatalf("expected 1 route, got %d", len(cfg.Routes))
	}
}

func TestLoad_not_found(t *testing.T) {
	_, err := Load("/nonexistent/config.yaml")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestLoad_cors(t *testing.T) {
	path := writeTemp(t, "config.yaml", `
cors: true
routes: []
`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.CORS {
		t.Error("expected CORS to be true")
	}
}

func TestLoad_include_merges_routes(t *testing.T) {
	dir := t.TempDir()
	// extra.yml defines one route
	os.WriteFile(filepath.Join(dir, "extra.yml"), []byte("routes:\n  - path: /extra\n    method: GET\n"), 0644)
	// main config includes it
	main := writeTemp(t, "config.yaml", "include:\n  - "+filepath.Join(dir, "extra.yml")+"\nroutes:\n  - path: /main\n    method: GET\n")
	cfg, err := Load(main)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Routes) != 2 {
		t.Fatalf("expected 2 routes, got %d", len(cfg.Routes))
	}
	paths := map[string]bool{cfg.Routes[0].Path: true, cfg.Routes[1].Path: true}
	if !paths["/main"] || !paths["/extra"] {
		t.Errorf("unexpected routes: %v", cfg.Routes)
	}
}

func TestLoad_include_glob(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.yml"), []byte("routes:\n  - path: /a\n    method: GET\n"), 0644)
	os.WriteFile(filepath.Join(dir, "b.yml"), []byte("routes:\n  - path: /b\n    method: GET\n"), 0644)
	main := writeTemp(t, "config.yaml", "include:\n  - "+filepath.Join(dir, "*.yml")+"\nroutes: []\n")
	cfg, err := Load(main)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Routes) != 2 {
		t.Fatalf("expected 2 routes from glob, got %d", len(cfg.Routes))
	}
}

func TestLoad_include_cycle_detection(t *testing.T) {
	dir := t.TempDir()
	aPath := filepath.Join(dir, "a.yml")
	bPath := filepath.Join(dir, "b.yml")
	// a includes b, b includes a
	os.WriteFile(aPath, []byte("include:\n  - "+bPath+"\nroutes:\n  - path: /a\n    method: GET\n"), 0644)
	os.WriteFile(bPath, []byte("include:\n  - "+aPath+"\nroutes:\n  - path: /b\n    method: GET\n"), 0644)
	cfg, err := Load(aPath)
	if err != nil {
		t.Fatal(err)
	}
	// cycle is silently broken; should get 2 unique routes
	if len(cfg.Routes) != 2 {
		t.Fatalf("expected 2 routes (cycle broken), got %d: %+v", len(cfg.Routes), cfg.Routes)
	}
}

func TestLoad_all_route_fields(t *testing.T) {
	path := writeTemp(t, "config.yaml", `
routes:
  - path: /test
    method: POST
    status: 201
    delay: 100
    headers:
      X-Foo: bar
    response:
      ok: true
    mode: random
    responses:
      - status: 200
        response: a
      - status: 500
        response: b
    match:
      - query:
          q: x
        body:
          role: admin
        status: 200
        response: matched
`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	r := cfg.Routes[0]
	if r.Delay != 100 {
		t.Errorf("expected delay 100, got %d", r.Delay)
	}
	if r.Headers["X-Foo"] != "bar" {
		t.Errorf("expected header X-Foo: bar")
	}
	if r.Mode != "random" {
		t.Errorf("expected mode random, got %s", r.Mode)
	}
	if len(r.Responses) != 2 {
		t.Errorf("expected 2 responses, got %d", len(r.Responses))
	}
	if len(r.Match) != 1 {
		t.Errorf("expected 1 match, got %d", len(r.Match))
	}
	if r.Match[0].Query["q"] != "x" {
		t.Errorf("expected match query q=x")
	}
}
