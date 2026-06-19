package examples

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/Saku0512/specter/cmd/validate"
	"github.com/Saku0512/specter/config"
)

func TestExampleNames(t *testing.T) {
	got := exampleNames()
	want := []string{"auth", "crud", "errors", "graphql", "openapi", "pagination", "sse", "webhooks"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected examples %v, got %v", want, got)
	}
}

func TestExampleForIsCaseInsensitive(t *testing.T) {
	example, ok := exampleFor(" GraphQL ")
	if !ok {
		t.Fatal("expected GraphQL example to resolve")
	}
	if example.Name != "graphql" {
		t.Fatalf("expected graphql, got %q", example.Name)
	}
}

func TestExamplesLoadAndValidate(t *testing.T) {
	for _, name := range exampleNames() {
		t.Run(name, func(t *testing.T) {
			example, ok := exampleFor(name)
			if !ok {
				t.Fatalf("missing example %q", name)
			}
			cfg, err := config.LoadBytes([]byte(example.Files["config.yml"]))
			if err != nil {
				t.Fatalf("example %q did not load: %v", name, err)
			}
			if len(cfg.Routes) == 0 {
				t.Fatalf("example %q should include routes", name)
			}
			if errs := validate.CheckNoFilesystem(cfg); len(errs) != 0 {
				t.Fatalf("example %q has validation errors: %v", name, errs)
			}
		})
	}
}

func TestWriteExampleWritesConfig(t *testing.T) {
	out := filepath.Join(t.TempDir(), "sample.yml")
	written, err := writeExample("auth", out, false)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(written, []string{out}) {
		t.Fatalf("expected only %s to be written, got %v", out, written)
	}
	cfg, err := config.Load(out)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := cfg.Scenarios["logged-in"]; !ok {
		t.Fatalf("expected auth scenario, got %v", cfg.Scenarios)
	}
}

func TestWriteExampleWritesOpenAPISidecar(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "config.yml")
	written, err := writeExample("openapi", out, false)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{out, filepath.Join(dir, "openapi.yml")}
	if !reflect.DeepEqual(written, want) {
		t.Fatalf("expected files %v, got %v", want, written)
	}
	if _, err := os.Stat(filepath.Join(dir, "openapi.yml")); err != nil {
		t.Fatal(err)
	}
}

func TestWriteExampleRejectsOverwriteWithoutForce(t *testing.T) {
	out := filepath.Join(t.TempDir(), "config.yml")
	if err := os.WriteFile(out, []byte("original"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := writeExample("crud", out, false)
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

func TestWriteExampleForceOverwrites(t *testing.T) {
	out := filepath.Join(t.TempDir(), "config.yml")
	if err := os.WriteFile(out, []byte("original"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := writeExample("crud", out, true); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "store_list: todos") {
		t.Fatalf("expected CRUD config, got %s", string(data))
	}
}

func TestWriteExampleRejectsUnknownExample(t *testing.T) {
	_, err := writeExample("missing", filepath.Join(t.TempDir(), "config.yml"), false)
	if err == nil {
		t.Fatal("expected unknown example error")
	}
	if !strings.Contains(err.Error(), "Available examples") {
		t.Fatalf("expected available examples in error, got %v", err)
	}
}
