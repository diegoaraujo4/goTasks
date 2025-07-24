package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"otel/internal/domain"
	"otel/internal/service"
	"otel/pkg/telemetry"

	"github.com/gorilla/mux"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// WeatherHandler handles HTTP requests for weather endpoints
type WeatherHandler struct {
	weatherService *service.WeatherService
	tracer         trace.Tracer
}

// NewWeatherHandler creates a new weather handler
func NewWeatherHandler(weatherService *service.WeatherService) *WeatherHandler {
	log.Printf("[ORCHESTRATOR] Initializing weather handler")
	return &WeatherHandler{
		weatherService: weatherService,
		tracer:         telemetry.GetTracer("otel-orchestration"),
	}
}

// GetWeatherByCEP godoc
// @Summary Obter temperatura por CEP
// @Description Recebe um CEP brasileiro válido (já validado pelo Gateway) e retorna a temperatura atual em Celsius, Fahrenheit e Kelvin
// @Tags weather
// @Accept json
// @Produce json
// @Param cep path string true "CEP brasileiro (8 dígitos, já validado)" example("01310100")
// @Success 200 {object} domain.WeatherResponse "Informações de temperatura"
// @Failure 404 {object} domain.ErrorResponse "CEP não encontrado"
// @Failure 500 {object} domain.ErrorResponse "Erro interno do servidor"
// @Router /weather/{cep} [get]
func (h *WeatherHandler) GetWeatherByCEP(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	clientIP := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		clientIP = forwarded
	}

	vars := mux.Vars(r)
	cep := vars["cep"]

	// Start a new span for this request
	ctx, span := h.tracer.Start(r.Context(), "orchestration.get_weather_by_cep")
	defer span.End()

	// Add attributes to the span
	span.SetAttributes(
		attribute.String("client.ip", clientIP),
		attribute.String("cep.input", cep),
		attribute.String("http.method", r.Method),
		attribute.String("http.url", r.URL.String()),
	)

	log.Printf("[ORCHESTRATOR] Received weather request for CEP: %s from %s", cep, clientIP)

	weather, err := h.weatherService.GetWeatherByCEP(ctx, cep)
	if err != nil {
		log.Printf("[ORCHESTRATOR] Error processing CEP %s from %s: %v", cep, clientIP, err)
		span.SetStatus(codes.Error, "Error processing CEP")
		span.RecordError(err)
		h.handleError(w, err)
		return
	}

	duration := time.Since(startTime)
	log.Printf("[ORCHESTRATOR] Successfully processed weather request for CEP: %s from %s in %v", cep, clientIP, duration)

	span.SetAttributes(
		attribute.String("weather.city", weather.City),
		attribute.Float64("weather.temp_c", weather.TempC),
		attribute.Int64("request.duration_ms", duration.Milliseconds()),
		attribute.Int("http.status_code", http.StatusOK),
	)
	span.SetStatus(codes.Ok, "Weather request processed successfully")

	h.sendJSON(w, http.StatusOK, weather)
}

// handleError handles different types of errors and sends appropriate HTTP responses
func (h *WeatherHandler) handleError(w http.ResponseWriter, err error) {
	var statusCode int
	var message string

	switch {
	// NOTE: CEP validation is now handled by the Gateway service
	// case errors.Is(err, service.ErrInvalidCEP):
	//	statusCode = http.StatusUnprocessableEntity
	//	message = service.ErrInvalidCEP.Error()
	//	log.Printf("[ORCHESTRATOR] Invalid CEP error: %v", err)
	case errors.Is(err, service.ErrCEPNotFound):
		statusCode = http.StatusNotFound
		message = service.ErrCEPNotFound.Error()
		log.Printf("[ORCHESTRATOR] CEP not found error: %v", err)
	case errors.Is(err, service.ErrWeatherDataUnavailable):
		statusCode = http.StatusInternalServerError
		message = service.ErrWeatherDataUnavailable.Error()
		log.Printf("[ORCHESTRATOR] Weather data unavailable error: %v", err)
	default:
		statusCode = http.StatusInternalServerError
		message = "internal server error"
		log.Printf("[ORCHESTRATOR] Unexpected error: %v", err)
	}

	log.Printf("[ORCHESTRATOR] Sending error response - Status: %d, Message: %s", statusCode, message)
	errorResponse := domain.ErrorResponse{Message: message}
	h.sendJSON(w, statusCode, errorResponse)
}

// sendJSON sends a JSON response
func (h *WeatherHandler) sendJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	log.Printf("[ORCHESTRATOR] Sending JSON response - Status: %d", statusCode)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("[ORCHESTRATOR] Error encoding JSON response: %v", err)
	}
}
