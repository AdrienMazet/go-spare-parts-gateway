package market

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/adrienmazet/go-spare-parts-gateway/api"
	"github.com/adrienmazet/go-spare-parts-gateway/internal/messaging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStoreRecordOfferFetchedUpsertsPriceStats(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		_ = db.Close()
	}()

	fetchedAt := time.Date(2026, 6, 9, 12, 0, 0, 0, time.UTC)
	event := messaging.OfferFetchedEvent{
		Reference: "BRK-PAD-4521",
		Supplier:  "TestSupplier",
		Price:     4599,
		Currency:  api.EUR,
		FetchedAt: fetchedAt,
	}

	mock.ExpectExec(regexp.QuoteMeta(`
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
	`)).
		WithArgs("BRK-PAD-4521", api.EUR, 4599, fetchedAt, 4599).
		WillReturnResult(sqlmock.NewResult(0, 1))

	store := NewStore(db)

	err = store.RecordOfferFetched(context.Background(), event)

	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
