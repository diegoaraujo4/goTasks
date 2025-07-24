package gateway

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"

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
	w.Header().Set("Content-Type", "application/json")

	// Parse request body
	var req CEPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "invalid request body"})
		return
	}

	// Validate CEP
	if !isValidCEP(req.CEP) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "invalid zipcode"})
		return
	}

	// Forward to orchestration service
	response, err := h.forwardToOrchestrationService(req.CEP)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "failed to process request"})
		return
	}

	// Return the response from orchestration service
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

// isValidCEP validates if the CEP is a valid 8-digit string
func isValidCEP(cep string) bool {
	// Check if it's a string and has exactly 8 digits
	if len(cep) != 8 {
		return false
	}

	// Check if all characters are digits
	matched, _ := regexp.MatchString(`^\d{8}$`, cep)
	return matched
}

// forwardToOrchestrationService forwards the CEP to the orchestration service
func (h *GatewayHandler) forwardToOrchestrationService(cep string) ([]byte, error) {
	// Format CEP for the orchestration service (add hyphen if needed)
	formattedCEP := validator.FormatCEP(cep)

	// Create the URL for the orchestration service
	url := fmt.Sprintf("%s/weather/%s", h.orchestrationServiceURL, formattedCEP)

	// Make HTTP request to orchestration service
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to call orchestration service: %w", err)
	}
	defer resp.Body.Close()

	// Read the response
	var buf bytes.Buffer
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// If orchestration service returns an error, forward it
	if resp.StatusCode != http.StatusOK {
		return buf.Bytes(), fmt.Errorf("orchestration service error: %d", resp.StatusCode)
	}

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
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "healthy",
		"service": "otel-gateway",
	})
}
