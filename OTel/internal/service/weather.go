package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"otel/internal/domain"
	"otel/pkg/telemetry"
	"otel/pkg/temperature"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// WeatherService implements the weather service business logic
type WeatherService struct {
	locationRepo    domain.LocationService
	weatherDataRepo domain.WeatherDataService
	tracer          trace.Tracer
}

// NewWeatherService creates a new weather service
func NewWeatherService(locationRepo domain.LocationService, weatherDataRepo domain.WeatherDataService) *WeatherService {
	log.Printf("[ORCHESTRATOR] Initializing weather service")
	return &WeatherService{
		locationRepo:    locationRepo,
		weatherDataRepo: weatherDataRepo,
		tracer:          telemetry.GetTracer("weather-service"),
	}
}

// GetWeatherByCEP gets weather information for a given CEP
func (s *WeatherService) GetWeatherByCEP(ctx context.Context, cep string) (*domain.WeatherResponse, error) {
	// Start span for the entire weather service operation
	ctx, span := s.tracer.Start(ctx, "weather_service.get_weather_by_cep")
	defer span.End()

	span.SetAttributes(attribute.String("cep.input", cep))
	log.Printf("[ORCHESTRATOR] Starting weather service for CEP: %s", cep)

	// Note: CEP validation is handled by the Gateway service
	// The CEP received here is already validated and formatted

	// Get location by CEP
	log.Printf("[ORCHESTRATOR] Fetching location for CEP: %s", cep)
	locationStart := time.Now()
	_, locationSpan := s.tracer.Start(ctx, "weather_service.get_location_by_cep")

	location, err := s.locationRepo.GetLocationByCEP(cep)
	locationDuration := time.Since(locationStart)

	if err != nil {
		log.Printf("[ORCHESTRATOR] Error fetching location for CEP %s: %v", cep, err)
		locationSpan.SetStatus(codes.Error, "Failed to fetch location")
		locationSpan.RecordError(err)
		locationSpan.End()
		span.SetStatus(codes.Error, "Failed to fetch location")
		span.RecordError(err)
		return nil, ErrCEPNotFound
	}

	locationSpan.SetAttributes(
		attribute.String("location.city", location.Localidade),
		attribute.String("location.state", location.UF),
		attribute.Int64("location.fetch_duration_ms", locationDuration.Milliseconds()),
	)
	locationSpan.SetStatus(codes.Ok, "Location fetched successfully")
	locationSpan.End()

	log.Printf("[ORCHESTRATOR] Location found: %s, %s", location.Localidade, location.UF)

	// Get weather data for the location
	locationQuery := fmt.Sprintf("%s,%s", location.Localidade, location.UF)
	log.Printf("[ORCHESTRATOR] Fetching weather for location: %s", locationQuery)

	weatherStart := time.Now()
	_, weatherSpan := s.tracer.Start(ctx, "weather_service.get_weather_by_location")

	weather, err := s.weatherDataRepo.GetWeatherByLocation(locationQuery)
	weatherDuration := time.Since(weatherStart)

	if err != nil {
		log.Printf("[ORCHESTRATOR] Error fetching weather for location %s: %v", locationQuery, err)
		weatherSpan.SetStatus(codes.Error, "Failed to fetch weather data")
		weatherSpan.RecordError(err)
		weatherSpan.End()
		span.SetStatus(codes.Error, "Failed to fetch weather data")
		span.RecordError(err)
		return nil, ErrWeatherDataUnavailable
	}

	weatherSpan.SetAttributes(
		attribute.String("weather.location_query", locationQuery),
		attribute.Float64("weather.temp_c_raw", weather.Current.TempC),
		attribute.Int64("weather.fetch_duration_ms", weatherDuration.Milliseconds()),
	)
	weatherSpan.SetStatus(codes.Ok, "Weather data fetched successfully")
	weatherSpan.End()

	log.Printf("[ORCHESTRATOR] Weather data fetched successfully - Temperature: %.1f°C", weather.Current.TempC)

	// Convert temperatures
	_, conversionSpan := s.tracer.Start(ctx, "weather_service.convert_temperatures")
	tempC := weather.Current.TempC
	tempF := temperature.ConvertCelsiusToFahrenheit(tempC)
	tempK := temperature.ConvertCelsiusToKelvin(tempC)

	conversionSpan.SetAttributes(
		attribute.Float64("temperature.celsius", tempC),
		attribute.Float64("temperature.fahrenheit", tempF),
		attribute.Float64("temperature.kelvin", tempK),
	)
	conversionSpan.SetStatus(codes.Ok, "Temperature conversion completed")
	conversionSpan.End()

	log.Printf("[ORCHESTRATOR] Temperature conversions - C: %.1f, F: %.1f, K: %.1f", tempC, tempF, tempK)

	response := &domain.WeatherResponse{
		City:  location.Localidade,
		TempC: tempC,
		TempF: tempF,
		TempK: tempK,
	}

	span.SetAttributes(
		attribute.String("response.city", response.City),
		attribute.Float64("response.temp_c", response.TempC),
		attribute.Float64("response.temp_f", response.TempF),
		attribute.Float64("response.temp_k", response.TempK),
	)
	span.SetStatus(codes.Ok, "Weather service completed successfully")

	log.Printf("[ORCHESTRATOR] Weather service completed successfully for CEP: %s", cep)
	return response, nil
}
