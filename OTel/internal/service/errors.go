package service

import "errors"

var (
	// ErrInvalidCEP is returned when the CEP format is invalid
	// NOTE: CEP validation is now handled by the Gateway service
	// ErrInvalidCEP = errors.New("invalid zipcode")

	// ErrCEPNotFound is returned when the CEP is not found
	ErrCEPNotFound = errors.New("can not find zipcode")

	// ErrWeatherDataUnavailable is returned when weather data cannot be retrieved
	ErrWeatherDataUnavailable = errors.New("error fetching weather data")
)
