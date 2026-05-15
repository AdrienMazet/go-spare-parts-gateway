package controllers

import (
	"testing"

	"github.com/adrienmazet/go-spare-parts-gateway/api"
	"github.com/stretchr/testify/assert"
)

func TestValidateISO8601Duration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{
			name:    "valid hours duration",
			value:   "PT2H",
			wantErr: false,
		},
		{
			name:    "valid 48 hours duration",
			value:   "PT48H",
			wantErr: false,
		},
		{
			name:    "valid days duration",
			value:   "P3D",
			wantErr: false,
		},
		{
			name:    "valid days and hours duration",
			value:   "P1DT4H",
			wantErr: false,
		},
		{
			name:    "valid minutes duration",
			value:   "PT30M",
			wantErr: false,
		},
		{
			name:    "valid seconds duration",
			value:   "PT45S",
			wantErr: false,
		},
		{
			name:    "invalid human duration",
			value:   "48h",
			wantErr: true,
		},
		{
			name:    "invalid missing prefix",
			value:   "2H",
			wantErr: true,
		},
		{
			name:    "invalid incomplete period",
			value:   "P",
			wantErr: true,
		},
		{
			name:    "invalid incomplete time period",
			value:   "PT",
			wantErr: true,
		},
		{
			name:    "invalid hours without time separator",
			value:   "P3H",
			wantErr: true,
		},
		{
			name:    "invalid days after time separator",
			value:   "PT2D",
			wantErr: true,
		},
		{
			name:    "invalid empty value",
			value:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := Validator.Var(tt.value, "iso8601duration")

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestValidateUppercaseReference(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{
			name:    "valid hyphenated reference",
			value:   "BRK-PAD-4521",
			wantErr: false,
		},
		{
			name:    "valid underscored reference",
			value:   "SUS_SPR_3341",
			wantErr: false,
		},
		{
			name:    "valid alphanumeric reference",
			value:   "A123",
			wantErr: false,
		},
		{
			name:    "valid numeric prefix",
			value:   "123ABC",
			wantErr: false,
		},
		{
			name:    "invalid lowercase reference",
			value:   "brk-pad-4521",
			wantErr: true,
		},
		{
			name:    "invalid starts with hyphen",
			value:   "-BRK-PAD",
			wantErr: true,
		},
		{
			name:    "invalid starts with underscore",
			value:   "_BRK-PAD",
			wantErr: true,
		},
		{
			name:    "invalid contains space",
			value:   "BRK PAD",
			wantErr: true,
		},
		{
			name:    "invalid contains dot",
			value:   "BRK.PAD",
			wantErr: true,
		},
		{
			name:    "invalid empty value",
			value:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := Validator.Var(tt.value, "uppercase_ref")

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestValidateSparePart(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		sparePart api.SparePart
		wantErr   bool
	}{
		{
			name:      "valid spare part",
			sparePart: validSparePart(),
			wantErr:   false,
		},
		{
			name: "invalid lowercase reference",
			sparePart: func() api.SparePart {
				sp := validSparePart()
				sp.Reference = "brk-pad-4521"

				return sp
			}(),
			wantErr: true,
		},
		{
			name: "invalid delivery delay",
			sparePart: func() api.SparePart {
				sp := validSparePart()
				sp.Offers[0].DeliveryDelay = "48h"

				return sp
			}(),
			wantErr: true,
		},
		{
			name: "invalid currency",
			sparePart: func() api.SparePart {
				sp := validSparePart()
				sp.Offers[0].Currency = api.AmountCurrency("EURO")

				return sp
			}(),
			wantErr: true,
		},
		{
			name: "invalid category",
			sparePart: func() api.SparePart {
				sp := validSparePart()
				sp.Category = api.SparePartCategory("UNKNOWN")

				return sp
			}(),
			wantErr: true,
		},
		{
			name: "invalid negative price",
			sparePart: func() api.SparePart {
				sp := validSparePart()
				sp.Offers[0].Price = -1

				return sp
			}(),
			wantErr: true,
		},
		{
			name: "invalid nil offers",
			sparePart: func() api.SparePart {
				sp := validSparePart()
				sp.Offers = nil

				return sp
			}(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := Validator.Struct(tt.sparePart)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
		})
	}
}

func validSparePart() api.SparePart {
	return api.SparePart{
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
}
