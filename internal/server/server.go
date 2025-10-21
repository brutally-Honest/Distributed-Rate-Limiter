package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/brutally-Honest/distributed-rate-limiter/internal/config"
)

type Server struct {
	httpServer *http.Server
	config     *config.Config
}

func New(cfg *config.Config) *Server {

	mux := http.NewServeMux()

	s := &Server{
		config: cfg,
		httpServer: &http.Server{
			Addr:    ":" + cfg.Server.Port,
			Handler: mux,
		},
	}
	mux.HandleFunc("/api", s.handleApi)
	mux.HandleFunc("/health", s.handleHealth)
	return s
}

type response struct {
	Msg  string    `json:"msg"`
	Time time.Time `json:"time"`
}

func (s *Server) handleApi(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	resp := response{
		Msg:  "Successfully Hit",
		Time: time.Now(),
	}
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	resp := response{
		Msg:  fmt.Sprintf("Health ok at %s", s.config.Server.InstanceId),
		Time: time.Now(),
	}
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) Start() error {
	fmt.Println("Server running on port", s.config.Server.Port)
	return s.httpServer.ListenAndServe()
}
