package controller

import (
	"strings"
	"unicode/utf8"
	"unitly/internal/domain"

	"github.com/go-playground/validator/v10"
)

var appValidator *validator.Validate

func init() {
	appValidator = validator.New(validator.WithRequiredStructEnabled())
	if appValidator != nil {
		appValidator.RegisterValidation("profile_name", validateProfileName)
	}
}

func validateProfileName(fl validator.FieldLevel) bool {
	value := strings.TrimSpace(fl.Field().String())
	l := utf8.RuneCountInString(value)

	if l < 2 || l > 20 {
		return false
	}

	return true
}

func validateStruct(s any) error {
	if err := appValidator.Struct(s); err != nil {
		return domain.NewError("VALIDATION_ERROR", err.Error())
	}

	return nil
}
