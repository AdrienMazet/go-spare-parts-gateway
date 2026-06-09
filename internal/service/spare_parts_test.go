package service

import (
	"errors"
	"testing"

	"github.com/adrienmazet/go-spare-parts-gateway/api"
	"github.com/adrienmazet/go-spare-parts-gateway/internal/repository"
	"github.com/adrienmazet/go-spare-parts-gateway/internal/service/mapper"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSparePartsServiceRetrieve(t *testing.T) {
	t.Parallel()

	mockController := gomock.NewController(t)
	defer mockController.Finish()

	expectedSparePart := &api.SparePart{
		ID:          "sp-001",
		Reference:   "BRK-PAD-4521",
		Label:       "Front Brake Pads",
		Brand:       "Brembo",
		Category:    api.BRAKING,
		Description: "High performance front brake pads for urban and highway use",
	}

	repo := repository.NewMockSparePartsRepo(mockController)
	repo.EXPECT().GetByReference("BRK-PAD-4521").Return(expectedSparePart, nil)

	sparePartsService := NewSparePartsService(repo, nil)

	sparePart, err := sparePartsService.Retrieve("BRK-PAD-4521")

	require.NoError(t, err)
	assert.Same(t, expectedSparePart, sparePart)
}

func TestSparePartsServiceRetrieveReturnsRepositoryError(t *testing.T) {
	t.Parallel()

	mockController := gomock.NewController(t)
	defer mockController.Finish()

	expectedErr := errors.New("repository failed")

	repo := repository.NewMockSparePartsRepo(mockController)
	repo.EXPECT().GetByReference("UNKNOWN-REF").Return(nil, expectedErr)

	sparePartsService := NewSparePartsService(repo, nil)

	sparePart, err := sparePartsService.Retrieve("UNKNOWN-REF")

	require.Nil(t, sparePart)
	assert.ErrorIs(t, err, expectedErr)
}

func TestSparePartsServiceMapper(t *testing.T) {
	t.Parallel()

	mockController := gomock.NewController(t)
	defer mockController.Finish()

	expectedMapper := mapper.NewMockSparePartsMapper(mockController)

	sparePartsService := NewSparePartsService(nil, expectedMapper)

	assert.Same(t, expectedMapper, sparePartsService.Mapper())
}
