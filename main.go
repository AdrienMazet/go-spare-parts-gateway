package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	nethttpmiddleware "github.com/oapi-codegen/nethttp-middleware"

	"github.com/adrienmazet/go-spare-parts-gateway/api"
	"github.com/adrienmazet/go-spare-parts-gateway/internal/config"
	"github.com/adrienmazet/go-spare-parts-gateway/internal/controllers"
	"github.com/adrienmazet/go-spare-parts-gateway/internal/db"
	"github.com/adrienmazet/go-spare-parts-gateway/internal/offer"
	"github.com/adrienmazet/go-spare-parts-gateway/internal/repository"
	"github.com/adrienmazet/go-spare-parts-gateway/internal/service"
	"github.com/adrienmazet/go-spare-parts-gateway/internal/service/mapper"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if err := db.Migrate(cfg.DatabaseURL, cfg.MigrationsPath); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	database, err := db.Open(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer func() {
		if err := database.Close(); err != nil {
			log.Printf("Failed to close database: %v", err)
		}
	}()

	if err := db.SeedSpareParts(database); err != nil {
		log.Fatalf("Failed to seed spare parts: %v", err)
	}

	offerProvider := offer.NewHTTPMultiProvider(cfg.OfferProviderURLs)
	sparePartsService := service.NewSparePartsService(
		repository.NewSparePartsRepo(database, offerProvider), mapper.NewSparePartsMapper())

	sparePartsHandler := controllers.NewSparePartHandler(sparePartsService)

	mux := http.NewServeMux()
	controllers.SetupRoutes(mux, sparePartsHandler)

	swagger, err := api.GetSwagger()
	if err != nil {
		log.Fatal(err)
	}

	swagger.Servers = nil

	middleware := nethttpmiddleware.OapiRequestValidatorWithOptions(
		swagger,
		&nethttpmiddleware.Options{
			ErrorHandlerWithOpts: controllers.OpenAPIErrorHandler,
		},
	)

	handler := middleware(mux)

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
		serverErr <- server.ListenAndServe()
	}()

	shutdownContext, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-serverErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server failed: %v", err)
		}
	case <-shutdownContext.Done():
		stop()

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Fatalf("Server shutdown failed: %v", err)
		}
	}
}
