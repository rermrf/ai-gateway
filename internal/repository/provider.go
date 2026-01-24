// Package repository 定义数据访问的存储库接口。
// 存储库层聚合了 DAO 和缓存，暴露统一的数据访问接口。
package repository

import (
	"context"
	"encoding/json"

	"ai-gateway/internal/domain"
	"ai-gateway/internal/repository/cache"
	"ai-gateway/internal/repository/dao"
)

// ProviderRepository 定义 Provider 实体的存储库接口。
type ProviderRepository interface {
	Create(ctx context.Context, provider *domain.Provider) error
	Update(ctx context.Context, provider *domain.Provider) error
	Delete(ctx context.Context, id int64) error
	GetByID(ctx context.Context, id int64) (*domain.Provider, error)
	GetByName(ctx context.Context, name string) (*domain.Provider, error)
	List(ctx context.Context) ([]domain.Provider, error)
	GetDefaultByType(ctx context.Context, providerType string) (*domain.Provider, error)
}

// providerRepository 是 ProviderRepository 的默认实现。
type providerRepository struct {
	dao   dao.ProviderDAO
	cache cache.ProviderCache
}

// NewProviderRepository 创建一个新的 ProviderRepository。
func NewProviderRepository(providerDAO dao.ProviderDAO, cache cache.ProviderCache) ProviderRepository {
	return &providerRepository{
		dao:   providerDAO,
		cache: cache,
	}
}

// toDAO 将 domain.Provider 转换为 dao.Provider
func (r *providerRepository) toDAO(p *domain.Provider) *dao.Provider {
	modelsJSON, _ := json.Marshal(p.Models)
	return &dao.Provider{
		ID:        p.ID,
		Name:      p.Name,
		Type:      p.Type,
		APIKey:    p.APIKey,
		BaseURL:   p.BaseURL,
		Models:    string(modelsJSON),
		TimeoutMs: p.TimeoutMs,
		IsDefault: p.IsDefault,
		Enabled:   p.Enabled,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}

// toDomain 将 dao.Provider 转换为 domain.Provider
func (r *providerRepository) toDomain(p *dao.Provider) *domain.Provider {
	if p == nil {
		return nil
	}
	var models []string
	if p.Models != "" {
		_ = json.Unmarshal([]byte(p.Models), &models)
	}
	return &domain.Provider{
		ID:        p.ID,
		Name:      p.Name,
		Type:      p.Type,
		APIKey:    p.APIKey,
		BaseURL:   p.BaseURL,
		Models:    models,
		TimeoutMs: p.TimeoutMs,
		IsDefault: p.IsDefault,
		Enabled:   p.Enabled,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}

func (r *providerRepository) Create(ctx context.Context, p *domain.Provider) error {
	daoProvider := r.toDAO(p)
	if err := r.dao.Create(ctx, daoProvider); err != nil {
		return err
	}
	p.ID = daoProvider.ID
	p.CreatedAt = daoProvider.CreatedAt
	p.UpdatedAt = daoProvider.UpdatedAt

	if r.cache != nil {
		_ = r.cache.Invalidate(ctx)
	}
	return nil
}

func (r *providerRepository) Update(ctx context.Context, p *domain.Provider) error {
	err := r.dao.Update(ctx, r.toDAO(p))
	if err == nil && r.cache != nil {
		_ = r.cache.Invalidate(ctx)
	}
	return err
}

func (r *providerRepository) Delete(ctx context.Context, id int64) error {
	err := r.dao.Delete(ctx, id)
	if err == nil && r.cache != nil {
		_ = r.cache.Invalidate(ctx)
	}
	return err
}

func (r *providerRepository) GetByID(ctx context.Context, id int64) (*domain.Provider, error) {
	daoProvider, err := r.dao.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return r.toDomain(daoProvider), nil
}

func (r *providerRepository) GetByName(ctx context.Context, name string) (*domain.Provider, error) {
	daoProvider, err := r.dao.GetByName(ctx, name)
	if err != nil {
		return nil, err
	}
	return r.toDomain(daoProvider), nil
}

func (r *providerRepository) List(ctx context.Context) ([]domain.Provider, error) {
	if r.cache != nil {
		if providers, ok := r.cache.GetAll(ctx); ok {
			return providers, nil
		}
	}

	daoProviders, err := r.dao.List(ctx)
	if err != nil {
		return nil, err
	}

	providers := make([]domain.Provider, len(daoProviders))
	for i, p := range daoProviders {
		providers[i] = *r.toDomain(&p)
	}

	if r.cache != nil {
		_ = r.cache.SetAll(ctx, providers)
	}

	return providers, nil
}

func (r *providerRepository) GetDefaultByType(ctx context.Context, providerType string) (*domain.Provider, error) {
	daoProvider, err := r.dao.GetDefaultByType(ctx, providerType)
	if err != nil {
		return nil, err
	}
	return r.toDomain(daoProvider), nil
}
