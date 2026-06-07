package controller

import "github.com/go-playground/validator/v10"

var appValidator *validator.Validate

func init() {
	appValidator = validator.New(validator.WithRequiredStructEnabled())
}

func validateStruct(s any) error {
	return appValidator.Struct(s)
}
