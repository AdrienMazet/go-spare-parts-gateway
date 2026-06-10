package offer

import (
	"context"
	"errors"
	"testing"

	"github.com/adrienmazet/go-spare-parts-gateway/api"
	"github.com/adrienmazet/go-spare-parts-gateway/internal/messaging"
	"github.com/stretchr/testify/assert"
)

func TestPublishingProviderPublishesOneEventPerOffer(t *testing.T) {
	t.Parallel()

	offers := []api.Offer{
		{
			ID:       "off-test-001",
			Supplier: "TestSupplier",
			Price:    4599,
			Currency: api.EUR,
		},
		{
			ID:       "off-test-002",
			Supplier: "OtherSupplier",
			Price:    4299,
			Currency: api.EUR,
		},
	}
	publisher := &fakeOfferFetchedPublisher{}
	provider := NewPublishingProvider(fakeProvider{offers: offers}, publisher)

	actualOffers := provider.GetByReference(context.Background(), "BRK-PAD-4521")

	assert.Equal(t, offers, actualOffers)
	assert.Len(t, publisher.events, 2)
	assert.Equal(t, "BRK-PAD-4521", publisher.events[0].Reference)
	assert.Equal(t, "TestSupplier", publisher.events[0].Supplier)
	assert.Equal(t, 4599, publisher.events[0].Price)
	assert.Equal(t, api.EUR, publisher.events[0].Currency)
}

func TestPublishingProviderKeepsOffersWhenPublishFails(t *testing.T) {
	t.Parallel()

	offers := []api.Offer{
		{
			ID:       "off-test-001",
			Supplier: "TestSupplier",
			Price:    4599,
			Currency: api.EUR,
		},
	}
	publisher := &fakeOfferFetchedPublisher{err: errors.New("kafka unavailable")}
	provider := NewPublishingProvider(fakeProvider{offers: offers}, publisher)

	actualOffers := provider.GetByReference(context.Background(), "BRK-PAD-4521")

	assert.Equal(t, offers, actualOffers)
	assert.Len(t, publisher.events, 1)
}

type fakeProvider struct {
	offers []api.Offer
}

func (p fakeProvider) GetByReference(ctx context.Context, reference string) []api.Offer {
	return p.offers
}

type fakeOfferFetchedPublisher struct {
	events []messaging.OfferFetchedEvent
	err    error
}

func (p *fakeOfferFetchedPublisher) PublishOfferFetched(ctx context.Context, event messaging.OfferFetchedEvent) error {
	p.events = append(p.events, event)
	return p.err
}
