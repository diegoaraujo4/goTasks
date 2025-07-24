package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"otel/internal/domain"
	"otel/internal/service"

	"github.com/gorilla/mux"
)

// WeatherHandler handles HTTP requests for weather endpoints
type WeatherHandler struct {
	weatherService *service.WeatherService
}

// NewWeatherHandler creates a new weather handler
func NewWeatherHandler(weatherService *service.WeatherService) *WeatherHandler {
	log.Printf("[ORCHESTRATOR] Initializing weather handler")
	return &WeatherHandler{
		weatherService: weatherService,
	}
}

// GetWeatherByCEP godoc
// @Summary Obter temperatura por CEP
// @Description Recebe um CEP brasileiro válido e retorna a temperatura atual em Celsius, Fahrenheit e Kelvin
// @Tags weather
// @Accept json
// @Produce json
// @Param cep path string true "CEP brasileiro (8 dígitos)" example("01310100")
// @Success 200 {object} domain.WeatherResponse "Informações de temperatura"
// @Failure 422 {object} domain.ErrorResponse "CEP inválido"
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

	log.Printf("[ORCHESTRATOR] Received weather request for CEP: %s from %s", cep, clientIP)

	weather, err := h.weatherService.GetWeatherByCEP(cep)
	if err != nil {
		log.Printf("[ORCHESTRATOR] Error processing CEP %s from %s: %v", cep, clientIP, err)
		h.handleError(w, err)
		return
	}

	duration := time.Since(startTime)
	log.Printf("[ORCHESTRATOR] Successfully processed weather request for CEP: %s from %s in %v", cep, clientIP, duration)
	h.sendJSON(w, http.StatusOK, weather)
}

// handleError handles different types of errors and sends appropriate HTTP responses
func (h *WeatherHandler) handleError(w http.ResponseWriter, err error) {
	var statusCode int
	var message string

	switch {
	case errors.Is(err, service.ErrInvalidCEP):
		statusCode = http.StatusUnprocessableEntity
		message = service.ErrInvalidCEP.Error()
		log.Printf("[ORCHESTRATOR] Invalid CEP error: %v", err)
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
