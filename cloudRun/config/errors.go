package config

import "errors"

var (
	// ErrMissingWeatherAPIKey is returned when the weather API key is not configured
	ErrMissingWeatherAPIKey = errors.New("WEATHER_API_KEY environment variable is required")
)
