package web3

import (
	"context"
	"database/sql"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
	"github.com/sqlc-dev/pqtype"
	"go.uber.org/zap"

	"urlshortener/internal/modules/wallet"
	"urlshortener/internal/modules/web3/dto"
	"urlshortener/internal/repository"
	web3client "urlshortener/internal/web3"
	"urlshortener/pkg/constants"
	"urlshortener/pkg/logger"
	"urlshortener/pkg/response"
)

var faucetAmountBig = new(big.Int).Mul(big.NewInt(constants.FaucetAmount), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))

type DepositHandlerImpl struct {
	db   *sql.DB
	repo *repository.Queries
}

func NewDepositHandlerImpl(db *sql.DB, repo *repository.Queries) *DepositHandlerImpl {
	return &DepositHandlerImpl{db: db, repo: repo}
}

type Web3Service struct {
	repo   repository.Querier
	client *web3client.ETHClient
	faucet *web3client.FaucetService
	redis  *redis.Client
}

func NewWeb3Service(repo repository.Querier, client *web3client.ETHClient, faucet *web3client.FaucetService, redisClient *redis.Client) *Web3Service {
	return &Web3Service{
		repo:   repo,
		client: client,
		faucet: faucet,
		redis:  redisClient,
	}
}

func (s *Web3Service) ClaimFaucet(ctx context.Context, userID uuid.UUID, walletAddr string) (*dto.FaucetClaimResponse, error) {
	count, err := s.repo.CountFaucetClaimsToday(ctx, userID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to check faucet claim count", zap.String("user_id", userID.String()), zap.Error(err))
		return nil, response.NewAppError(fiber.StatusInternalServerError, "Failed to check faucet eligibility")
	}
	if count >= int64(constants.FaucetClaimLimit) {
		logger.Ctx(ctx).Warn("Faucet claim rejected: cooldown active", zap.String("user_id", userID.String()), zap.Int64("current_claims_today", count))
		return nil, response.NewAppError(fiber.StatusTooManyRequests, "faucet cooldown active, 24h between claims")
	}

	result, err := s.faucet.Claim(ctx, walletAddr, constants.FaucetAmount)
	if err != nil {
		if exErr, ok := err.(*web3client.FaucetExhaustedError); ok {
			logger.Ctx(ctx).Warn("Faucet claim failed: faucet exhausted", zap.String("wallet", walletAddr), zap.Error(err))
			return nil, response.NewAppError(fiber.StatusServiceUnavailable, exErr.Error())
		}
		logger.Ctx(ctx).Error("Faucet claim failed", zap.String("wallet", walletAddr), zap.Error(err))
		return nil, response.NewAppError(fiber.StatusInternalServerError, "Faucet claim failed")
	}

	nonce, _ := new(big.Int).SetString(result.Nonce, 10)
	deadline, _ := new(big.Int).SetString(result.Deadline, 10)

	logger.Ctx(ctx).Info("Faucet claim permit generated successfully",
		zap.String("user_id", userID.String()),
		zap.String("wallet", walletAddr),
		zap.String("amount", result.Amount),
	)

	if s.redis != nil {
		key := fmt.Sprintf("faucet:pending:%s", userID.String())
		_ = s.redis.Set(ctx, key, strings.ToLower(walletAddr), 15*time.Minute).Err()
	}

	return &dto.FaucetClaimResponse{
		Wallet:     result.Wallet,
		Amount:     result.Amount,
		Nonce:      nonce.String(),
		Deadline:   deadline.String(),
		Signature:  result.Signature,
		FaucetAddr: result.FaucetAddr,
		ChainID:    result.ChainID,
	}, nil
}

