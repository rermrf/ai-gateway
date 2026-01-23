package modelrate

import (
	"context"

	"go.uber.org/zap"

	"ai-gateway/internal/domain"
	"ai-gateway/internal/repository"
)

// Service 模型费率服务接口
//
//go:generate mockgen -source=./modelrate.go -destination=./mocks/modelrate.mock.go -package=modelratemocks -typed Service
type Service interface {
	Create(ctx context.Context, rate *domain.ModelRate) error
	Update(ctx context.Context, rate *domain.ModelRate) error
	Delete(ctx context.Context, id int64) error
	GetByID(ctx context.Context, id int64) (*domain.ModelRate, error)
	List(ctx context.Context) ([]domain.ModelRate, error)
	GetRateForModel(ctx context.Context, modelName string) (promptPrice, completionPrice float64, err error)
}

type service struct {
	repo   repository.ModelRateRepository
	logger *zap.Logger
}

func NewService(repo repository.ModelRateRepository, logger *zap.Logger) Service {
	return &service{
		repo:   repo,
		logger: logger.Named("service.modelrate"),
	}
}

func (s *service) Create(ctx context.Context, rate *domain.ModelRate) error {
	s.logger.Info("creating model rate", zap.String("pattern", rate.ModelPattern))
	return s.repo.Create(ctx, rate)
}

func (s *service) Update(ctx context.Context, rate *domain.ModelRate) error {
	s.logger.Info("updating model rate", zap.Int64("id", rate.ID))
	return s.repo.Update(ctx, rate)
}

func (s *service) Delete(ctx context.Context, id int64) error {
	s.logger.Info("deleting model rate", zap.Int64("id", id))
	return s.repo.Delete(ctx, id)
}

func (s *service) GetByID(ctx context.Context, id int64) (*domain.ModelRate, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *service) List(ctx context.Context) ([]domain.ModelRate, error) {
	return s.repo.List(ctx)
}

// GetRateForModel 获取指定模型的费率
// 匹配逻辑：完全匹配 > 前缀匹配（通配符） > 默认（0.0）
// 目前简单实现：遍历所有启用的规则，找到最长匹配
func (s *service) GetRateForModel(ctx context.Context, modelName string) (float64, float64, error) {
	rates, err := s.repo.GetAllEnabled(ctx)
	if err != nil {
		s.logger.Error("failed to get enabled rates", zap.Error(err))
		return 0, 0, nil // 出错降级为默认费率
	}

	var bestMatch *domain.ModelRate
	matchToLen := 0

	for i := range rates {
		rate := &rates[i]
		pattern := rate.ModelPattern

		// 1. 精确匹配
		if pattern == modelName {
			return rate.PromptPrice, rate.CompletionPrice, nil
		}

		// 2. 通配符匹配 (简单实现：只支持末尾 *)
		if len(pattern) > 0 && pattern[len(pattern)-1] == '*' {
			prefix := pattern[:len(pattern)-1]
			if len(modelName) >= len(prefix) && modelName[:len(prefix)] == prefix {
				if len(prefix) > matchToLen {
					matchToLen = len(prefix)
					bestMatch = rate
				}
			}
		} else {
			// 如果没有通配符，且不等于 modelName（上面判断过），就是不匹配
		}
	}

	if bestMatch != nil {
		return bestMatch.PromptPrice, bestMatch.CompletionPrice, nil
	}

	return 0, 0, nil
}
