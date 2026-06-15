package handler

import (
	"encoding/json"
	"net/http"
	"time"
)

// HealthHandler serves liveness checks for Docker, Railway, and load balancers.
type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Health responds with a simple JSON payload so we know the gateway is up.
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"service": "gateway",
		"time":    time.Now().UTC().Format(time.RFC3339),
	})
}
