package wallet

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"urlshortener/internal/modules/wallet/dto"
	"urlshortener/pkg/helper"
	"urlshortener/pkg/logger"
	"urlshortener/pkg/response"
	tjvalidator "urlshortener/pkg/validator"
)

type WalletServicer interface {
	GetWallet(ctx context.Context, userID uuid.UUID) (*dto.WalletWithTransactionsResponse, error)
	RequestWithdraw(ctx context.Context, userID uuid.UUID, req dto.WithdrawRequest) (*dto.WithdrawalPermitResponse, error)
}

type WalletHandler struct {
	svc      WalletServicer
	validate *validator.Validate
}

func NewWalletHandler(svc WalletServicer) *WalletHandler {
	return &WalletHandler{svc: svc, validate: tjvalidator.New()}
}

func (h *WalletHandler) RequestWithdraw(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromCtx(c)
	if err != nil {
		logger.Ctx(c.UserContext()).Warn("RequestWithdraw failed: unauthorized access", zap.Error(err))
		return response.Unauthorized(c, err.Error())
	}

	var req dto.WithdrawRequest
	if err := c.BodyParser(&req); err != nil {
		logger.Ctx(c.UserContext()).Warn("RequestWithdraw failed: invalid request body", zap.Error(err))
		return response.HandleError(c, response.NewAppError(400, "Invalid request body"), "RequestWithdraw")
	}

	if err := h.validate.Struct(&req); err != nil {
		logger.Ctx(c.UserContext()).Warn("RequestWithdraw failed: validation error", zap.Error(err))
		return response.HandleError(c, response.NewAppError(400, err.Error()), "RequestWithdraw")
	}

	permit, err := h.svc.RequestWithdraw(c.Context(), userID, req)
	if err != nil {
		logger.Ctx(c.UserContext()).Error("RequestWithdraw failed: service layer error", zap.Error(err))
		return response.HandleError(c, err, "RequestWithdraw")
	}

	logger.Ctx(c.UserContext()).Info("Withdrawal requested successfully",
		zap.String("amount", req.Amount.String()),
		zap.String("wallet_addr", req.WalletAddr),
	)

	return response.OK(c, permit, "Withdrawal permit created - submit on-chain to claim")
}

func (h *WalletHandler) GetWallet(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromCtx(c)
	if err != nil {
		logger.Ctx(c.UserContext()).Warn("GetWallet failed: unauthorized access", zap.Error(err))
		return response.Unauthorized(c, err.Error())
	}

	resp, err := h.svc.GetWallet(c.Context(), userID)
	if err != nil {
		logger.Ctx(c.UserContext()).Error("GetWallet failed: service layer error", zap.Error(err))
		return response.HandleError(c, err, "GetWallet")
	}

	logger.Ctx(c.UserContext()).Info("Wallet details fetched successfully")

	return response.OK(c, resp, "Wallet fetched successfully")
}
