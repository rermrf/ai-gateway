// Package usage 提供使用统计相关业务逻辑服务。
package usage

import (
	"context"

	"ai-gateway/internal/domain"
	"ai-gateway/internal/pkg/logger"
	"ai-gateway/internal/repository"
	"ai-gateway/internal/service/wallet"
)

// Service 使用统计服务接口。
//
//go:generate mockgen -source=./usage.go -destination=./mocks/usage.mock.go -package=usagemocks -typed Service
type Service interface {
	// GetGlobalStats 获取全局使用统计（管理员）
	GetGlobalStats(ctx context.Context) (*domain.UsageStats, error)
	// GetDailyUsage 获取全局每日使用统计（管理员）
	GetGlobalDailyUsage(ctx context.Context, days int) ([]domain.DailyUsage, error)
	// LogRequest 记录请求使用情况
	LogRequest(ctx context.Context, log *domain.UsageLog) error
	// ListLogs 获取日志列表
	ListLogs(ctx context.Context, page, pageSize int, filters map[string]interface{}) ([]*domain.UsageLog, int64, error)
	// GetLeaderboard 获取使用量排行榜
	GetLeaderboard(ctx context.Context, dimension string, limit, days int) ([]domain.UsageLeaderboardEntry, error)
}

// service 使用统计服务实现。
type service struct {
	usageLogRepo repository.UsageLogRepository
	walletSvc    wallet.Service
	logger       logger.Logger
}

// NewService 创建使用统计服务实例。
func NewService(
	usageLogRepo repository.UsageLogRepository,
	walletSvc wallet.Service,
	l logger.Logger,
) Service {
	return &service{
		usageLogRepo: usageLogRepo,
		walletSvc:    walletSvc,
		logger:       l.With(logger.String("service", "usage")),
	}
}

// GetGlobalStats 获取全局使用统计。
func (s *service) GetGlobalStats(ctx context.Context) (*domain.UsageStats, error) {
	s.logger.Debug("getting global usage stats")

	stats, err := s.usageLogRepo.GetGlobalStats(ctx)
	if err != nil {
		s.logger.Error("failed to get global stats", logger.Error(err))
		return nil, err
	}

	return stats, nil
}

// GetGlobalDailyUsage 获取全局每日使用统计。
func (s *service) GetGlobalDailyUsage(ctx context.Context, days int) ([]domain.DailyUsage, error) {
	s.logger.Debug("getting global daily usage", logger.Int("days", days))

	// TODO: 需要在 repository 层添加 GetGlobalDailyUsage 方法
	// 目前可以返回空切片或实现一个聚合逻辑
	return []domain.DailyUsage{}, nil
}

// LogRequest 记录请求使用情况。
func (s *service) LogRequest(ctx context.Context, log *domain.UsageLog) error {
	// 异步记录，避免阻塞
	// 注意：这里我们使用传入的 ctx，但在调用方应该传入一个独立的 context 或者
	// 我们在这里创建一个新的 context，但这取决于调用方是否已经处理了。
	// 简单起见，我们直接调用 repo，假设调用方已经处理了异步逻辑或者就在同步链路中记录
	// 为了不影响主请求延迟，建议调用方 go func() 调用，或者最好在这里 go func
	// 但在这里 go func 需要 context 不被 cancel。

	// 1. 尝试扣费 (同步执行，确保有余额)
	// 注意：如果 LogRequest 是异步调用的，这里也会异步执行。
	// 1. 尝试扣费 (同步执行，确保有余额)
	// 注意：如果 LogRequest 是异步调用的，这里也会异步执行。

	totalTokens := log.InputTokens + log.OutputTokens
	if log.UserID > 0 && totalTokens > 0 {
		if err := s.walletSvc.Deduct(ctx, log.UserID, log.InputTokens, log.OutputTokens, log.Model); err != nil {
			// 扣费失败仅记录日志，暂不阻断（取决于策略，如果是预付费严格校验，应该在网关入口检查余额）
			// 但这里是 LogRequest，请求已经完成了。
			s.logger.Error("failed to deduct wallet balance",
				logger.Int64("userID", log.UserID),
				logger.String("model", log.Model),
				logger.Int("tokens", totalTokens),
				logger.Error(err),
			)
		}
	}

	// 暂时同步写入，在这个阶段确保数据一致性优先
	if err := s.usageLogRepo.Create(ctx, log); err != nil {
		s.logger.Error("failed to log request usage", logger.Error(err))
		return err
	}
	return nil
}

// ListLogs 获取日志列表。
func (s *service) ListLogs(ctx context.Context, page, pageSize int, filters map[string]interface{}) ([]*domain.UsageLog, int64, error) {
	return s.usageLogRepo.List(ctx, page, pageSize, filters)
}

// GetLeaderboard 获取使用量排行榜。
func (s *service) GetLeaderboard(ctx context.Context, dimension string, limit, days int) ([]domain.UsageLeaderboardEntry, error) {
	switch dimension {
	case "user_id":
		return s.usageLogRepo.GetTopUsers(ctx, limit, days)
	case "api_key_id":
		return s.usageLogRepo.GetTopAPIKeys(ctx, limit, days)
	case "client_ip":
		return s.usageLogRepo.GetTopClientIPs(ctx, limit, days)
	default:
		return nil, domain.ErrInvalidParameter
	}
}
