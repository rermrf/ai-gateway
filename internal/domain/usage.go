// Package domain 定义领域模型和业务实体。
package domain

import (
	"time"
)

// UsageLog 使用记录领域实体。
type UsageLog struct {
	ID           int64     `json:"id"`
	UserID       int64     `json:"userId"`
	APIKeyID     *int64    `json:"apiKeyId,omitempty"`
	Model        string    `json:"model"`
	Provider     string    `json:"provider"`
	InputTokens  int       `json:"inputTokens"`
	OutputTokens int       `json:"outputTokens"`
	LatencyMs    int       `json:"latencyMs"`
	StatusCode   int       `json:"statusCode"`
	CreatedAt    time.Time `json:"createdAt"`
}

// TotalTokens 返回总 Token 数。
func (u *UsageLog) TotalTokens() int {
	return u.InputTokens + u.OutputTokens
}

// IsSuccess 判断请求是否成功。
func (u *UsageLog) IsSuccess() bool {
	return u.StatusCode >= 200 && u.StatusCode < 300
}

// UsageStats 使用统计汇总。
type UsageStats struct {
	TotalRequests int64 `json:"totalRequests"`
	TotalInputs   int64 `json:"totalInputTokens"`
	TotalOutputs  int64 `json:"totalOutputTokens"`
	AvgLatencyMs  int64 `json:"avgLatencyMs"`
	SuccessCount  int64 `json:"successCount"`
	ErrorCount    int64 `json:"errorCount"`
}

// TotalTokens 返回总 Token 数。
func (s *UsageStats) TotalTokens() int64 {
	return s.TotalInputs + s.TotalOutputs
}

// SuccessRate 返回成功率。
func (s *UsageStats) SuccessRate() float64 {
	if s.TotalRequests == 0 {
		return 0
	}
	return float64(s.SuccessCount) / float64(s.TotalRequests) * 100
}

// DailyUsage 每日使用统计。
type DailyUsage struct {
	Date         string `json:"date"`
	Requests     int64  `json:"requests"`
	InputTokens  int64  `json:"inputTokens"`
	OutputTokens int64  `json:"outputTokens"`
}
