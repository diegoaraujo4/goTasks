package config

import (
	"os"
)

// Config holds all configuration for the application
type Config struct {
	WeatherAPIKey string
	Port          string
}

// New creates a new configuration instance
func New() *Config {
	return &Config{
		WeatherAPIKey: getEnv("WEATHER_API_KEY", ""),
		Port:          getEnv("PORT", "8081"),
	}
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.WeatherAPIKey == "" {
		return ErrMissingWeatherAPIKey
	}
	return nil
}
