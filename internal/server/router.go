package server

import (
	"net/http"

	"github.com/brutally-Honest/distributed-rate-limiter/internal/config"
	"github.com/brutally-Honest/distributed-rate-limiter/internal/handlers"
)

func SetUpRoutes(cfg *config.Config) http.Handler {
	mux := http.NewServeMux()
	handlers := handlers.New(cfg)

	mux.Handle("/api", http.HandlerFunc(handlers.HandleApi))
	mux.Handle("/health", http.HandlerFunc(handlers.HandleHealth))

	return mux
}
