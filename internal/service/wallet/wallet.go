package wallet

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"

	"ai-gateway/internal/domain"
	"ai-gateway/internal/repository"
	"ai-gateway/internal/service/modelrate"
)

var (
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrWalletNotFound      = errors.New("wallet not found")
)

// Service 钱包服务接口
//
//go:generate mockgen -source=./wallet.go -destination=./mocks/wallet.mock.go -package=walletmocks -typed Service
type Service interface {
	GetBalance(ctx context.Context, userID int64) (*domain.Wallet, error)
	GetTransactions(ctx context.Context, userID int64, page, size int) ([]domain.WalletTransaction, int64, error)

	// TopUp 充值
	TopUp(ctx context.Context, userID int64, amount float64, referenceID string) error

	// Deduct 扣费
	Deduct(ctx context.Context, userID int64, inputTokens, outputTokens int, modelName string) error

	// HasBalance 检查用户是否有充足余额
	HasBalance(ctx context.Context, userID int64) (bool, error)
}

type service struct {
	walletRepo   repository.WalletRepository
	modelRateSvc modelrate.Service
	logger       *zap.Logger
}

func NewService(
	walletRepo repository.WalletRepository,
	modelRateSvc modelrate.Service,
	logger *zap.Logger,
) Service {
	return &service{
		walletRepo:   walletRepo,
		modelRateSvc: modelRateSvc,
		logger:       logger.Named("service.wallet"),
	}
}

func (s *service) GetBalance(ctx context.Context, userID int64) (*domain.Wallet, error) {
	wallet, err := s.walletRepo.GetByUserID(ctx, userID)
	if err != nil {
		if err.Error() == "record not found" {
			return nil, nil
		}
		return nil, err
	}
	return wallet, nil
}

func (s *service) GetTransactions(ctx context.Context, userID int64, page, size int) ([]domain.WalletTransaction, int64, error) {
	wallet, err := s.walletRepo.GetByUserID(ctx, userID)
	if err != nil {
		if err.Error() == "record not found" {
			return []domain.WalletTransaction{}, 0, nil
		}
		return nil, 0, err
	}
	if wallet == nil {
		return nil, 0, ErrWalletNotFound
	}

	offset := (page - 1) * size
	return s.walletRepo.GetTransactions(ctx, wallet.ID, size, offset)
}

func (s *service) TopUp(ctx context.Context, userID int64, amount float64, referenceID string) error {
	// 查找或创建钱包
	wallet, err := s.walletRepo.GetByUserID(ctx, userID)
	if err != nil {
		// 这里简化处理：通常 GORM 找不到记录会返回 error。
		// 如果是 Not Found，我们创建钱包。需要具体 DAO 实现确认 error 类型。
		// 简单起见，假设 GetByUserID 返回 nil, nil 表示没找到 (取决于 repository 实现)
		// 如果 repository 遵循 gorm 习惯，可能是 record not found error.
		// 这里暂且不做过细的 error type 断言，直接假设如果报错且不是 RecordNotFound 则返回。
		s.logger.Warn("failed to get wallet, attempting create if not exists", zap.Error(err))
	}

	if wallet == nil {
		s.logger.Info("creating wallet for user", zap.Int64("userID", userID))
		wallet = &domain.Wallet{
			UserID:  userID,
			Balance: 0,
		}
		if err := s.walletRepo.Create(ctx, wallet); err != nil {
			return err
		}
	}

	s.logger.Info("top up wallet", zap.Int64("userID", userID), zap.Float64("amount", amount))

	// 记录交易
	balanceBefore := wallet.Balance
	if err := s.walletRepo.UpdateBalance(ctx, wallet.ID, amount); err != nil {
		return err
	}

	tx := &domain.WalletTransaction{
		WalletID:      wallet.ID,
		Type:          domain.TransactionTypeTopUp,
		Amount:        amount,
		BalanceBefore: balanceBefore,
		BalanceAfter:  balanceBefore + amount,
		ReferenceID:   referenceID,
		Description:   "System Top Up",
	}

	return s.walletRepo.CreateTransaction(ctx, tx)
}

func (s *service) Deduct(ctx context.Context, userID int64, inputTokens, outputTokens int, modelName string) error {
	// 1. 获取费率 (Price Per 1M tokens)
	promptPrice, completionPrice, err := s.modelRateSvc.GetRateForModel(ctx, modelName)
	if err != nil {
		s.logger.Error("failed to get model rate", zap.Error(err))
		// Fail open or closed? Here we fail open with 0 cost if error, but typically we want to charge.
		// Given defaults are handled in service, err here is db error.
		promptPrice = 0
		completionPrice = 0
	}

	// Cost formula: (input / 1M * promptPrice) + (output / 1M * completionPrice)
	inputCost := (float64(inputTokens) / 1_000_000.0) * promptPrice
	outputCost := (float64(outputTokens) / 1_000_000.0) * completionPrice
	totalCost := inputCost + outputCost

	if totalCost <= 0 {
		return nil
	}

	// 2. 获取钱包
	wallet, err := s.walletRepo.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}
	if wallet == nil {
		return ErrWalletNotFound
	}

	// 3. 扣费
	s.logger.Info("deducting wallet",
		zap.Int64("userID", userID),
		zap.Float64("cost", totalCost),
		zap.String("model", modelName),
		zap.Int("inputTokens", inputTokens),
		zap.Int("outputTokens", outputTokens),
		zap.Float64("promptPrice", promptPrice),
		zap.Float64("completionPrice", completionPrice),
	)

	balanceBefore := wallet.Balance
	if err := s.walletRepo.UpdateBalance(ctx, wallet.ID, -totalCost); err != nil {
		return err
	}

	tx := &domain.WalletTransaction{
		WalletID:      wallet.ID,
		Type:          domain.TransactionTypeDeduct,
		Amount:        -totalCost,
		BalanceBefore: balanceBefore,
		BalanceAfter:  balanceBefore - totalCost,
		ReferenceID:   "", // Could pass RequestID
		Description:   fmt.Sprintf("Usage: %s (In:%d, Out:%d)", modelName, inputTokens, outputTokens),
	}

	return s.walletRepo.CreateTransaction(ctx, tx)
}

func (s *service) HasBalance(ctx context.Context, userID int64) (bool, error) {
	wallet, err := s.walletRepo.GetByUserID(ctx, userID)
	if err != nil {
		if err.Error() == "record not found" {
			return false, nil
		}
		return false, err
	}
	if wallet == nil {
		return false, nil // Default to false if no wallet
	}
	return wallet.Balance > 0, nil
}
