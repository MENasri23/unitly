package controller

import (
	"unitly/internal/domain"

	"github.com/go-playground/validator/v10"
)

var appValidator *validator.Validate

func init() {
	appValidator = validator.New(validator.WithRequiredStructEnabled())
}

func validateStruct(s any) error {
	if err := appValidator.Struct(s); err != nil {
		return domain.NewError(domain.ErrorCode("VALIDATION_ERROR"), err.Error())
	}

	return nil
}
