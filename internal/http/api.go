package handlers

import (
	"encoding/json"
	"net/http"
	"time"
)

func (h *Handlers) HandleApi(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	resp := Resp{
		Msg:        "Successfully Hit",
		Time:       time.Now(),
		InstanceId: h.cfg.Server.InstanceId,
	}
	json.NewEncoder(w).Encode(resp)
}
