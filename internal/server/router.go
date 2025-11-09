package server

import (
	"net/http"
	"time"

	"github.com/brutally-Honest/distributed-rate-limiter/internal/config"
	handlers "github.com/brutally-Honest/distributed-rate-limiter/internal/http"
	"github.com/brutally-Honest/distributed-rate-limiter/internal/middlewares"
)

func SetUpRoutes(cfg *config.Config, limiterMiddleware middlewares.Middleware, startTime time.Time) http.Handler {
	mux := http.NewServeMux()
	httpHandlers := handlers.New(cfg, startTime)

	mux.Handle("/api", limiterMiddleware(http.HandlerFunc(httpHandlers.HandleApi)))
	mux.Handle("/health", http.HandlerFunc(httpHandlers.HandleHealth))

	return mux
}
