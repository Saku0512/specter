// Package export provides the `specter export` subcommand, which reads the
// request history of a running specter instance and generates a starter
// config.yml from the observed (method, path) pairs.
package export

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/Saku0512/specter/config"
	"gopkg.in/yaml.v3"
)

// requestEntry mirrors the JSON shape returned by GET /__specter/requests.
type requestEntry struct {
	Method string `json:"method"`
	Path   string `json:"path"`
}

// Run implements the export subcommand.
func Run(args []string) {
	fs := flag.NewFlagSet("export", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: specter export [flags]\n\n")
		fmt.Fprintf(os.Stderr, "Reads the request history from a running specter instance and writes a\n")
		fmt.Fprintf(os.Stderr, "starter config.yml for the observed routes.\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		fs.PrintDefaults()
	}
	from := fs.String("from", "http://localhost:8080", "specter base URL to read history from")
	output := fs.String("o", "exported.yml", "output config file")
	force := fs.Bool("f", false, "overwrite output file if it exists")
	fs.Parse(args)

	if !*force {
		if _, err := os.Stat(*output); err == nil {
			fmt.Fprintf(os.Stderr, "error: %s already exists, use -f to overwrite\n", *output)
			os.Exit(1)
		}
	}

	entries, err := fetchHistory(*from)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if len(entries) == 0 {
		fmt.Println("no requests recorded — nothing to export")
		return
	}

	cfg := buildConfig(entries)
	data, err := yaml.Marshal(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to marshal config: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(*output, data, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ exported %d route(s) → %s\n", len(cfg.Routes), *output)
}

func fetchHistory(baseURL string) ([]requestEntry, error) {
	baseURL = strings.TrimRight(baseURL, "/")
	resp, err := http.Get(baseURL + "/__specter/requests")
	if err != nil {
		return nil, fmt.Errorf("GET %s/__specter/requests: %w", baseURL, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d from %s/__specter/requests", resp.StatusCode, baseURL)
	}
	var entries []requestEntry
	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return entries, nil
}

func buildConfig(entries []requestEntry) config.Config {
	seen := map[string]bool{}
	var routes []config.Route
	for _, e := range entries {
		key := strings.ToUpper(e.Method) + " " + e.Path
		if seen[key] {
			continue
		}
		seen[key] = true
		routes = append(routes, config.Route{
			Path:   e.Path,
			Method: strings.ToUpper(e.Method),
			Status: http.StatusOK,
			// Response intentionally left nil — user fills in the mock body.
		})
	}
	sort.Slice(routes, func(i, j int) bool {
		if routes[i].Path == routes[j].Path {
			return routes[i].Method < routes[j].Method
		}
		return routes[i].Path < routes[j].Path
	})
	return config.Config{Routes: routes}
}
