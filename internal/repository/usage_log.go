// Package repository 定义数据访问的存储库接口。
package repository

import (
	"context"

	"ai-gateway/internal/domain"
	"ai-gateway/internal/repository/dao"
)

// UsageLogRepository 定义使用记录的存储库接口。
type UsageLogRepository interface {
	Create(ctx context.Context, log *domain.UsageLog) error
	GetStatsByUserID(ctx context.Context, userID int64) (*domain.UsageStats, error)
	GetDailyUsageByUserID(ctx context.Context, userID int64, days int) ([]domain.DailyUsage, error)
	GetGlobalStats(ctx context.Context) (*domain.UsageStats, error)
}

// usageLogRepository 是 UsageLogRepository 的默认实现。
type usageLogRepository struct {
	dao dao.UsageLogDAO
}

// NewUsageLogRepository 创建一个新的 UsageLogRepository。
func NewUsageLogRepository(usageLogDAO dao.UsageLogDAO) UsageLogRepository {
	return &usageLogRepository{dao: usageLogDAO}
}

// toDAO 将 domain.UsageLog 转换为 dao.UsageLog。
func (r *usageLogRepository) toDAO(log *domain.UsageLog) *dao.UsageLog {
	return &dao.UsageLog{
		ID:           log.ID,
		UserID:       log.UserID,
		APIKeyID:     log.APIKeyID,
		Model:        log.Model,
		Provider:     log.Provider,
		InputTokens:  log.InputTokens,
		OutputTokens: log.OutputTokens,
		LatencyMs:    log.LatencyMs,
		StatusCode:   log.StatusCode,
		CreatedAt:    log.CreatedAt,
	}
}

func (r *usageLogRepository) Create(ctx context.Context, log *domain.UsageLog) error {
	daoLog := r.toDAO(log)
	if err := r.dao.Create(ctx, daoLog); err != nil {
		return err
	}
	log.ID = daoLog.ID
	log.CreatedAt = daoLog.CreatedAt
	return nil
}

func (r *usageLogRepository) GetStatsByUserID(ctx context.Context, userID int64) (*domain.UsageStats, error) {
	daoStats, err := r.dao.GetStatsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &domain.UsageStats{
		TotalRequests: daoStats.TotalRequests,
		TotalInputs:   daoStats.TotalInputs,
		TotalOutputs:  daoStats.TotalOutputs,
		AvgLatencyMs:  daoStats.AvgLatencyMs,
		SuccessCount:  daoStats.SuccessCount,
		ErrorCount:    daoStats.ErrorCount,
	}, nil
}

func (r *usageLogRepository) GetDailyUsageByUserID(ctx context.Context, userID int64, days int) ([]domain.DailyUsage, error) {
	daoUsage, err := r.dao.GetDailyUsageByUserID(ctx, userID, days)
	if err != nil {
		return nil, err
	}
	usage := make([]domain.DailyUsage, len(daoUsage))
	for i, u := range daoUsage {
		usage[i] = domain.DailyUsage{
			Date:         u.Date,
			Requests:     u.Requests,
			InputTokens:  u.InputTokens,
			OutputTokens: u.OutputTokens,
		}
	}
	return usage, nil
}

func (r *usageLogRepository) GetGlobalStats(ctx context.Context) (*domain.UsageStats, error) {
	daoStats, err := r.dao.GetGlobalStats(ctx)
	if err != nil {
		return nil, err
	}
	return &domain.UsageStats{
		TotalRequests: daoStats.TotalRequests,
		TotalInputs:   daoStats.TotalInputs,
		TotalOutputs:  daoStats.TotalOutputs,
		AvgLatencyMs:  daoStats.AvgLatencyMs,
		SuccessCount:  daoStats.SuccessCount,
		ErrorCount:    daoStats.ErrorCount,
	}, nil
}
