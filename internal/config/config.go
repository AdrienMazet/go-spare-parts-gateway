package config

import (
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

// Config contains application runtime configuration.
type Config struct {
	ServerPort              string   `envconfig:"SERVER_PORT" default:"8080"`
	LogLevel                string   `envconfig:"LOG_LEVEL" default:"info"`
	DatabaseURL             string   `envconfig:"DATABASE_URL" default:"postgres://postgres:postgres@localhost:5432/spare_parts?sslmode=disable"`
	MigrationsPath          string   `envconfig:"MIGRATIONS_PATH" default:"internal/db/migrations"`
	OfferProviderURLs       []string `envconfig:"OFFER_PROVIDER_URLS" default:"http://localhost:8081,http://localhost:8082,http://localhost:8083,http://localhost:8084"`
	KafkaBrokers            []string `envconfig:"KAFKA_BROKERS" default:"localhost:9096"`
	KafkaOfferFetchedTopic  string   `envconfig:"KAFKA_OFFER_FETCHED_TOPIC" default:"offer-fetched"`
	KafkaOfferConsumerGroup string   `envconfig:"KAFKA_OFFER_CONSUMER_GROUP" default:"offer-price-worker"`
	OTLPEndpoint            string   `envconfig:"OTEL_EXPORTER_OTLP_ENDPOINT"`
	WorkerMetricsPort       string   `envconfig:"WORKER_METRICS_PORT" default:"9091"`
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
