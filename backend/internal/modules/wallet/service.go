package wallet

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/sqlc-dev/pqtype"
	"go.uber.org/zap"

	"urlshortener/internal/modules/wallet/dto"
	"urlshortener/internal/repository"
	web3client "urlshortener/internal/web3"
	"urlshortener/pkg/constants"
	"urlshortener/pkg/helper"
	"urlshortener/pkg/logger"
	"urlshortener/pkg/response"
)

type WalletService struct {
	repo        repository.Querier
	db          *sql.DB
	withdrawer  *web3client.WithdrawerService
	platformFee float64
}

func NewWalletService(repo repository.Querier, db *sql.DB, withdrawer *web3client.WithdrawerService, platformFee float64) *WalletService {
	return &WalletService{repo: repo, db: db, withdrawer: withdrawer, platformFee: platformFee}
}

func (s *WalletService) GetWallet(ctx context.Context, userID uuid.UUID, page, perPage int, q, sortBy, sortDir string) (*dto.WalletWithTransactionsResponse, error) {
	wallet, err := s.repo.GetWalletByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, response.NewAppError(404, "Wallet not found")
		}
		logger.Ctx(ctx).Error("Failed to get wallet", zap.Error(err))
		return nil, response.NewAppError(500, "Internal server error")
	}

	hasFilter := q != "" || sortBy != "created_at" || sortDir != "desc"

	var total int64
	if hasFilter {
		total, err = s.repo.CountTransactionsByUserFiltered(ctx, repository.CountTransactionsByUserFilteredParams{
			UserID: userID,
			Q:      sql.NullString{String: q, Valid: q != ""},
		})
	} else {
		total, err = s.repo.CountTransactionsByUser(ctx, userID)
	}
	if err != nil {
		logger.Ctx(ctx).Error("Failed to count transactions", zap.Error(err))
		return nil, response.NewAppError(500, "Failed to load transactions")
	}

	offset := int32((page - 1) * perPage)

	var txs []repository.Transaction
	if hasFilter {
		txs, err = s.repo.ListTransactionsByUserFiltered(ctx, repository.ListTransactionsByUserFilteredParams{
			UserID:  userID,
			Q:       sql.NullString{String: q, Valid: q != ""},
			SortBy:  sortBy,
			SortDir: sortDir,
			Limit:   int32(perPage),
			Offset:  offset,
		})
	} else {
		txs, err = s.repo.ListTransactionsByUser(ctx, repository.ListTransactionsByUserParams{
			UserID: userID,
			Limit:  int32(perPage),
			Offset: offset,
		})
	}
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list transactions", zap.Error(err))
		return nil, response.NewAppError(500, "Failed to load transactions")
	}

	frozen, err := s.getFrozenBalance(ctx, userID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to calculate budget freeze", zap.Error(err))
		frozen = decimal.Zero
	}

	wr := dto.MapWalletToResponse(wallet)
	available := wr.Balance
	totalBalance := available.Add(frozen)

	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}

	return &dto.WalletWithTransactionsResponse{
		Balance:      totalBalance,
		Available:    available,
		Frozen:       frozen,
		UpdatedAt:    wr.UpdatedAt,
		Transactions: dto.MapTransactionsToResponse(txs),
		Total:        total,
		Page:         page,
		PerPage:      perPage,
		TotalPages:   totalPages,
	}, nil
}

