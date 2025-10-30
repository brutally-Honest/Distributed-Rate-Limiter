package server

import (
	"net/http"

	"github.com/brutally-Honest/distributed-rate-limiter/internal/config"
	handlers "github.com/brutally-Honest/distributed-rate-limiter/internal/http"
)

func SetUpRoutes(cfg *config.Config) http.Handler {
	mux := http.NewServeMux()
	httpHandlers := handlers.New(cfg)

	mux.Handle("/api", http.HandlerFunc(httpHandlers.HandleApi))
	mux.Handle("/health", http.HandlerFunc(httpHandlers.HandleHealth))

	return mux
}
