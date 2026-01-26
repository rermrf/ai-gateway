// Package provider 提供 Provider 管理相关业务逻辑服务。
package provider

import (
	"context"

	"ai-gateway/internal/domain"
	"ai-gateway/internal/pkg/logger"
	"ai-gateway/internal/repository"
)

// Service Provider 管理服务接口。
//
//go:generate mockgen -source=./provider.go -destination=./mocks/provider.mock.go -package=providermocks Service
type Service interface {
	// List 获取所有 Provider
	List(ctx context.Context) ([]domain.Provider, error)
	// GetByID 根据ID获取 Provider
	GetByID(ctx context.Context, id int64) (*domain.Provider, error)
	// Create 创建 Provider
	Create(ctx context.Context, provider *domain.Provider) error
	// Update 更新 Provider
	Update(ctx context.Context, provider *domain.Provider) error
	// Delete 删除 Provider
	Delete(ctx context.Context, id int64) error
}

// service Provider 管理服务实现。
type service struct {
	providerRepo repository.ProviderRepository
	logger       logger.Logger
}

// NewService 创建 Provider 管理服务实例。
func NewService(
	providerRepo repository.ProviderRepository,
	l logger.Logger,
) Service {
	return &service{
		providerRepo: providerRepo,
		logger:       l.With(logger.String("service", "provider")),
	}
}

// List 获取所有 Provider。
func (s *service) List(ctx context.Context) ([]domain.Provider, error) {
	s.logger.Debug("listing all providers")
	return s.providerRepo.List(ctx)
}

// GetByID 根据ID获取 Provider。
func (s *service) GetByID(ctx context.Context, id int64) (*domain.Provider, error) {
	s.logger.Debug("getting provider by id", logger.Int64("id", id))
	return s.providerRepo.GetByID(ctx, id)
}

// Create 创建 Provider。
func (s *service) Create(ctx context.Context, provider *domain.Provider) error {
	s.logger.Info("creating provider", logger.String("name", provider.Name))
	return s.providerRepo.Create(ctx, provider)
}

// Update 更新 Provider。
func (s *service) Update(ctx context.Context, provider *domain.Provider) error {
	s.logger.Info("updating provider", logger.Int64("id", provider.ID))
	return s.providerRepo.Update(ctx, provider)
}

// Delete 删除 Provider。
func (s *service) Delete(ctx context.Context, id int64) error {
	s.logger.Info("deleting provider", logger.Int64("id", id))
	return s.providerRepo.Delete(ctx, id)
}
