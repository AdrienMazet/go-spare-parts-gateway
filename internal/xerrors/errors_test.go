package xerrors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorIsMatchesByKind(t *testing.T) {
	err := ErrorEntityNotFound.Msgf("spare part with id %s not found", "sp-999")

	assert.ErrorIs(t, err, ErrorEntityNotFound)
}

func TestErrorString(t *testing.T) {
	err := ErrorEntityNotFound.Msgf("spare part with id %s not found", "sp-999")

	assert.Equal(t, "ServiceError(NOT_FOUND): spare part with id sp-999 not found", err.Error())
}

func TestErrorWrapPreservesCause(t *testing.T) {
	cause := errors.New("repository unavailable")
	err := ErrorEntityNotFound.Wrap(cause)

	assert.ErrorIs(t, err, ErrorEntityNotFound)
	assert.ErrorIs(t, err, cause)
}

func TestErrorMsgfDoesNotMutateTemplate(t *testing.T) {
	err := ErrorEntityNotFound.Msgf("spare part with id %s not found", "sp-999")

	assert.NotEqual(t, ErrorEntityNotFound.Msg, err.Msg)
	assert.Equal(t, "Not Found", ErrorEntityNotFound.Msg)
}
