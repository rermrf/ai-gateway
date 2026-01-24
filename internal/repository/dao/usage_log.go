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
	ClientIP     string    `gorm:"size:45;index" json:"clientIp"`
	UserAgent    string    `gorm:"size:512" json:"userAgent"`
	RequestID    string    `gorm:"size:64" json:"requestId"`
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

// LeaderboardEntry 排行榜条目 (DAO)。
type LeaderboardEntry struct {
	Value        string `json:"value"`
	RequestCount int64  `json:"requestCount"`
	InputTokens  int64  `json:"inputTokens"`
	OutputTokens int64  `json:"outputTokens"`
}

// UsageLogDAO 定义使用记录的数据访问操作。
type UsageLogDAO interface {
	Create(ctx context.Context, log *UsageLog) error
	GetStatsByUserID(ctx context.Context, userID int64) (*UsageStats, error)
	GetDailyUsageByUserID(ctx context.Context, userID int64, days int) ([]DailyUsage, error)

	GetGlobalStats(ctx context.Context) (*UsageStats, error)
	List(ctx context.Context, page, pageSize int, filters map[string]interface{}) ([]*UsageLog, int64, error)
	GetTopUsers(ctx context.Context, limit, days int) ([]LeaderboardEntry, error)
	GetTopAPIKeys(ctx context.Context, limit, days int) ([]LeaderboardEntry, error)
	GetTopClientIPs(ctx context.Context, limit, days int) ([]LeaderboardEntry, error)
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
			CAST(COALESCE(SUM(input_tokens), 0) AS UNSIGNED) as total_inputs,
			CAST(COALESCE(SUM(output_tokens), 0) AS UNSIGNED) as total_outputs,
			CAST(COALESCE(AVG(latency_ms), 0) AS UNSIGNED) as avg_latency_ms,
			CAST(SUM(CASE WHEN status_code >= 200 AND status_code < 300 THEN 1 ELSE 0 END) AS UNSIGNED) as success_count,
			CAST(SUM(CASE WHEN status_code >= 400 THEN 1 ELSE 0 END) AS UNSIGNED) as error_count
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
			CAST(COALESCE(SUM(input_tokens), 0) AS UNSIGNED) as input_tokens,
			CAST(COALESCE(SUM(output_tokens), 0) AS UNSIGNED) as output_tokens
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
			CAST(COALESCE(SUM(input_tokens), 0) AS UNSIGNED) as total_inputs,
			CAST(COALESCE(SUM(output_tokens), 0) AS UNSIGNED) as total_outputs,
			CAST(COALESCE(AVG(latency_ms), 0) AS UNSIGNED) as avg_latency_ms,
			CAST(SUM(CASE WHEN status_code >= 200 AND status_code < 300 THEN 1 ELSE 0 END) AS UNSIGNED) as success_count,
			CAST(SUM(CASE WHEN status_code >= 400 THEN 1 ELSE 0 END) AS UNSIGNED) as error_count
		`).
		Scan(&stats).Error
	return &stats, err
}

func (d *GormUsageLogDAO) List(ctx context.Context, page, pageSize int, filters map[string]interface{}) ([]*UsageLog, int64, error) {
	var logs []*UsageLog
	var total int64

	query := d.db.WithContext(ctx).Model(&UsageLog{})

	if userID, ok := filters["user_id"]; ok && userID != nil {
		query = query.Where("user_id = ?", userID)
	}
	if clientIP, ok := filters["client_ip"]; ok && clientIP != "" {
		query = query.Where("client_ip = ?", clientIP)
	}
	if apiKeyID, ok := filters["api_key_id"]; ok && apiKeyID != nil {
		query = query.Where("api_key_id = ?", apiKeyID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&logs).Error

	return logs, total, err
}

func (d *GormUsageLogDAO) GetTopUsers(ctx context.Context, limit, days int) ([]LeaderboardEntry, error) {
	var entries []LeaderboardEntry
	err := d.db.WithContext(ctx).Model(&UsageLog{}).
		Select(`
			CAST(user_id AS CHAR) as value,
			COUNT(*) as request_count,
			CAST(COALESCE(SUM(input_tokens), 0) AS UNSIGNED) as input_tokens,
			CAST(COALESCE(SUM(output_tokens), 0) AS UNSIGNED) as output_tokens
		`).
		Where("created_at >= DATE_SUB(NOW(), INTERVAL ? DAY)", days).
		Group("user_id").
		Order("request_count DESC").
		Limit(limit).
		Scan(&entries).Error
	return entries, err
}

func (d *GormUsageLogDAO) GetTopAPIKeys(ctx context.Context, limit, days int) ([]LeaderboardEntry, error) {
	var entries []LeaderboardEntry
	err := d.db.WithContext(ctx).Model(&UsageLog{}).
		Select(`
			CAST(api_key_id AS CHAR) as value,
			COUNT(*) as request_count,
			CAST(COALESCE(SUM(input_tokens), 0) AS UNSIGNED) as input_tokens,
			CAST(COALESCE(SUM(output_tokens), 0) AS UNSIGNED) as output_tokens
		`).
		Where("created_at >= DATE_SUB(NOW(), INTERVAL ? DAY) AND api_key_id IS NOT NULL", days).
		Group("api_key_id").
		Order("request_count DESC").
		Limit(limit).
		Scan(&entries).Error
	return entries, err
}

func (d *GormUsageLogDAO) GetTopClientIPs(ctx context.Context, limit, days int) ([]LeaderboardEntry, error) {
	var entries []LeaderboardEntry
	err := d.db.WithContext(ctx).Model(&UsageLog{}).
		Select(`
			client_ip as value,
			COUNT(*) as request_count,
			CAST(COALESCE(SUM(input_tokens), 0) AS UNSIGNED) as input_tokens,
			CAST(COALESCE(SUM(output_tokens), 0) AS UNSIGNED) as output_tokens
		`).
		Where("created_at >= DATE_SUB(NOW(), INTERVAL ? DAY) AND client_ip != ''", days).
		Group("client_ip").
		Order("request_count DESC").
		Limit(limit).
		Scan(&entries).Error
	return entries, err
}

var _ UsageLogDAO = (*GormUsageLogDAO)(nil)
