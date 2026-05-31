package wallet

import (
	"context"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"urlshortener/internal/modules/wallet/dto"
	"urlshortener/pkg/constants"
	"urlshortener/pkg/helper"
	"urlshortener/pkg/logger"
	"urlshortener/pkg/response"
	tjvalidator "urlshortener/pkg/validator"
)

type WalletServicer interface {
	GetWallet(ctx context.Context, userID uuid.UUID, page, perPage int, q, sortBy, sortDir string) (*dto.WalletWithTransactionsResponse, error)
	RequestWithdraw(ctx context.Context, userID uuid.UUID, req dto.WithdrawRequest) (*dto.WithdrawalPermitResponse, error)
	CreatePendingTransaction(ctx context.Context, userID uuid.UUID, req dto.CreatePendingTransactionRequest) (*dto.TransactionResponse, error)
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
	if err := helper.ParseAndValidate(c, h.validate, &req); err != nil {
		logger.Ctx(c.UserContext()).Warn("RequestWithdraw failed: validation error", zap.Error(err))
		return response.HandleError(c, err, "RequestWithdraw")
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

	page := c.QueryInt("page", constants.DefaultPage)
	perPage := c.QueryInt("per_page", constants.DefaultPerPage)
	if page < 1 {
		page = constants.DefaultPage
	}
	if perPage < 1 || perPage > constants.MaxPerPage {
		perPage = constants.DefaultPerPage
	}

	q := c.Query("q", "")
	sortBy := c.Query("sort_by", "created_at")
	sortDir := c.Query("sort_dir", "desc")

	resp, err := h.svc.GetWallet(c.Context(), userID, page, perPage, q, sortBy, sortDir)
	if err != nil {
		logger.Ctx(c.UserContext()).Error("GetWallet failed: service layer error", zap.Error(err))
		return response.HandleError(c, err, "GetWallet")
	}

	logger.Ctx(c.UserContext()).Info("Wallet details fetched successfully")

	return response.OK(c, resp, "Wallet fetched successfully")
}

func (h *WalletHandler) CreatePendingTransaction(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromCtx(c)
	if err != nil {
		logger.Ctx(c.UserContext()).Warn("CreatePendingTransaction failed: unauthorized access", zap.Error(err))
		return response.Unauthorized(c, err.Error())
	}

	var req dto.CreatePendingTransactionRequest
	if err := helper.ParseAndValidate(c, h.validate, &req); err != nil {
		logger.Ctx(c.UserContext()).Warn("CreatePendingTransaction failed: validation error", zap.Error(err))
		return response.HandleError(c, err, "CreatePendingTransaction")
	}

	txResp, err := h.svc.CreatePendingTransaction(c.Context(), userID, req)
	if err != nil {
		logger.Ctx(c.UserContext()).Error("CreatePendingTransaction failed: service layer error", zap.Error(err))
		return response.HandleError(c, err, "CreatePendingTransaction")
	}

	logger.Ctx(c.UserContext()).Info("Pending transaction registered successfully",
		zap.String("tx_hash", req.TxHash),
		zap.String("type", req.Type),
	)

	return response.OK(c, txResp, "Pending transaction registered successfully")
}

func (h *WalletHandler) ConnectWebSocket(c *websocket.Conn) {
	userIDStr, ok := c.Locals("user_id").(string)
	if !ok {
		zap.L().Warn("WebSocket upgrade aborted: unauthorized")
		_ = c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "Unauthorized"))
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		zap.L().Warn("WebSocket upgrade aborted: invalid user ID")
		return
	}

	cl := &client{
		userID: userID,
		conn:   c,
	}

	GlobalHub.register <- cl
	defer func() {
		GlobalHub.unregister <- cl
		_ = c.Close()
	}()

	zap.L().Info("WebSocket connected successfully for user", zap.String("user_id", userID.String()))

	done := make(chan struct{})
	defer close(done)

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := c.WriteMessage(websocket.PingMessage, nil); err != nil {
					return
				}
			case <-done:
				return
			}
		}
	}()

	for {
		_, _, err := c.ReadMessage()
		if err != nil {
			zap.L().Debug("WebSocket connection closed by client",
				zap.String("user_id", userID.String()),
				zap.Error(err),
			)
			break
		}
	}
}
