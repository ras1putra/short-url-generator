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
	repo       repository.Querier
	db         *sql.DB
	withdrawer *web3client.WithdrawerService
}

func NewWalletService(repo repository.Querier, db *sql.DB, withdrawer *web3client.WithdrawerService) *WalletService {
	return &WalletService{repo: repo, db: db, withdrawer: withdrawer}
}

func (s *WalletService) GetWallet(ctx context.Context, userID uuid.UUID) (*dto.WalletWithTransactionsResponse, error) {
	wallet, err := s.repo.GetWalletByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, response.NewAppError(404, "Wallet not found")
		}
		logger.Ctx(ctx).Error("Failed to get wallet", zap.Error(err))
		return nil, response.NewAppError(500, "Internal server error")
	}

	txs, err := s.repo.ListTransactionsByUser(ctx, repository.ListTransactionsByUserParams{
		UserID: userID,
		Limit:  int32(constants.WalletDefaultLimit),
		Offset: 0,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list transactions", zap.Error(err))
		return nil, response.NewAppError(500, "Failed to load transactions")
	}

	// Calculate frozen (allocated) budget from all active/paused campaigns
	var frozen decimal.Decimal
	ads, err := s.repo.ListAdsByAdvertiser(ctx, userID)
	if err == nil {
		for _, ad := range ads {
			if ad.Status != constants.AdStatusDeleted {
				val, err := decimal.NewFromString(ad.RemainingBudget)
				if err == nil {
					frozen = frozen.Add(val)
				}
			}
		}
	} else {
		logger.Ctx(ctx).Error("Failed to list ads for budget freeze calculation", zap.Error(err))
	}

	wr := dto.MapWalletToResponse(wallet)
	available := wr.Balance.Sub(frozen)
	if available.IsNegative() {
		available = decimal.Zero
	}

	return &dto.WalletWithTransactionsResponse{
		Balance:      wr.Balance,
		Available:    available,
		Frozen:       frozen,
		UpdatedAt:    wr.UpdatedAt,
		Transactions: dto.MapTransactionsToResponse(txs),
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

	balance := helper.ParseDecimal(wallet.Balance)
	if balance.LessThan(req.Amount) {
		return nil, response.NewAppError(400, "Insufficient balance")
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

	nonceBig, _ := new(big.Int).SetString(permit.Nonce, 10)
	metaBytes, _ := json.Marshal(map[string]interface{}{
		"nonce":    nonceBig.String(),
		"deadline": permit.Deadline,
		"contract": permit.Contract,
	})

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin tx: %w", err)
	}
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		} else if err != nil {
			_ = tx.Rollback()
		}
	}()

	q := s.repo
	if queriesInstance, ok := s.repo.(*repository.Queries); ok {
		q = queriesInstance.WithTx(tx)
	} else {
		_ = tx.Rollback()
		return nil, fmt.Errorf("repository not transaction-compatible")
	}

	negAmount := req.Amount.Neg()

	_, err = q.UpdateWalletBalance(ctx, repository.UpdateWalletBalanceParams{
		UserID:  userID,
		Balance: helper.FormatDecimal(negAmount),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update wallet balance: %w", err)
	}

	_, err = q.CreateTransaction(ctx, repository.CreateTransactionParams{
		UserID: userID,
		Amount: helper.FormatDecimal(negAmount),
		Type:   constants.TxTypeWithdrawal,
		TxHash: sql.NullString{},
		Metadata: pqtype.NullRawMessage{
			RawMessage: metaBytes,
			Valid:      true,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create withdrawal tx: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit withdrawal tx: %w", err)
	}

	return &dto.WithdrawalPermitResponse{
		Wallet:    permit.Wallet,
		Amount:    permit.Amount,
		Nonce:     permit.Nonce,
		Deadline:  permit.Deadline,
		Signature: permit.Signature,
		Contract:  permit.Contract,
		ChainID:   permit.ChainID,
	}, nil
}
