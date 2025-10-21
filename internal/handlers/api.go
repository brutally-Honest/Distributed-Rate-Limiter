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
		Msg:  "Successfully Hit",
		Time: time.Now(),
	}
	json.NewEncoder(w).Encode(resp)
}
