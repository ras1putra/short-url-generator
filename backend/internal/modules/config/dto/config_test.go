package dto

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigResponse_JSON(t *testing.T) {
	resp := ConfigResponse{
		ContractPayment:    "0xpayment",
		ContractToken:      "0xtoken",
		ContractFaucet:     "0xfaucet",
		ContractWithdrawer: "0xwithdrawer",
		TokenSymbol:        "SURL",
		TokenDecimals:      18,
		PaymentChain: PaymentChainResponse{
			ChainID:     31337,
			ChainName:   "Hardhat",
			RPCURL:      "http://localhost:8545",
			ExplorerURL: "http://localhost:5100",
			Currency: CurrencyResponse{
				Name:     "SURL",
				Symbol:   "SURL",
				Decimals: 18,
			},
		},
	}
	data, err := json.Marshal(resp)
	assert.NoError(t, err)
	assert.Contains(t, string(data), `"contract_payment":"0xpayment"`)
	assert.Contains(t, string(data), `"chain_id":31337`)
	assert.Contains(t, string(data), `"token_symbol":"SURL"`)
}

func TestPaymentChainResponse_JSON(t *testing.T) {
	resp := PaymentChainResponse{
		ChainID:     1,
		ChainName:   "Ethereum",
		RPCURL:      "https://eth.llamarpc.com",
		ExplorerURL: "https://etherscan.io",
		Currency: CurrencyResponse{
			Name:     "Ether",
			Symbol:   "ETH",
			Decimals: 18,
		},
	}
	data, err := json.Marshal(resp)
	assert.NoError(t, err)
	assert.Contains(t, string(data), `"chain_id":1`)
	assert.Contains(t, string(data), `"chain_name":"Ethereum"`)
}

func TestCurrencyResponse_JSON(t *testing.T) {
	resp := CurrencyResponse{Name: "USDC", Symbol: "USDC", Decimals: 6}
	data, err := json.Marshal(resp)
	assert.NoError(t, err)
	assert.Contains(t, string(data), `"decimals":6`)
}
