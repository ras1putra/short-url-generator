package config

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"urlshortener/internal/config"
	"urlshortener/pkg/response"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestGetConfig(t *testing.T) {
	_ = zap.ReplaceGlobals(zap.NewNop())

	cfg := &config.Config{
		ContractPayment: "0x123",
		ContractToken:   "0x456",
		ContractFaucet:  "0x789",
		TokenSymbol:     "TST",
		TokenDecimals:   18,
		PlatformFee:     0.005,
		ChainID:         31337,
		ChainName:       "Hardhat",
		ChainRPCURL:     "http://localhost:8545",
		ExplorerURL:     "http://localhost:5100",
	}

	handler := NewHandler(cfg)

	app := fiber.New(fiber.Config{ErrorHandler: response.ErrorHandler})
	app.Get("/api/config", handler.GetConfig)

	req := httptest.NewRequest("GET", "/api/config", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var body map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&body)
	require.NoError(t, err)
	assert.Equal(t, "Config fetched", body["message"])

	data, ok := body["data"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "0x123", data["contract_payment"])
	assert.Equal(t, "0x456", data["contract_token"])
	assert.Equal(t, "0x789", data["contract_faucet"])
	assert.Equal(t, "TST", data["token_symbol"])
	assert.Equal(t, float64(18), data["token_decimals"])

	chain, ok := data["payment_chain"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, float64(31337), chain["chain_id"])
	assert.Equal(t, "Hardhat", chain["chain_name"])
	assert.Equal(t, "http://localhost:8545", chain["rpc_url"])
	assert.Equal(t, "http://localhost:5100", chain["explorer_url"])
}
