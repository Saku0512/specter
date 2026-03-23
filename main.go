package main

import (
	"flag"
	"log"
	"github.com/Saku0512/specter/config"
	"github.com/Saku0512/specter/server"
)

func main() {
	configPath := flag.String("c", "config.yaml", "path to config file")
	port := flag.String("p", "8080", "port to listen on")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	r := server.New(cfg)

	log.Printf("👻 Specter running on :%s", *port)
	r.Run(":" + *port)
}
