package partspec

import (
	"net/http"

	"github.com/adrienmazet/go-spare-parts-gateway/external/dummyprovider"
)

// NewHandler creates the PartSpec dummy provider HTTP handler.
func NewHandler() http.Handler {
	return dummyprovider.NewHandler(dummyprovider.Config{
		Provider:      "partspec",
		Supplier:      "PartSpec",
		Currency:      "EUR",
		MinPrice:      1500,
		MaxPrice:      22000,
		MaxStock:      40,
		DeliveryTimes: []string{"PT12H", "P2D", "P5D"},
	})
}

// ListenAndServe starts the PartSpec dummy provider.
func ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, NewHandler())
}
