package init_cmd

import (
	"os"
	"path/filepath"
	"strings"
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

func TestWriteConfig_rejectsUnknownTemplate(t *testing.T) {
	out := filepath.Join(t.TempDir(), "config.yml")
	_, err := writeConfig(out, "missing", false)
	if err == nil {
		t.Fatal("expected error for unknown template")
	}
	if !strings.Contains(err.Error(), "Available templates") {
		t.Fatalf("expected available templates in error, got %v", err)
	}
}

func TestWriteConfig_doesNotOverwriteWithoutForce(t *testing.T) {
	out := filepath.Join(t.TempDir(), "config.yml")
	if err := os.WriteFile(out, []byte("original"), 0644); err != nil {
		t.Fatal(err)
	}
	_, err := writeConfig(out, "basic", false)
	if err == nil {
		t.Fatal("expected overwrite error")
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "original" {
		t.Fatalf("expected original file to be preserved, got %q", string(data))
	}
}

func TestWriteConfig_forceOverwrites(t *testing.T) {
	out := filepath.Join(t.TempDir(), "config.yml")
	if err := os.WriteFile(out, []byte("original"), 0644); err != nil {
		t.Fatal(err)
	}
	name, err := writeConfig(out, "AUTH", true)
	if err != nil {
		t.Fatal(err)
	}
	if name != "auth" {
		t.Fatalf("expected normalized template name auth, got %q", name)
	}
	cfg, err := config.Load(out)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := cfg.Scenarios["logged-in-admin"]; !ok {
		t.Fatalf("expected auth template scenario, got %v", cfg.Scenarios)
	}
}
