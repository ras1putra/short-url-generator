package dto

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFaucetClaimRequest_Validation(t *testing.T) {
	req := FaucetClaimRequest{WalletAddr: "0x1234567890abcdef1234567890abcdef12345678"}
	assert.Len(t, req.WalletAddr, 42)
}

func TestFaucetClaimResponse_JSON(t *testing.T) {
	resp := FaucetClaimResponse{
		Wallet:     "0x1234",
		Amount:     "20",
		Nonce:      "1",
		Deadline:   "1234567890",
		Signature:  "0xsig",
		FaucetAddr: "0xfaucet",
		ChainID:    31337,
	}
	data, err := json.Marshal(resp)
	assert.NoError(t, err)
	assert.Contains(t, string(data), `"wallet":"0x1234"`)
	assert.Contains(t, string(data), `"chain_id":31337`)
}

func TestFaucetConfirmRequest_Validation(t *testing.T) {
	req := FaucetConfirmRequest{
		TxHash:     "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		WalletAddr: "0x1234567890abcdef1234567890abcdef12345678",
	}
	assert.Len(t, req.TxHash, 66)
	assert.Len(t, req.WalletAddr, 42)
}

func TestFaucetConfirmResponse_JSON(t *testing.T) {
	resp := FaucetConfirmResponse{Status: "confirmed", TxHash: "0xtxhash"}
	data, err := json.Marshal(resp)
	assert.NoError(t, err)
	assert.Contains(t, string(data), `"status":"confirmed"`)
	assert.Contains(t, string(data), `"tx_hash":"0xtxhash"`)
}

func TestFaucetHistoryItem_JSON(t *testing.T) {
	item := FaucetHistoryItem{ID: "1", Amount: "20", TxHash: "0xtx", ClaimedAt: "2024-01-01T00:00:00Z"}
	data, err := json.Marshal(item)
	assert.NoError(t, err)
	assert.Contains(t, string(data), `"id":"1"`)
}

func TestFaucetHistoryResponse_JSON(t *testing.T) {
	resp := FaucetHistoryResponse{
		Claims:     []FaucetHistoryItem{{ID: "1", Amount: "20"}},
		Total:      1,
		Page:       1,
		PerPage:    10,
		TotalPages: 1,
	}
	data, err := json.Marshal(resp)
	assert.NoError(t, err)
	assert.Contains(t, string(data), `"total":1`)
}

func TestDevETHResponse_JSON(t *testing.T) {
	resp := DevETHResponse{TxHash: "0xtx"}
	data, err := json.Marshal(resp)
	assert.NoError(t, err)
	assert.Contains(t, string(data), `"tx_hash":"0xtx"`)
}

func TestDepositStatusResponse_JSON(t *testing.T) {
	resp := DepositStatusResponse{Contract: "0xcontract", Status: "active", LastBlock: 100}
	data, err := json.Marshal(resp)
	assert.NoError(t, err)
	assert.Contains(t, string(data), `"last_block":100`)
}
