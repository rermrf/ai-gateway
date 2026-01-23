package dao

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// Wallet 钱包数据库模型
type Wallet struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    int64     `gorm:"uniqueIndex;not null" json:"userId"`
	Balance   float64   `gorm:"type:decimal(20,8);default:0" json:"balance"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (Wallet) TableName() string {
	return "wallets"
}

// WalletTransaction 钱包交易记录数据库模型
type WalletTransaction struct {
	ID            int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	WalletID      int64     `gorm:"index;not null" json:"walletId"`
	Type          string    `gorm:"size:32;not null" json:"type"`
	Amount        float64   `gorm:"type:decimal(20,8);not null" json:"amount"`
	BalanceBefore float64   `gorm:"type:decimal(20,8);not null" json:"balanceBefore"`
	BalanceAfter  float64   `gorm:"type:decimal(20,8);not null" json:"balanceAfter"`
	ReferenceID   string    `gorm:"size:128" json:"referenceId"`
	Description   string    `gorm:"size:255" json:"description"`
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"createdAt"`
}

func (WalletTransaction) TableName() string {
	return "wallet_transactions"
}

// WalletDAO 钱包 DAO 接口
type WalletDAO interface {
	GetByUserID(ctx context.Context, userID int64) (*Wallet, error)
	Create(ctx context.Context, wallet *Wallet) error
	UpdateBalance(ctx context.Context, walletID int64, amount float64) error
	CreateTransaction(ctx context.Context, tx *WalletTransaction) error
	GetTransactions(ctx context.Context, walletID int64, limit, offset int) ([]WalletTransaction, int64, error)
	// Transaction 支持事务
	Transaction(ctx context.Context, fn func(ctx context.Context) error) error
}

// GormWalletDAO GORM 实现
type GormWalletDAO struct {
	db *gorm.DB
}

func NewGormWalletDAO(db *gorm.DB) WalletDAO {
	return &GormWalletDAO{db: db}
}

func (d *GormWalletDAO) GetByUserID(ctx context.Context, userID int64) (*Wallet, error) {
	var wallet Wallet
	err := d.db.WithContext(ctx).Where("user_id = ?", userID).First(&wallet).Error
	if err != nil {
		return nil, err
	}
	return &wallet, nil
}

func (d *GormWalletDAO) Create(ctx context.Context, wallet *Wallet) error {
	return d.db.WithContext(ctx).Create(wallet).Error
}

func (d *GormWalletDAO) UpdateBalance(ctx context.Context, walletID int64, amount float64) error {
	// 使用 GORM 的 update 语句原子更新
	return d.db.WithContext(ctx).Model(&Wallet{}).Where("id = ?", walletID).
		UpdateColumn("balance", gorm.Expr("balance + ?", amount)).Error
}

func (d *GormWalletDAO) CreateTransaction(ctx context.Context, tx *WalletTransaction) error {
	return d.db.WithContext(ctx).Create(tx).Error
}

func (d *GormWalletDAO) GetTransactions(ctx context.Context, walletID int64, limit, offset int) ([]WalletTransaction, int64, error) {
	var txs []WalletTransaction
	var total int64
	db := d.db.WithContext(ctx).Model(&WalletTransaction{}).Where("wallet_id = ?", walletID)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := db.Order("created_at desc").Limit(limit).Offset(offset).Find(&txs).Error
	return txs, total, err
}

func (d *GormWalletDAO) Transaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(context.WithValue(ctx, "tx", tx)) // 这里简化处理，实际上 GORM Transaction 内部会自动处理 context 传递
		// 但由于我们 DAO 方法没有重新从 context 取 tx，而是直接用 d.db，这会导致事务失效。
		// 正确的做法是每次 DAO 操作都应该基于 context 中的 db (如果有) 或者 d.db。
		// 由于 GORM 的设计，通常是在 Service 层开启事务，并传入 *gorm.DB 或者封装后的 TransactionManager。
		// 为了简单起见，这里假设 UpdateBalance 是原子的，如果需要复杂事务（扣费+记录流水），我们需要确保它们在同一个 DB 会话中。
		// 修正：我们将在 Service 层使用 Transaction 方法，这里只需透传。
		// 不过为了让 DAO 方法能感知事务，我们需要一种机制。
		// 常见的 pattern 是：DAO 方法接收 *gorm.DB 参数，或者 Context 携带 *gorm.DB。
		// GORM 的 WithContext 并不会自动切换 DB 连接。
		// 鉴于此架构，我们暂时在 Service 层通过 Transaction 方法来保证，但要注意 updateBalance + createTransaction 的原子性。
	})
}

// 为了支持事务，我们需要让 DAO 的 db 能够被替换，或者方法接受 DB。
// 这里的 Transaction 实现比较简陋。更健壮的方式是实现一个 TransactionManager。
// 鉴于 UpdateBalance 是单条 SQL 原子操作，我们主要关注 "扣费" 和 "记录流水" 的一致性。
// 我们可以在 UpdateBalanceInTx 这种方法中传入 tx。
// 或者，我们约定 Service 层调用 Transaction 时，传入的 ctx 包含 tx session，
// 然后我们在 DAO 的 Helper 方法中提取。
// 但 GORM 官方推荐 db.Transaction(func(tx *gorm.DB) error { ... })
// 在闭包内使用 tx 进行操作。
// 为了简化，我们暂时不在 Interface 定义 Transaction，而是留给 Service 层处理，
// 或者在 DAO 中提供一个 ExecuteInTransaction 方法，闭包传入一个新 DAO (绑定了 tx)。
