package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Saku0512/specter/config"
	gen_cmd "github.com/Saku0512/specter/cmd/gen"
	init_cmd "github.com/Saku0512/specter/cmd/init"
	record_cmd "github.com/Saku0512/specter/cmd/record"
	validate_cmd "github.com/Saku0512/specter/cmd/validate"
	"github.com/Saku0512/specter/server"
	"github.com/fsnotify/fsnotify"
)

var version = "dev"

func printRoutes(cfg *config.Config) {
	fmt.Printf("\nregistered %d route(s):\n", len(cfg.Routes))
	for _, r := range cfg.Routes {
		fmt.Printf("  %-8s %s\n", r.Method, r.Path)
	}
	fmt.Println()
}

func usage() {
	fmt.Fprintf(os.Stderr, `👻 specter %s — lightweight mock API server

Usage:
  specter [flags]
  specter gen -i openapi.yml [-o config.yml]

Flags:
  -c <path>    Path to config file (default: config.yaml)
  -p <port>    Port to listen on (default: 8080)
  --host       Host to listen on (default: all interfaces)
  --cert       TLS certificate file (enables HTTPS)
  --key        TLS key file (enables HTTPS)
  --verbose    Log request headers and body
  -v, --version  Show version
  -h, --help   Show this help

Commands:
  init         Create a starter config.yml
  gen          Generate config from an OpenAPI spec
  validate     Validate a config file
  record       Proxy a real API and record responses to config.yml

Environment variables:
  SPECTER_CONFIG   Path to config file
  SPECTER_PORT     Port to listen on
  SPECTER_HOST     Host to listen on
  SPECTER_CERT     TLS certificate file
  SPECTER_KEY      TLS key file
  SPECTER_VERBOSE  Set to 1 or true to enable verbose logging

Examples:
  specter -c config.yml -p 3000
  specter gen -i openapi.yml -o config.yml
  SPECTER_PORT=3000 specter -c config.yml

`, version)
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "gen" {
		gen_cmd.Run(os.Args[2:])
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "validate" {
		validate_cmd.Run(os.Args[2:])
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "init" {
		init_cmd.Run(os.Args[2:])
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "record" {
		record_cmd.Run(os.Args[2:])
		return
	}

	flag.Usage = usage

	configPath := flag.String("c", "config.yaml", "path to config file")
	port := flag.String("p", "8080", "port to listen on")
	host := flag.String("host", "", "host to listen on (default: all interfaces)")
	cert := flag.String("cert", "", "TLS certificate file")
	key := flag.String("key", "", "TLS key file")
	verbose := flag.Bool("verbose", false, "log request headers and body")
	v := flag.Bool("v", false, "show version")
	flag.BoolVar(v, "version", false, "show version")
	flag.Parse()

	// Fall back to environment variables for flags not explicitly set
	set := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) { set[f.Name] = true })

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
	if !set["verbose"] {
		if val := os.Getenv("SPECTER_VERBOSE"); val == "1" || val == "true" {
			*verbose = true
		}
	}
	if !set["host"] {
		if val := os.Getenv("SPECTER_HOST"); val != "" {
			*host = val
		}
	}
	if !set["cert"] {
		if val := os.Getenv("SPECTER_CERT"); val != "" {
			*cert = val
		}
	}
	if !set["key"] {
		if val := os.Getenv("SPECTER_KEY"); val != "" {
			*key = val
		}
	}

	if *v {
		fmt.Println("specter", version)
		os.Exit(0)
	}

	if len(os.Args) == 1 {
		usage()
		os.Exit(0)
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	srv := server.New(cfg, *verbose)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalf("failed to create watcher: %v", err)
	}
	defer watcher.Close()

	if err := watcher.Add(*configPath); err != nil {
		log.Printf("warning: could not watch config file: %v", err)
	}

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
					newCfg, err := config.Load(*configPath)
					if err != nil {
						log.Printf("reload failed: %v", err)
						continue
					}
					srv.Reload(newCfg)
					log.Println("config reloaded")
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Printf("watcher error: %v", err)
			}
		}
	}()

	httpSrv := &http.Server{
		Addr:    *host + ":" + *port,
		Handler: srv,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	printRoutes(cfg)

	go func() {
		if *cert != "" && *key != "" {
			log.Printf("👻 Specter running on :%s (TLS)", *port)
			if err := httpSrv.ListenAndServeTLS(*cert, *key); err != nil && err != http.ErrServerClosed {
				log.Fatalf("server error: %v", err)
			}
		} else {
			log.Printf("👻 Specter running on :%s", *port)
			if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("server error: %v", err)
			}
		}
	}()

	<-quit
	log.Println("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpSrv.Shutdown(ctx); err != nil {
		log.Fatalf("shutdown error: %v", err)
	}
	log.Println("bye 👋")
}
