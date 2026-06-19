package init_cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Saku0512/specter/config"
)

func TestTemplateNames(t *testing.T) {
	got := templateNames()
	want := []string{"auth", "basic", "crud", "openapi"}
	if len(got) != len(want) {
		t.Fatalf("expected %d templates, got %d: %v", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("expected templates %v, got %v", want, got)
		}
	}
}

func TestTemplateFor_isCaseInsensitive(t *testing.T) {
	tmpl, ok := templateFor(" CRUD ")
	if !ok {
		t.Fatal("expected CRUD template to resolve")
	}
	if tmpl != crudTemplate {
		t.Fatal("expected CRUD template")
	}
}

func TestTemplatesLoadAsConfig(t *testing.T) {
	for _, name := range templateNames() {
		t.Run(name, func(t *testing.T) {
			tmpl, ok := templateFor(name)
			if !ok {
				t.Fatalf("missing template %q", name)
			}
			path := filepath.Join(t.TempDir(), "config.yml")
			if err := os.WriteFile(path, []byte(tmpl), 0644); err != nil {
				t.Fatal(err)
			}
			cfg, err := config.Load(path)
			if err != nil {
				t.Fatalf("template %q did not load: %v", name, err)
			}
			if len(cfg.Routes) == 0 {
				t.Fatalf("template %q should include routes", name)
			}
		})
	}
}

func TestRun_writesSelectedTemplate(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "crud.yml")
	Run([]string{"--template", "crud", "-o", out})

	cfg, err := config.Load(out)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := cfg.Scenarios["seeded"]; !ok {
		t.Fatalf("expected crud template to include seeded scenario, got %v", cfg.Scenarios)
	}
	if len(cfg.Routes) != 6 {
		t.Fatalf("expected 6 CRUD routes, got %d", len(cfg.Routes))
	}
}
