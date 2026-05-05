package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"urlshortener/pkg/response"
)

func TestNew(t *testing.T) {
	v := New()
	assert.NotNil(t, v)
}

func TestValidateStruct_Valid(t *testing.T) {
	type ValidStruct struct {
		Name  string `validate:"required,min=2"`
		Email string `validate:"required,email"`
	}

	err := ValidateStruct(ValidStruct{Name: "John", Email: "john@example.com"})
	assert.NoError(t, err)
}

func TestValidateStruct_Invalid(t *testing.T) {
	type InvalidStruct struct {
		Name  string `validate:"required,min=2"`
		Email string `validate:"required,email"`
	}

	err := ValidateStruct(InvalidStruct{Name: "J", Email: "bad-email"})
	assert.Error(t, err)

	appErr, ok := err.(*response.AppError)
	assert.True(t, ok)
	assert.Equal(t, 400, appErr.Code)
	assert.Contains(t, appErr.Message, "Validation failed")
}

func TestValidateStruct_EmptyStruct(t *testing.T) {
	type EmptyStruct struct{}

	err := ValidateStruct(EmptyStruct{})
	assert.NoError(t, err)
}

func TestValidateStruct_RequiredField(t *testing.T) {
	type ReqStruct struct {
		Name string `validate:"required"`
	}

	err := ValidateStruct(ReqStruct{Name: ""})
	assert.Error(t, err)
	assert.Contains(t, err.(*response.AppError).Message, "Validation failed")
}