package service

import (
	"context"

	"github.com/adrienmazet/go-spare-parts-gateway/api"
	"github.com/adrienmazet/go-spare-parts-gateway/internal/repository"
	"github.com/adrienmazet/go-spare-parts-gateway/internal/service/mapper"
)

//go:generate mockgen -source=$GOFILE -destination=mock_$GOFILE -package=$GOPACKAGE
type SparePartsService interface {
	Retrieve(ctx context.Context, reference string) (*api.SparePart, error)
	Mapper() mapper.SparePartsMapper
}

type sparePartsService struct {
	repository repository.SparePartsRepo
	mapper     mapper.SparePartsMapper
}

func NewSparePartsService(r repository.SparePartsRepo, m mapper.SparePartsMapper) SparePartsService {
	return sparePartsService{r, m}
}

// Retrieve retrieves a spare part by its reference.
func (s sparePartsService) Retrieve(ctx context.Context, reference string) (*api.SparePart, error) {
	sp, err := s.repository.GetByReference(ctx, reference)

	if err != nil {
		return nil, err
	}

	return sp, nil
}

func (s sparePartsService) Mapper() mapper.SparePartsMapper {
	return s.mapper
}
