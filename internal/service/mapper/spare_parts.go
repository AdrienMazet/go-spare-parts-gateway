package mapper

import (
	"github.com/adrienmazet/go-spare-parts-gateway/api"
)

//go:generate mockgen -source=$GOFILE -destination=mock_$GOFILE -package=$GOPACKAGE
type SparePartsMapper interface {
	ModelToResponse(sp api.SparePart) *api.SparePart
}

type sparePartsMapper struct{}

func NewSparePartsMapper() SparePartsMapper {
	return &sparePartsMapper{}
}

func (m sparePartsMapper) ModelToResponse(sp api.SparePart) *api.SparePart {
	return &sp
}
