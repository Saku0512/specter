package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/Saku0512/specter/config"
	"github.com/Saku0512/specter/server"
)

var version = "dev"

func main() {
	configPath := flag.String("c", "config.yaml", "path to config file")
	port := flag.String("p", "8080", "port to listen on")
	v := flag.Bool("v", false, "show version")
	flag.BoolVar(v, "version", false, "show version")
	flag.Parse()

	if *v {
		fmt.Println("specter", version)
		os.Exit(0)
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	r := server.New(cfg)

	log.Printf("👻 Specter running on :%s", *port)
	r.Run(":" + *port)
}
