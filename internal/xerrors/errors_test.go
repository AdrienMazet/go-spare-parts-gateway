package xerrors

import (
	"errors"
	"testing"
)

func TestErrorIsMatchesByKind(t *testing.T) {
	err := ErrorEntityNotFound.Msgf("spare part with id %s not found", "sp-999")

	if !errors.Is(err, ErrorEntityNotFound) {
		t.Fatal("expected error to match ErrorEntityNotFound")
	}
}

func TestErrorWrapPreservesCause(t *testing.T) {
	cause := errors.New("repository unavailable")
	err := ErrorEntityNotFound.Wrap(cause)

	if !errors.Is(err, ErrorEntityNotFound) {
		t.Fatal("expected wrapped error to match ErrorEntityNotFound")
	}

	if !errors.Is(err, cause) {
		t.Fatal("expected wrapped error to preserve cause")
	}
}

func TestErrorMsgfDoesNotMutateTemplate(t *testing.T) {
	err := ErrorEntityNotFound.Msgf("spare part with id %s not found", "sp-999")

	if err.Msg == ErrorEntityNotFound.Msg {
		t.Fatal("expected formatted error message to differ from template")
	}

	if ErrorEntityNotFound.Msg != "Not Found" {
		t.Fatalf("expected template message to remain unchanged, got %q", ErrorEntityNotFound.Msg)
	}
}