func (s *Web3Service) ConfirmFaucet(ctx context.Context, userID uuid.UUID, req *dto.FaucetConfirmRequest) (*dto.FaucetConfirmResponse, error) {
	if s.redis != nil {
		key := fmt.Sprintf("faucet:pending:%s", userID.String())
		expectedWallet, err := s.redis.Get(ctx, key).Result()
		if err != nil {
			if err == redis.Nil {
				return nil, response.NewAppError(fiber.StatusBadRequest, "No pending faucet claim request for this user")
			}
			return nil, response.NewAppError(fiber.StatusInternalServerError, "Failed to validate faucet request context")
		}
		if !strings.EqualFold(expectedWallet, req.WalletAddr) {
			return nil, response.NewAppError(fiber.StatusBadRequest, "Wallet address does not match the pending faucet request")
		}
	}

	if err := s.faucet.VerifyClaim(ctx, req.TxHash, req.WalletAddr, faucetAmountBig); err != nil {
		logger.Ctx(ctx).Error("Faucet claim verification failed",
			zap.String("tx_hash", req.TxHash),
			zap.String("wallet", req.WalletAddr),
			zap.Error(err),
		)
		return nil, response.NewAppError(fiber.StatusUnprocessableEntity, err.Error())
	}

	_, err := s.repo.CreateFaucetClaim(ctx, repository.CreateFaucetClaimParams{
		UserID: userID,
		Amount: faucetAmountBig.String(),
		TxHash: sql.NullString{String: req.TxHash, Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to record faucet claim in DB",
			zap.String("user_id", userID.String()),
			zap.String("tx_hash", req.TxHash),
			zap.Error(err),
		)
		return nil, response.NewAppError(fiber.StatusInternalServerError, "Failed to record faucet claim")
	}

	logger.Ctx(ctx).Info("Faucet claim confirmed",
		zap.String("user_id", userID.String()),
		zap.String("tx_hash", req.TxHash),
	)

	if s.redis != nil {
		key := fmt.Sprintf("faucet:pending:%s", userID.String())
		_ = s.redis.Del(ctx, key).Err()
	}

	return &dto.FaucetConfirmResponse{Status: "confirmed", TxHash: req.TxHash}, nil
}

func (s *Web3Service) ClaimDevETH(ctx context.Context, walletAddr string) (string, error) {
	txHash, err := s.faucet.SendDevETH(ctx, walletAddr)
	if err != nil {
		if exErr, ok := err.(*web3client.FaucetExhaustedError); ok {
			logger.Ctx(ctx).Warn("ClaimDevETH failed: faucet exhausted", zap.String("wallet", walletAddr), zap.Error(err))
			return "", response.NewAppError(fiber.StatusServiceUnavailable, exErr.Error())
		}
		logger.Ctx(ctx).Error("Failed to send dev ETH", zap.String("wallet", walletAddr), zap.Error(err))
		return "", response.NewAppError(fiber.StatusInternalServerError, "Failed to send dev ETH")
	}

	logger.Ctx(ctx).Info("Dev ETH claim tx sent successfully", zap.String("wallet", walletAddr), zap.String("tx_hash", txHash))
	return txHash, nil
}

func (s *Web3Service) GetFaucetHistory(ctx context.Context, userID uuid.UUID, page, limit int32, q, sortBy, sortDir string) ([]dto.FaucetHistoryItem, int64, error) {
	logger.Ctx(ctx).Info("Retrieving faucet claim history from DB", zap.String("user_id", userID.String()), zap.Int32("page", page), zap.Int32("limit", limit))

	hasFilter := q != "" || sortBy != "claimed_at" || sortDir != "desc"

	if !hasFilter {
		claims, err := s.repo.GetFaucetClaimByUser(ctx, repository.GetFaucetClaimByUserParams{
			UserID: userID,
			Limit:  limit,
			Offset: (page - 1) * limit,
		})
		if err != nil {
			logger.Ctx(ctx).Error("Failed to get faucet claim history", zap.String("user_id", userID.String()), zap.Error(err))
			return nil, 0, response.NewAppError(fiber.StatusInternalServerError, "Failed to retrieve faucet claim history")
		}

		total, err := s.repo.CountFaucetClaims(ctx, userID)
		if err != nil {
			logger.Ctx(ctx).Error("Failed to count faucet claims", zap.String("user_id", userID.String()), zap.Error(err))
			return nil, 0, response.NewAppError(fiber.StatusInternalServerError, "Failed to count faucet claims")
		}

		history := make([]dto.FaucetHistoryItem, len(claims))
		for i, c := range claims {
			history[i] = dto.FaucetHistoryItem{
				ID:        c.ID.String(),
				Amount:    c.Amount,
				TxHash:    c.TxHash.String,
				ClaimedAt: c.ClaimedAt.Format(time.RFC3339),
			}
		}

		return history, total, nil
	}

	offset := (page - 1) * limit
	filterParams := repository.GetFaucetClaimByUserFilteredParams{
		UserID:  userID,
		Q:       sql.NullString{String: q, Valid: q != ""},
		SortBy:  sortBy,
		SortDir: sortDir,
		Limit:   limit,
		Offset:  offset,
	}

	claims, err := s.repo.GetFaucetClaimByUserFiltered(ctx, filterParams)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get faucet claim history", zap.String("user_id", userID.String()), zap.Error(err))
		return nil, 0, response.NewAppError(fiber.StatusInternalServerError, "Failed to retrieve faucet claim history")
	}

	total, err := s.repo.CountFaucetClaimsByUserFiltered(ctx, repository.CountFaucetClaimsByUserFilteredParams{
		UserID: userID,
		Q:      sql.NullString{String: q, Valid: q != ""},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to count faucet claims", zap.String("user_id", userID.String()), zap.Error(err))
		return nil, 0, response.NewAppError(fiber.StatusInternalServerError, "Failed to count faucet claims")
	}

	history := make([]dto.FaucetHistoryItem, len(claims))
	for i, c := range claims {
		history[i] = dto.FaucetHistoryItem{
			ID:        c.ID.String(),
			Amount:    c.Amount,
			TxHash:    c.TxHash.String,
			ClaimedAt: c.ClaimedAt.Format(time.RFC3339),
		}
	}

	return history, total, nil
}

func (s *Web3Service) GetDepositStatus(ctx context.Context) (*dto.DepositStatusResponse, error) {
	lastBlock, err := s.client.GetLastProcessedBlock(ctx)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get last processed block from ETH client", zap.Error(err))
		return nil, response.NewAppError(fiber.StatusInternalServerError, "Failed to get deposit status")
	}

	logger.Ctx(ctx).Info("Fetched last processed block status", zap.Uint64("last_block", lastBlock))

	return &dto.DepositStatusResponse{
		Status:    constants.AdStatusActive,
		LastBlock: lastBlock,
	}, nil
}

