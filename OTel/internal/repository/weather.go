package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"otel/internal/domain"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// WeatherAPIRepository handles communication with Weather API
type WeatherAPIRepository struct {
	client  *http.Client
	apiKey  string
	baseURL string
}

// NewWeatherAPIRepository creates a new Weather API repository
func NewWeatherAPIRepository(apiKey string) *WeatherAPIRepository {
	return &WeatherAPIRepository{
		client: &http.Client{
			Transport: otelhttp.NewTransport(http.DefaultTransport),
			Timeout:   10 * time.Second,
		},
		apiKey:  apiKey,
		baseURL: "https://api.weatherapi.com/v1",
	}
}

// GetWeatherByLocation fetches weather data from Weather API
func (r *WeatherAPIRepository) GetWeatherByLocation(location string) (*domain.WeatherAPIResponse, error) {
	// URL encode the location to handle special characters
	encodedLocation := url.QueryEscape(location)
	url := fmt.Sprintf("%s/current.json?key=%s&q=%s&aqi=no", r.baseURL, r.apiKey, encodedLocation)

	// Create request with context for tracing
	req, err := http.NewRequestWithContext(context.Background(), "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch weather data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("weather API returned status %d for location: %s", resp.StatusCode, location)
	}

	var weatherResp domain.WeatherAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&weatherResp); err != nil {
		return nil, fmt.Errorf("failed to decode weather response: %w", err)
	}

	return &weatherResp, nil
}
