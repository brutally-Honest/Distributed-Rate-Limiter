package server

import (
	"fmt"
	"net/http"

	"github.com/brutally-Honest/distributed-rate-limiter/internal/config"
	"github.com/brutally-Honest/distributed-rate-limiter/internal/middlewares"
)

type Server struct {
	httpServer *http.Server
	config     *config.Config
}

func New(cfg *config.Config) *Server {

	router := SetUpRoutes(cfg)
	handlersWithMiddleware := middlewares.Chain(
		middlewares.Logger(),
	)(router)

	s := &Server{
		config: cfg,
		httpServer: &http.Server{
			Addr:    ":" + cfg.Server.Port,
			Handler: handlersWithMiddleware,
		},
	}

	return s
}

func (s *Server) Start() error {
	fmt.Println("Server running on port", s.config.Server.Port)
	return s.httpServer.ListenAndServe()
}
