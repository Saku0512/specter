package doctor

import (
	"flag"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/Saku0512/specter/cmd/validate"
	"github.com/Saku0512/specter/config"
	"github.com/getkin/kin-openapi/openapi3"
	"gopkg.in/yaml.v3"
)

type Level string

const (
	LevelOK    Level = "ok"
	LevelWarn  Level = "warn"
	LevelError Level = "error"
)

type Diagnostic struct {
	Level   Level
	Check   string
	Message string
}

type Options struct {
	ConfigPath string
	Host       string
	Port       string
	UIPort     string
}

type fileConfig struct {
	OpenAPI string         `yaml:"openapi,omitempty"`
	Include []string       `yaml:"include,omitempty"`
	Routes  []config.Route `yaml:"routes"`
}

func Run(args []string) {
	fs := flag.NewFlagSet("doctor", flag.ExitOnError)
	configPath := fs.String("c", "config.yaml", "path to config file")
	port := fs.String("p", "8080", "port to check")
	host := fs.String("host", "", "host to check")
	uiPort := fs.String("ui-port", "4444", "port for the web UI (0 to disable)")
	fs.Parse(args)

	set := make(map[string]bool)
	fs.Visit(func(f *flag.Flag) { set[f.Name] = true })
	if !set["c"] {
		if val := os.Getenv("SPECTER_CONFIG"); val != "" {
			*configPath = val
		}
	}
	if !set["p"] {
		if val := os.Getenv("SPECTER_PORT"); val != "" {
			*port = val
		}
	}
	if !set["host"] {
		if val := os.Getenv("SPECTER_HOST"); val != "" {
			*host = val
		}
	}
	if !set["ui-port"] {
		if val := os.Getenv("SPECTER_UI_PORT"); val != "" {
			*uiPort = val
		}
	}

	diagnostics := Diagnose(Options{
		ConfigPath: *configPath,
		Host:       *host,
		Port:       *port,
		UIPort:     *uiPort,
	})
	Print(diagnostics)
	if HasErrors(diagnostics) {
		os.Exit(1)
	}
}

func Diagnose(opts Options) []Diagnostic {
	opts = normalizeOptions(opts)
	var diagnostics []Diagnostic

	resolved, tried, err := resolveConfigPath(opts.ConfigPath)
	if err != nil {
		return []Diagnostic{{
			Level:   LevelError,
			Check:   "config path",
			Message: fmt.Sprintf("config %q was not found (tried: %s)", opts.ConfigPath, strings.Join(tried, ", ")),
		}}
	}
	diagnostics = append(diagnostics, Diagnostic{Level: LevelOK, Check: "config path", Message: fmt.Sprintf("found %s", resolved)})

	cfg, err := config.Load(resolved)
	if err != nil {
		diagnostics = append(diagnostics, Diagnostic{Level: LevelError, Check: "config load", Message: err.Error()})
		return diagnostics
	}
	diagnostics = append(diagnostics, Diagnostic{Level: LevelOK, Check: "config load", Message: fmt.Sprintf("loaded %d route(s)", len(cfg.Routes))})

	if errs := validate.CheckNoFilesystem(cfg); len(errs) > 0 {
		for _, err := range errs {
			diagnostics = append(diagnostics, Diagnostic{Level: LevelError, Check: "config validation", Message: err})
		}
	} else {
		diagnostics = append(diagnostics, Diagnostic{Level: LevelOK, Check: "config validation", Message: "no semantic errors"})
	}

	diagnostics = append(diagnostics, inspectConfigTree(resolved)...)
	diagnostics = append(diagnostics, checkRouteConflicts(cfg)...)
	diagnostics = append(diagnostics, checkPorts(opts)...)
	return diagnostics
}

func normalizeOptions(opts Options) Options {
	if opts.ConfigPath == "" {
		opts.ConfigPath = "config.yaml"
	}
	if opts.Port == "" {
		opts.Port = "8080"
	}
	if opts.UIPort == "" {
		opts.UIPort = "4444"
	}
	return opts
}

func resolveConfigPath(path string) (string, []string, error) {
	candidates := []string{path}
	if path == "config.yaml" {
		candidates = append(candidates, "config.yml")
	} else if path == "config.yml" {
		candidates = append(candidates, "config.yaml")
	}
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate, candidates, nil
		} else if !os.IsNotExist(err) {
			return "", candidates, err
		}
	}
	return "", candidates, os.ErrNotExist
}

