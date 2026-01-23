package domain

import "time"

// ModelRate 模型费率配置
type ModelRate struct {
	ID              int64     `json:"id"`
	ModelPattern    string    `json:"modelPattern"`    // 模型匹配模式，支持通配符
	PromptPrice     float64   `json:"promptPrice"`     // 输入价格（每 1M tokens）
	CompletionPrice float64   `json:"completionPrice"` // 输出价格（每 1M tokens）
	Enabled         bool      `json:"enabled"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}
