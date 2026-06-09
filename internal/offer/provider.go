package offer

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/adrienmazet/go-spare-parts-gateway/api"
	"github.com/adrienmazet/go-spare-parts-gateway/external/dummyprovider"
)

// Provider retrieves supplier offers for a spare part reference.
type Provider interface {
	GetByReference(reference string) []api.Offer
}

// HTTPProvider retrieves offers from HTTP provider services.
type HTTPProvider struct {
	client  *http.Client
	baseURL string
}

// NewHTTPProvider creates an HTTP offer provider.
func NewHTTPProvider(baseURL string) HTTPProvider {
	return HTTPProvider{
		client: &http.Client{
			Timeout: 2 * time.Second,
		},
		baseURL: strings.TrimRight(baseURL, "/"),
	}
}

// GetByReference retrieves offers from one HTTP provider.
func (p HTTPProvider) GetByReference(reference string) []api.Offer {
	req, err := http.NewRequest(http.MethodGet, p.baseURL+"/offers/"+reference, nil)
	if err != nil {
		return nil
	}

	res, err := p.client.Do(req)
	if err != nil {
		return nil
	}
	defer func() {
		_ = res.Body.Close()
	}()

	if res.StatusCode != http.StatusOK {
		return nil
	}

	var providerResponse dummyprovider.Response
	if err := json.NewDecoder(res.Body).Decode(&providerResponse); err != nil {
		return nil
	}

	return toAPIOffers(providerResponse.Offers)
}

// MultiProvider retrieves offers from several providers concurrently.
type MultiProvider struct {
	providers []Provider
}

// NewMultiProvider creates a provider that aggregates offers from all providers.
func NewMultiProvider(providers []Provider) MultiProvider {
	return MultiProvider{providers: providers}
}

// NewHTTPMultiProvider creates a provider from provider base URLs.
func NewHTTPMultiProvider(baseURLs []string) MultiProvider {
	providers := make([]Provider, 0, len(baseURLs))
	for _, baseURL := range baseURLs {
		if baseURL == "" {
			continue
		}

		providers = append(providers, NewHTTPProvider(baseURL))
	}

	return NewMultiProvider(providers)
}

// GetByReference retrieves offers from every configured provider.
func (p MultiProvider) GetByReference(reference string) []api.Offer {
	offers := make(chan []api.Offer, len(p.providers))

	wg := sync.WaitGroup{}
	wg.Add(len(p.providers))

	for _, provider := range p.providers {
		go func() {
			defer wg.Done()
			offers <- provider.GetByReference(reference)
		}()
	}

	wg.Wait()
	close(offers)

	aggregatedOffers := make([]api.Offer, 0)
	for providerOffers := range offers {
		aggregatedOffers = append(aggregatedOffers, providerOffers...)
	}

	return aggregatedOffers
}

func toAPIOffers(providerOffers []dummyprovider.Offer) []api.Offer {
	offers := make([]api.Offer, 0, len(providerOffers))
	for _, offer := range providerOffers {
		offers = append(offers, api.Offer{
			ID:            offer.ID,
			Supplier:      offer.Supplier,
			Price:         offer.Price,
			Currency:      api.AmountCurrency(offer.Currency),
			StockQuantity: offer.StockQuantity,
			DeliveryDelay: offer.DeliveryDelay,
		})
	}

	return offers
}
