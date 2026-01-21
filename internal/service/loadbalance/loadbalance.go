// Package loadbalance 提供负载均衡管理相关业务逻辑服务。
package loadbalance

import (
	"context"

	"go.uber.org/zap"

	"ai-gateway/internal/domain"
	"ai-gateway/internal/repository"
)

// Service 负载均衡管理服务接口。
//
//go:generate mockgen -source=./loadbalance.go -destination=./mocks/loadbalance.mock.go -package=loadbalancemocks -typed Service
type Service interface {
	// ListGroups 获取所有负载均衡组
	ListGroups(ctx context.Context) ([]domain.LoadBalanceGroup, error)
	// GetGroupByID 获取指定ID的负载均衡组
	GetGroupByID(ctx context.Context, id int64) (*domain.LoadBalanceGroup, error)
	// CreateGroup 创建负载均衡组
	CreateGroup(ctx context.Context, group *domain.LoadBalanceGroup) error
	// UpdateGroup 更新负载均衡组
	UpdateGroup(ctx context.Context, group *domain.LoadBalanceGroup) error
	// DeleteGroup 删除负载均衡组
	DeleteGroup(ctx context.Context, id int64) error
}

// service 负载均衡管理服务实现。
type service struct {
	loadBalanceRepo repository.LoadBalanceRepository
	logger          *zap.Logger
}

// NewService 创建负载均衡管理服务实例。
func NewService(
	loadBalanceRepo repository.LoadBalanceRepository,
	logger *zap.Logger,
) Service {
	return &service{
		loadBalanceRepo: loadBalanceRepo,
		logger:          logger.Named("service.loadbalance"),
	}
}

// ListGroups 获取所有负载均衡组。
func (s *service) ListGroups(ctx context.Context) ([]domain.LoadBalanceGroup, error) {
	s.logger.Debug("listing all load balance groups")
	return s.loadBalanceRepo.ListGroups(ctx)
}

// GetGroupByID 获取指定ID的负载均衡组。
func (s *service) GetGroupByID(ctx context.Context, id int64) (*domain.LoadBalanceGroup, error) {
	s.logger.Debug("getting load balance group by id", zap.Int64("id", id))
	return s.loadBalanceRepo.GetGroupByID(ctx, id)
}

// CreateGroup 创建负载均衡组。
func (s *service) CreateGroup(ctx context.Context, group *domain.LoadBalanceGroup) error {
	s.logger.Info("creating load balance group", zap.String("name", group.Name))
	return s.loadBalanceRepo.CreateGroup(ctx, group)
}

// UpdateGroup 更新负载均衡组。
func (s *service) UpdateGroup(ctx context.Context, group *domain.LoadBalanceGroup) error {
	s.logger.Info("updating load balance group", zap.Int64("id", group.ID), zap.String("name", group.Name))
	return s.loadBalanceRepo.UpdateGroup(ctx, group)
}

// DeleteGroup 删除负载均衡组。
func (s *service) DeleteGroup(ctx context.Context, id int64) error {
	s.logger.Info("deleting load balance group", zap.Int64("id", id))
	return s.loadBalanceRepo.DeleteGroup(ctx, id)
}
