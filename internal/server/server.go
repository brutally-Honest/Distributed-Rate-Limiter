package server

import (
	"fmt"
	"net/http"

	"github.com/brutally-Honest/distributed-rate-limiter/internal/config"
)

type Server struct {
	httpServer *http.Server
	config     *config.Config
}

func New(cfg *config.Config) *Server {

	mux := http.NewServeMux()
	handlers := NewHandlers(cfg)
	s := &Server{
		config: cfg,
		httpServer: &http.Server{
			Addr:    ":" + cfg.Server.Port,
			Handler: mux,
		},
	}
	mux.HandleFunc("/api", handlers.HandleApi)
	mux.HandleFunc("/health", handlers.HandleHealth)
	return s
}

func (s *Server) Start() error {
	fmt.Println("Server running on port", s.config.Server.Port)
	return s.httpServer.ListenAndServe()
}
