package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/brutally-Honest/distributed-rate-limiter/internal/config"
)

type Handlers struct {
	config *config.Config
}

func NewHandlers(cfg *config.Config) *Handlers {
	return &Handlers{
		config: cfg,
	}
}

type response struct {
	Msg  string    `json:"msg"`
	Time time.Time `json:"time"`
}

func (h *Handlers) HandleApi(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	resp := response{
		Msg:  "Successfully Hit",
		Time: time.Now(),
	}
	json.NewEncoder(w).Encode(resp)
}

func (h *Handlers) HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	resp := response{
		Msg:  fmt.Sprintf("Health ok at %s", h.config.Server.InstanceId),
		Time: time.Now(),
	}
	json.NewEncoder(w).Encode(resp)
}
