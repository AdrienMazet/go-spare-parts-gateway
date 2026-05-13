package service

import (
	"github.com/adrienmazet/go-spare-parts-gateway/api"
	"github.com/adrienmazet/go-spare-parts-gateway/internal/repository"
	"github.com/adrienmazet/go-spare-parts-gateway/internal/service/mapper"
)

//go:generate mockgen -source=$GOFILE -destination=mock_$GOFILE -package=$GOPACKAGE
type SparePartsService interface {
	Retrieve(ID string) (*api.SparePart, error)
	Mapper() mapper.SparePartsMapper
}

type sparePartsService struct {
	repository repository.SparePartsRepo
	mapper     mapper.SparePartsMapper
}

func NewSparePartsService(r repository.SparePartsRepo, m mapper.SparePartsMapper) SparePartsService {
	return sparePartsService{r, m}
}

// Retrieve retrieves a spare part by its id
func (s sparePartsService) Retrieve(id string) (*api.SparePart, error) {
	sp, err := s.repository.GetById(id)

	if err != nil {
		return nil, err
	}

	return sp, nil
}

func (s sparePartsService) Mapper() mapper.SparePartsMapper {
	return s.mapper
}
