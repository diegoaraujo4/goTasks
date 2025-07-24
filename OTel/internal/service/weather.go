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
	return &WeatherService{
		locationRepo:    locationRepo,
		weatherDataRepo: weatherDataRepo,
	}
}

// GetWeatherByCEP gets weather information for a given CEP
func (s *WeatherService) GetWeatherByCEP(cep string) (*domain.WeatherResponse, error) {
	// Validate CEP format
	if !validator.ValidateCEP(cep) {
		return nil, ErrInvalidCEP
	}

	// Clean CEP (remove dashes and spaces)
	cleanCEP := validator.CleanCEP(cep)

	// Get location by CEP
	location, err := s.locationRepo.GetLocationByCEP(cleanCEP)
	if err != nil {
		log.Printf("Error fetching location for CEP %s: %v", cleanCEP, err)
		return nil, ErrCEPNotFound
	}

	// Get weather data for the location
	locationQuery := fmt.Sprintf("%s,%s", location.Localidade, location.UF)
	log.Printf("Fetching weather for location: %s", locationQuery)
	weather, err := s.weatherDataRepo.GetWeatherByLocation(locationQuery)
	if err != nil {
		log.Printf("Error fetching weather for location %s: %v", locationQuery, err)
		return nil, ErrWeatherDataUnavailable
	}

	// Convert temperatures
	tempC := weather.Current.TempC
	tempF := temperature.ConvertCelsiusToFahrenheit(tempC)
	tempK := temperature.ConvertCelsiusToKelvin(tempC)

	return &domain.WeatherResponse{
		TempC: tempC,
		TempF: tempF,
		TempK: tempK,
	}, nil
}
