package validator

import (
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

type testSlugReq struct {
	Slug string `validate:"slug"`
}

func TestNew_RegistersSlugValidator(t *testing.T) {
	v := New()
	err := v.Struct(testSlugReq{Slug: "valid-slug-123"})
	assert.NoError(t, err)

	err = v.Struct(testSlugReq{Slug: "invalid slug!!! "})
	assert.Error(t, err)
}

func TestFormatErrors_NonValidationError(t *testing.T) {
	msg := FormatErrors(assert.AnError)
	assert.Equal(t, "Validation failed", msg)
}

func TestFormatErrors_Required(t *testing.T) {
	v := validator.New()
	type testReq struct {
		Name string `validate:"required"`
	}
	err := v.Struct(testReq{})
	msg := FormatErrors(err)
	assert.True(t, strings.Contains(msg, "Name"))
	assert.True(t, strings.Contains(msg, "required"))
}

func TestFormatErrors_Email(t *testing.T) {
	v := validator.New()
	type testReq struct {
		Email string `validate:"email"`
	}
	err := v.Struct(testReq{Email: "invalid"})
	msg := FormatErrors(err)
	assert.True(t, strings.Contains(msg, "Email"))
	assert.True(t, strings.Contains(msg, "valid email"))
}

func TestFormatErrors_Url(t *testing.T) {
	v := validator.New()
	type testReq struct {
		URL string `validate:"url"`
	}
	err := v.Struct(testReq{URL: "not-a-url"})
	msg := FormatErrors(err)
	assert.True(t, strings.Contains(msg, "URL"))
	assert.True(t, strings.Contains(msg, "valid URL"))
}

func TestFormatErrors_Min(t *testing.T) {
	v := validator.New()
	type testReq struct {
		Code string `validate:"min=3"`
	}
	err := v.Struct(testReq{Code: "ab"})
	msg := FormatErrors(err)
	assert.True(t, strings.Contains(msg, "Code"))
	assert.True(t, strings.Contains(msg, "least"))
}

func TestFormatErrors_Max(t *testing.T) {
	v := validator.New()
	type testReq struct {
		Code string `validate:"max=3"`
	}
	err := v.Struct(testReq{Code: "abcd"})
	msg := FormatErrors(err)
	assert.True(t, strings.Contains(msg, "Code"))
	assert.True(t, strings.Contains(msg, "most"))
}

func TestFormatErrors_OneOf(t *testing.T) {
	v := validator.New()
	type testReq struct {
		Role string `validate:"oneof=admin user"`
	}
	err := v.Struct(testReq{Role: "superadmin"})
	msg := FormatErrors(err)
	assert.True(t, strings.Contains(msg, "Role"))
	assert.True(t, strings.Contains(msg, "one of"))
}

func TestFormatErrors_Slug(t *testing.T) {
	v := New()
	type testReq struct {
		Slug string `validate:"slug"`
	}
	err := v.Struct(testReq{Slug: "invalid slug with spaces"})
	msg := FormatErrors(err)
	assert.True(t, strings.Contains(msg, "Slug"))
	assert.True(t, strings.Contains(msg, "letters, numbers"))
}

func TestFormatErrors_Default(t *testing.T) {
	v := validator.New()
	type testReq struct {
		Field string `validate:"numeric"`
	}
	err := v.Struct(testReq{Field: "abc"})
	msg := FormatErrors(err)
	assert.True(t, strings.Contains(msg, "Field"))
	assert.True(t, strings.Contains(msg, "invalid"))
}

type testEthAddrReq struct {
	Address string `validate:"eth_addr"`
}

type testEthTxHashReq struct {
	TxHash string `validate:"eth_tx_hash"`
}

type testDecimalGtReq struct {
	Decimal float64 `validate:"decimal_gt=0"`
}

func TestEthAddr_Valid(t *testing.T) {
	v := New()
	err := v.Struct(testEthAddrReq{Address: "0x742d35Cc6634C0532925a3b844Bc9e7595f2bD18"})
	assert.NoError(t, err)
}

func TestEthAddr_Invalid(t *testing.T) {
	v := New()
	err := v.Struct(testEthAddrReq{Address: "invalid-address"})
	assert.Error(t, err)
}

func TestEthAddr_Empty(t *testing.T) {
	v := New()
	err := v.Struct(testEthAddrReq{Address: ""})
	assert.Error(t, err)
}

func TestEthAddr_ShortAddress(t *testing.T) {
	v := New()
	err := v.Struct(testEthAddrReq{Address: "0x1234"})
	assert.Error(t, err)
}

func TestEthTxHash_Valid(t *testing.T) {
	v := New()
	err := v.Struct(testEthTxHashReq{TxHash: "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"})
	assert.NoError(t, err)
}

func TestEthTxHash_Invalid(t *testing.T) {
	v := New()
	err := v.Struct(testEthTxHashReq{TxHash: "invalid-hash"})
	assert.Error(t, err)
}

func TestEthTxHash_NoPrefix(t *testing.T) {
	v := New()
	err := v.Struct(testEthTxHashReq{TxHash: "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"})
	assert.Error(t, err)
}

func TestEthTxHash_WrongLength(t *testing.T) {
	v := New()
	err := v.Struct(testEthTxHashReq{TxHash: "0x1234"})
	assert.Error(t, err)
}

func TestEthTxHash_UpperCase(t *testing.T) {
	v := New()
	err := v.Struct(testEthTxHashReq{TxHash: "0xABCDEF1234567890ABCDEF1234567890ABCDEF1234567890ABCDEF1234567890"})
	assert.NoError(t, err)
}

func TestFormatErrors_EthAddr(t *testing.T) {
	v := New()
	err := v.Struct(testEthAddrReq{Address: "0xinvalid"})
	msg := FormatErrors(err)
	assert.Contains(t, msg, "valid EVM address")
}

func TestFormatErrors_EthTxHash(t *testing.T) {
	v := New()
	err := v.Struct(testEthTxHashReq{TxHash: "0xshort"})
	msg := FormatErrors(err)
	assert.Contains(t, msg, "valid 0x-prefixed")
}

func TestSlugValidator_InvalidChars(t *testing.T) {
	v := New()
	err := v.Struct(testSlugReq{Slug: "slug with spaces"})
	assert.Error(t, err)
	msg := FormatErrors(err)
	assert.Contains(t, msg, "letters, numbers")
}

func TestSlugValidator_LeadingDash(t *testing.T) {
	v := New()
	err := v.Struct(testSlugReq{Slug: "-leading-dash"})
	assert.Error(t, err)
}

func TestSlugValidator_TrailingDash(t *testing.T) {
	v := New()
	err := v.Struct(testSlugReq{Slug: "trailing-dash-"})
	assert.Error(t, err)
}

func TestSlugValidator_Empty(t *testing.T) {
	v := New()
	err := v.Struct(testSlugReq{Slug: ""})
	assert.Error(t, err)
}
