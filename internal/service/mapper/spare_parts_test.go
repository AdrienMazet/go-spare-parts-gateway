package mapper

import (
	"testing"

	"github.com/adrienmazet/go-spare-parts-gateway/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSparePartsMapperModelToResponse(t *testing.T) {
	t.Parallel()

	sparePart := api.SparePart{
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
		},
	}

	mapper := NewSparePartsMapper()

	response := mapper.ModelToResponse(sparePart)

	require.NotNil(t, response)
	assert.NotSame(t, &sparePart, response)
	assert.Equal(t, sparePart, *response)
}
