package web3

import "fmt"

type FaucetExhaustedError struct {
	FaucetAddress string
	Message       string
}

func (e *FaucetExhaustedError) Error() string {
	return fmt.Sprintf("Faucet exhausted: %s. Please contact administrator to fund the faucet address: %s", e.Message, e.FaucetAddress)
}
