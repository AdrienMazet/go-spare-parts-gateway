package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	nethttpmiddleware "github.com/oapi-codegen/nethttp-middleware"

	"github.com/adrienmazet/go-spare-parts-gateway/api"
	"github.com/adrienmazet/go-spare-parts-gateway/internal/config"
	"github.com/adrienmazet/go-spare-parts-gateway/internal/controllers"
	"github.com/adrienmazet/go-spare-parts-gateway/internal/db"
	"github.com/adrienmazet/go-spare-parts-gateway/internal/messaging"
	"github.com/adrienmazet/go-spare-parts-gateway/internal/offer"
	"github.com/adrienmazet/go-spare-parts-gateway/internal/repository"
	"github.com/adrienmazet/go-spare-parts-gateway/internal/service"
	"github.com/adrienmazet/go-spare-parts-gateway/internal/service/mapper"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	configureLogger(cfg.LogLevel)
	slog.Info("starting spare parts gateway", "port", cfg.ServerPort, "offer_provider_count", len(cfg.OfferProviderURLs))

	if err := db.Migrate(cfg.DatabaseURL, cfg.MigrationsPath); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}
	slog.Info("database migrations applied", "path", cfg.MigrationsPath)

	database, err := db.Open(cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to open database", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := database.Close(); err != nil {
			slog.Warn("failed to close database", "error", err)
		}
	}()
	slog.Info("database connection ready")

	if err := db.SeedSpareParts(database); err != nil {
		slog.Error("failed to seed spare parts", "error", err)
		os.Exit(1)
	}
	slog.Info("spare parts seed data refreshed")

	offerFetchedPublisher, err := messaging.NewKafkaOfferFetchedPublisher(cfg.KafkaBrokers, cfg.KafkaOfferFetchedTopic)
	if err != nil {
		slog.Error("failed to create kafka publisher", "error", err)
		os.Exit(1)
	}
	defer offerFetchedPublisher.Close()
	pingContext, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	if err := offerFetchedPublisher.Ping(pingContext); err != nil {
		slog.Warn("kafka broker not reachable", "brokers", cfg.KafkaBrokers, "topic", cfg.KafkaOfferFetchedTopic, "error", err)
	} else {
		slog.Info("kafka broker reachable", "brokers", cfg.KafkaBrokers, "topic", cfg.KafkaOfferFetchedTopic)
	}
	cancel()
	topicContext, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	if err := offerFetchedPublisher.EnsureTopic(topicContext); err != nil {
		slog.Warn("failed to ensure kafka topic", "topic", cfg.KafkaOfferFetchedTopic, "error", err)
	} else {
		slog.Info("kafka topic ready", "topic", cfg.KafkaOfferFetchedTopic)
	}
	cancel()

	offerProvider := offer.NewPublishingProvider(
		offer.NewHTTPMultiProvider(cfg.OfferProviderURLs),
		offerFetchedPublisher,
	)
	sparePartsService := service.NewSparePartsService(
		repository.NewSparePartsRepo(database, offerProvider), mapper.NewSparePartsMapper())

	sparePartsHandler := controllers.NewSparePartHandler(sparePartsService)

	mux := http.NewServeMux()
	controllers.SetupRoutes(mux, sparePartsHandler)

	swagger, err := api.GetSwagger()
	if err != nil {
		slog.Error("failed to load OpenAPI spec", "error", err)
		os.Exit(1)
	}

	swagger.Servers = nil

	middleware := nethttpmiddleware.OapiRequestValidatorWithOptions(
		swagger,
		&nethttpmiddleware.Options{
			ErrorHandlerWithOpts: controllers.OpenAPIErrorHandler,
		},
	)

	handler := controllers.LogRequests(middleware(mux))

	server := &http.Server{
		Addr:              ":" + cfg.ServerPort,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		slog.Info("http server listening", "addr", server.Addr)
		serverErr <- server.ListenAndServe()
	}()

	shutdownContext, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-serverErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server failed", "error", err)
			os.Exit(1)
		}
	case <-shutdownContext.Done():
		stop()
		slog.Info("shutdown signal received")

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			slog.Error("server shutdown failed", "error", err)
			os.Exit(1)
		}
		slog.Info("server shutdown complete")
	}
}

func configureLogger(logLevel string) {
	level := slog.LevelInfo
	switch strings.ToLower(logLevel) {
	case "debug":
		level = slog.LevelDebug
	case "warn", "warning":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})))
}
