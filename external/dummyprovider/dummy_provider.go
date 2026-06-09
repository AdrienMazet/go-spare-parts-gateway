package dummyprovider

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net/http"
)

// Config defines the behavior of a dummy external provider.
type Config struct {
	Provider      string
	Supplier      string
	Currency      string
	MinPrice      int
	MaxPrice      int
	MaxStock      int
	DeliveryTimes []string
}

// Offer is the dummy offer payload returned by external providers.
type Offer struct {
	ID            string `json:"id"`
	Reference     string `json:"reference"`
	Supplier      string `json:"supplier"`
	Price         int    `json:"price"`
	Currency      string `json:"currency"`
	StockQuantity int    `json:"stockQuantity"`
	DeliveryDelay string `json:"deliveryDelay"`
}

// Response is the dummy provider response payload.
type Response struct {
	Provider  string  `json:"provider"`
	Reference string  `json:"reference"`
	Offers    []Offer `json:"offers"`
}

// NewHandler creates an HTTP handler returning random offers for a reference.
func NewHandler(config Config) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /offers/{reference}", func(w http.ResponseWriter, r *http.Request) {
		reference := r.PathValue("reference")
		if reference == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing reference"})
			return
		}

		writeJSON(w, http.StatusOK, Response{
			Provider:  config.Provider,
			Reference: reference,
			Offers:    randomOffers(config, reference),
		})
	})

	return mux
}

func randomOffers(config Config, reference string) []Offer {
	count := rand.IntN(3) + 1
	offers := make([]Offer, 0, count)

	for range count {
		offers = append(offers, Offer{
			ID:            fmt.Sprintf("off-%s-%03d", config.Provider, rand.IntN(1000)),
			Reference:     reference,
			Supplier:      config.Supplier,
			Price:         randomPrice(config.MinPrice, config.MaxPrice),
			Currency:      config.Currency,
			StockQuantity: rand.IntN(config.MaxStock + 1),
			DeliveryDelay: config.DeliveryTimes[rand.IntN(len(config.DeliveryTimes))],
		})
	}

	return offers
}

func randomPrice(minPrice, maxPrice int) int {
	if maxPrice <= minPrice {
		return minPrice
	}

	return minPrice + rand.IntN(maxPrice-minPrice+1)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}
