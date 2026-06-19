package doctor

import (
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDiagnoseValidConfig(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "fixture.json"), `{"ok":true}`)
	configPath := filepath.Join(dir, "config.yml")
	writeFile(t, configPath, `
routes:
  - path: /ping
    method: GET
    file: fixture.json
`)

	diagnostics := Diagnose(Options{ConfigPath: configPath, Port: freePort(t), UIPort: "0"})

	if HasErrors(diagnostics) {
		t.Fatalf("expected no errors, got %#v", diagnostics)
	}
	assertDiagnostic(t, diagnostics, LevelOK, "files", "all referenced files exist")
}

func TestDiagnoseMissingConfig(t *testing.T) {
	diagnostics := Diagnose(Options{ConfigPath: filepath.Join(t.TempDir(), "missing.yml"), Port: freePort(t), UIPort: "0"})

	if !HasErrors(diagnostics) {
		t.Fatalf("expected errors, got %#v", diagnostics)
	}
	assertDiagnostic(t, diagnostics, LevelError, "config path", "was not found")
}

func TestDiagnoseValidationErrors(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.yml")
	writeFile(t, configPath, `
routes:
  - path: /bad
    method: WAT
`)

	diagnostics := Diagnose(Options{ConfigPath: configPath, Port: freePort(t), UIPort: "0"})

	assertDiagnostic(t, diagnostics, LevelError, "config validation", "invalid method")
}

func TestDiagnoseWarnsForUnmatchedInclude(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.yml")
	writeFile(t, configPath, `
include:
  - routes/*.yml
routes: []
`)

	diagnostics := Diagnose(Options{ConfigPath: configPath, Port: freePort(t), UIPort: "0"})

	assertDiagnostic(t, diagnostics, LevelWarn, "include", "matched no files")
}

func TestDiagnoseChecksIncludedRouteFilesRelativeToIncludedConfig(t *testing.T) {
	dir := t.TempDir()
	routesDir := filepath.Join(dir, "routes")
	if err := os.Mkdir(routesDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(routesDir, "fixture.json"), `{"ok":true}`)
	writeFile(t, filepath.Join(routesDir, "extra.yml"), `
routes:
  - path: /included
    method: GET
    file: fixture.json
`)
	configPath := filepath.Join(dir, "config.yml")
	writeFile(t, configPath, `
include:
  - routes/*.yml
routes: []
`)

	diagnostics := Diagnose(Options{ConfigPath: configPath, Port: freePort(t), UIPort: "0"})

	if HasErrors(diagnostics) {
		t.Fatalf("expected included fixture path to resolve from included file, got %#v", diagnostics)
	}
	assertDiagnostic(t, diagnostics, LevelOK, "files", "all referenced files exist")
}

func TestDiagnoseReportsMissingRouteFiles(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.yml")
	writeFile(t, configPath, `
routes:
  - path: /missing
    method: GET
    responses:
      - file: missing.json
`)

	diagnostics := Diagnose(Options{ConfigPath: configPath, Port: freePort(t), UIPort: "0"})

	assertDiagnostic(t, diagnostics, LevelError, "files", `responses[0] file "missing.json" not found`)
}

func TestDiagnoseChecksOpenAPIRelativePath(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "openapi.yml"), `
openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths: {}
`)
	configPath := filepath.Join(dir, "config.yml")
	writeFile(t, configPath, `
openapi: openapi.yml
routes: []
`)

	diagnostics := Diagnose(Options{ConfigPath: configPath, Port: freePort(t), UIPort: "0"})

	assertDiagnostic(t, diagnostics, LevelOK, "openapi", "openapi.yml is valid")
}

func TestDiagnoseReportsMissingOpenAPI(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.yml")
	writeFile(t, configPath, `
openapi: openapi.yml
routes: []
`)

	diagnostics := Diagnose(Options{ConfigPath: configPath, Port: freePort(t), UIPort: "0"})

	assertDiagnostic(t, diagnostics, LevelError, "openapi", `openapi "openapi.yml" not found`)
}

func TestDiagnoseWarnsForDuplicateRoutes(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.yml")
	writeFile(t, configPath, `
routes:
  - path: /users
    method: GET
  - path: /users
    method: get
    priority: 1
`)

	diagnostics := Diagnose(Options{ConfigPath: configPath, Port: freePort(t), UIPort: "0"})

	assertDiagnostic(t, diagnostics, LevelWarn, "routes", "GET /users is defined 2 times")
}

func TestDiagnoseReportsPortConflict(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	_, port, err := net.SplitHostPort(ln.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(t.TempDir(), "config.yml")
	writeFile(t, configPath, "routes: []\n")

	diagnostics := Diagnose(Options{ConfigPath: configPath, Port: port, UIPort: "0"})

	assertDiagnostic(t, diagnostics, LevelError, "server port", "unavailable")
}

func TestDiagnoseReportsSameServerAndUIPort(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.yml")
	writeFile(t, configPath, "routes: []\n")
	port := freePort(t)

	diagnostics := Diagnose(Options{ConfigPath: configPath, Port: port, UIPort: port})

	assertDiagnostic(t, diagnostics, LevelError, "ui port", "both use")
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func freePort(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	_, port, err := net.SplitHostPort(ln.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	return port
}

func assertDiagnostic(t *testing.T, diagnostics []Diagnostic, level Level, check, substring string) {
	t.Helper()
	for _, diagnostic := range diagnostics {
		if diagnostic.Level == level && diagnostic.Check == check && strings.Contains(diagnostic.Message, substring) {
			return
		}
	}
	t.Fatalf("missing %s %s containing %q in %#v", level, check, substring, diagnostics)
}
