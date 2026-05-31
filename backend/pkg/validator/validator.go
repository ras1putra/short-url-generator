package validator

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-playground/validator/v10"
	"github.com/shopspring/decimal"
)

var slugRegex = regexp.MustCompile(`^[a-zA-Z0-9]+(?:-[a-zA-Z0-9]+)*$`)

func New() *validator.Validate {
	v := validator.New()
	_ = v.RegisterValidation("slug", func(fl validator.FieldLevel) bool {
		return slugRegex.MatchString(fl.Field().String())
	})
	_ = v.RegisterValidation("decimal_gt", func(fl validator.FieldLevel) bool {
		val, ok := fl.Field().Interface().(decimal.Decimal)
		if !ok {
			return false
		}
		param := fl.Param()
		if param == "" {
			return false
		}
		compareVal, err := decimal.NewFromString(param)
		if err != nil {
			return false
		}
		return val.GreaterThan(compareVal)
	})
	_ = v.RegisterValidation("eth_addr", func(fl validator.FieldLevel) bool {
		return common.IsHexAddress(fl.Field().String())
	})
	_ = v.RegisterValidation("eth_tx_hash", func(fl validator.FieldLevel) bool {
		return regexp.MustCompile(`^0x[0-9a-fA-F]{64}$`).MatchString(fl.Field().String())
	})
	return v
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
		case "decimal_gt":
			errs = append(errs, fmt.Sprintf("%s must be greater than %s", field, e.Param()))
		case "eth_addr":
			errs = append(errs, fmt.Sprintf("%s must be a valid EVM address", field))
		case "eth_tx_hash":
			errs = append(errs, fmt.Sprintf("%s must be a valid 0x-prefixed 32-byte transaction hash", field))
		default:
			errs = append(errs, fmt.Sprintf("%s is invalid", field))
		}
	}

	return "Validation failed: " + strings.Join(errs, ", ")
}
