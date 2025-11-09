package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

func (h *Handlers) HandleHealth(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	redisStatus := "disconnected"
	var redisLatency string

	if h.cfg.Redis.Client != nil {
		start := time.Now()
		err := h.cfg.Redis.Client.Ping(ctx).Err()
		latency := time.Since(start)

		if err == nil {
			redisStatus = "connected"
			redisLatency = latency.String()
		}
	}

	overallStatus := "healthy"
	httpStatus := http.StatusOK

	if redisStatus == "disconnected" {
		overallStatus = "unhealthy"
		httpStatus = http.StatusServiceUnavailable
	}

	resp := map[string]interface{}{
		"status":     overallStatus,
		"timestamp":  time.Now().Format(time.RFC3339),
		"uptime":     time.Since(h.startTime).String(),
		"instanceId": h.cfg.Server.InstanceId,
		"services": map[string]interface{}{
			"redis": map[string]interface{}{
				"status":  redisStatus,
				"latency": redisLatency,
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	json.NewEncoder(w).Encode(resp)
}
