package main

import (
	"log"
	"os"

	"github.com/brutally-Honest/distributed-rate-limiter/internal/config"
	"github.com/brutally-Honest/distributed-rate-limiter/internal/server"
)

func main() {

	cfg, err := config.Load()
	if err != nil {
		log.Printf("App error %v", err)
		os.Exit(1)
	}

	server, err := server.New(cfg)
	if err != nil {
		log.Printf("Server error %v", err)
		os.Exit(1)
	}

	if err := server.Start(); err != nil {
		log.Printf("Failed to start server: %v", err)
		os.Exit(1)
	}
}
