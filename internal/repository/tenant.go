// Package repository 定义数据访问的存储库接口。
package repository

import (
	"context"

	"ai-gateway/internal/repository/dao"
)

// TenantRepository 定义租户的存储库接口。
type TenantRepository interface {
	Create(ctx context.Context, tenant *dao.Tenant) error
	Update(ctx context.Context, tenant *dao.Tenant) error
	Delete(ctx context.Context, id int64) error
	GetByID(ctx context.Context, id int64) (*dao.Tenant, error)
	GetBySlug(ctx context.Context, slug string) (*dao.Tenant, error)
	List(ctx context.Context) ([]dao.Tenant, error)
}

// tenantRepository 是 TenantRepository 的默认实现。
type tenantRepository struct {
	dao dao.TenantDAO
}

// NewTenantRepository 创建一个新的 TenantRepository。
func NewTenantRepository(tenantDAO dao.TenantDAO) TenantRepository {
	return &tenantRepository{dao: tenantDAO}
}

func (r *tenantRepository) Create(ctx context.Context, tenant *dao.Tenant) error {
	return r.dao.Create(ctx, tenant)
}

func (r *tenantRepository) Update(ctx context.Context, tenant *dao.Tenant) error {
	return r.dao.Update(ctx, tenant)
}

func (r *tenantRepository) Delete(ctx context.Context, id int64) error {
	return r.dao.Delete(ctx, id)
}

func (r *tenantRepository) GetByID(ctx context.Context, id int64) (*dao.Tenant, error) {
	return r.dao.GetByID(ctx, id)
}

func (r *tenantRepository) GetBySlug(ctx context.Context, slug string) (*dao.Tenant, error) {
	return r.dao.GetBySlug(ctx, slug)
}

func (r *tenantRepository) List(ctx context.Context) ([]dao.Tenant, error) {
	return r.dao.List(ctx)
}
