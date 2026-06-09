package db

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/adrienmazet/go-spare-parts-gateway/api"
)

var seedSpareParts = []api.SparePart{
	{
		ID:          "sp-001",
		Reference:   "BRK-PAD-4521",
		Label:       "Front Brake Pads",
		Brand:       "Brembo",
		Category:    api.BRAKING,
		Description: "High performance front brake pads for urban and highway use",
	},
	{
		ID:          "sp-002",
		Reference:   "ENG-FLT-7823",
		Label:       "Oil Filter",
		Brand:       "Mann",
		Category:    api.FILTERS,
		Description: "Standard oil filter compatible with most 4-cylinder engines",
	},
	{
		ID:          "sp-003",
		Reference:   "SUS-SPR-3341",
		Label:       "Rear Coil Spring",
		Brand:       "KYB",
		Category:    api.SUSPENSION,
		Description: "OEM replacement rear coil spring for medium sedans",
	},
}

// SeedSpareParts refreshes dummy spare part rows in one transaction.
func SeedSpareParts(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin seed transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			return
		}
	}()

	if _, err := tx.Exec(`DELETE FROM spare_parts`); err != nil {
		return fmt.Errorf("delete spare parts: %w", err)
	}

	for _, sparePart := range seedSpareParts {
		if _, err := tx.Exec(`
			INSERT INTO spare_parts (id, reference, label, brand, category, description)
			VALUES ($1, $2, $3, $4, $5, $6)
		`,
			sparePart.ID,
			sparePart.Reference,
			sparePart.Label,
			sparePart.Brand,
			sparePart.Category,
			sparePart.Description,
		); err != nil {
			return fmt.Errorf("insert spare part %s: %w", sparePart.Reference, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit seed transaction: %w", err)
	}

	return nil
}
