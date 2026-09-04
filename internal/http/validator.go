package http

import (
	"errors"
	"reflect"
	"strconv"
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
	registerTrimmedNameValidator(v, transEn)
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

func registerTrimmedNameValidator(v *validator.Validate, transEn ut.Translator) {
	v.RegisterValidation("trimmed_name", trimmedNameValidator)
	v.RegisterTranslation(
		"trimmed_name",
		transEn,
		func(ut ut.Translator) error {
			return ut.Add("trimmed_name", "{0} must be between {1} and {2} characters long after trimming", true)
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			params := strings.Split(fe.Param(), ":")

			minVal, maxVal := "?", "?"
			if len(params) == 2 {
				minVal, maxVal = params[0], params[1]
			} else {
				return fe.Tag()
			}

			t, err := ut.T(fe.Tag(), fe.Field(), minVal, maxVal)
			if err != nil {
				return fe.Tag()
			}
			return t
		},
	)
}

// TrimmedNameValidator validates that a string, after stripping leading and trailing
// whitespace, falls within [min, max] character length.
//
// Supported tag formats:
//   - validate:"trimmed_name=2:50"
func trimmedNameValidator(fl validator.FieldLevel) bool {
	field := fl.Field()

	// 1. Ensure target field is a string
	if field.Kind() != reflect.String {
		return false
	}

	params := strings.Split(fl.Param(), ":")

	if len(params) != 2 {
		return false
	}

	minLen, err := strconv.Atoi(params[0])
	if err != nil {
		return false
	}

	maxLen, err := strconv.Atoi(params[1])
	if err != nil {
		return false
	}

	trimmed := strings.TrimSpace(field.String())
	len := utf8.RuneCountInString(trimmed)

	// 5. Verify constraints
	return len >= minLen && len <= maxLen
}
