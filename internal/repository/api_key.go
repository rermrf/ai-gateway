// Package repository 定义数据访问的存储库接口。
package repository

import (
	"context"
	"time"

	"ai-gateway/internal/domain"
	"ai-gateway/internal/repository/cache"
	"ai-gateway/internal/repository/dao"
)

// APIKeyRepository 定义 API 密钥的存储库接口。
type APIKeyRepository interface {
	Create(ctx context.Context, key *domain.APIKey) error
	Update(ctx context.Context, key *domain.APIKey) error
	Delete(ctx context.Context, id int64) error
	GetByID(ctx context.Context, id int64) (*domain.APIKey, error)
	GetByKey(ctx context.Context, key string) (*domain.APIKey, error)
	List(ctx context.Context) ([]domain.APIKey, error)
	ListByUserID(ctx context.Context, userID int64) ([]domain.APIKey, error)
	Validate(ctx context.Context, key string) (bool, *domain.APIKey, error)
	UpdateLastUsed(ctx context.Context, id int64) error
	IncrementUsage(ctx context.Context, id int64, amount float64) error
}

// apiKeyRepository 是 APIKeyRepository 的默认实现。
type apiKeyRepository struct {
	dao   dao.APIKeyDAO
	cache cache.APIKeyCache
}

// NewAPIKeyRepository 创建一个新的 APIKeyRepository。
func NewAPIKeyRepository(apiKeyDAO dao.APIKeyDAO, cache cache.APIKeyCache) APIKeyRepository {
	return &apiKeyRepository{
		dao:   apiKeyDAO,
		cache: cache,
	}
}

// toDAO 将 domain.APIKey 转换为 dao.APIKey。
func (r *apiKeyRepository) toDAO(key *domain.APIKey) *dao.APIKey {
	return &dao.APIKey{
		ID:         key.ID,
		UserID:     key.UserID,
		Key:        key.Key,
		KeyHash:    key.KeyHash,
		Name:       key.Name,
		Enabled:    key.Enabled,
		Quota:      key.Quota,
		UsedAmount: key.UsedAmount,
		ExpiresAt:  key.ExpiresAt,
		LastUsedAt: key.LastUsedAt,
		CreatedAt:  key.CreatedAt,
	}
}

// toDomain 将 dao.APIKey 转换为 domain.APIKey。
func (r *apiKeyRepository) toDomain(key *dao.APIKey) *domain.APIKey {
	if key == nil {
		return nil
	}
	return &domain.APIKey{
		ID:         key.ID,
		UserID:     key.UserID,
		Key:        key.Key,
		KeyHash:    key.KeyHash,
		Name:       key.Name,
		Enabled:    key.Enabled,
		Quota:      key.Quota,
		UsedAmount: key.UsedAmount,
		ExpiresAt:  key.ExpiresAt,
		LastUsedAt: key.LastUsedAt,
		CreatedAt:  key.CreatedAt,
	}
}

func (r *apiKeyRepository) Create(ctx context.Context, key *domain.APIKey) error {
	daoKey := r.toDAO(key)
	if err := r.dao.Create(ctx, daoKey); err != nil {
		return err
	}
	key.ID = daoKey.ID
	key.CreatedAt = daoKey.CreatedAt
	return nil
}

func (r *apiKeyRepository) Update(ctx context.Context, key *domain.APIKey) error {
	if err := r.dao.Update(ctx, r.toDAO(key)); err != nil {
		return err
	}
	//失效缓存
	if r.cache != nil {
		_ = r.cache.Delete(ctx, key.Key)
	}
	return nil
}

func (r *apiKeyRepository) Delete(ctx context.Context, id int64) error {
	// 获取 key 以便删除缓存
	key, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if key != nil && r.cache != nil {
		_ = r.cache.Delete(ctx, key.Key)
	}
	return r.dao.Delete(ctx, id)
}

func (r *apiKeyRepository) GetByID(ctx context.Context, id int64) (*domain.APIKey, error) {
	daoKey, err := r.dao.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return r.toDomain(daoKey), nil
}

func (r *apiKeyRepository) GetByKey(ctx context.Context, key string) (*domain.APIKey, error) {
	// 1. 尝试从缓存获取
	if r.cache != nil {
		cached, err := r.cache.Get(ctx, key)
		if err == nil && cached != nil {
			return cached, nil
		}
	}

	// 2. 从数据库获取
	daoKey, err := r.dao.GetByKey(ctx, key)
	if err != nil {
		return nil, err
	}
	domainKey := r.toDomain(daoKey)

	// 3. 写入缓存
	if r.cache != nil && domainKey != nil {
		// 忽略缓存设置错误
		_ = r.cache.Set(ctx, domainKey)
	}

	return domainKey, nil
}

func (r *apiKeyRepository) List(ctx context.Context) ([]domain.APIKey, error) {
	daoKeys, err := r.dao.List(ctx)
	if err != nil {
		return nil, err
	}
	keys := make([]domain.APIKey, len(daoKeys))
	for i, k := range daoKeys {
		keys[i] = *r.toDomain(&k)
	}
	return keys, nil
}

func (r *apiKeyRepository) ListByUserID(ctx context.Context, userID int64) ([]domain.APIKey, error) {
	daoKeys, err := r.dao.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	keys := make([]domain.APIKey, len(daoKeys))
	for i, k := range daoKeys {
		keys[i] = *r.toDomain(&k)
	}
	return keys, nil
}

func (r *apiKeyRepository) Validate(ctx context.Context, key string) (bool, *domain.APIKey, error) {
	daoKey, err := r.dao.GetByKey(ctx, key)
	if err != nil {
		return false, nil, err
	}
	if daoKey == nil || !daoKey.Enabled {
		return false, nil, nil
	}
	if daoKey.ExpiresAt != nil && daoKey.ExpiresAt.Before(time.Now()) {
		return false, nil, nil
	}

	domainKey := r.toDomain(daoKey)
	if domainKey.IsQuotaExceeded() {
		return false, nil, nil
	}

	return true, domainKey, nil
}

func (r *apiKeyRepository) UpdateLastUsed(ctx context.Context, id int64) error {
	return r.dao.UpdateLastUsed(ctx, id)
}

func (r *apiKeyRepository) IncrementUsage(ctx context.Context, id int64, amount float64) error {
	return r.dao.IncrementUsage(ctx, id, amount)
}
