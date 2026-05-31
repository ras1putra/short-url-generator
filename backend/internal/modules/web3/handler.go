package web3

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"urlshortener/internal/modules/web3/dto"
	"urlshortener/pkg/constants"
	"urlshortener/pkg/helper"
	"urlshortener/pkg/logger"
	"urlshortener/pkg/response"
	tjvalidator "urlshortener/pkg/validator"
)

type Web3Servicer interface {
	ClaimFaucet(ctx context.Context, userID uuid.UUID, walletAddr string) (*dto.FaucetClaimResponse, error)
	ConfirmFaucet(ctx context.Context, userID uuid.UUID, req *dto.FaucetConfirmRequest) (*dto.FaucetConfirmResponse, error)
	ClaimDevETH(ctx context.Context, walletAddr string) (string, error)
	GetFaucetHistory(ctx context.Context, userID uuid.UUID, page, limit int32, q, sortBy, sortDir string) ([]dto.FaucetHistoryItem, int64, error)
	GetDepositStatus(ctx context.Context) (*dto.DepositStatusResponse, error)
}

type Web3Handler struct {
	svc      Web3Servicer
	validate *validator.Validate
	isDev    bool
}

func NewWeb3Handler(svc Web3Servicer, isDev bool) *Web3Handler {
	return &Web3Handler{svc: svc, validate: tjvalidator.New(), isDev: isDev}
}

func (h *Web3Handler) ClaimFaucet(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromCtx(c)
	if err != nil {
		logger.Ctx(c.UserContext()).Warn("ClaimFaucet failed: unauthorized access", zap.Error(err))
		return response.Unauthorized(c, err.Error())
	}

	var req dto.FaucetClaimRequest
	if err := helper.ParseAndValidate(c, h.validate, &req); err != nil {
		logger.Ctx(c.UserContext()).Warn("ClaimFaucet failed: validation error", zap.Error(err))
		return err
	}

	resp, err := h.svc.ClaimFaucet(c.Context(), userID, req.WalletAddr)
	if err != nil {
		logger.Ctx(c.UserContext()).Error("ClaimFaucet failed: service layer error", zap.Error(err))
		return response.HandleError(c, err, "ClaimFaucet")
	}

	logger.Ctx(c.UserContext()).Info("Faucet claimed successfully",
		zap.String("wallet", req.WalletAddr),
	)

	return response.OK(c, resp, "Faucet requested successfully")
}

func (h *Web3Handler) ConfirmFaucet(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromCtx(c)
	if err != nil {
		logger.Ctx(c.UserContext()).Warn("ConfirmFaucet failed: unauthorized access", zap.Error(err))
		return response.Unauthorized(c, err.Error())
	}

	var req dto.FaucetConfirmRequest
	if err := helper.ParseAndValidate(c, h.validate, &req); err != nil {
		logger.Ctx(c.UserContext()).Warn("ConfirmFaucet failed: validation error", zap.Error(err))
		return err
	}

	resp, err := h.svc.ConfirmFaucet(c.Context(), userID, &req)
	if err != nil {
		logger.Ctx(c.UserContext()).Error("ConfirmFaucet failed: service layer error", zap.Error(err))
		return response.HandleError(c, err, "ConfirmFaucet")
	}

	logger.Ctx(c.UserContext()).Info("Faucet claim confirmed successfully",
		zap.String("wallet", req.WalletAddr),
		zap.String("tx_hash", req.TxHash),
	)

	return response.OK(c, resp, "Faucet claim confirmed")
}

func (h *Web3Handler) GetFaucetHistory(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromCtx(c)
	if err != nil {
		logger.Ctx(c.UserContext()).Warn("GetFaucetHistory failed: unauthorized access", zap.Error(err))
		return response.Unauthorized(c, err.Error())
	}

	page := c.QueryInt("page", constants.DefaultPage)
	limit := c.QueryInt("per_page", constants.ClaimsPerPage)
	if page < 1 {
		page = constants.DefaultPage
	}
	if limit < 1 || limit > constants.MaxPerPage {
		limit = constants.ClaimsPerPage
	}

	q := c.Query("q", "")
	sortBy := c.Query("sort_by", "claimed_at")
	sortDir := c.Query("sort_dir", "desc")

	history, total, err := h.svc.GetFaucetHistory(c.Context(), userID, int32(page), int32(limit), q, sortBy, sortDir)
	if err != nil {
		logger.Ctx(c.UserContext()).Error("GetFaucetHistory failed: service layer error", zap.Error(err))
		return response.HandleError(c, err, "GetFaucetHistory")
	}

	totalPages := (int(total) + limit - 1) / limit

	logger.Ctx(c.UserContext()).Info("Faucet history retrieved successfully",
		zap.Int("page", page),
		zap.Int("per_page", limit),
		zap.Int64("total", total),
		zap.String("q", q),
		zap.String("sort_by", sortBy),
		zap.String("sort_dir", sortDir),
	)

	return response.OK(c, dto.FaucetHistoryResponse{
		Claims:     history,
		Total:      total,
		Page:       page,
		PerPage:    limit,
		TotalPages: totalPages,
	}, "Faucet claim history retrieved")
}

func (h *Web3Handler) DepositStatus(c *fiber.Ctx) error {
	resp, err := h.svc.GetDepositStatus(c.Context())
	if err != nil {
		logger.Ctx(c.UserContext()).Error("DepositStatus failed: service layer error", zap.Error(err))
		return response.HandleError(c, err, "DepositStatus")
	}

	logger.Ctx(c.UserContext()).Info("Deposit listener status fetched successfully",
		zap.String("status", resp.Status),
		zap.Uint64("last_block", resp.LastBlock),
	)

	return response.OK(c, resp, "Deposit listener status")
}

func (h *Web3Handler) ClaimDevETH(c *fiber.Ctx) error {
	if !h.isDev {
		logger.Ctx(c.UserContext()).Warn("ClaimDevETH failed: dev endpoint invoked in production environment")
		return response.Forbidden(c, "This endpoint is only available in development mode")
	}

	_, err := helper.UserIDFromCtx(c)
	if err != nil {
		logger.Ctx(c.UserContext()).Warn("ClaimDevETH failed: unauthorized access", zap.Error(err))
		return response.Unauthorized(c, err.Error())
	}

	var req dto.FaucetClaimRequest
	if err := helper.ParseAndValidate(c, h.validate, &req); err != nil {
		logger.Ctx(c.UserContext()).Warn("ClaimDevETH failed: validation error", zap.Error(err))
		return err
	}

	txHash, err := h.svc.ClaimDevETH(c.Context(), req.WalletAddr)
	if err != nil {
		logger.Ctx(c.UserContext()).Error("ClaimDevETH failed: service layer error", zap.Error(err))
		return response.HandleError(c, err, "ClaimDevETH")
	}

	logger.Ctx(c.UserContext()).Info("Dev ETH claimed successfully",
		zap.String("wallet", req.WalletAddr),
		zap.String("tx_hash", txHash),
	)

	return response.OK(c, dto.DevETHResponse{TxHash: txHash}, "Dev ETH sent successfully")
}
