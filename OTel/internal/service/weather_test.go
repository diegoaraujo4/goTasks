package service

import (
	"context"
	"testing"

	"otel/internal/domain"
)

// MockLocationRepo for testing
type MockLocationRepo struct {
	shouldFail bool
}

func (m *MockLocationRepo) GetLocationByCEP(cep string) (*domain.ViaCEPResponse, error) {
	if m.shouldFail {
		return nil, ErrCEPNotFound
	}

	// Return different cities to test URL encoding scenarios
	switch cep {
	case "01310100":
		return &domain.ViaCEPResponse{
			CEP:        "01310-100",
			Localidade: "São Paulo", // Special characters
			UF:         "SP",
			Erro:       false,
		}, nil
	case "20040020":
		return &domain.ViaCEPResponse{
			CEP:        "20040-020",
			Localidade: "Rio de Janeiro", // Space in name
			UF:         "RJ",
			Erro:       false,
		}, nil
	case "30112000":
		return &domain.ViaCEPResponse{
			CEP:        "30112-000",
			Localidade: "Belo Horizonte",
			UF:         "MG",
			Erro:       false,
		}, nil
	}
	return nil, ErrCEPNotFound
}

// MockWeatherRepo for testing
type MockWeatherRepo struct {
	shouldFail bool
}

func (m *MockWeatherRepo) GetWeatherByLocation(location string) (*domain.WeatherAPIResponse, error) {
	if m.shouldFail {
		return nil, ErrWeatherDataUnavailable
	}

	// Verify that different location formats are handled
	tempMap := map[string]float64{
		"São Paulo,SP":      25.5,
		"Rio de Janeiro,RJ": 28.0,
		"Belo Horizonte,MG": 22.0,
	}

	if temp, exists := tempMap[location]; exists {
		return &domain.WeatherAPIResponse{
			Current: struct {
				TempC float64 `json:"temp_c"`
			}{
				TempC: temp,
			},
		}, nil
	}

	return nil, ErrWeatherDataUnavailable
}

func TestWeatherService_GetWeatherByCEP_Success(t *testing.T) {
	locationRepo := &MockLocationRepo{}
	weatherRepo := &MockWeatherRepo{}
	service := NewWeatherService(locationRepo, weatherRepo)

	testCases := []struct {
		cep          string
		expectedTemp float64
		expectedCity string
		description  string
	}{
		{
			cep:          "01310100",
			expectedTemp: 25.5,
			expectedCity: "São Paulo",
			description:  "São Paulo with special characters",
		},
		{
			cep:          "20040020",
			expectedTemp: 28.0,
			expectedCity: "Rio de Janeiro",
			description:  "Rio de Janeiro with spaces",
		},
		{
			cep:          "30112000",
			expectedTemp: 22.0,
			expectedCity: "Belo Horizonte",
			description:  "Belo Horizonte normal name",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			result, err := service.GetWeatherByCEP(context.TODO(), tc.cep)
			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}

			if result.TempC != tc.expectedTemp {
				t.Errorf("Expected temp_C to be %v, got %v", tc.expectedTemp, result.TempC)
			}

			if result.City != tc.expectedCity {
				t.Errorf("Expected city to be %v, got %v", tc.expectedCity, result.City)
			}

			// Verify temperature conversions
			expectedTempF := (tc.expectedTemp * 9 / 5) + 32
			if result.TempF != expectedTempF {
				t.Errorf("Expected temp_F to be %v, got %v", expectedTempF, result.TempF)
			}

			expectedTempK := tc.expectedTemp + 273
			if result.TempK != expectedTempK {
				t.Errorf("Expected temp_K to be %v, got %v", expectedTempK, result.TempK)
			}
		})
	}
}

func TestWeatherService_GetWeatherByCEP_CEPNotFound(t *testing.T) {
	locationRepo := &MockLocationRepo{shouldFail: true}
	weatherRepo := &MockWeatherRepo{}
	service := NewWeatherService(locationRepo, weatherRepo)

	_, err := service.GetWeatherByCEP(context.TODO(), "99999999")
	if err != ErrCEPNotFound {
		t.Errorf("Expected ErrCEPNotFound, got %v", err)
	}
}

func TestWeatherService_GetWeatherByCEP_WeatherDataUnavailable(t *testing.T) {
	locationRepo := &MockLocationRepo{}
	weatherRepo := &MockWeatherRepo{shouldFail: true}
	service := NewWeatherService(locationRepo, weatherRepo)

	_, err := service.GetWeatherByCEP(context.TODO(), "01310100")
	if err != ErrWeatherDataUnavailable {
		t.Errorf("Expected ErrWeatherDataUnavailable, got %v", err)
	}
}

// Test to ensure our fixes for location name handling work
func TestWeatherService_LocationQueryConstruction(t *testing.T) {
	// This test verifies that the service processes different CEPs correctly
	// The actual URL encoding is tested in the repository tests
	locationRepo := &MockLocationRepo{}
	weatherRepo := &MockWeatherRepo{}
	service := NewWeatherService(locationRepo, weatherRepo)

	testCases := []struct {
		cep         string
		description string
	}{
		{
			cep:         "01310100",
			description: "CEP for São Paulo (special characters)",
		},
		{
			cep:         "20040020",
			description: "CEP for Rio de Janeiro (spaces in name)",
		},
		{
			cep:         "30112000",
			description: "CEP for Belo Horizonte (normal name)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			result, err := service.GetWeatherByCEP(context.TODO(), tc.cep)
			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}

			// Verify we get a valid response for all location types
			if result.TempC <= 0 {
				t.Error("Expected positive temperature")
			}

			if result.TempF <= result.TempC {
				t.Error("Expected Fahrenheit to be higher than Celsius for positive temps")
			}

			if result.TempK <= result.TempC {
				t.Error("Expected Kelvin to be higher than Celsius")
			}
		})
	}
}
