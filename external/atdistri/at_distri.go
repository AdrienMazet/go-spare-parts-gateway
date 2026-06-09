package atdistri

import (
	"net/http"

	"github.com/adrienmazet/go-spare-parts-gateway/external/dummyprovider"
)

// NewHandler creates the ATDistri dummy provider HTTP handler.
func NewHandler() http.Handler {
	return dummyprovider.NewHandler(dummyprovider.Config{
		Provider:      "atdistri",
		Supplier:      "ATDistri",
		Currency:      "USD",
		MinPrice:      700,
		MaxPrice:      15000,
		MaxStock:      60,
		DeliveryTimes: []string{"PT6H", "PT48H", "P4D"},
	})
}

// ListenAndServe starts the ATDistri dummy provider.
func ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, NewHandler())
}
