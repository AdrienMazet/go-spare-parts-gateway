package main

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/adrienmazet/go-spare-parts-gateway/internal/config"
	"github.com/adrienmazet/go-spare-parts-gateway/internal/db"
	"github.com/adrienmazet/go-spare-parts-gateway/internal/market"
	"github.com/adrienmazet/go-spare-parts-gateway/internal/messaging"
	"github.com/twmb/franz-go/pkg/kgo"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	configureLogger(cfg.LogLevel)

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
	var event messaging.OfferFetchedEvent
	if err := json.Unmarshal(record.Value, &event); err != nil {
		return err
	}

	if err := store.RecordOfferFetched(ctx, event); err != nil {
		return err
	}

	slog.Info(
		"offer price event processed",
		"reference", event.Reference,
		"supplier", event.Supplier,
		"price", event.Price,
		"currency", event.Currency,
	)

	return nil
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
