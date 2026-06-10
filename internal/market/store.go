package market

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/adrienmazet/go-spare-parts-gateway/internal/messaging"
)

// Store persists offer price aggregates.
type Store struct {
	db *sql.DB
}

// NewStore creates a market store.
func NewStore(db *sql.DB) Store {
	return Store{db: db}
}

// RecordOfferFetched updates market-like price statistics for an offer event.
func (s Store) RecordOfferFetched(ctx context.Context, event messaging.OfferFetchedEvent) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO offer_price_stats (
			reference,
			currency,
			observed_count,
			min_price,
			max_price,
			latest_price,
			average_price,
			last_observed_at
		)
		VALUES ($1, $2, 1, $3, $3, $3, $5::numeric, $4)
		ON CONFLICT (reference, currency) DO UPDATE SET
			observed_count = offer_price_stats.observed_count + 1,
			min_price = LEAST(offer_price_stats.min_price, EXCLUDED.latest_price),
			max_price = GREATEST(offer_price_stats.max_price, EXCLUDED.latest_price),
			latest_price = EXCLUDED.latest_price,
			average_price = (
				(offer_price_stats.average_price * offer_price_stats.observed_count) + EXCLUDED.latest_price
			) / (offer_price_stats.observed_count + 1),
			last_observed_at = EXCLUDED.last_observed_at
	`,
		event.Reference,
		event.Currency,
		event.Price,
		event.FetchedAt,
		event.Price,
	)
	if err != nil {
		return fmt.Errorf("record offer fetched event: %w", err)
	}

	return nil
}
