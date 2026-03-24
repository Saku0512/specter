package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"flag"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	export_cmd "github.com/Saku0512/specter/cmd/export"
	gen_cmd "github.com/Saku0512/specter/cmd/gen"
	init_cmd "github.com/Saku0512/specter/cmd/init"
	record_cmd "github.com/Saku0512/specter/cmd/record"
	validate_cmd "github.com/Saku0512/specter/cmd/validate"
	"github.com/Saku0512/specter/config"
	"github.com/Saku0512/specter/server"
	"github.com/fsnotify/fsnotify"
)

var version = "dev"

// generateSelfSignedCert creates an in-memory ECDSA P-256 self-signed certificate
// valid for 365 days, covering localhost and 127.0.0.1.
func generateSelfSignedCert() (tls.Certificate, error) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return tls.Certificate{}, err
	}

	serial, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	tmpl := &x509.Certificate{
		SerialNumber: serial,
		Subject:      pkix.Name{Organization: []string{"specter (self-signed)"}},
		DNSNames:     []string{"localhost"},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &priv.PublicKey, priv)
	if err != nil {
		return tls.Certificate{}, err
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyDER, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return tls.Certificate{}, err
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})

	return tls.X509KeyPair(certPEM, keyPEM)
}

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
  --ui-port    Port for the web UI (default: 4444, set to 0 to disable)
  --verbose    Log request headers and body
  -v, --version  Show version
  -h, --help   Show this help

Commands:
  init         Create a starter config.yml
  gen          Generate config from an OpenAPI spec
  validate     Validate a config file
  record       Proxy a real API and record responses to config.yml
  export       Generate a starter config from a running specter's request history

Environment variables:
  SPECTER_CONFIG    Path to config file
  SPECTER_PORT      Port to listen on
  SPECTER_HOST      Host to listen on
  SPECTER_CERT      TLS certificate file
  SPECTER_KEY       TLS key file
  SPECTER_VERBOSE   Set to 1 or true to enable verbose logging
  SPECTER_UI_PORT   Port for the web UI (0 to disable)

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
	if len(os.Args) > 1 && os.Args[1] == "export" {
		export_cmd.Run(os.Args[2:])
		return
	}

	flag.Usage = usage

	configPath := flag.String("c", "config.yaml", "path to config file")
	port := flag.String("p", "8080", "port to listen on")
	host := flag.String("host", "", "host to listen on (default: all interfaces)")
	tlsAuto := flag.Bool("tls", false, "enable HTTPS with auto-generated self-signed certificate")
	cert := flag.String("cert", "", "TLS certificate file")
	key := flag.String("key", "", "TLS key file")
	uiPort := flag.String("ui-port", "4444", "port for the web UI (0 to disable)")
	verbose := flag.Bool("verbose", false, "log request headers and body")
	random := flag.Bool("random", false, "generate random responses from OpenAPI spec (requires --openapi or openapi: in config)")
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
	if !set["ui-port"] {
		if val := os.Getenv("SPECTER_UI_PORT"); val != "" {
			*uiPort = val
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

	srv := server.New(cfg, *verbose, *random)

	if *uiPort != "0" && *uiPort != "" {
		scheme := "http"
		if *tlsAuto || (*cert != "" && *key != "") {
			scheme = "https"
		}
		apiAddr := fmt.Sprintf("%s://localhost:%s", scheme, *port)
		go server.StartUI("localhost:"+*uiPort, apiAddr)
	}

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

	useTLS := *tlsAuto || (*cert != "" && *key != "")

	go func() {
		if useTLS {
			if *cert != "" && *key != "" {
				// Use provided cert/key files
				log.Printf("👻 Specter running on https://localhost:%s", *port)
				if err := httpSrv.ListenAndServeTLS(*cert, *key); err != nil && err != http.ErrServerClosed {
					log.Fatalf("server error: %v", err)
				}
			} else {
				// Auto-generate self-signed certificate
				tlsCert, err := generateSelfSignedCert()
				if err != nil {
					log.Fatalf("failed to generate TLS certificate: %v", err)
				}
				httpSrv.TLSConfig = &tls.Config{Certificates: []tls.Certificate{tlsCert}}
				log.Printf("👻 Specter running on https://localhost:%s (self-signed cert)", *port)
				if err := httpSrv.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
					log.Fatalf("server error: %v", err)
				}
			}
		} else {
			log.Printf("👻 Specter running on http://localhost:%s", *port)
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
