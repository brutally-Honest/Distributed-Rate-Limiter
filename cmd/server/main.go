package main

import (
	"log"

	"github.com/brutally-Honest/distributed-rate-limiter/internal/config"
	"github.com/brutally-Honest/distributed-rate-limiter/internal/server"
)

func main() {

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Configuration error %v", err)
	}

	srv, err := server.New(cfg)
	if err != nil {
		log.Fatalf("Server initialization error %v", err)
	}

	if err := server.RunWithGracefulShutdown(srv); err != nil {
		log.Fatalf("Server shutdown error: %v", err)
	}

}
