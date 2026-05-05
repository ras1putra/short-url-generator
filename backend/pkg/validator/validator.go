package validator

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
	"urlshortener/pkg/response"
)

var slugRegex = regexp.MustCompile(`^[a-zA-Z0-9]+(?:-[a-zA-Z0-9]+)*$`)

func New() *validator.Validate {
	v := validator.New()
	_ = v.RegisterValidation("slug", func(fl validator.FieldLevel) bool {
		return slugRegex.MatchString(fl.Field().String())
	})
	return v
}

func ValidateStruct(s interface{}) error {
	validate := New()
	err := validate.Struct(s)
	if err != nil {
		return response.NewAppError(400, FormatErrors(err))
	}
	return nil
}

func FormatErrors(err error) string {
	verrs, ok := err.(validator.ValidationErrors)
	if !ok {
		return "Validation failed"
	}

	var errs []string
	for _, e := range verrs {
		field := e.Field()
		switch e.Tag() {
		case "required":
			errs = append(errs, fmt.Sprintf("%s is required", field))
		case "email":
			errs = append(errs, fmt.Sprintf("%s must be a valid email", field))
		case "url":
			errs = append(errs, fmt.Sprintf("%s must be a valid URL", field))
		case "min":
			errs = append(errs, fmt.Sprintf("%s must be at least %s", field, e.Param()))
		case "max":
			errs = append(errs, fmt.Sprintf("%s must be at most %s", field, e.Param()))
		case "oneof":
			errs = append(errs, fmt.Sprintf("%s must be one of: %s", field, e.Param()))
		case "slug":
			errs = append(errs, fmt.Sprintf("%s can only contain letters, numbers, and single dashes between characters", field))
		default:
			errs = append(errs, fmt.Sprintf("%s is invalid", field))
		}
	}

	return "Validation failed: " + strings.Join(errs, ", ")
}
