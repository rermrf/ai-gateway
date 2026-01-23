package domain

import "time"

// Wallet 用户钱包
type Wallet struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"userId"`
	Balance   float64   `json:"balance"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// TransactionType 交易类型
type TransactionType string

const (
	TransactionTypeTopUp  TransactionType = "top_up" // 充值
	TransactionTypeDeduct TransactionType = "deduct" // 扣费
	TransactionTypeRefund TransactionType = "refund" // 退款
)

// WalletTransaction 钱包交易记录
type WalletTransaction struct {
	ID            int64           `json:"id"`
	WalletID      int64           `json:"walletId"`
	Type          TransactionType `json:"type"`
	Amount        float64         `json:"amount"`
	BalanceBefore float64         `json:"balanceBefore"`
	BalanceAfter  float64         `json:"balanceAfter"`
	ReferenceID   string          `json:"referenceId"` // 关联ID，如请求ID或管理员操作ID
	Description   string          `json:"description"`
	CreatedAt     time.Time       `json:"createdAt"`
}
