package repository

import (
	"github.com/adrienmazet/go-spare-parts-gateway/api"
	"github.com/adrienmazet/go-spare-parts-gateway/internal/xerrors"
)

// SparePartsRepo provides necessary methods to store and retrieve spare parts
//
//go:generate mockgen -source=$GOFILE -destination=mock_$GOFILE -package=$GOPACKAGE
type SparePartsRepo interface {
	GetById(id string) (*api.SparePart, error)
}

type sparePartsRepo struct {
}

// GetById retrieves spare part by id
func (c sparePartsRepo) GetById(id string) (*api.SparePart, error) {
	spareParts := getSpareParts()

	for i := range spareParts {
		if spareParts[i].ID == id {
			return &spareParts[i], nil
		}
	}

	return nil, xerrors.ErrorEntityNotFound.Msgf("spare part with id %s not found", id)
}

// TODO : retrieve spare parts from external data sources
func getSpareParts() []api.SparePart {
	return []api.SparePart{
		{
			ID:          "sp-001",
			Reference:   "BRK-PAD-4521",
			Label:       "Front Brake Pads",
			Brand:       "Brembo",
			Category:    api.BRAKING,
			Description: "High performance front brake pads for urban and highway use",
			Offers: []api.Offer{
				{
					ID:            "off-001",
					Supplier:      "PartsPro",
					Price:         4599,
					Currency:      api.EUR,
					StockQuantity: 42,
					DeliveryDelay: "PT48H",
				},
				{
					ID:            "off-002",
					Supplier:      "AutoStock",
					Price:         3999,
					Currency:      api.EUR,
					StockQuantity: 8,
					DeliveryDelay: "P3D",
				},
			},
		},
		{
			ID:          "sp-002",
			Reference:   "ENG-FLT-7823",
			Label:       "Oil Filter",
			Brand:       "Mann",
			Category:    api.FILTERS,
			Description: "Standard oil filter compatible with most 4-cylinder engines",
			Offers: []api.Offer{
				{
					ID:            "off-003",
					Supplier:      "PartsPro",
					Price:         899,
					Currency:      api.EUR,
					StockQuantity: 150,
					DeliveryDelay: "PT2H",
				},
			},
		},
		{
			ID:          "sp-003",
			Reference:   "SUS-SPR-3341",
			Label:       "Rear Coil Spring",
			Brand:       "KYB",
			Category:    api.SUSPENSION,
			Description: "OEM replacement rear coil spring for medium sedans",
			Offers:      []api.Offer{},
		},
	}
}

func NewSparePartsRepo() SparePartsRepo {
	return &sparePartsRepo{}
}