func (h *DepositHandlerImpl) HandleDeposit(ctx context.Context, event *web3client.DepositEvent) error {
	refID := event.RefID
	if strings.HasPrefix(refID, "0x") && len(refID) >= 34 {
		refID = refID[2:34]
	}

	userID, err := uuid.Parse(refID)
	if err != nil {
		logger.Ctx(ctx).Error("Invalid refId format in deposit event",
			zap.String("refId", event.RefID),
			zap.Error(err),
		)
		return err
	}

	amountDec := decimal.NewFromBigInt(event.Amount, -18)
	amountStr := amountDec.StringFixed(8)

	logger.Ctx(ctx).Info("Processing incoming web3 deposit event",
		zap.String("user_id", userID.String()),
		zap.String("amount", amountStr),
		zap.String("tx_hash", event.TxHash),
		zap.Uint64("block", event.Block),
	)

	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		logger.Ctx(ctx).Error("Database error starting transaction in HandleDeposit", zap.Error(err))
		return err
	}
	defer func() { _ = tx.Rollback() }()

	q := h.repo.WithTx(tx)

	dbTx, err := q.GetTransactionByHash(ctx, sql.NullString{String: event.TxHash, Valid: true})
	if err == nil {
		// Transaction exists!
		if dbTx.Status == constants.TxStatusPending {
			// Update status to CONFIRMED
			_, err = q.UpdateTransactionStatus(ctx, repository.UpdateTransactionStatusParams{
				TxHash: sql.NullString{String: event.TxHash, Valid: true},
				Status: constants.TxStatusConfirmed,
			})
			if err != nil {
				logger.Ctx(ctx).Error("Failed to update pending transaction status to CONFIRMED",
					zap.String("tx_hash", event.TxHash),
					zap.Error(err),
				)
				return err
			}

			// Credit wallet balance
			if _, err = q.UpdateWalletBalance(ctx, repository.UpdateWalletBalanceParams{
				UserID:  userID,
				Balance: amountStr,
			}); err != nil {
				logger.Ctx(ctx).Error("Failed to update wallet balance after confirming transaction",
					zap.String("tx_hash", event.TxHash),
					zap.Error(err),
				)
				return err
			}

			if err := tx.Commit(); err != nil {
				logger.Ctx(ctx).Error("Failed to commit confirmed transaction",
					zap.String("tx_hash", event.TxHash),
					zap.Error(err),
				)
				return err
			}

			logger.Ctx(ctx).Info("Pending deposit confirmed successfully",
				zap.String("user_id", userID.String()),
				zap.String("amount", amountStr),
				zap.String("tx_hash", event.TxHash),
			)

			// Notify user
			wallet.GlobalHub.BroadcastToUser(userID, constants.WSEventWalletUpdate, nil)

			return nil
		}

		logger.Ctx(ctx).Warn("Deposit already processed (already CONFIRMED)",
			zap.String("tx_hash", event.TxHash),
		)
		return nil
	} else if err != sql.ErrNoRows {
		logger.Ctx(ctx).Error("Database error checking transaction existence by hash in HandleDeposit", zap.String("tx_hash", event.TxHash), zap.Error(err))
		return err
	}

	_, err = q.CreateTransaction(ctx, repository.CreateTransactionParams{
		UserID: userID,
		Amount: amountStr,
		Type:   constants.TxTypeDeposit,
		TxHash: sql.NullString{String: event.TxHash, Valid: true},
		Metadata: pqtype.NullRawMessage{
			RawMessage: []byte(fmt.Sprintf(`{"tx_hash":"%s","block":%d}`, event.TxHash, event.Block)),
			Valid:      true,
		},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Database error creating deposit transaction records in HandleDeposit", zap.String("tx_hash", event.TxHash), zap.Error(err))
		return err
	}

	if _, err = q.UpdateWalletBalance(ctx, repository.UpdateWalletBalanceParams{
		UserID:  userID,
		Balance: amountStr,
	}); err != nil {
		logger.Ctx(ctx).Error("Failed to update wallet balance after recording transaction",
			zap.String("tx_hash", event.TxHash),
			zap.Error(err),
		)
		return err
	}

	if err := tx.Commit(); err != nil {
		logger.Ctx(ctx).Error("Failed to commit deposit transaction",
			zap.String("tx_hash", event.TxHash),
			zap.Error(err),
		)
		return err
	}

	logger.Ctx(ctx).Info("Deposit processed successfully (first-time create)",
		zap.String("user_id", userID.String()),
		zap.String("amount", amountStr),
		zap.String("tx_hash", event.TxHash),
	)

	wallet.GlobalHub.BroadcastToUser(userID, constants.WSEventWalletUpdate, nil)

	return nil
}
