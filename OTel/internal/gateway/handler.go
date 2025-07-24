package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"otel/pkg/telemetry"
	"otel/pkg/validator"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// CEPRequest represents the input request structure
type CEPRequest struct {
	CEP string `json:"cep"`
}

// ErrorResponse represents the error response structure
type ErrorResponse struct {
	Message string `json:"message"`
}

// OrchestrationResponse represents a response from the orchestration service
type OrchestrationResponse struct {
	Body       []byte
	StatusCode int
}

// GatewayHandler handles HTTP requests for the gateway service
type GatewayHandler struct {
	orchestrationServiceURL string
	tracer                  trace.Tracer
	httpClient              *http.Client
}

// NewGatewayHandler creates a new gateway handler
func NewGatewayHandler(orchestrationServiceURL string) *GatewayHandler {
	log.Printf("[GATEWAY] Initializing gateway handler with orchestration URL: %s", orchestrationServiceURL)

	// Create HTTP client with OpenTelemetry instrumentation
	httpClient := &http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
		Timeout:   30 * time.Second,
	}

	return &GatewayHandler{
		orchestrationServiceURL: orchestrationServiceURL,
		tracer:                  telemetry.GetTracer("otel-gateway"),
		httpClient:              httpClient,
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

	// Start a new span for this request
	ctx, span := h.tracer.Start(r.Context(), "gateway.process_cep")
	defer span.End()

	// Add attributes to the span
	span.SetAttributes(
		attribute.String("client.ip", clientIP),
		attribute.String("http.method", r.Method),
		attribute.String("http.url", r.URL.String()),
	)

	log.Printf("[GATEWAY] Received CEP request from %s", clientIP)

	w.Header().Set("Content-Type", "application/json")

	// Parse request body
	var req CEPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[GATEWAY] Failed to parse request body from %s: %v", clientIP, err)
		span.SetStatus(codes.Error, "Failed to parse request body")
		span.RecordError(err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "invalid request body"})
		return
	}

	log.Printf("[GATEWAY] Processing CEP: %s from %s", req.CEP, clientIP)
	span.SetAttributes(attribute.String("cep.input", req.CEP))

	// Start CEP validation span
	_, validationSpan := h.tracer.Start(ctx, "gateway.validate_cep")
	validationStart := time.Now()

	// Validate CEP
	if !validator.ValidateCEP(req.CEP) {
		validationSpan.SetStatus(codes.Error, "Invalid CEP format")
		validationSpan.End()
		log.Printf("[GATEWAY] Invalid CEP format: %s from %s", req.CEP, clientIP)
		span.SetStatus(codes.Error, "Invalid CEP format")
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "invalid zipcode"})
		return
	}

	validationDuration := time.Since(validationStart)
	validationSpan.SetAttributes(
		attribute.String("cep.validated", req.CEP),
		attribute.Int64("validation.duration_ms", validationDuration.Milliseconds()),
	)
	validationSpan.SetStatus(codes.Ok, "CEP validation successful")
	validationSpan.End()

	log.Printf("[GATEWAY] CEP validation successful: %s", req.CEP)

	// Forward to orchestration service
	orchestrationResp, err := h.forwardToOrchestrationService(ctx, req.CEP)
	if err != nil {
		log.Printf("[GATEWAY] Failed to forward request to orchestration service: %v", err)
		span.SetStatus(codes.Error, "Failed to forward request to orchestration service")
		span.RecordError(err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "failed to process request"})
		return
	}

	// Handle different response status codes from orchestration service
	if orchestrationResp.StatusCode != http.StatusOK {
		log.Printf("[GATEWAY] Orchestration service returned error status %d", orchestrationResp.StatusCode)
		span.SetAttributes(attribute.Int("orchestration.status_code", orchestrationResp.StatusCode))
		span.SetStatus(codes.Error, fmt.Sprintf("Orchestration service returned status %d", orchestrationResp.StatusCode))

		// Forward the exact status code and response from orchestration service
		w.WriteHeader(orchestrationResp.StatusCode)
		w.Write(orchestrationResp.Body)
		return
	}

	// Return the successful response from orchestration service
	duration := time.Since(startTime)
	log.Printf("[GATEWAY] Successfully processed CEP: %s from %s in %v", req.CEP, clientIP, duration)

	span.SetAttributes(
		attribute.Int64("request.duration_ms", duration.Milliseconds()),
		attribute.Int("http.status_code", http.StatusOK),
	)
	span.SetStatus(codes.Ok, "Request processed successfully")

	w.WriteHeader(http.StatusOK)
	w.Write(orchestrationResp.Body)
}

// forwardToOrchestrationService forwards the CEP to the orchestration service
func (h *GatewayHandler) forwardToOrchestrationService(ctx context.Context, cep string) (*OrchestrationResponse, error) {
	// Start span for orchestration service call
	_, span := h.tracer.Start(ctx, "gateway.call_orchestration_service")
	defer span.End()

	// Format CEP for the orchestration service (add hyphen if needed)
	formattedCEP := validator.FormatCEP(cep)
	log.Printf("[GATEWAY] Formatted CEP: %s -> %s", cep, formattedCEP)

	// Create the URL for the orchestration service
	url := fmt.Sprintf("%s/weather/%s", h.orchestrationServiceURL, formattedCEP)
	log.Printf("[GATEWAY] Calling orchestration service at: %s", url)

	span.SetAttributes(
		attribute.String("orchestration.url", url),
		attribute.String("cep.formatted", formattedCEP),
	)

	// Create HTTP request with context
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		span.SetStatus(codes.Error, "Failed to create HTTP request")
		span.RecordError(err)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Make HTTP request to orchestration service
	requestStart := time.Now()
	resp, err := h.httpClient.Do(req)
	if err != nil {
		log.Printf("[GATEWAY] HTTP request to orchestration service failed: %v", err)
		span.SetStatus(codes.Error, "HTTP request failed")
		span.RecordError(err)
		return nil, fmt.Errorf("failed to call orchestration service: %w", err)
	}
	defer resp.Body.Close()

	requestDuration := time.Since(requestStart)
	log.Printf("[GATEWAY] Orchestration service responded with status %d in %v", resp.StatusCode, requestDuration)

	span.SetAttributes(
		attribute.Int("http.status_code", resp.StatusCode),
		attribute.Int64("request.duration_ms", requestDuration.Milliseconds()),
	)

	// Read the response
	var buf bytes.Buffer
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		log.Printf("[GATEWAY] Failed to read response body: %v", err)
		span.SetStatus(codes.Error, "Failed to read response body")
		span.RecordError(err)
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// If orchestration service returns an error, forward it
	if resp.StatusCode != http.StatusOK {
		log.Printf("[GATEWAY] Orchestration service returned error status %d: %s", resp.StatusCode, buf.String())
		span.SetStatus(codes.Error, fmt.Sprintf("Orchestration service error: %d", resp.StatusCode))
		return &OrchestrationResponse{
			Body:       buf.Bytes(),
			StatusCode: resp.StatusCode,
		}, nil
	}

	span.SetAttributes(attribute.Int("response.size_bytes", buf.Len()))
	span.SetStatus(codes.Ok, "Successfully received response from orchestration service")

	log.Printf("[GATEWAY] Successfully received response from orchestration service: %d bytes", buf.Len())
	return &OrchestrationResponse{
		Body:       buf.Bytes(),
		StatusCode: resp.StatusCode,
	}, nil
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
