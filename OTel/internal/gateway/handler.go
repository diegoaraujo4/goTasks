package gateway

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"otel/pkg/validator"
)

// CEPRequest represents the input request structure
type CEPRequest struct {
	CEP string `json:"cep"`
}

// ErrorResponse represents the error response structure
type ErrorResponse struct {
	Message string `json:"message"`
}

// GatewayHandler handles HTTP requests for the gateway service
type GatewayHandler struct {
	orchestrationServiceURL string
}

// NewGatewayHandler creates a new gateway handler
func NewGatewayHandler(orchestrationServiceURL string) *GatewayHandler {
	log.Printf("[GATEWAY] Initializing gateway handler with orchestration URL: %s", orchestrationServiceURL)
	return &GatewayHandler{
		orchestrationServiceURL: orchestrationServiceURL,
	}
}

// ProcessCEP handles the CEP input validation and forwarding
// @Summary Process CEP input
// @Description Validates CEP input and forwards to orchestration service
// @Tags gateway
// @Accept json
// @Produce json
// @Param cep body CEPRequest true "CEP input"
// @Success 200 {object} map[string]interface{} "Success response from orchestration service"
// @Failure 422 {object} ErrorResponse "Invalid zipcode"
// @Failure 400 {object} ErrorResponse "Bad request"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /cep [post]
func (h *GatewayHandler) ProcessCEP(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	clientIP := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		clientIP = forwarded
	}

	log.Printf("[GATEWAY] Received CEP request from %s", clientIP)

	w.Header().Set("Content-Type", "application/json")

	// Parse request body
	var req CEPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[GATEWAY] Failed to parse request body from %s: %v", clientIP, err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "invalid request body"})
		return
	}

	log.Printf("[GATEWAY] Processing CEP: %s from %s", req.CEP, clientIP)

	// Validate CEP
	if !validator.ValidateCEP(req.CEP) {
		log.Printf("[GATEWAY] Invalid CEP format: %s from %s", req.CEP, clientIP)
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "invalid zipcode"})
		return
	}

	log.Printf("[GATEWAY] CEP validation successful: %s", req.CEP)

	// Forward to orchestration service
	response, err := h.forwardToOrchestrationService(req.CEP)
	if err != nil {
		log.Printf("[GATEWAY] Failed to forward request to orchestration service: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "failed to process request"})
		return
	}

	// Return the response from orchestration service
	duration := time.Since(startTime)
	log.Printf("[GATEWAY] Successfully processed CEP: %s from %s in %v", req.CEP, clientIP, duration)
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

// forwardToOrchestrationService forwards the CEP to the orchestration service
func (h *GatewayHandler) forwardToOrchestrationService(cep string) ([]byte, error) {
	// Format CEP for the orchestration service (add hyphen if needed)
	formattedCEP := validator.FormatCEP(cep)
	log.Printf("[GATEWAY] Formatted CEP: %s -> %s", cep, formattedCEP)

	// Create the URL for the orchestration service
	url := fmt.Sprintf("%s/weather/%s", h.orchestrationServiceURL, formattedCEP)
	log.Printf("[GATEWAY] Calling orchestration service at: %s", url)

	// Make HTTP request to orchestration service
	requestStart := time.Now()
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("[GATEWAY] HTTP request to orchestration service failed: %v", err)
		return nil, fmt.Errorf("failed to call orchestration service: %w", err)
	}
	defer resp.Body.Close()

	requestDuration := time.Since(requestStart)
	log.Printf("[GATEWAY] Orchestration service responded with status %d in %v", resp.StatusCode, requestDuration)

	// Read the response
	var buf bytes.Buffer
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		log.Printf("[GATEWAY] Failed to read response body: %v", err)
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// If orchestration service returns an error, forward it
	if resp.StatusCode != http.StatusOK {
		log.Printf("[GATEWAY] Orchestration service returned error status %d: %s", resp.StatusCode, buf.String())
		return buf.Bytes(), fmt.Errorf("orchestration service error: %d", resp.StatusCode)
	}

	log.Printf("[GATEWAY] Successfully received response from orchestration service: %d bytes", buf.Len())
	return buf.Bytes(), nil
}

// HealthCheck handles health check requests
// @Summary Health check
// @Description Check if the gateway service is healthy
// @Tags health
// @Produce json
// @Success 200 {object} map[string]string "Service is healthy"
// @Router /health [get]
func (h *GatewayHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	clientIP := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		clientIP = forwarded
	}

	log.Printf("[GATEWAY] Health check requested from %s", clientIP)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "healthy",
		"service": "otel-gateway",
	})

	log.Printf("[GATEWAY] Health check response sent to %s", clientIP)
}
