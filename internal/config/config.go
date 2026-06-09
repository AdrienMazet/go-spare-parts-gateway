package config

import (
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

// Config contains application runtime configuration.
type Config struct {
	ServerPort        string   `envconfig:"SERVER_PORT" default:"8080"`
	DatabaseURL       string   `envconfig:"DATABASE_URL" default:"postgres://postgres:postgres@localhost:5432/spare_parts?sslmode=disable"`
	MigrationsPath    string   `envconfig:"MIGRATIONS_PATH" default:"internal/db/migrations"`
	OfferProviderURLs []string `envconfig:"OFFER_PROVIDER_URLS" default:"http://localhost:8081,http://localhost:8082,http://localhost:8083,http://localhost:8084"`
}

// Load reads configuration from .env and environment variables.
func Load() (Config, error) {
	_ = godotenv.Load()

	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}
