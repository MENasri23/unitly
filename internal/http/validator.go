package http

import (
	"errors"
	"reflect"
	"strings"
	"unicode/utf8"
	"unitly/internal/domain"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
)

var v *validator.Validate
var uni *ut.UniversalTranslator

func init() {
	v = validator.New(validator.WithRequiredStructEnabled())
	v.RegisterTagNameFunc(jsonTagFieldName)

	enLocale := en.New()
	uni = ut.New(enLocale, enLocale)
	transEn, _ := uni.GetTranslator("en")
	en_translations.RegisterDefaultTranslations(v, transEn)

	registerProfileNameValidation(v, transEn)

}

func jsonTagFieldName(fld reflect.StructField) string {
	name, _, _ := strings.Cut(fld.Tag.Get("json"), ",")
	if name == "-" {
		return ""
	}
	return name
}

func registerProfileNameValidation(v *validator.Validate, transEn ut.Translator) {
	v.RegisterValidation("profile_name", validateProfileName)
	v.RegisterTranslation(
		"profile_name",
		transEn,
		func(ut ut.Translator) error {
			return ut.Add("profile_name", "{0} must be between 2 and 20 characters long", true)
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			t, err := ut.T(fe.Tag(), fe.Field())
			if err != nil {
				return fe.Tag()
			}
			return t
		},
	)

	v.RegisterTagNameFunc(jsonTagFieldName)
}

func validateProfileName(fl validator.FieldLevel) bool {
	value := strings.TrimSpace(fl.Field().String())
	l := utf8.RuneCountInString(value)

	return l >= 2 && l <= 20
}

func ValidateStruct(s any) error {
	if err := v.Struct(s); err != nil {
		return domain.NewError("VALIDATION_ERROR", translate(err))
	}

	return nil
}

func translate(err error) string {
	var errs validator.ValidationErrors
	if errors.As(err, &errs) {
		trans, _ := uni.GetTranslator("en")
		for _, e := range errs {
			translated := e.Translate(trans)
			if translated != "" {
				return translated
			}
		}
	}
	return err.Error()
}
