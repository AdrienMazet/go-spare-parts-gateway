package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/adrienmazet/go-spare-parts-gateway/api"
	"github.com/adrienmazet/go-spare-parts-gateway/internal/observability"
	"github.com/twmb/franz-go/pkg/kerr"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/kmsg"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// OfferFetchedEvent is emitted each time an external offer is fetched.
type OfferFetchedEvent struct {
	Reference string             `json:"reference"`
	Supplier  string             `json:"supplier"`
	Price     int                `json:"price"`
	Currency  api.AmountCurrency `json:"currency"`
	FetchedAt time.Time          `json:"fetchedAt"`
}

// OfferFetchedPublisher publishes offer fetched events.
type OfferFetchedPublisher interface {
	PublishOfferFetched(ctx context.Context, event OfferFetchedEvent) error
}

// KafkaOfferFetchedPublisher publishes offer fetched events to Kafka.
type KafkaOfferFetchedPublisher struct {
	client *kgo.Client
	topic  string
}

// NewKafkaOfferFetchedPublisher creates a Kafka-backed offer event publisher.
func NewKafkaOfferFetchedPublisher(brokers []string, topic string) (*KafkaOfferFetchedPublisher, error) {
	client, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		kgo.DefaultProduceTopic(topic),
	)
	if err != nil {
		return nil, fmt.Errorf("create kafka producer: %w", err)
	}

	return &KafkaOfferFetchedPublisher{
		client: client,
		topic:  topic,
	}, nil
}

// Ping checks that at least one Kafka broker is reachable.
func (p *KafkaOfferFetchedPublisher) Ping(ctx context.Context) error {
	if err := p.client.Ping(ctx); err != nil {
		return fmt.Errorf("ping kafka broker: %w", err)
	}

	return nil
}

// EnsureTopic creates the publisher topic if it does not already exist.
func (p *KafkaOfferFetchedPublisher) EnsureTopic(ctx context.Context) error {
	return EnsureTopic(ctx, p.client, p.topic)
}

// EnsureTopic creates a Kafka topic if it does not already exist.
func EnsureTopic(ctx context.Context, client *kgo.Client, topicName string) error {
	req := kmsg.NewPtrCreateTopicsRequest()
	topic := kmsg.NewCreateTopicsRequestTopic()
	topic.Topic = topicName
	topic.NumPartitions = 1
	topic.ReplicationFactor = 1
	req.Topics = append(req.Topics, topic)

	resp, err := req.RequestWith(ctx, client)
	if err != nil {
		return fmt.Errorf("create kafka topic %s: %w", topicName, err)
	}
	if len(resp.Topics) == 0 {
		return fmt.Errorf("create kafka topic %s: empty response", topicName)
	}

	if err := kerr.ErrorForCode(resp.Topics[0].ErrorCode); err != nil && !errors.Is(err, kerr.TopicAlreadyExists) {
		return fmt.Errorf("create kafka topic %s: %w", topicName, err)
	}

	return nil
}

// PublishOfferFetched queues an offer fetched event for asynchronous publishing.
func (p *KafkaOfferFetchedPublisher) PublishOfferFetched(ctx context.Context, event OfferFetchedEvent) error {
	ctx, span := observability.Tracer("messaging").Start(ctx, "kafka.publish.offer_fetched")
	span.SetAttributes(
		attribute.String("messaging.system", "kafka"),
		attribute.String("messaging.destination.name", p.topic),
		attribute.String("spare_part.reference", event.Reference),
		attribute.String("supplier", event.Supplier),
	)

	payload, err := json.Marshal(event)
	if err != nil {
		observability.RecordKafkaEvent("publish", p.topic, err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		span.End()
		return fmt.Errorf("marshal offer fetched event: %w", err)
	}

	record := &kgo.Record{
		Topic: p.topic,
		Key:   []byte(event.Reference),
		Value: payload,
	}

	p.client.Produce(context.WithoutCancel(ctx), record, func(_ *kgo.Record, err error) {
		defer span.End()
		observability.RecordKafkaEvent("publish", p.topic, err)

		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			slog.Warn(
				"failed to publish offer fetched event",
				"reference", event.Reference,
				"supplier", event.Supplier,
				"error", err,
			)
			return
		}

		span.SetStatus(codes.Ok, "")
		slog.Debug(
			"offer fetched event published",
			"reference", event.Reference,
			"supplier", event.Supplier,
		)
	})

	return nil
}

// Close closes the Kafka client.
func (p *KafkaOfferFetchedPublisher) Close() {
	p.client.Close()
}
