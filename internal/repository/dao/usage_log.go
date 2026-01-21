// Package dao 提供数据访问对象 (DAO) 接口和模型。
package dao

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// UsageLog 是使用记录的数据库模型。
type UsageLog struct {
	ID           int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID       int64     `gorm:"index;not null" json:"userId"`
	APIKeyID     *int64    `gorm:"index" json:"apiKeyId,omitempty"`
	Model        string    `gorm:"size:64" json:"model"`
	Provider     string    `gorm:"size:32" json:"provider"`
	InputTokens  int       `gorm:"default:0" json:"inputTokens"`
	OutputTokens int       `gorm:"default:0" json:"outputTokens"`
	LatencyMs    int       `gorm:"" json:"latencyMs"`
	StatusCode   int       `gorm:"" json:"statusCode"`
	CreatedAt    time.Time `gorm:"autoCreateTime;index" json:"createdAt"`
}

// TableName 返回 UsageLog 的表名。
func (UsageLog) TableName() string {
	return "usage_logs"
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

// DailyUsage 每日使用统计。
type DailyUsage struct {
	Date         string `json:"date"`
	Requests     int64  `json:"requests"`
	InputTokens  int64  `json:"inputTokens"`
	OutputTokens int64  `json:"outputTokens"`
}

// UsageLogDAO 定义使用记录的数据访问操作。
type UsageLogDAO interface {
	Create(ctx context.Context, log *UsageLog) error
	GetStatsByUserID(ctx context.Context, userID int64) (*UsageStats, error)
	GetDailyUsageByUserID(ctx context.Context, userID int64, days int) ([]DailyUsage, error)
	GetGlobalStats(ctx context.Context) (*UsageStats, error)
}

// GormUsageLogDAO 是 UsageLogDAO 的 GORM 实现。
type GormUsageLogDAO struct {
	db *gorm.DB
}

// NewGormUsageLogDAO 创建一个新的基于 GORM 的 UsageLogDAO。
func NewGormUsageLogDAO(db *gorm.DB) UsageLogDAO {
	return &GormUsageLogDAO{db: db}
}

func (d *GormUsageLogDAO) Create(ctx context.Context, log *UsageLog) error {
	return d.db.WithContext(ctx).Create(log).Error
}

func (d *GormUsageLogDAO) GetStatsByUserID(ctx context.Context, userID int64) (*UsageStats, error) {
	var stats UsageStats
	err := d.db.WithContext(ctx).Model(&UsageLog{}).
		Select(`
			COUNT(*) as total_requests,
			COALESCE(SUM(input_tokens), 0) as total_inputs,
			COALESCE(SUM(output_tokens), 0) as total_outputs,
			COALESCE(AVG(latency_ms), 0) as avg_latency_ms,
			SUM(CASE WHEN status_code >= 200 AND status_code < 300 THEN 1 ELSE 0 END) as success_count,
			SUM(CASE WHEN status_code >= 400 THEN 1 ELSE 0 END) as error_count
		`).
		Where("user_id = ?", userID).
		Scan(&stats).Error
	return &stats, err
}

func (d *GormUsageLogDAO) GetDailyUsageByUserID(ctx context.Context, userID int64, days int) ([]DailyUsage, error) {
	var usage []DailyUsage
	err := d.db.WithContext(ctx).Model(&UsageLog{}).
		Select(`
			DATE(created_at) as date,
			COUNT(*) as requests,
			COALESCE(SUM(input_tokens), 0) as input_tokens,
			COALESCE(SUM(output_tokens), 0) as output_tokens
		`).
		Where("user_id = ? AND created_at >= DATE_SUB(NOW(), INTERVAL ? DAY)", userID, days).
		Group("DATE(created_at)").
		Order("date DESC").
		Scan(&usage).Error
	return usage, err
}

func (d *GormUsageLogDAO) GetGlobalStats(ctx context.Context) (*UsageStats, error) {
	var stats UsageStats
	err := d.db.WithContext(ctx).Model(&UsageLog{}).
		Select(`
			COUNT(*) as total_requests,
			COALESCE(SUM(input_tokens), 0) as total_inputs,
			COALESCE(SUM(output_tokens), 0) as total_outputs,
			COALESCE(AVG(latency_ms), 0) as avg_latency_ms,
			SUM(CASE WHEN status_code >= 200 AND status_code < 300 THEN 1 ELSE 0 END) as success_count,
			SUM(CASE WHEN status_code >= 400 THEN 1 ELSE 0 END) as error_count
		`).
		Scan(&stats).Error
	return &stats, err
}

var _ UsageLogDAO = (*GormUsageLogDAO)(nil)
