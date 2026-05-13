// Package xerrors defines application errors shared across internal packages.
package xerrors

import (
	"errors"
	"fmt"
)

var (
	// ErrorEntityNotFound identifies missing domain entities.
	ErrorEntityNotFound = New("NOT_FOUND", "Not Found")

	// ErrorInvalidEntity identifies invalid domain entities.
	ErrorInvalidEntity = New("INVALID_ENTITY", "Invalid entity")
)

// Error is an application error with a stable kind and a client-facing message.
type Error struct {
	Kind string
	Msg  string
	Err  error
}

// New creates an application error template.
func New(kind, defaultMsg string) *Error {
	return &Error{
		Kind: kind,
		Msg:  defaultMsg,
	}
}

// Error returns the formatted error string.
func (e *Error) Error() string {
	return fmt.Sprintf("ServiceError(%s): %s", e.Kind, e.Msg)
}

// Unwrap returns the wrapped error.
func (e *Error) Unwrap() error {
	return e.Err
}

// Is reports whether target has the same application error kind.
func (e *Error) Is(target error) bool {
	var targetError *Error
	if errors.As(target, &targetError) {
		return targetError.Kind == e.Kind
	}

	return false
}

func (e *Error) clone() *Error {
	err := *e
	return &err
}

// Msgf returns a copy of e with a formatted message.
func (e *Error) Msgf(format string, args ...any) *Error {
	err := e.clone()
	err.Msg = fmt.Sprintf(format, args...)
	return err
}

// Wrap returns a copy of e wrapping err.
func (e *Error) Wrap(err error) *Error {
	newErr := e.clone()
	newErr.Err = err
	return newErr
}
