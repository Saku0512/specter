package record

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/Saku0512/specter/config"
	"gopkg.in/yaml.v3"
)

type recordedRoute struct {
	method      string
	path        string
	status      int
	contentType string
	body        []byte
}

type responseRecorder struct {
	http.ResponseWriter
	status  int
	buf     bytes.Buffer
	headers http.Header
}

func (r *responseRecorder) WriteHeader(status int) {
	r.status = status
	r.headers = r.ResponseWriter.Header().Clone()
	r.ResponseWriter.WriteHeader(status)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	r.buf.Write(b)
	return r.ResponseWriter.Write(b)
}

func Run(args []string) {
	fs := flag.NewFlagSet("record", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: specter record -t <target URL> [flags]\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		fs.PrintDefaults()
	}
	target := fs.String("t", "", "target URL to proxy to (required)")
	output := fs.String("o", "config.yml", "output config file")
	port := fs.String("p", "8080", "port to listen on")
	force := fs.Bool("f", false, "overwrite output file if it exists")
	fs.Parse(args)

	if *target == "" {
		fmt.Fprintln(os.Stderr, "error: -t <target URL> is required")
		fs.Usage()
		os.Exit(1)
	}

	targetURL, err := url.Parse(*target)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: invalid target URL: %v\n", err)
		os.Exit(1)
	}

	if !*force {
		if _, err := os.Stat(*output); err == nil {
			fmt.Fprintf(os.Stderr, "error: %s already exists, use -f to overwrite\n", *output)
			os.Exit(1)
		}
	}

	var mu sync.Mutex
	var routes []recordedRoute
	seen := map[string]bool{}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	proxy.Director = func(req *http.Request) {
		req.URL.Scheme = targetURL.Scheme
		req.URL.Host = targetURL.Host
		req.Host = targetURL.Host
		if targetURL.Path != "" && targetURL.Path != "/" {
			req.URL.Path = targetURL.Path + req.URL.Path
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Skip CORS preflight
		if r.Method == http.MethodOptions {
			proxy.ServeHTTP(w, r)
			return
		}

		rec := &responseRecorder{ResponseWriter: w, status: http.StatusOK}
		proxy.ServeHTTP(rec, r)

		key := r.Method + " " + r.URL.Path
		mu.Lock()
		if !seen[key] {
			seen[key] = true
			ct := rec.headers.Get("Content-Type")
			if ct == "" {
				ct = w.Header().Get("Content-Type")
			}
			routes = append(routes, recordedRoute{
				method:      r.Method,
				path:        r.URL.Path,
				status:      rec.status,
				contentType: ct,
				body:        rec.buf.Bytes(),
			})
			log.Printf("recorded %s %s → %d", r.Method, r.URL.Path, rec.status)
		} else {
			log.Printf("skipped  %s %s (already recorded)", r.Method, r.URL.Path)
		}
		mu.Unlock()
	})

	srv := &http.Server{
		Addr:    ":" + *port,
		Handler: mux,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("👻 Specter recording on :%s → %s", *port, *target)
		log.Printf("   Send requests, then press Ctrl+C to save %s", *output)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-quit
	fmt.Println()

	if len(routes) == 0 {
		fmt.Println("no routes recorded, exiting without writing file")
		return
	}

	cfg := buildConfig(routes)
	data, err := yaml.Marshal(cfg)
	if err != nil {
		log.Fatalf("failed to marshal config: %v", err)
	}

	if err := os.WriteFile(*output, data, 0o644); err != nil {
		log.Fatalf("failed to write %s: %v", *output, err)
	}

	fmt.Printf("✓ recorded %d route(s) → %s\n", len(routes), *output)
}

func buildConfig(routes []recordedRoute) config.Config {
	var cfgRoutes []config.Route
	for _, r := range routes {
		route := config.Route{
			Path:   r.path,
			Method: r.method,
			Status: r.status,
		}

		isJSON := strings.Contains(r.contentType, "application/json")
		if isJSON && len(r.body) > 0 {
			var v any
			if err := json.Unmarshal(r.body, &v); err == nil {
				route.Response = v
			} else {
				route.Response = string(r.body)
			}
		} else if len(r.body) > 0 {
			route.Response = string(r.body)
			if r.contentType != "" && !isJSON {
				route.ContentType = strings.Split(r.contentType, ";")[0]
			}
		}

		cfgRoutes = append(cfgRoutes, route)
	}
	return config.Config{Routes: cfgRoutes}
}
