// Package repository 定义数据访问的存储库接口。
// 存储库层聚合了 DAO 和缓存，暴露统一的数据访问接口。
package repository

import (
	"context"

	"ai-gateway/internal/repository/dao"
)

// ProviderRepository 定义 Provider 实体的存储库接口。
type ProviderRepository interface {
	Create(ctx context.Context, provider *dao.Provider) error
	Update(ctx context.Context, provider *dao.Provider) error
	Delete(ctx context.Context, id int64) error
	GetByID(ctx context.Context, id int64) (*dao.Provider, error)
	GetByName(ctx context.Context, name string) (*dao.Provider, error)
	List(ctx context.Context) ([]dao.Provider, error)
	GetDefaultByType(ctx context.Context, providerType string) (*dao.Provider, error)
}

// providerRepository 是 ProviderRepository 的默认实现。
type providerRepository struct {
	dao dao.ProviderDAO
}

// NewProviderRepository 创建一个新的 ProviderRepository。
func NewProviderRepository(providerDAO dao.ProviderDAO) ProviderRepository {
	return &providerRepository{dao: providerDAO}
}

func (r *providerRepository) Create(ctx context.Context, p *dao.Provider) error {
	return r.dao.Create(ctx, p)
}

func (r *providerRepository) Update(ctx context.Context, p *dao.Provider) error {
	return r.dao.Update(ctx, p)
}

func (r *providerRepository) Delete(ctx context.Context, id int64) error {
	return r.dao.Delete(ctx, id)
}

func (r *providerRepository) GetByID(ctx context.Context, id int64) (*dao.Provider, error) {
	return r.dao.GetByID(ctx, id)
}

func (r *providerRepository) GetByName(ctx context.Context, name string) (*dao.Provider, error) {
	return r.dao.GetByName(ctx, name)
}

func (r *providerRepository) List(ctx context.Context) ([]dao.Provider, error) {
	return r.dao.List(ctx)
}

func (r *providerRepository) GetDefaultByType(ctx context.Context, providerType string) (*dao.Provider, error) {
	return r.dao.GetDefaultByType(ctx, providerType)
}
