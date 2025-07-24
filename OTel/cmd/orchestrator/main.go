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

	"otel/config"
	"otel/internal/handler"
	"otel/internal/repository"
	"otel/internal/service"
	"otel/pkg/telemetry"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
)

// @title OTEL Orchestration Service
// @version 1.0
// @description Serviço de orquestração para consulta de temperatura por CEP brasileiro
// @description Recebe um CEP válido e retorna a temperatura atual em Celsius, Fahrenheit e Kelvin.
// @termsOfService http://swagger.io/terms/

// @contact.name OTEL Orchestration Support
// @contact.email support@otel-orchestration.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8081
// @BasePath /
// @schemes http https

// @tag.name weather
// @tag.description Operações relacionadas ao clima

// @tag.name health
// @tag.description Health check da aplicação

func main() {
	log.Printf("[MAIN] Starting OTEL Orchestration Service...")

	// Initialize OpenTelemetry tracing
	zipkinURL := os.Getenv("ZIPKIN_URL")
	if zipkinURL == "" {
		zipkinURL = "http://localhost:9411/api/v2/spans" // Default Zipkin URL
	}

	shutdown, err := telemetry.InitTracer("otel-orchestration", zipkinURL)
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

	// Load configuration
	log.Printf("[MAIN] Loading configuration...")
	cfg := config.New()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("[MAIN] Configuration validation failed: %v", err)
	}
	log.Printf("[MAIN] Configuration loaded successfully - Port: %s", cfg.Port)

	// Initialize repositories
	log.Printf("[MAIN] Initializing repositories...")
	locationRepo := repository.NewViaCEPRepository()
	weatherRepo := repository.NewWeatherAPIRepository(cfg.WeatherAPIKey)
	log.Printf("[MAIN] Repositories initialized successfully")

	// Initialize services
	log.Printf("[MAIN] Initializing services...")
	weatherService := service.NewWeatherService(locationRepo, weatherRepo)
	log.Printf("[MAIN] Services initialized successfully")

	// Initialize handlers
	log.Printf("[MAIN] Initializing handlers...")
	weatherHandler := handler.NewWeatherHandler(weatherService)
	healthHandler := handler.NewHealthHandler()
	log.Printf("[MAIN] Handlers initialized successfully")

	// Setup router
	log.Printf("[MAIN] Setting up routes...")
	r := mux.NewRouter()

	// Add OpenTelemetry middleware for automatic instrumentation
	r.Use(otelmux.Middleware("otel-orchestration"))

	// Add logging middleware
	r.Use(loggingMiddleware)

	// API endpoints
	r.HandleFunc("/weather/{cep}", weatherHandler.GetWeatherByCEP).Methods("GET")
	r.HandleFunc("/health", healthHandler.HealthCheck).Methods("GET")

	// Swagger documentation
	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	log.Printf("[MAIN] Routes configured: GET /weather/{cep}, GET /health, /swagger/")

	log.Printf("[MAIN] OTEL Orchestration Service starting on port %s", cfg.Port)
	log.Printf("[MAIN] Zipkin URL: %s", zipkinURL)
	log.Printf("[MAIN] Swagger documentation available at: http://localhost:%s/swagger/index.html", cfg.Port)
	log.Printf("[MAIN] Server ready to accept connections...")

	// Setup graceful shutdown
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
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
