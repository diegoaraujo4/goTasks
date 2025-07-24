package service

import (
	"fmt"
	"log"

	"otel/internal/domain"
	"otel/pkg/temperature"
	"otel/pkg/validator"
)

// WeatherService implements the weather service business logic
type WeatherService struct {
	locationRepo    domain.LocationService
	weatherDataRepo domain.WeatherDataService
}

// NewWeatherService creates a new weather service
func NewWeatherService(locationRepo domain.LocationService, weatherDataRepo domain.WeatherDataService) *WeatherService {
	log.Printf("[ORCHESTRATOR] Initializing weather service")
	return &WeatherService{
		locationRepo:    locationRepo,
		weatherDataRepo: weatherDataRepo,
	}
}

// GetWeatherByCEP gets weather information for a given CEP
func (s *WeatherService) GetWeatherByCEP(cep string) (*domain.WeatherResponse, error) {
	log.Printf("[ORCHESTRATOR] Starting weather service for CEP: %s", cep)

	// Validate CEP format
	if !validator.ValidateCEP(cep) {
		log.Printf("[ORCHESTRATOR] Invalid CEP format: %s", cep)
		return nil, ErrInvalidCEP
	}

	// Clean CEP (remove dashes and spaces)
	cleanCEP := validator.CleanCEP(cep)
	log.Printf("[ORCHESTRATOR] Cleaned CEP: %s -> %s", cep, cleanCEP)

	// Get location by CEP
	log.Printf("[ORCHESTRATOR] Fetching location for CEP: %s", cleanCEP)
	location, err := s.locationRepo.GetLocationByCEP(cleanCEP)
	if err != nil {
		log.Printf("[ORCHESTRATOR] Error fetching location for CEP %s: %v", cleanCEP, err)
		return nil, ErrCEPNotFound
	}
	log.Printf("[ORCHESTRATOR] Location found: %s, %s", location.Localidade, location.UF)

	// Get weather data for the location
	locationQuery := fmt.Sprintf("%s,%s", location.Localidade, location.UF)
	log.Printf("[ORCHESTRATOR] Fetching weather for location: %s", locationQuery)
	weather, err := s.weatherDataRepo.GetWeatherByLocation(locationQuery)
	if err != nil {
		log.Printf("[ORCHESTRATOR] Error fetching weather for location %s: %v", locationQuery, err)
		return nil, ErrWeatherDataUnavailable
	}
	log.Printf("[ORCHESTRATOR] Weather data fetched successfully - Temperature: %.1f°C", weather.Current.TempC)

	// Convert temperatures
	tempC := weather.Current.TempC
	tempF := temperature.ConvertCelsiusToFahrenheit(tempC)
	tempK := temperature.ConvertCelsiusToKelvin(tempC)

	log.Printf("[ORCHESTRATOR] Temperature conversions - C: %.1f, F: %.1f, K: %.1f", tempC, tempF, tempK)

	response := &domain.WeatherResponse{
		City:  location.Localidade,
		TempC: tempC,
		TempF: tempF,
		TempK: tempK,
	}

	log.Printf("[ORCHESTRATOR] Weather service completed successfully for CEP: %s", cep)
	return response, nil
}