func (s *WalletService) RequestWithdraw(ctx context.Context, userID uuid.UUID, req dto.WithdrawRequest) (*dto.WithdrawalPermitResponse, error) {
	wallet, err := s.repo.GetWalletByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, response.NewAppError(404, "Wallet not found")
		}
		logger.Ctx(ctx).Error("Failed to get wallet", zap.Error(err))
		return nil, response.NewAppError(500, "Internal server error")
	}

	balanceDec := helper.ParseDecimal(wallet.Balance)
	available := balanceDec

	fee := decimal.NewFromFloat(s.platformFee)
	totalDeduction := req.Amount.Add(fee)

	if available.LessThan(totalDeduction) {
		return nil, response.NewAppError(400, fmt.Sprintf("Insufficient available balance: requested %s SURL, platform fee %s SURL, total required %s SURL, but you only have %s SURL available", req.Amount.StringFixed(4), fee.StringFixed(4), totalDeduction.StringFixed(4), available.StringFixed(4)))
	}

	amountBig := req.Amount.Mul(decimal.NewFromInt(10).Pow(decimal.NewFromInt(18)))
	amountWei := new(big.Int)
	amountWei, ok := amountWei.SetString(amountBig.String(), 10)
	if !ok {
		return nil, response.NewAppError(500, "Failed to parse amount")
	}

	permit, err := s.withdrawer.CreatePermit(ctx, req.WalletAddr, amountWei)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create withdrawal permit", zap.Error(err))
		return nil, response.NewAppError(500, "Failed to create withdrawal permit")
	}

	requestID := uuid.NewString()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, response.NewAppError(500, "Failed to start withdrawal reservation")
	}
	defer func() { _ = tx.Rollback() }()

	q := s.repo.(*repository.Queries).WithTx(tx)

	metaBytes, _ := json.Marshal(map[string]interface{}{
		"request_id":  requestID,
		"wallet_addr": req.WalletAddr,
		"nonce":       permit.Nonce,
		"deadline":    permit.Deadline,
		"amount":      req.Amount.String(),
		"fee":         fee.String(),
	})

	negTotalDeduction := totalDeduction.Neg()
	if _, err = q.UpdateWalletBalance(ctx, repository.UpdateWalletBalanceParams{
		UserID:  userID,
		Balance: helper.FormatDecimal(negTotalDeduction),
	}); err != nil {
		logger.Ctx(ctx).Error("Database error reserving withdrawal balance", zap.Error(err))
		return nil, response.NewAppError(500, "Failed to reserve withdrawal balance")
	}

	negAmount := req.Amount.Neg()
	if _, err = q.CreatePendingTransaction(ctx, repository.CreatePendingTransactionParams{
		UserID: userID,
		Amount: helper.FormatDecimal(negAmount),
		Type:   constants.TxTypeWithdrawal,
		TxHash: sql.NullString{},
		Metadata: pqtype.NullRawMessage{
			RawMessage: metaBytes,
			Valid:      true,
		},
	}); err != nil {
		logger.Ctx(ctx).Error("Database error creating pending withdrawal transaction", zap.Error(err))
		return nil, response.NewAppError(500, "Failed to create pending withdrawal")
	}

	if fee.IsPositive() {
		negFee := fee.Neg()
		if _, err = q.CreatePendingTransaction(ctx, repository.CreatePendingTransactionParams{
			UserID: userID,
			Amount: helper.FormatDecimal(negFee),
			Type:   constants.TxTypeWithdrawalFee,
			TxHash: sql.NullString{},
			Metadata: pqtype.NullRawMessage{
				RawMessage: metaBytes,
				Valid:      true,
			},
		}); err != nil {
			logger.Ctx(ctx).Error("Database error creating pending withdrawal fee transaction", zap.Error(err))
			return nil, response.NewAppError(500, "Failed to create pending withdrawal fee")
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, response.NewAppError(500, "Failed to finalize withdrawal reservation")
	}

	return &dto.WithdrawalPermitResponse{
		RequestID: requestID,
		Wallet:    permit.Wallet,
		Amount:    permit.Amount,
		Nonce:     permit.Nonce,
		Deadline:  permit.Deadline,
		Signature: permit.Signature,
		Contract:  permit.Contract,
		ChainID:   permit.ChainID,
	}, nil
}

