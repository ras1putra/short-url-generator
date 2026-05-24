package web3

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFaucetExhaustedError_Error(t *testing.T) {
	err := &FaucetExhaustedError{
		FaucetAddress: "0x1234567890abcdef",
		Message:       "Insufficient pool balance",
	}
	msg := err.Error()
	assert.Contains(t, msg, "Faucet exhausted")
	assert.Contains(t, msg, "Insufficient pool balance")
	assert.Contains(t, msg, "0x1234567890abcdef")
}
