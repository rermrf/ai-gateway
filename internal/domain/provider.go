package domain

import "time"

// Provider 提供商领域实体。
type Provider struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"` // openai, anthropic
	APIKey    string    `json:"apiKey"`
	BaseURL   string    `json:"baseURL"`
	Models    []string  `json:"models"` // 支持的模型列表
	TimeoutMs int       `json:"timeoutMs"`
	IsDefault bool      `json:"isDefault"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
