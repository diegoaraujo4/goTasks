package handler

import (
	"log"
	"net/http"
)

// HealthHandler handles health check requests
type HealthHandler struct{}

// NewHealthHandler creates a new health handler
func NewHealthHandler() *HealthHandler {
	log.Printf("[ORCHESTRATOR] Initializing health handler")
	return &HealthHandler{}
}

// HealthCheck godoc
// @Summary Health check
// @Description Verifica se a aplicação está funcionando
// @Tags health
// @Produce plain
// @Success 200 {string} string "OK"
// @Router /health [get]
func (h *HealthHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	clientIP := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		clientIP = forwarded
	}

	log.Printf("[ORCHESTRATOR] Health check requested from %s", clientIP)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))

	log.Printf("[ORCHESTRATOR] Health check response sent to %s", clientIP)
}
