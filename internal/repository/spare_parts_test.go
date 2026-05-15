package repository

import (
	"errors"
	"testing"

	"github.com/adrienmazet/go-spare-parts-gateway/internal/xerrors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSparePartsRepoGetById(t *testing.T) {
	t.Parallel()

	repo := NewSparePartsRepo()

	sparePart, err := repo.GetById("sp-001")

	require.NoError(t, err)
	require.NotNil(t, sparePart)
	assert.Equal(t, "sp-001", sparePart.ID)
	assert.Equal(t, "BRK-PAD-4521", sparePart.Reference)
	assert.Len(t, sparePart.Offers, 2)
}

func TestSparePartsRepoGetByIdNotFound(t *testing.T) {
	t.Parallel()

	repo := NewSparePartsRepo()

	sparePart, err := repo.GetById("sp-999")

	require.Nil(t, sparePart)
	require.Error(t, err)
	assert.ErrorIs(t, err, xerrors.ErrorEntityNotFound)

	var appError *xerrors.Error
	require.True(t, errors.As(err, &appError))
	assert.Equal(t, "spare part with id sp-999 not found", appError.Msg)
}
