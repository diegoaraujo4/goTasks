package main

import (
	"log"
	"net/http"
	"os"
	"time"

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

// @host localhost:8080
// @BasePath /
// @schemes http https

// @tag.name gateway
// @tag.description CEP input processing operations

// @tag.name health
// @tag.description Health check operations

func main() {
	log.Printf("[MAIN] Starting OTEL Gateway Service...")

	// Get orchestration service URL from environment
	orchestrationURL := os.Getenv("ORCHESTRATION_SERVICE_URL")
	if orchestrationURL == "" {
		orchestrationURL = "http://localhost:8081" // Default to local orchestration service
		log.Printf("[MAIN] Using default orchestration URL: %s", orchestrationURL)
	} else {
		log.Printf("[MAIN] Using orchestration URL from environment: %s", orchestrationURL)
	}

	// Get port from environment
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port for gateway
		log.Printf("[MAIN] Using default port: %s", port)
	} else {
		log.Printf("[MAIN] Using port from environment: %s", port)
	}

	// Initialize gateway handler
	log.Printf("[MAIN] Initializing gateway handler...")
	gatewayHandler := gateway.NewGatewayHandler(orchestrationURL)

	// Create router
	log.Printf("[MAIN] Setting up routes...")
	r := mux.NewRouter()

	// Add logging middleware
	r.Use(loggingMiddleware)

	// Gateway routes
	r.HandleFunc("/cep", gatewayHandler.ProcessCEP).Methods("POST")
	r.HandleFunc("/health", gatewayHandler.HealthCheck).Methods("GET")

	log.Printf("[MAIN] Routes configured: POST /cep, GET /health")

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

	log.Printf("[MAIN] OTEL Gateway Service starting on port %s", port)
	log.Printf("[MAIN] Orchestration service URL: %s", orchestrationURL)
	log.Printf("[MAIN] Server ready to accept connections...")
	log.Fatal(http.ListenAndServe(":"+port, r))
}

// loggingMiddleware logs all incoming requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		clientIP := r.RemoteAddr
		if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
			clientIP = forwarded
		}

		log.Printf("[REQUEST] %s %s from %s - User-Agent: %s",
			r.Method, r.URL.Path, clientIP, r.Header.Get("User-Agent"))

		// Create a custom ResponseWriter to capture status code
		lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(lrw, r)

		duration := time.Since(start)
		log.Printf("[RESPONSE] %s %s - Status: %d, Duration: %v, Client: %s",
			r.Method, r.URL.Path, lrw.statusCode, duration, clientIP)
	})
}

// loggingResponseWriter wraps http.ResponseWriter to capture status code
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}