func inspectConfigTree(root string) []Diagnostic {
	var diagnostics []Diagnostic
	seen := make(map[string]bool)
	inspectConfigFile(root, true, seen, &diagnostics)
	if !hasLevelForCheck(diagnostics, LevelError, "include") && !hasLevelForCheck(diagnostics, LevelWarn, "include") {
		diagnostics = append(diagnostics, Diagnostic{Level: LevelOK, Check: "include", Message: "all include patterns resolved"})
	}
	if !hasLevelForCheck(diagnostics, LevelError, "files") {
		diagnostics = append(diagnostics, Diagnostic{Level: LevelOK, Check: "files", Message: "all referenced files exist"})
	}
	if !hasLevelForCheck(diagnostics, LevelError, "openapi") && !hasLevelForCheck(diagnostics, LevelWarn, "openapi") {
		diagnostics = append(diagnostics, Diagnostic{Level: LevelOK, Check: "openapi", Message: "OpenAPI spec is not configured"})
	}
	return diagnostics
}

func inspectConfigFile(path string, isRoot bool, seen map[string]bool, diagnostics *[]Diagnostic) {
	abs, _ := filepath.Abs(path)
	if seen[abs] {
		return
	}
	seen[abs] = true

	data, err := os.ReadFile(path)
	if err != nil {
		*diagnostics = append(*diagnostics, Diagnostic{Level: LevelError, Check: "include", Message: fmt.Sprintf("%s: %v", path, err)})
		return
	}

	var fc fileConfig
	if err := yaml.Unmarshal(data, &fc); err != nil {
		*diagnostics = append(*diagnostics, Diagnostic{Level: LevelError, Check: "config load", Message: fmt.Sprintf("%s: %v", path, err)})
		return
	}

	dir := filepath.Dir(path)
	if isRoot && fc.OpenAPI != "" {
		checkOpenAPI(path, dir, fc.OpenAPI, diagnostics)
	}
	checkRouteFiles(path, dir, fc.Routes, diagnostics)
	for _, pattern := range fc.Include {
		resolvedPattern := resolveRelative(dir, pattern)
		matches, err := filepath.Glob(resolvedPattern)
		if err != nil {
			*diagnostics = append(*diagnostics, Diagnostic{Level: LevelError, Check: "include", Message: fmt.Sprintf("%s: invalid include pattern %q: %v", path, pattern, err)})
			continue
		}
		if len(matches) == 0 {
			*diagnostics = append(*diagnostics, Diagnostic{Level: LevelWarn, Check: "include", Message: fmt.Sprintf("%s: include pattern %q matched no files", path, pattern)})
			continue
		}
		sort.Strings(matches)
		for _, match := range matches {
			inspectConfigFile(match, false, seen, diagnostics)
		}
	}
}

func checkOpenAPI(source, dir, spec string, diagnostics *[]Diagnostic) {
	path := resolveRelative(dir, spec)
	if _, err := os.Stat(path); err != nil {
		*diagnostics = append(*diagnostics, Diagnostic{Level: LevelError, Check: "openapi", Message: fmt.Sprintf("%s: openapi %q not found", source, spec)})
		return
	}
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromFile(path)
	if err != nil {
		*diagnostics = append(*diagnostics, Diagnostic{Level: LevelError, Check: "openapi", Message: fmt.Sprintf("%s: failed to load %q: %v", source, spec, err)})
		return
	}
	if err := doc.Validate(loader.Context); err != nil {
		*diagnostics = append(*diagnostics, Diagnostic{Level: LevelError, Check: "openapi", Message: fmt.Sprintf("%s: invalid OpenAPI spec %q: %v", source, spec, err)})
		return
	}
	*diagnostics = append(*diagnostics, Diagnostic{Level: LevelOK, Check: "openapi", Message: fmt.Sprintf("%s: %s is valid", source, spec)})
}

func checkRouteFiles(source, dir string, routes []config.Route, diagnostics *[]Diagnostic) {
	for i, route := range routes {
		prefix := routePrefix(i, route)
		checkFileReference(source, dir, route.File, prefix+": file", diagnostics)
		for j, response := range route.Responses {
			checkFileReference(source, dir, response.File, fmt.Sprintf("%s: responses[%d] file", prefix, j), diagnostics)
		}
		for j, step := range route.Timeline {
			checkFileReference(source, dir, step.File, fmt.Sprintf("%s: timeline[%d] file", prefix, j), diagnostics)
		}
		for j, match := range route.Match {
			checkFileReference(source, dir, match.File, fmt.Sprintf("%s: match[%d] file", prefix, j), diagnostics)
		}
	}
}

