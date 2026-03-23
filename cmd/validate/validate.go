package validate

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/Saku0512/specter/config"
)

var validMethods = map[string]bool{
	"GET": true, "POST": true, "PUT": true, "PATCH": true,
	"DELETE": true, "HEAD": true, "OPTIONS": true,
}

var validModes = map[string]bool{
	"": true, "sequential": true, "random": true,
}

func Run(args []string) {
	fs := flag.NewFlagSet("validate", flag.ExitOnError)
	configPath := fs.String("c", "config.yaml", "path to config file")
	fs.Parse(args)

	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "✗ failed to load %s: %v\n", *configPath, err)
		os.Exit(1)
	}

	errs := check(cfg)
	if len(errs) == 0 {
		fmt.Printf("✓ %s is valid (%d routes)\n", *configPath, len(cfg.Routes))
		return
	}

	fmt.Fprintf(os.Stderr, "✗ %s has %d error(s):\n", *configPath, len(errs))
	for _, e := range errs {
		fmt.Fprintf(os.Stderr, "  - %s\n", e)
	}
	os.Exit(1)
}

func check(cfg *config.Config) []string {
	var errs []string

	if cfg.Proxy != "" {
		if _, err := url.ParseRequestURI(cfg.Proxy); err != nil {
			errs = append(errs, fmt.Sprintf("invalid proxy URL %q: %v", cfg.Proxy, err))
		}
	}

	for i, r := range cfg.Routes {
		prefix := fmt.Sprintf("route %d", i+1)
		if r.Path != "" && r.Method != "" {
			prefix = fmt.Sprintf("route %d (%s %s)", i+1, r.Method, r.Path)
		}

		if r.Path == "" {
			errs = append(errs, prefix+": missing path")
		}
		if r.Method == "" {
			errs = append(errs, prefix+": missing method")
		} else if !validMethods[strings.ToUpper(r.Method)] {
			errs = append(errs, prefix+fmt.Sprintf(": invalid method %q", r.Method))
		}
		if r.Status != 0 && (r.Status < 100 || r.Status > 599) {
			errs = append(errs, prefix+fmt.Sprintf(": invalid status %d", r.Status))
		}
		if !validModes[r.Mode] {
			errs = append(errs, prefix+fmt.Sprintf(": invalid mode %q (must be sequential or random)", r.Mode))
		}
		if r.Delay < 0 {
			errs = append(errs, prefix+": delay must be non-negative")
		}
		if r.RateLimit < 0 {
			errs = append(errs, prefix+": rate_limit must be non-negative")
		}
		if r.RateReset < 0 {
			errs = append(errs, prefix+": rate_reset must be non-negative")
		}
		if r.RateReset > 0 && r.RateLimit == 0 {
			errs = append(errs, prefix+": rate_reset requires rate_limit to be set")
		}
		if r.File != "" {
			if _, err := os.Stat(r.File); err != nil {
				errs = append(errs, prefix+fmt.Sprintf(": file %q not found", r.File))
			}
		}
		for j, resp := range r.Responses {
			if resp.Status != 0 && (resp.Status < 100 || resp.Status > 599) {
				errs = append(errs, prefix+fmt.Sprintf(": responses[%d] invalid status %d", j, resp.Status))
			}
			if resp.File != "" {
				if _, err := os.Stat(resp.File); err != nil {
					errs = append(errs, prefix+fmt.Sprintf(": responses[%d] file %q not found", j, resp.File))
				}
			}
		}
		for j, m := range r.Match {
			if len(m.Query) == 0 && len(m.Body) == 0 && len(m.Headers) == 0 {
				errs = append(errs, prefix+fmt.Sprintf(": match[%d] must have at least one query, body, or headers condition", j))
			}
			if m.File != "" {
				if _, err := os.Stat(m.File); err != nil {
					errs = append(errs, prefix+fmt.Sprintf(": match[%d] file %q not found", j, m.File))
				}
			}
		}
		if wh := r.Webhook; wh != nil {
			if wh.URL == "" {
				errs = append(errs, prefix+": webhook missing url")
			} else if _, err := url.ParseRequestURI(wh.URL); err != nil {
				errs = append(errs, prefix+fmt.Sprintf(": webhook invalid url %q: %v", wh.URL, err))
			}
			if wh.Method != "" && !validMethods[strings.ToUpper(wh.Method)] {
				errs = append(errs, prefix+fmt.Sprintf(": webhook invalid method %q", wh.Method))
			}
			if wh.Delay < 0 {
				errs = append(errs, prefix+": webhook delay must be non-negative")
			}
		}
	}

	return errs
}
