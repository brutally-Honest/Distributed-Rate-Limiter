package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func (h *Handlers) HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	resp := Resp{
		Msg:  fmt.Sprintf("Health ok at %s", h.cfg.Server.InstanceId),
		Time: time.Now(),
	}
	json.NewEncoder(w).Encode(resp)
}
