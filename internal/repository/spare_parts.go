package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/adrienmazet/go-spare-parts-gateway/api"
	"github.com/adrienmazet/go-spare-parts-gateway/internal/offer"
	"github.com/adrienmazet/go-spare-parts-gateway/internal/xerrors"
)

// SparePartsRepo provides necessary methods to store and retrieve spare parts
//
//go:generate mockgen -source=$GOFILE -destination=mock_$GOFILE -package=$GOPACKAGE
type SparePartsRepo interface {
	GetByReference(reference string) (*api.SparePart, error)
}

type sparePartsRepo struct {
	db            *sql.DB
	offerProvider offer.Provider
}

// GetByReference retrieves spare part by reference.
func (c sparePartsRepo) GetByReference(reference string) (*api.SparePart, error) {
	sparePart, err := c.getSparePart(reference)
	if err != nil {
		return nil, err
	}

	sparePart.Offers = c.offerProvider.GetByReference(sparePart.Reference)

	return sparePart, nil
}

func (c sparePartsRepo) getSparePart(reference string) (*api.SparePart, error) {
	row := c.db.QueryRow(`
		SELECT id, reference, label, brand, category, description
		FROM spare_parts
		WHERE reference = $1
	`, reference)

	var sparePart api.SparePart
	if err := row.Scan(
		&sparePart.ID,
		&sparePart.Reference,
		&sparePart.Label,
		&sparePart.Brand,
		&sparePart.Category,
		&sparePart.Description,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, xerrors.ErrorEntityNotFound.Msgf("spare part with reference %s not found", reference)
		}

		return nil, fmt.Errorf("query spare part by reference: %w", err)
	}

	return &sparePart, nil
}

func NewSparePartsRepo(db *sql.DB, offerProvider offer.Provider) SparePartsRepo {
	return &sparePartsRepo{
		db:            db,
		offerProvider: offerProvider,
	}
}
