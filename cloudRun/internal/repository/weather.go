package repository

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"cloudrun/internal/domain"
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
			Timeout: 10 * time.Second,
		},
		apiKey:  apiKey,
		baseURL: "http://api.weatherapi.com/v1",
	}
}

// GetWeatherByLocation fetches weather data from Weather API
func (r *WeatherAPIRepository) GetWeatherByLocation(location string) (*domain.WeatherAPIResponse, error) {
	url := fmt.Sprintf("%s/current.json?key=%s&q=%s&aqi=no", r.baseURL, r.apiKey, location)

	resp, err := r.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch weather data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("weather API returned status %d", resp.StatusCode)
	}

	var weatherResp domain.WeatherAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&weatherResp); err != nil {
		return nil, fmt.Errorf("failed to decode weather response: %w", err)
	}

	return &weatherResp, nil
}
