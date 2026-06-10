package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/adrienmazet/go-spare-parts-gateway/api"
	"github.com/adrienmazet/go-spare-parts-gateway/internal/observability"
	"github.com/adrienmazet/go-spare-parts-gateway/internal/offer"
	"github.com/adrienmazet/go-spare-parts-gateway/internal/xerrors"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// SparePartsRepo provides necessary methods to store and retrieve spare parts
//
//go:generate mockgen -source=$GOFILE -destination=mock_$GOFILE -package=$GOPACKAGE
type SparePartsRepo interface {
	GetByReference(ctx context.Context, reference string) (*api.SparePart, error)
}

type sparePartsRepo struct {
	db            *sql.DB
	offerProvider offer.Provider
}

// GetByReference retrieves spare part by reference.
func (c sparePartsRepo) GetByReference(ctx context.Context, reference string) (*api.SparePart, error) {
	sparePart, err := c.getSparePart(ctx, reference)
	if err != nil {
		return nil, err
	}

	sparePart.Offers = c.offerProvider.GetByReference(ctx, sparePart.Reference)

	return sparePart, nil
}

func (c sparePartsRepo) getSparePart(ctx context.Context, reference string) (*api.SparePart, error) {
	ctx, span := observability.Tracer("repository").Start(ctx, "repository.get_spare_part")
	defer span.End()
	span.SetAttributes(attribute.String("spare_part.reference", reference))

	startedAt := time.Now()
	row := c.db.QueryRowContext(ctx, `
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
		observability.RecordDBOperation("spare_parts.select_by_reference", err, time.Since(startedAt))
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		if errors.Is(err, sql.ErrNoRows) {
			return nil, xerrors.ErrorEntityNotFound.Msgf("spare part with reference %s not found", reference)
		}

		return nil, fmt.Errorf("query spare part by reference: %w", err)
	}

	observability.RecordDBOperation("spare_parts.select_by_reference", nil, time.Since(startedAt))
	return &sparePart, nil
}

func NewSparePartsRepo(db *sql.DB, offerProvider offer.Provider) SparePartsRepo {
	return &sparePartsRepo{
		db:            db,
		offerProvider: offerProvider,
	}
}
