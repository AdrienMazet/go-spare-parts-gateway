package hexapart

import (
	"net/http"

	"github.com/adrienmazet/go-spare-parts-gateway/external/dummyprovider"
)

// NewHandler creates the HexaPart dummy provider HTTP handler.
func NewHandler() http.Handler {
	return dummyprovider.NewHandler(dummyprovider.Config{
		Provider:      "hexapart",
		Supplier:      "HexaPart",
		Currency:      "EUR",
		MinPrice:      1200,
		MaxPrice:      12000,
		MaxStock:      80,
		DeliveryTimes: []string{"PT24H", "PT48H", "P3D"},
	})
}

// ListenAndServe starts the HexaPart dummy provider.
func ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, NewHandler())
}
