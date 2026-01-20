// Package repository 定义数据访问的存储库接口。
package repository

import (
	"context"
	"time"

	"ai-gateway/internal/repository/dao"
)

// APIKeyRepository 定义 API 密钥的存储库接口。
type APIKeyRepository interface {
	Create(ctx context.Context, key *dao.APIKey) error
	Update(ctx context.Context, key *dao.APIKey) error
	Delete(ctx context.Context, id int64) error
	GetByID(ctx context.Context, id int64) (*dao.APIKey, error)
	GetByKey(ctx context.Context, key string) (*dao.APIKey, error)
	List(ctx context.Context) ([]dao.APIKey, error)
	ListByTenantID(ctx context.Context, tenantID int64) ([]dao.APIKey, error)
	ListByUserID(ctx context.Context, userID int64) ([]dao.APIKey, error)
	Validate(ctx context.Context, key string) (bool, error)
}

// apiKeyRepository 是 APIKeyRepository 的默认实现。
type apiKeyRepository struct {
	dao dao.APIKeyDAO
}

// NewAPIKeyRepository 创建一个新的 APIKeyRepository。
func NewAPIKeyRepository(apiKeyDAO dao.APIKeyDAO) APIKeyRepository {
	return &apiKeyRepository{dao: apiKeyDAO}
}

func (r *apiKeyRepository) Create(ctx context.Context, key *dao.APIKey) error {
	return r.dao.Create(ctx, key)
}

func (r *apiKeyRepository) Update(ctx context.Context, key *dao.APIKey) error {
	return r.dao.Update(ctx, key)
}

func (r *apiKeyRepository) Delete(ctx context.Context, id int64) error {
	return r.dao.Delete(ctx, id)
}

func (r *apiKeyRepository) GetByID(ctx context.Context, id int64) (*dao.APIKey, error) {
	return r.dao.GetByID(ctx, id)
}

func (r *apiKeyRepository) GetByKey(ctx context.Context, key string) (*dao.APIKey, error) {
	return r.dao.GetByKey(ctx, key)
}

func (r *apiKeyRepository) List(ctx context.Context) ([]dao.APIKey, error) {
	return r.dao.List(ctx)
}

func (r *apiKeyRepository) ListByTenantID(ctx context.Context, tenantID int64) ([]dao.APIKey, error) {
	return r.dao.ListByTenantID(ctx, tenantID)
}

func (r *apiKeyRepository) ListByUserID(ctx context.Context, userID int64) ([]dao.APIKey, error) {
	return r.dao.ListByUserID(ctx, userID)
}

func (r *apiKeyRepository) Validate(ctx context.Context, key string) (bool, error) {
	apiKey, err := r.dao.GetByKey(ctx, key)
	if err != nil {
		return false, err
	}
	if apiKey == nil || !apiKey.Enabled {
		return false, nil
	}
	if apiKey.ExpiresAt != nil && apiKey.ExpiresAt.Before(time.Now()) {
		return false, nil
	}
	return true, nil
}
