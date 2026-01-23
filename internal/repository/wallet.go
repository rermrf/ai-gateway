package repository

import (
	"context"

	"ai-gateway/internal/domain"
	"ai-gateway/internal/repository/dao"

	"gorm.io/gorm"
)

// WalletRepository 钱包仓储接口
type WalletRepository interface {
	GetByUserID(ctx context.Context, userID int64) (*domain.Wallet, error)
	Create(ctx context.Context, wallet *domain.Wallet) error
	UpdateBalance(ctx context.Context, walletID int64, amount float64) error
	CreateTransaction(ctx context.Context, tx *domain.WalletTransaction) error
	GetTransactions(ctx context.Context, walletID int64, limit, offset int) ([]domain.WalletTransaction, int64, error)
	// WithTx 支持事务
	WithTx(tx *gorm.DB) WalletRepository
}

type walletRepository struct {
	dao dao.WalletDAO
}

func NewWalletRepository(dao dao.WalletDAO) WalletRepository {
	return &walletRepository{dao: dao}
}

// WithTx 返回一个新的 Repository 实例，使用传入的事务 DB
// 这是一个简化的事务支持方式，依赖于 DAO 支持或是我们直接在这里 Hack
// 由于 DAO 接口没有暴露 WithDB，我们需要转换类型或者在 DAO 增加方法
// 这里为了简单，我们假设 DAO 能够处理，或者我们暂时不实现 WithTx
// 更好的方式是：Repository 方法接收 context，Service 层负责在 context 中注入 tx
// 目前先留空 WithTx，后面在 Service 层通过 TransactionManager 统一处理
func (r *walletRepository) WithTx(tx *gorm.DB) WalletRepository {
	// TODO: 实现基于 TX 的 DAO 切换
	// 实际项目中，通常会有一个 NewWalletDAO(tx)
	return r
}

func (r *walletRepository) toDomainWallet(w *dao.Wallet) *domain.Wallet {
	if w == nil {
		return nil
	}
	return &domain.Wallet{
		ID:        w.ID,
		UserID:    w.UserID,
		Balance:   w.Balance,
		CreatedAt: w.CreatedAt,
		UpdatedAt: w.UpdatedAt,
	}
}

func (r *walletRepository) toDAOWallet(w *domain.Wallet) *dao.Wallet {
	if w == nil {
		return nil
	}
	return &dao.Wallet{
		ID:        w.ID,
		UserID:    w.UserID,
		Balance:   w.Balance,
		CreatedAt: w.CreatedAt,
		UpdatedAt: w.UpdatedAt,
	}
}

func (r *walletRepository) toDomainTransaction(tx *dao.WalletTransaction) *domain.WalletTransaction {
	if tx == nil {
		return nil
	}
	return &domain.WalletTransaction{
		ID:            tx.ID,
		WalletID:      tx.WalletID,
		Type:          domain.TransactionType(tx.Type),
		Amount:        tx.Amount,
		BalanceBefore: tx.BalanceBefore,
		BalanceAfter:  tx.BalanceAfter,
		ReferenceID:   tx.ReferenceID,
		Description:   tx.Description,
		CreatedAt:     tx.CreatedAt,
	}
}

func (r *walletRepository) toDAOTransaction(tx *domain.WalletTransaction) *dao.WalletTransaction {
	if tx == nil {
		return nil
	}
	return &dao.WalletTransaction{
		ID:            tx.ID,
		WalletID:      tx.WalletID,
		Type:          string(tx.Type),
		Amount:        tx.Amount,
		BalanceBefore: tx.BalanceBefore,
		BalanceAfter:  tx.BalanceAfter,
		ReferenceID:   tx.ReferenceID,
		Description:   tx.Description,
		CreatedAt:     tx.CreatedAt,
	}
}

func (r *walletRepository) GetByUserID(ctx context.Context, userID int64) (*domain.Wallet, error) {
	w, err := r.dao.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return r.toDomainWallet(w), nil
}

func (r *walletRepository) Create(ctx context.Context, wallet *domain.Wallet) error {
	daoWallet := r.toDAOWallet(wallet)
	if err := r.dao.Create(ctx, daoWallet); err != nil {
		return err
	}
	wallet.ID = daoWallet.ID
	wallet.CreatedAt = daoWallet.CreatedAt
	wallet.UpdatedAt = daoWallet.UpdatedAt
	return nil
}

func (r *walletRepository) UpdateBalance(ctx context.Context, walletID int64, amount float64) error {
	return r.dao.UpdateBalance(ctx, walletID, amount)
}

func (r *walletRepository) CreateTransaction(ctx context.Context, tx *domain.WalletTransaction) error {
	daoTx := r.toDAOTransaction(tx)
	if err := r.dao.CreateTransaction(ctx, daoTx); err != nil {
		return err
	}
	tx.ID = daoTx.ID
	tx.CreatedAt = daoTx.CreatedAt
	return nil
}

func (r *walletRepository) GetTransactions(ctx context.Context, walletID int64, limit, offset int) ([]domain.WalletTransaction, int64, error) {
	daoTxs, total, err := r.dao.GetTransactions(ctx, walletID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	txs := make([]domain.WalletTransaction, len(daoTxs))
	for i, tx := range daoTxs {
		txs[i] = *r.toDomainTransaction(&tx)
	}
	return txs, total, nil
}
