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
	List(ctx context.Context, page, pageSize int, filters map[string]interface{}) ([]*domain.UsageLog, int64, error)
	GetTopUsers(ctx context.Context, limit, days int) ([]domain.UsageLeaderboardEntry, error)
	GetTopAPIKeys(ctx context.Context, limit, days int) ([]domain.UsageLeaderboardEntry, error)
	GetTopClientIPs(ctx context.Context, limit, days int) ([]domain.UsageLeaderboardEntry, error)
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
		ClientIP:     log.ClientIP,
		UserAgent:    log.UserAgent,
		RequestID:    log.RequestID,
		CreatedAt:    log.CreatedAt,
	}
}

// toDomain 将 dao.UsageLog 转换为 domain.UsageLog。
func (r *usageLogRepository) toDomain(log *dao.UsageLog) *domain.UsageLog {
	return &domain.UsageLog{
		ID:           log.ID,
		UserID:       log.UserID,
		APIKeyID:     log.APIKeyID,
		Model:        log.Model,
		Provider:     log.Provider,
		InputTokens:  log.InputTokens,
		OutputTokens: log.OutputTokens,
		LatencyMs:    log.LatencyMs,
		StatusCode:   log.StatusCode,
		ClientIP:     log.ClientIP,
		UserAgent:    log.UserAgent,
		RequestID:    log.RequestID,
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

func (r *usageLogRepository) List(ctx context.Context, page, pageSize int, filters map[string]interface{}) ([]*domain.UsageLog, int64, error) {
	daoLogs, total, err := r.dao.List(ctx, page, pageSize, filters)
	if err != nil {
		return nil, 0, err
	}

	logs := make([]*domain.UsageLog, len(daoLogs))
	for i, l := range daoLogs {
		logs[i] = r.toDomain(l)
	}

	return logs, total, nil
}

func (r *usageLogRepository) GetTopUsers(ctx context.Context, limit, days int) ([]domain.UsageLeaderboardEntry, error) {
	daoEntries, err := r.dao.GetTopUsers(ctx, limit, days)
	if err != nil {
		return nil, err
	}
	return r.toDomainLeaderboard(daoEntries, "user_id"), nil
}

func (r *usageLogRepository) GetTopAPIKeys(ctx context.Context, limit, days int) ([]domain.UsageLeaderboardEntry, error) {
	daoEntries, err := r.dao.GetTopAPIKeys(ctx, limit, days)
	if err != nil {
		return nil, err
	}
	return r.toDomainLeaderboard(daoEntries, "api_key_id"), nil
}

func (r *usageLogRepository) GetTopClientIPs(ctx context.Context, limit, days int) ([]domain.UsageLeaderboardEntry, error) {
	daoEntries, err := r.dao.GetTopClientIPs(ctx, limit, days)
	if err != nil {
		return nil, err
	}
	return r.toDomainLeaderboard(daoEntries, "client_ip"), nil
}

func (r *usageLogRepository) toDomainLeaderboard(daoEntries []dao.LeaderboardEntry, dimension string) []domain.UsageLeaderboardEntry {
	entries := make([]domain.UsageLeaderboardEntry, len(daoEntries))
	for i, e := range daoEntries {
		entries[i] = domain.UsageLeaderboardEntry{
			Dimension:    dimension,
			Value:        e.Value,
			RequestCount: e.RequestCount,
			InputTokens:  e.InputTokens,
			OutputTokens: e.OutputTokens,
			TotalTokens:  e.InputTokens + e.OutputTokens,
		}
	}
	return entries
}
