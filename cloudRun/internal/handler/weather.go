package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"cloudrun/internal/domain"
	"cloudrun/internal/service"

	"github.com/gorilla/mux"
)

// WeatherHandler handles HTTP requests for weather endpoints
type WeatherHandler struct {
	weatherService *service.WeatherService
}

// NewWeatherHandler creates a new weather handler
func NewWeatherHandler(weatherService *service.WeatherService) *WeatherHandler {
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
	vars := mux.Vars(r)
	cep := vars["cep"]

	weather, err := h.weatherService.GetWeatherByCEP(cep)
	if err != nil {
		h.handleError(w, err)
		return
	}

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
	case errors.Is(err, service.ErrCEPNotFound):
		statusCode = http.StatusNotFound
		message = service.ErrCEPNotFound.Error()
	case errors.Is(err, service.ErrWeatherDataUnavailable):
		statusCode = http.StatusInternalServerError
		message = service.ErrWeatherDataUnavailable.Error()
	default:
		statusCode = http.StatusInternalServerError
		message = "internal server error"
	}

	errorResponse := domain.ErrorResponse{Message: message}
	h.sendJSON(w, statusCode, errorResponse)
}

// sendJSON sends a JSON response
func (h *WeatherHandler) sendJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}
