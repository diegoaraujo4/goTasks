package main

import (
	"log"
	"net/http"
	"os"

	"otel/internal/gateway"

	"github.com/gorilla/mux"
)

// @title OTEL Gateway Service
// @version 1.0
// @description Gateway service for CEP input validation and forwarding
// @description Validates CEP input and forwards to orchestration service.
// @termsOfService http://swagger.io/terms/

// @contact.name Gateway API Support
// @contact.email support@otel-gateway.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8081
// @BasePath /
// @schemes http https

// @tag.name gateway
// @tag.description CEP input processing operations

// @tag.name health
// @tag.description Health check operations

func main() {
	// Get orchestration service URL from environment
	orchestrationURL := os.Getenv("ORCHESTRATION_SERVICE_URL")
	if orchestrationURL == "" {
		orchestrationURL = "http://localhost:8080" // Default to local orchestration service
	}

	// Get port from environment
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081" // Default port for gateway
	}

	// Initialize gateway handler
	gatewayHandler := gateway.NewGatewayHandler(orchestrationURL)

	// Create router
	r := mux.NewRouter()

	// Gateway routes
	r.HandleFunc("/cep", gatewayHandler.ProcessCEP).Methods("POST")
	r.HandleFunc("/health", gatewayHandler.HealthCheck).Methods("GET")

	// CORS middleware
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	})

	log.Printf("OTEL Gateway Service starting on port %s", port)
	log.Printf("Orchestration service URL: %s", orchestrationURL)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