func (s *WalletService) CreatePendingTransaction(ctx context.Context, userID uuid.UUID, req dto.CreatePendingTransactionRequest) (*dto.TransactionResponse, error) {
	if req.Type == constants.TxTypeWithdrawal {
		if req.WalletAddr == "" {
			return nil, response.NewAppError(400, "Wallet address is required for withdrawal confirmation")
		}
		if req.RequestID == "" {
			return nil, response.NewAppError(400, "Request ID is required for withdrawal confirmation")
		}

		amountBig := req.Amount.Mul(decimal.NewFromInt(10).Pow(decimal.NewFromInt(18)))
		amountWei := new(big.Int)
		amountWei, ok := amountWei.SetString(amountBig.StringFixed(0), 10)
		if !ok {
			return nil, response.NewAppError(500, "Failed to parse withdrawal amount to wei")
		}

		err := s.withdrawer.VerifyWithdrawal(ctx, req.TxHash, req.WalletAddr, amountWei)
		if err != nil {
			logger.Ctx(ctx).Error("Failed to verify on-chain withdrawal", zap.String("tx_hash", req.TxHash), zap.Error(err))
			return nil, response.NewAppError(400, fmt.Sprintf("On-chain withdrawal verification failed: %s", err.Error()))
		}

		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			logger.Ctx(ctx).Error("Database error starting transaction for withdrawal confirmation", zap.Error(err))
			return nil, response.NewAppError(500, "Internal server error")
		}
		defer func() { _ = tx.Rollback() }()

		q := s.repo
		if queriesInstance, ok := s.repo.(*repository.Queries); ok {
			q = queriesInstance.WithTx(tx)
		} else {
			return nil, response.NewAppError(500, "Internal server error")
		}

		pendingWithdrawal, err := q.GetPendingWithdrawalByRequestID(ctx, repository.GetPendingWithdrawalByRequestIDParams{
			UserID:  userID,
			Column2: req.RequestID,
		})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, response.NewAppError(404, "Pending withdrawal request not found or already finalized")
			}
			logger.Ctx(ctx).Error("Database error finding pending withdrawal by request ID", zap.Error(err))
			return nil, response.NewAppError(500, "Internal server error")
		}

		pendingAmount := helper.ParseDecimal(pendingWithdrawal.Amount).Abs()
		if !pendingAmount.Equal(req.Amount) {
			return nil, response.NewAppError(400, "Withdrawal amount does not match pending request")
		}

		dbTx, err := q.UpdateTransactionHashAndStatusByID(ctx, repository.UpdateTransactionHashAndStatusByIDParams{
			ID:     pendingWithdrawal.ID,
			TxHash: sql.NullString{String: req.TxHash, Valid: true},
			Status: constants.TxStatusConfirmed,
		})
		if err != nil {
			logger.Ctx(ctx).Error("Database error finalising withdrawal transaction", zap.Error(err))
			return nil, response.NewAppError(500, "Internal server error")
		}

		feeRows, err := q.ListPendingWithdrawalFeesByRequestID(ctx, repository.ListPendingWithdrawalFeesByRequestIDParams{
			UserID:  userID,
			Column2: req.RequestID,
		})
		if err != nil {
			logger.Ctx(ctx).Error("Database error listing pending withdrawal fee rows", zap.Error(err))
			return nil, response.NewAppError(500, "Internal server error")
		}
		for _, feeRow := range feeRows {
			_, err = q.UpdateTransactionHashAndStatusByID(ctx, repository.UpdateTransactionHashAndStatusByIDParams{
				ID:     feeRow.ID,
				TxHash: sql.NullString{String: req.TxHash, Valid: true},
				Status: constants.TxStatusConfirmed,
			})
			if err != nil {
				logger.Ctx(ctx).Error("Database error finalising withdrawal fee transaction", zap.Error(err))
				return nil, response.NewAppError(500, "Internal server error")
			}
		}

		if err := tx.Commit(); err != nil {
			logger.Ctx(ctx).Error("Database error committing withdrawal confirmation transaction", zap.Error(err))
			return nil, response.NewAppError(500, "Internal server error")
		}

		resp := dto.MapTransactionToResponse(dbTx)
		GlobalHub.BroadcastToUser(userID, constants.WSEventWalletUpdate, nil)
		return &resp, nil
	}

	amount := req.Amount
	amountStr := helper.FormatDecimal(amount)
	metaBytes, _ := json.Marshal(map[string]interface{}{
		"tx_hash": req.TxHash,
	})

	tx, err := s.repo.CreatePendingTransaction(ctx, repository.CreatePendingTransactionParams{
		UserID: userID,
		Amount: amountStr,
		Type:   req.Type,
		TxHash: sql.NullString{String: req.TxHash, Valid: req.TxHash != ""},
		Metadata: pqtype.NullRawMessage{
			RawMessage: metaBytes,
			Valid:      true,
		},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create pending transaction", zap.Error(err))
		return nil, response.NewAppError(500, "Failed to register pending transaction")
	}

	resp := dto.MapTransactionToResponse(tx)

	GlobalHub.BroadcastToUser(userID, constants.WSEventWalletUpdate, nil)

	return &resp, nil
}

func (s *WalletService) getFrozenBalance(ctx context.Context, userID uuid.UUID) (decimal.Decimal, error) {
	var frozen decimal.Decimal
	ads, err := s.repo.ListAdsByAdvertiser(ctx, userID)
	if err != nil {
		if err != sql.ErrNoRows {
			logger.Ctx(ctx).Error("Database error listing advertiser campaigns for frozen balance", zap.Error(err))
		}
		return decimal.Zero, err
	}
	for _, ad := range ads {
		if ad.Status != constants.AdStatusDeleted {
			val, err := decimal.NewFromString(ad.RemainingBudget)
			if err == nil {
				frozen = frozen.Add(val)
			}
		}
	}
	return frozen, nil
}
