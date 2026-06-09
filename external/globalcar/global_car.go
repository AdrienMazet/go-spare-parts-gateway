package globalcar

import (
	"net/http"

	"github.com/adrienmazet/go-spare-parts-gateway/external/dummyprovider"
)

// NewHandler creates the GlobalCar dummy provider HTTP handler.
func NewHandler() http.Handler {
	return dummyprovider.NewHandler(dummyprovider.Config{
		Provider:      "globalcar",
		Supplier:      "GlobalCar",
		Currency:      "USD",
		MinPrice:      900,
		MaxPrice:      18000,
		MaxStock:      120,
		DeliveryTimes: []string{"PT2H", "PT24H", "PT72H"},
	})
}

// ListenAndServe starts the GlobalCar dummy provider.
func ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, NewHandler())
}
