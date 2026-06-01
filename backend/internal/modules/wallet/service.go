package wallet

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
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
	operator    *web3client.OperatorService
	platformFee float64
	tokenSymbol string
}

func NewWalletService(repo repository.Querier, db *sql.DB, operator *web3client.OperatorService, platformFee float64, tokenSymbol string) *WalletService {
	return &WalletService{repo: repo, db: db, operator: operator, platformFee: platformFee, tokenSymbol: tokenSymbol}
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

func (s *WalletService) RequestWithdraw(ctx context.Context, userID uuid.UUID, req dto.WithdrawRequest) (*dto.WithdrawResponse, error) {
	fee := decimal.NewFromFloat(s.platformFee)
	totalDeduction := req.Amount.Add(fee)

	amountBig := req.Amount.Mul(decimal.NewFromInt(10).Pow(decimal.NewFromInt(18)))
	amountWei := new(big.Int)
	amountWei, ok := amountWei.SetString(amountBig.String(), 10)
	if !ok {
		return nil, response.NewAppError(500, "Failed to parse amount")
	}

	tokenBalance, err := s.operator.TokenBalance(ctx)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to check operator token balance", zap.Error(err))
		return nil, response.NewAppError(503, "Withdrawal temporarily unavailable: cannot verify pool balance")
	}
	if tokenBalance.Cmp(amountWei) < 0 {
		return nil, response.NewAppError(503, "Withdrawal temporarily unavailable: platform pool insufficient. Please contact support.")
	}

	ethBalance, err := s.operator.EthBalance(ctx)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to check operator ETH balance", zap.Error(err))
		return nil, response.NewAppError(503, "Withdrawal temporarily unavailable: cannot verify gas balance")
	}
	minEth := new(big.Int).Mul(big.NewInt(10), new(big.Int).Exp(big.NewInt(10), big.NewInt(15), nil))
	if ethBalance.Cmp(minEth) < 0 {
		return nil, response.NewAppError(503, "Withdrawal temporarily unavailable: platform gas wallet depleted. Please contact support.")
	}

	dbTx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, response.NewAppError(500, "Failed to start withdrawal transaction")
	}
	defer func() { _ = dbTx.Rollback() }()

	q := s.repo.(*repository.Queries).WithTx(dbTx)

	wallet, err := q.GetWalletByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, response.NewAppError(404, "Wallet not found")
		}
		logger.Ctx(ctx).Error("Failed to get wallet", zap.Error(err))
		return nil, response.NewAppError(500, "Internal server error")
	}

	balanceDec := helper.ParseDecimal(wallet.Balance)
	if balanceDec.LessThan(totalDeduction) {
		return nil, response.NewAppError(400, fmt.Sprintf("Insufficient available balance: requested %s %s, platform fee %s %s, total required %s %s, but you only have %s %s available", req.Amount.StringFixed(4), s.tokenSymbol, fee.StringFixed(4), s.tokenSymbol, totalDeduction.StringFixed(4), s.tokenSymbol, balanceDec.StringFixed(4), s.tokenSymbol))
	}

	negTotalDeduction := totalDeduction.Neg()
	if _, err = q.UpdateWalletBalance(ctx, repository.UpdateWalletBalanceParams{
		UserID:  userID,
		Balance: helper.FormatDecimal(negTotalDeduction),
	}); err != nil {
		logger.Ctx(ctx).Error("Database error reserving withdrawal balance", zap.Error(err))
		return nil, response.NewAppError(500, "Failed to reserve withdrawal balance")
	}

	to := common.HexToAddress(req.WalletAddr)
	txHash, err := s.operator.SendERC20(ctx, to, amountWei)
	if err != nil {
		logger.Ctx(ctx).Error("On-chain transfer failed", zap.String("wallet", req.WalletAddr), zap.String("amount", req.Amount.String()), zap.String("tx_hash", txHash), zap.Error(err))
		return nil, response.NewAppError(502, "Withdrawal failed due to a network error. Your balance has been restored. Please contact support.")
	}

	metaBytes, _ := json.Marshal(map[string]interface{}{
		"wallet_addr": req.WalletAddr,
		"amount":      req.Amount.String(),
		"fee":         fee.String(),
	})

	negAmount := req.Amount.Neg()
	if _, err = q.CreateTransaction(ctx, repository.CreateTransactionParams{
		UserID: userID,
		Amount: helper.FormatDecimal(negAmount),
		Type:   constants.TxTypeWithdrawal,
		TxHash: sql.NullString{String: txHash, Valid: true},
		Metadata: pqtype.NullRawMessage{
			RawMessage: metaBytes,
			Valid:      true,
		},
	}); err != nil {
		logger.Ctx(ctx).Error("Database error creating withdrawal transaction", zap.Error(err))
		return nil, response.NewAppError(500, "Failed to create withdrawal record")
	}

	if fee.IsPositive() {
		negFee := fee.Neg()
		if _, err = q.CreateTransaction(ctx, repository.CreateTransactionParams{
			UserID: userID,
			Amount: helper.FormatDecimal(negFee),
			Type:   constants.TxTypeWithdrawalFee,
			TxHash: sql.NullString{},
			Metadata: pqtype.NullRawMessage{
				RawMessage: metaBytes,
				Valid:      true,
			},
		}); err != nil {
			logger.Ctx(ctx).Error("Database error creating withdrawal fee transaction", zap.Error(err))
			return nil, response.NewAppError(500, "Failed to create withdrawal fee record")
		}
	}

	if err = dbTx.Commit(); err != nil {
		logger.Ctx(ctx).Error("Database error committing withdrawal transaction", zap.Error(err))
		return nil, response.NewAppError(500, "Failed to finalize withdrawal")
	}

	GlobalHub.BroadcastToUser(userID, constants.WSEventWalletUpdate, nil)

	return &dto.WithdrawResponse{
		TxHash: txHash,
		Amount: req.Amount,
		Wallet: req.WalletAddr,
	}, nil
}

func (s *WalletService) CreatePendingTransaction(ctx context.Context, userID uuid.UUID, req dto.CreatePendingTransactionRequest) (*dto.TransactionResponse, error) {
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