func checkFileReference(source, dir, ref, label string, diagnostics *[]Diagnostic) {
	if ref == "" {
		return
	}
	if _, err := os.Stat(resolveRelative(dir, ref)); err != nil {
		*diagnostics = append(*diagnostics, Diagnostic{Level: LevelError, Check: "files", Message: fmt.Sprintf("%s: %s %q not found", source, label, ref)})
	}
}

func resolveRelative(dir, path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(dir, path)
}

func routePrefix(i int, route config.Route) string {
	if route.Path != "" && route.Method != "" {
		return fmt.Sprintf("route %d (%s %s)", i+1, route.Method, route.Path)
	}
	return fmt.Sprintf("route %d", i+1)
}

func checkRouteConflicts(cfg *config.Config) []Diagnostic {
	type routeRef struct {
		index    int
		priority int
	}
	groups := make(map[string][]routeRef)
	for i, route := range cfg.Routes {
		if route.Path == "" || route.Method == "" {
			continue
		}
		key := strings.ToUpper(route.Method) + " " + route.Path
		groups[key] = append(groups[key], routeRef{index: i + 1, priority: route.Priority})
	}

	var diagnostics []Diagnostic
	keys := make([]string, 0, len(groups))
	for key, refs := range groups {
		if len(refs) > 1 {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	for _, key := range keys {
		refs := groups[key]
		parts := make([]string, 0, len(refs))
		for _, ref := range refs {
			parts = append(parts, fmt.Sprintf("#%d(priority=%d)", ref.index, ref.priority))
		}
		diagnostics = append(diagnostics, Diagnostic{
			Level:   LevelWarn,
			Check:   "routes",
			Message: fmt.Sprintf("%s is defined %d times (%s); higher priority or earlier routes may shadow later ones", key, len(refs), strings.Join(parts, ", ")),
		})
	}
	if len(diagnostics) == 0 {
		diagnostics = append(diagnostics, Diagnostic{Level: LevelOK, Check: "routes", Message: "no duplicate method/path routes"})
	}
	return diagnostics
}

func checkPorts(opts Options) []Diagnostic {
	var diagnostics []Diagnostic
	checkPort(opts.Host, opts.Port, "server port", &diagnostics)
	if opts.UIPort == "0" {
		diagnostics = append(diagnostics, Diagnostic{Level: LevelOK, Check: "ui port", Message: "disabled"})
		return diagnostics
	}
	if opts.Port == opts.UIPort {
		diagnostics = append(diagnostics, Diagnostic{Level: LevelError, Check: "ui port", Message: fmt.Sprintf("server port and UI port both use %s", opts.Port)})
		return diagnostics
	}
	checkPort(opts.Host, opts.UIPort, "ui port", &diagnostics)
	return diagnostics
}

func checkPort(host, port, check string, diagnostics *[]Diagnostic) bool {
	if _, err := strconv.Atoi(port); err != nil {
		*diagnostics = append(*diagnostics, Diagnostic{Level: LevelError, Check: check, Message: fmt.Sprintf("invalid port %q", port)})
		return false
	}
	listenHost := host
	if listenHost == "" {
		listenHost = "127.0.0.1"
	}
	ln, err := net.Listen("tcp", net.JoinHostPort(listenHost, port))
	if err != nil {
		*diagnostics = append(*diagnostics, Diagnostic{Level: LevelError, Check: check, Message: fmt.Sprintf("%s unavailable on %s: %v", port, listenHost, err)})
		return false
	}
	_ = ln.Close()
	*diagnostics = append(*diagnostics, Diagnostic{Level: LevelOK, Check: check, Message: fmt.Sprintf("%s available on %s", port, listenHost)})
	return true
}

func hasLevelForCheck(diagnostics []Diagnostic, level Level, check string) bool {
	for _, diagnostic := range diagnostics {
		if diagnostic.Level == level && diagnostic.Check == check {
			return true
		}
	}
	return false
}

func HasErrors(diagnostics []Diagnostic) bool {
	for _, diagnostic := range diagnostics {
		if diagnostic.Level == LevelError {
			return true
		}
	}
	return false
}

func Print(diagnostics []Diagnostic) {
	fmt.Println("👻 specter doctor")
	errors := 0
	warnings := 0
	for _, diagnostic := range diagnostics {
		switch diagnostic.Level {
		case LevelError:
			errors++
			fmt.Printf("✗ %-18s %s\n", diagnostic.Check, diagnostic.Message)
		case LevelWarn:
			warnings++
			fmt.Printf("! %-18s %s\n", diagnostic.Check, diagnostic.Message)
		default:
			fmt.Printf("✓ %-18s %s\n", diagnostic.Check, diagnostic.Message)
		}
	}
	fmt.Printf("\n%d error(s), %d warning(s)\n", errors, warnings)
}
