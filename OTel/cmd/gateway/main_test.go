package main

import (
	"os"
	"testing"

	"otel/internal/gateway"

	"github.com/gorilla/mux"
)

func TestMain_EnvironmentVariables(t *testing.T) {
	tests := []struct {
		name                  string
		orchestrationURL      string
		port                  string
		expectedOrchestration string
		expectedPort          string
	}{
		{
			name:                  "Default values when env vars not set",
			orchestrationURL:      "",
			port:                  "",
			expectedOrchestration: "http://localhost:8081",
			expectedPort:          "8080",
		},
		{
			name:                  "Custom values from env vars",
			orchestrationURL:      "http://custom-service:9000",
			port:                  "3000",
			expectedOrchestration: "http://custom-service:9000",
			expectedPort:          "3000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original env vars
			originalOrchestrationURL := os.Getenv("ORCHESTRATION_SERVICE_URL")
			originalPort := os.Getenv("PORT")

			// Set test env vars
			if tt.orchestrationURL != "" {
				os.Setenv("ORCHESTRATION_SERVICE_URL", tt.orchestrationURL)
			} else {
				os.Unsetenv("ORCHESTRATION_SERVICE_URL")
			}

			if tt.port != "" {
				os.Setenv("PORT", tt.port)
			} else {
				os.Unsetenv("PORT")
			}

			// Test environment variable reading
			orchestrationURL := os.Getenv("ORCHESTRATION_SERVICE_URL")
			if orchestrationURL == "" {
				orchestrationURL = "http://localhost:8081"
			}

			port := os.Getenv("PORT")
			if port == "" {
				port = "8080"
			}

			// Verify values
			if orchestrationURL != tt.expectedOrchestration {
				t.Errorf("Expected orchestration URL %s, got %s", tt.expectedOrchestration, orchestrationURL)
			}

			if port != tt.expectedPort {
				t.Errorf("Expected port %s, got %s", tt.expectedPort, port)
			}

			// Restore original env vars
			if originalOrchestrationURL != "" {
				os.Setenv("ORCHESTRATION_SERVICE_URL", originalOrchestrationURL)
			} else {
				os.Unsetenv("ORCHESTRATION_SERVICE_URL")
			}

			if originalPort != "" {
				os.Setenv("PORT", originalPort)
			} else {
				os.Unsetenv("PORT")
			}
		})
	}
}

func TestMainServerSetup(t *testing.T) {
	t.Run("Server setup doesn't panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Server setup panicked: %v", r)
			}
		}()

		// Set test environment
		os.Setenv("ORCHESTRATION_SERVICE_URL", "http://localhost:8081")
		os.Setenv("PORT", "8080")

		// Simulate the main function setup
		orchestrationURL := os.Getenv("ORCHESTRATION_SERVICE_URL")
		if orchestrationURL == "" {
			orchestrationURL = "http://localhost:8081"
		}

		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}

		// Initialize gateway handler
		gatewayHandler := gateway.NewGatewayHandler(orchestrationURL)
		if gatewayHandler == nil {
			t.Error("Gateway handler creation failed")
		}

		// Create router
		r := mux.NewRouter()
		r.HandleFunc("/cep", gatewayHandler.ProcessCEP).Methods("POST")
		r.HandleFunc("/health", gatewayHandler.HealthCheck).Methods("GET")

		if r == nil {
			t.Error("Router creation failed")
		}

		// Clean up
		os.Unsetenv("ORCHESTRATION_SERVICE_URL")
		os.Unsetenv("PORT")
	})
}
