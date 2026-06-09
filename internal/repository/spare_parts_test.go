package repository

import (
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/adrienmazet/go-spare-parts-gateway/api"
	"github.com/adrienmazet/go-spare-parts-gateway/internal/xerrors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSparePartsRepoGetByReference(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		_ = db.Close()
	}()

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, reference, label, brand, category, description
		FROM spare_parts
		WHERE reference = $1
	`)).
		WithArgs("BRK-PAD-4521").
		WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"reference",
			"label",
			"brand",
			"category",
			"description",
		}).AddRow(
			"sp-001",
			"BRK-PAD-4521",
			"Front Brake Pads",
			"Brembo",
			"BRAKING",
			"High performance front brake pads for urban and highway use",
		))

	expectedOffers := []api.Offer{
		{
			ID:            "off-test-001",
			Supplier:      "TestSupplier",
			Price:         4599,
			Currency:      api.EUR,
			StockQuantity: 42,
			DeliveryDelay: "PT48H",
		},
	}

	repo := NewSparePartsRepo(db, fakeOfferProvider{offers: expectedOffers})

	sparePart, err := repo.GetByReference("BRK-PAD-4521")

	require.NoError(t, err)
	require.NotNil(t, sparePart)
	assert.Equal(t, "sp-001", sparePart.ID)
	assert.Equal(t, "BRK-PAD-4521", sparePart.Reference)
	assert.Equal(t, expectedOffers, sparePart.Offers)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSparePartsRepoGetByReferenceNotFound(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		_ = db.Close()
	}()

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, reference, label, brand, category, description
		FROM spare_parts
		WHERE reference = $1
	`)).
		WithArgs("UNKNOWN-REF").
		WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"reference",
			"label",
			"brand",
			"category",
			"description",
		}))

	repo := NewSparePartsRepo(db, fakeOfferProvider{})

	sparePart, err := repo.GetByReference("UNKNOWN-REF")

	require.Nil(t, sparePart)
	require.Error(t, err)
	assert.ErrorIs(t, err, xerrors.ErrorEntityNotFound)

	var appError *xerrors.Error
	require.True(t, errors.As(err, &appError))
	assert.Equal(t, "spare part with reference UNKNOWN-REF not found", appError.Msg)
	assert.NoError(t, mock.ExpectationsWereMet())
}

type fakeOfferProvider struct {
	offers []api.Offer
}

func (p fakeOfferProvider) GetByReference(reference string) []api.Offer {
	return p.offers
}
