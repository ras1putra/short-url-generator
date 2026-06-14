package config

import (
	"github.com/gofiber/fiber/v2"

	"urlshortener/internal/config"
	"urlshortener/internal/modules/config/dto"
	"urlshortener/pkg/response"
)

type Handler struct {
	cfg *config.Config
}

func NewHandler(cfg *config.Config) *Handler {
	return &Handler{cfg: cfg}
}

func (h *Handler) GetConfig(c *fiber.Ctx) error {
	resp := dto.ConfigResponse{
		ContractPayment:  h.cfg.ContractPayment,
		ContractToken:    h.cfg.ContractToken,
		ContractFaucet:   h.cfg.ContractFaucet,
		ContractNFTPass:  h.cfg.ContractNFTPass,
		TokenSymbol:      h.cfg.TokenSymbol,
		TokenDecimals:    h.cfg.TokenDecimals,
		PlatformFee:      h.cfg.PlatformFee,
		TurnstileSiteKey: h.cfg.TurnstileSiteKey,
		PaymentChain: dto.PaymentChainResponse{
			ChainID:     h.cfg.ChainID,
			ChainName:   h.cfg.ChainName,
			RPCURL:      h.cfg.ChainRPCURL,
			ExplorerURL: h.cfg.ExplorerURL,
			Currency: dto.CurrencyResponse{
				Name:     "Ether",
				Symbol:   "ETH",
				Decimals: 18,
			},
		},
	}
	return response.OK(c, resp, "Config fetched")
}
