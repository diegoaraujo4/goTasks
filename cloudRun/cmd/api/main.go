package main

import (
	"log"
	"net/http"

	_ "cloudrun/docs" // Import docs for swagger

	"cloudrun/config"
	"cloudrun/internal/handler"
	"cloudrun/internal/repository"
	"cloudrun/internal/service"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title Weather API
// @version 1.0
// @description API para consulta de temperatura por CEP brasileiro
// @description Recebe um CEP válido e retorna a temperatura atual em Celsius, Fahrenheit e Kelvin.
// @termsOfService http://swagger.io/terms/

// @contact.name Suporte da API
// @contact.email support@weatherapi.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /
// @schemes http https

// @tag.name weather
// @tag.description Operações relacionadas ao clima

// @tag.name health
// @tag.description Health check da aplicação

func main() {
	// Load configuration
	cfg := config.New()
	if err := cfg.Validate(); err != nil {
		log.Fatal(err)
	}

	// Initialize repositories
	locationRepo := repository.NewViaCEPRepository()
	weatherRepo := repository.NewWeatherAPIRepository(cfg.WeatherAPIKey)

	// Initialize services
	weatherService := service.NewWeatherService(locationRepo, weatherRepo)

	// Initialize handlers
	weatherHandler := handler.NewWeatherHandler(weatherService)
	healthHandler := handler.NewHealthHandler()

	// Setup router
	r := mux.NewRouter()

	// API endpoints
	r.HandleFunc("/weather/{cep}", weatherHandler.GetWeatherByCEP).Methods("GET")
	r.HandleFunc("/health", healthHandler.HealthCheck).Methods("GET")

	// Swagger documentation
	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	log.Printf("Server starting on port %s", cfg.Port)
	log.Printf("Swagger documentation available at: http://localhost:%s/swagger/index.html", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, r))
}
