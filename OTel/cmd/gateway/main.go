package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "otel/docs" // Import docs for swagger
	"otel/internal/gateway"
	"otel/pkg/telemetry"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
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

	// Initialize OpenTelemetry tracing
	zipkinURL := os.Getenv("ZIPKIN_URL")
	if zipkinURL == "" {
		zipkinURL = "http://localhost:9411/api/v2/spans" // Default Zipkin URL
	}

	shutdown, err := telemetry.InitTracer("otel-gateway", zipkinURL)
	if err != nil {
		log.Fatalf("[MAIN] Failed to initialize tracer: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := shutdown(ctx); err != nil {
			log.Printf("[MAIN] Error shutting down tracer: %v", err)
		}
	}()

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

	// Add OpenTelemetry middleware for automatic instrumentation
	r.Use(otelmux.Middleware("otel-gateway"))

	// Add logging middleware
	r.Use(loggingMiddleware)

	// Gateway routes
	r.HandleFunc("/cep", gatewayHandler.ProcessCEP).Methods("POST")
	r.HandleFunc("/health", gatewayHandler.HealthCheck).Methods("GET")

	// Swagger documentation
	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	log.Printf("[MAIN] Routes configured: POST /cep, GET /health, /swagger/")

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
	log.Printf("[MAIN] Zipkin URL: %s", zipkinURL)
	log.Printf("[MAIN] Swagger documentation available at: http://localhost:%s/swagger/index.html", port)
	log.Printf("[MAIN] Server ready to accept connections...")

	// Setup graceful shutdown
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// Channel to listen for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[MAIN] Server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	<-c
	log.Printf("[MAIN] Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("[MAIN] Server shutdown error: %v", err)
	}

	log.Printf("[MAIN] Server shutdown complete")
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
