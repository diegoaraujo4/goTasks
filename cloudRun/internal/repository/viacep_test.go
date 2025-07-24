package repository

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"cloudrun/internal/domain"
)

func TestNewViaCEPRepository(t *testing.T) {
	repo := NewViaCEPRepository()

	expectedBaseURL := "https://viacep.com.br/ws"
	if repo.baseURL != expectedBaseURL {
		t.Errorf("Expected base URL to be %s, got %s", expectedBaseURL, repo.baseURL)
	}

	if repo.client.Timeout.Seconds() != 10 {
		t.Errorf("Expected timeout to be 10 seconds, got %v", repo.client.Timeout.Seconds())
	}
}

func TestGetLocationByCEP_Success(t *testing.T) {
	// Mock server with successful ViaCEP response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the URL format
		expectedPath := "/01310100/json/"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path to be %s, got %s", expectedPath, r.URL.Path)
		}

		response := domain.ViaCEPResponse{
			CEP:        "01310-100",
			Logradouro: "Avenida Paulista",
			Bairro:     "Bela Vista",
			Localidade: "São Paulo",
			UF:         "SP",
			Erro:       false,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	repo := &ViaCEPRepository{
		client:  &http.Client{},
		baseURL: server.URL,
	}

	result, err := repo.GetLocationByCEP("01310100")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.CEP != "01310-100" {
		t.Errorf("Expected CEP to be '01310-100', got %s", result.CEP)
	}

	if result.Localidade != "São Paulo" {
		t.Errorf("Expected Localidade to be 'São Paulo', got %s", result.Localidade)
	}

	if result.UF != "SP" {
		t.Errorf("Expected UF to be 'SP', got %s", result.UF)
	}

	if result.Erro {
		t.Error("Expected Erro to be false")
	}
}

func TestGetLocationByCEP_NotFound(t *testing.T) {
	// Mock server returning CEP not found
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := domain.ViaCEPResponse{
			Erro: true,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	repo := &ViaCEPRepository{
		client:  &http.Client{},
		baseURL: server.URL,
	}

	_, err := repo.GetLocationByCEP("99999999")
	if err == nil {
		t.Fatal("Expected error for CEP not found")
	}

	if !strings.Contains(err.Error(), "CEP not found") {
		t.Errorf("Expected error to contain 'CEP not found', got %v", err.Error())
	}
}

func TestGetLocationByCEP_HTTPError(t *testing.T) {
	// Mock server returning HTTP error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	repo := &ViaCEPRepository{
		client:  &http.Client{},
		baseURL: server.URL,
	}

	_, err := repo.GetLocationByCEP("01310100")
	if err == nil {
		t.Fatal("Expected error for HTTP 500 response")
	}

	if !strings.Contains(err.Error(), "ViaCEP API returned status 500") {
		t.Errorf("Expected error to contain status 500, got %v", err.Error())
	}
}

func TestGetLocationByCEP_InvalidJSON(t *testing.T) {
	// Mock server returning invalid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{invalid json`))
	}))
	defer server.Close()

	repo := &ViaCEPRepository{
		client:  &http.Client{},
		baseURL: server.URL,
	}

	_, err := repo.GetLocationByCEP("01310100")
	if err == nil {
		t.Fatal("Expected error for invalid JSON response")
	}

	if !strings.Contains(err.Error(), "failed to decode ViaCEP response") {
		t.Errorf("Expected error to contain decode error message, got %v", err.Error())
	}
}

func TestGetLocationByCEP_NetworkError(t *testing.T) {
	// Use an invalid URL to simulate network error
	repo := &ViaCEPRepository{
		client:  &http.Client{},
		baseURL: "http://invalid-url-that-does-not-exist.local",
	}

	_, err := repo.GetLocationByCEP("01310100")
	if err == nil {
		t.Fatal("Expected network error")
	}

	if !strings.Contains(err.Error(), "failed to fetch location data") {
		t.Errorf("Expected error to contain fetch error message, got %v", err.Error())
	}
}

// Test that verifies URL construction with different CEPs
func TestGetLocationByCEP_URLConstruction(t *testing.T) {
	testCases := []struct {
		cep          string
		expectedPath string
		description  string
	}{
		{
			cep:          "01310100",
			expectedPath: "/01310100/json/",
			description:  "Standard CEP without dash",
		},
		{
			cep:          "12345678",
			expectedPath: "/12345678/json/",
			description:  "Different CEP format",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			var capturedPath string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedPath = r.URL.Path

				response := domain.ViaCEPResponse{
					CEP:        tc.cep,
					Localidade: "Test City",
					UF:         "TS",
					Erro:       false,
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
			}))
			defer server.Close()

			repo := &ViaCEPRepository{
				client:  &http.Client{},
				baseURL: server.URL,
			}

			_, err := repo.GetLocationByCEP(tc.cep)
			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}

			if capturedPath != tc.expectedPath {
				t.Errorf("Expected path to be %s, got %s", tc.expectedPath, capturedPath)
			}
		})
	}
}
