package offer

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/adrienmazet/go-spare-parts-gateway/api"
	"github.com/adrienmazet/go-spare-parts-gateway/external/dummyprovider"
	"github.com/adrienmazet/go-spare-parts-gateway/internal/messaging"
	"github.com/adrienmazet/go-spare-parts-gateway/internal/observability"
)

// Provider retrieves supplier offers for a spare part reference.
type Provider interface {
	GetByReference(ctx context.Context, reference string) []api.Offer
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
			Timeout:   2 * time.Second,
			Transport: observability.InstrumentHTTPClientTransport(nil),
		},
		baseURL: strings.TrimRight(baseURL, "/"),
	}
}

// GetByReference retrieves offers from one HTTP provider.
func (p HTTPProvider) GetByReference(ctx context.Context, reference string) []api.Offer {
	startedAt := time.Now()
	status := "request_error"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.baseURL+"/offers/"+reference, nil)
	if err != nil {
		observability.RecordExternalProviderRequest(p.baseURL, status, time.Since(startedAt))
		return nil
	}

	res, err := p.client.Do(req)
	if err != nil {
		observability.RecordExternalProviderRequest(p.baseURL, status, time.Since(startedAt))
		return nil
	}
	defer func() {
		_ = res.Body.Close()
	}()

	status = res.Status
	if res.StatusCode != http.StatusOK {
		observability.RecordExternalProviderRequest(p.baseURL, status, time.Since(startedAt))
		return nil
	}

	var providerResponse dummyprovider.Response
	if err := json.NewDecoder(res.Body).Decode(&providerResponse); err != nil {
		observability.RecordExternalProviderRequest(p.baseURL, "decode_error", time.Since(startedAt))
		return nil
	}

	observability.RecordExternalProviderRequest(p.baseURL, status, time.Since(startedAt))
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
func (p MultiProvider) GetByReference(ctx context.Context, reference string) []api.Offer {
	offers := make(chan []api.Offer, len(p.providers))

	wg := sync.WaitGroup{}
	wg.Add(len(p.providers))

	for _, provider := range p.providers {
		go func() {
			defer wg.Done()
			offers <- provider.GetByReference(ctx, reference)
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

// PublishingProvider publishes one event for each fetched offer.
type PublishingProvider struct {
	provider  Provider
	publisher messaging.OfferFetchedPublisher
}

// NewPublishingProvider creates an offer provider decorator that emits events.
func NewPublishingProvider(provider Provider, publisher messaging.OfferFetchedPublisher) PublishingProvider {
	return PublishingProvider{
		provider:  provider,
		publisher: publisher,
	}
}

// GetByReference retrieves offers and publishes one event per offer.
func (p PublishingProvider) GetByReference(ctx context.Context, reference string) []api.Offer {
	offers := p.provider.GetByReference(ctx, reference)

	for _, fetchedOffer := range offers {
		event := messaging.OfferFetchedEvent{
			Reference: reference,
			Supplier:  fetchedOffer.Supplier,
			Price:     fetchedOffer.Price,
			Currency:  fetchedOffer.Currency,
			FetchedAt: time.Now().UTC(),
		}

		if err := p.publisher.PublishOfferFetched(ctx, event); err != nil {
			slog.Warn("failed to publish offer fetched event", "reference", reference, "supplier", fetchedOffer.Supplier, "error", err)
		}
	}

	return offers
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
