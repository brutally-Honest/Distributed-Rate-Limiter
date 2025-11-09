package handlers

import (
	"encoding/json"
	"net/http"
	"time"
)

func (h *Handlers) HandleApi(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	resp := map[string]interface{}{
		"Msg":        "Successfully Hit",
		"Time":       time.Now().Format(time.RFC3339),
		"InstanceId": h.cfg.Server.InstanceId,
	}
	json.NewEncoder(w).Encode(resp)
}
