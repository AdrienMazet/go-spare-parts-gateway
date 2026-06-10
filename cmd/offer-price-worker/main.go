package main

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/adrienmazet/go-spare-parts-gateway/internal/config"
	"github.com/adrienmazet/go-spare-parts-gateway/internal/db"
	"github.com/adrienmazet/go-spare-parts-gateway/internal/market"
	"github.com/adrienmazet/go-spare-parts-gateway/internal/messaging"
	"github.com/adrienmazet/go-spare-parts-gateway/internal/observability"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	observability.ConfigureLogger(cfg.LogLevel)

	shutdownTracer, err := observability.InitTracer(context.Background(), "offer-price-worker", cfg.OTLPEndpoint)
	if err != nil {
		slog.Warn("failed to initialize tracing", "error", err)
	} else {
		defer func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := shutdownTracer(ctx); err != nil {
				slog.Warn("failed to shutdown tracer", "error", err)
			}
		}()
	}

	if err := db.Migrate(cfg.DatabaseURL, cfg.MigrationsPath); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

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

	client, err := kgo.NewClient(
		kgo.SeedBrokers(cfg.KafkaBrokers...),
		kgo.ConsumerGroup(cfg.KafkaOfferConsumerGroup),
		kgo.ConsumeTopics(cfg.KafkaOfferFetchedTopic),
	)
	if err != nil {
		slog.Error("failed to create kafka consumer", "error", err)
		os.Exit(1)
	}
	defer client.Close()

	topicContext, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	if err := messaging.EnsureTopic(topicContext, client, cfg.KafkaOfferFetchedTopic); err != nil {
		cancel()
		slog.Error("failed to ensure kafka topic", "topic", cfg.KafkaOfferFetchedTopic, "error", err)
		os.Exit(1)
	}
	cancel()
	slog.Info("kafka topic ready", "topic", cfg.KafkaOfferFetchedTopic)

	store := market.NewStore(database)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	metricsServer := startMetricsServer(cfg.WorkerMetricsPort)
	defer func() {
		shutdownContext, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := metricsServer.Shutdown(shutdownContext); err != nil {
			slog.Warn("failed to shutdown metrics server", "error", err)
		}
	}()

	slog.Info(
		"offer price worker started",
		"topic", cfg.KafkaOfferFetchedTopic,
		"consumer_group", cfg.KafkaOfferConsumerGroup,
	)

	if err := consume(ctx, client, store); err != nil && !errors.Is(err, context.Canceled) {
		slog.Error("offer price worker stopped with error", "error", err)
		os.Exit(1)
	}

	slog.Info("offer price worker stopped")
}

func consume(ctx context.Context, client *kgo.Client, store market.Store) error {
	for {
		fetches := client.PollFetches(ctx)
		if err := fetches.Err(); err != nil {
			return err
		}

		fetches.EachRecord(func(record *kgo.Record) {
			if err := handleRecord(ctx, store, record); err != nil {
				slog.Warn("failed to process offer fetched event", "error", err)
			}
		})

		if err := client.CommitUncommittedOffsets(ctx); err != nil {
			return err
		}
	}
}

func handleRecord(ctx context.Context, store market.Store, record *kgo.Record) error {
	ctx, span := observability.Tracer("worker").Start(ctx, "kafka.consume.offer_fetched")
	defer span.End()
	span.SetAttributes(
		attribute.String("messaging.system", "kafka"),
		attribute.String("messaging.destination.name", record.Topic),
	)

	var event messaging.OfferFetchedEvent
	if err := json.Unmarshal(record.Value, &event); err != nil {
		observability.RecordKafkaEvent("consume", record.Topic, err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	span.SetAttributes(
		attribute.String("spare_part.reference", event.Reference),
		attribute.String("supplier", event.Supplier),
		attribute.Int("price", event.Price),
		attribute.String("currency", string(event.Currency)),
	)

	if err := store.RecordOfferFetched(ctx, event); err != nil {
		observability.RecordKafkaEvent("consume", record.Topic, err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	observability.RecordKafkaEvent("consume", record.Topic, nil)
	span.SetStatus(codes.Ok, "")

	slog.Info(
		"offer price event processed",
		"reference", event.Reference,
		"supplier", event.Supplier,
		"price", event.Price,
		"currency", event.Currency,
	)

	return nil
}

func startMetricsServer(port string) *http.Server {
	mux := http.NewServeMux()
	mux.Handle("/metrics", observability.MetricsHandler())

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		slog.Info("metrics server listening", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("metrics server failed", "error", err)
		}
	}()

	return server
}
