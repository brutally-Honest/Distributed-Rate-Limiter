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
		log.Fatalf("App error %v", err)
		os.Exit(1)
	}

	srv, err := server.New(cfg)
	if err != nil {
		log.Fatalf("Server error %v", err)
		os.Exit(1)
	}

	if err := server.RunWithGracefulShutdown(srv); err != nil {
		log.Fatalf("Server error: %v", err)
	}

}
