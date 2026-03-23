package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Saku0512/specter/config"
	"github.com/Saku0512/specter/gen"
	"github.com/Saku0512/specter/server"
	"github.com/fsnotify/fsnotify"
)

var version = "dev"

func usage() {
	fmt.Fprintf(os.Stderr, `👻 specter %s — lightweight mock API server

Usage:
  specter [flags]
  specter gen -i openapi.yml [-o config.yml]

Flags:
  -c <path>    Path to config file (default: config.yaml)
  -p <port>    Port to listen on (default: 8080)
  --verbose    Log request headers and body
  -v, --version  Show version
  -h, --help   Show this help

Commands:
  gen          Generate config from an OpenAPI spec

Examples:
  specter -c config.yml -p 3000
  specter gen -i openapi.yml -o config.yml

`, version)
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "gen" {
		gen.Run(os.Args[2:])
		return
	}

	flag.Usage = usage

	configPath := flag.String("c", "config.yaml", "path to config file")
	port := flag.String("p", "8080", "port to listen on")
	verbose := flag.Bool("verbose", false, "log request headers and body")
	v := flag.Bool("v", false, "show version")
	flag.BoolVar(v, "version", false, "show version")
	flag.Parse()

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

	log.Printf("👻 Specter running on :%s", *port)
	if err := http.ListenAndServe(":"+*port, srv); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
