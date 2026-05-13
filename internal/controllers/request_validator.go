package controllers

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

var Validator = newValidator()

func newValidator() *validator.Validate {
	v := validator.New(validator.WithRequiredStructEnabled())

	if err := v.RegisterValidation("iso8601duration", validateISO8601Duration); err != nil {
		panic(err)
	}

	if err := v.RegisterValidation("uppercase_ref", validateUppercaseReference); err != nil {
		panic(err)
	}

	return v
}

var iso8601DurationPattern = regexp.MustCompile(`^P(\d+D)?(T(\d+H)?(\d+M)?(\d+S)?)?$`)

func validateISO8601Duration(fl validator.FieldLevel) bool {
	s := fl.Field().String()
	return s != "P" && s != "PT" && iso8601DurationPattern.MatchString(s)
}

var uppercaseReferencePattern = regexp.MustCompile(`^[A-Z0-9][A-Z0-9-_]*$`)

func validateUppercaseReference(fl validator.FieldLevel) bool {
	return uppercaseReferencePattern.MatchString(fl.Field().String())
}
