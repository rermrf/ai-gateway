// Package routingrule 提供路由规则管理相关业务逻辑服务。
package routingrule

import (
	"context"

	"ai-gateway/internal/domain"
	"ai-gateway/internal/pkg/logger"
	"ai-gateway/internal/repository"
)

// Service 路由规则管理服务接口。
//
//go:generate mockgen -source=./routingrule.go -destination=./mocks/routingrule.mock.go -package=routingrulemocks Service
type Service interface {
	// List 获取所有路由规则
	List(ctx context.Context) ([]domain.RoutingRule, error)
	// Create 创建路由规则
	Create(ctx context.Context, rule *domain.RoutingRule) error
	// Update 更新路由规则
	Update(ctx context.Context, rule *domain.RoutingRule) error
	// Delete 删除路由规则
	Delete(ctx context.Context, id int64) error
}

// service 路由规则管理服务实现。
type service struct {
	routingRuleRepo repository.RoutingRuleRepository
	logger          logger.Logger
}

// NewService 创建路由规则管理服务实例。
func NewService(
	routingRuleRepo repository.RoutingRuleRepository,
	l logger.Logger,
) Service {
	return &service{
		routingRuleRepo: routingRuleRepo,
		logger:          l.With(logger.String("service", "routingrule")),
	}
}

// List 获取所有路由规则。
func (s *service) List(ctx context.Context) ([]domain.RoutingRule, error) {
	s.logger.Debug("listing all routing rules")
	return s.routingRuleRepo.List(ctx)
}

// Create 创建路由规则。
func (s *service) Create(ctx context.Context, rule *domain.RoutingRule) error {
	s.logger.Info("creating routing rule", logger.String("pattern", rule.Pattern))
	return s.routingRuleRepo.Create(ctx, rule)
}

// Update 更新路由规则。
func (s *service) Update(ctx context.Context, rule *domain.RoutingRule) error {
	s.logger.Info("updating routing rule", logger.Int64("id", rule.ID))
	return s.routingRuleRepo.Update(ctx, rule)
}

// Delete 删除路由规则。
func (s *service) Delete(ctx context.Context, id int64) error {
	s.logger.Info("deleting routing rule", logger.Int64("id", id))
	return s.routingRuleRepo.Delete(ctx, id)
}
