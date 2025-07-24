package repository

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"otel/internal/domain"
)

func TestNewWeatherAPIRepository(t *testing.T) {
	apiKey := "test_api_key"
	repo := NewWeatherAPIRepository(apiKey)

	if repo.apiKey != apiKey {
		t.Errorf("Expected API key to be %s, got %s", apiKey, repo.apiKey)
	}

	// Test that HTTPS is used (this was the main issue we fixed)
	expectedBaseURL := "https://api.weatherapi.com/v1"
	if repo.baseURL != expectedBaseURL {
		t.Errorf("Expected base URL to be %s, got %s", expectedBaseURL, repo.baseURL)
	}

	if repo.client.Timeout.Seconds() != 10 {
		t.Errorf("Expected timeout to be 10 seconds, got %v", repo.client.Timeout.Seconds())
	}
}

func TestGetWeatherByLocation_URLEncoding(t *testing.T) {
	// Test server that captures the request URL
	var capturedURL string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL.String()

		// Return a valid weather response
		response := domain.WeatherAPIResponse{
			Current: struct {
				TempC float64 `json:"temp_c"`
			}{
				TempC: 25.0,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create repository with test server URL
	repo := &WeatherAPIRepository{
		client:  &http.Client{},
		apiKey:  "test_key",
		baseURL: server.URL,
	}

	// Test with location containing special characters (São Paulo)
	location := "São Paulo,SP"
	_, err := repo.GetWeatherByLocation(location)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify that the location was properly URL encoded
	// Note: Go's url.QueryEscape encodes spaces as + and special chars correctly
	if !strings.Contains(capturedURL, "S%C3%A3o+Paulo%2CSP") {
		t.Errorf("Expected URL to contain URL-encoded location 'S%%C3%%A3o+Paulo%%2CSP', got %s", capturedURL)
	}

	// Verify other URL components
	if !strings.Contains(capturedURL, "key=test_key") {
		t.Errorf("Expected URL to contain API key, got %s", capturedURL)
	}

	if !strings.Contains(capturedURL, "aqi=no") {
		t.Errorf("Expected URL to contain aqi=no parameter, got %s", capturedURL)
	}

	if !strings.Contains(capturedURL, "/current.json") {
		t.Errorf("Expected URL to contain /current.json endpoint, got %s", capturedURL)
	}
}

func TestGetWeatherByLocation_Success(t *testing.T) {
	// Mock server with successful response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := domain.WeatherAPIResponse{
			Current: struct {
				TempC float64 `json:"temp_c"`
			}{
				TempC: 22.5,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	repo := &WeatherAPIRepository{
		client:  &http.Client{},
		apiKey:  "test_key",
		baseURL: server.URL,
	}

	result, err := repo.GetWeatherByLocation("Test Location")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.Current.TempC != 22.5 {
		t.Errorf("Expected temperature to be 22.5, got %v", result.Current.TempC)
	}
}

func TestGetWeatherByLocation_HTTPError(t *testing.T) {
	// Mock server that returns HTTP error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": {"message": "API key invalid"}}`))
	}))
	defer server.Close()

	repo := &WeatherAPIRepository{
		client:  &http.Client{},
		apiKey:  "invalid_key",
		baseURL: server.URL,
	}

	_, err := repo.GetWeatherByLocation("Test Location")
	if err == nil {
		t.Fatal("Expected error for HTTP 401 response")
	}

	if !strings.Contains(err.Error(), "weather API returned status 401") {
		t.Errorf("Expected error to contain status 401, got %v", err.Error())
	}

	if !strings.Contains(err.Error(), "Test Location") {
		t.Errorf("Expected error to contain location name, got %v", err.Error())
	}
}

func TestGetWeatherByLocation_InvalidJSON(t *testing.T) {
	// Mock server that returns invalid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{invalid json`))
	}))
	defer server.Close()

	repo := &WeatherAPIRepository{
		client:  &http.Client{},
		apiKey:  "test_key",
		baseURL: server.URL,
	}

	_, err := repo.GetWeatherByLocation("Test Location")
	if err == nil {
		t.Fatal("Expected error for invalid JSON response")
	}

	if !strings.Contains(err.Error(), "failed to decode weather response") {
		t.Errorf("Expected error to contain decode error message, got %v", err.Error())
	}
}

func TestGetWeatherByLocation_NetworkError(t *testing.T) {
	// Use an invalid URL to simulate network error
	repo := &WeatherAPIRepository{
		client:  &http.Client{},
		apiKey:  "test_key",
		baseURL: "http://invalid-url-that-does-not-exist.local",
	}

	_, err := repo.GetWeatherByLocation("Test Location")
	if err == nil {
		t.Fatal("Expected network error")
	}

	if !strings.Contains(err.Error(), "failed to fetch weather data") {
		t.Errorf("Expected error to contain fetch error message, got %v", err.Error())
	}
}

// Test that verifies we're using HTTPS (regression test for the main issue we fixed)
func TestWeatherAPIRepository_UsesHTTPS(t *testing.T) {
	repo := NewWeatherAPIRepository("test_key")

	if !strings.HasPrefix(repo.baseURL, "https://") {
		t.Errorf("Expected base URL to use HTTPS, got %s", repo.baseURL)
	}

	expectedURL := "https://api.weatherapi.com/v1"
	if repo.baseURL != expectedURL {
		t.Errorf("Expected base URL to be %s, got %s", expectedURL, repo.baseURL)
	}
}

// Test for various special characters that might need URL encoding
func TestGetWeatherByLocation_SpecialCharactersEncoding(t *testing.T) {
	testCases := []struct {
		location        string
		expectedEncoded string
		description     string
	}{
		{
			location:        "São Paulo,SP",
			expectedEncoded: "S%C3%A3o+Paulo%2CSP",
			description:     "Portuguese characters with tilde",
		},
		{
			location:        "Ribeirão Preto,SP",
			expectedEncoded: "Ribeir%C3%A3o+Preto%2CSP",
			description:     "Portuguese characters with tilde and ã",
		},
		{
			location:        "Brasília,DF",
			expectedEncoded: "Bras%C3%ADlia%2CDF",
			description:     "Portuguese characters with í",
		},
		{
			location:        "Location with spaces",
			expectedEncoded: "Location+with+spaces",
			description:     "Spaces should be encoded as plus signs",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			var capturedURL string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedURL = r.URL.String()

				response := domain.WeatherAPIResponse{
					Current: struct {
						TempC float64 `json:"temp_c"`
					}{
						TempC: 20.0,
					},
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
			}))
			defer server.Close()

			repo := &WeatherAPIRepository{
				client:  &http.Client{},
				apiKey:  "test_key",
				baseURL: server.URL,
			}

			_, err := repo.GetWeatherByLocation(tc.location)
			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}

			if !strings.Contains(capturedURL, tc.expectedEncoded) {
				t.Errorf("Expected URL to contain encoded location '%s', got %s", tc.expectedEncoded, capturedURL)
			}
		})
	}
}
