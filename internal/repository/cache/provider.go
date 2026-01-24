package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"

	"ai-gateway/internal/domain"
)

// ProviderCache 定义 Provider 缓存接口
type ProviderCache interface {
	GetAll(ctx context.Context) ([]domain.Provider, bool)
	SetAll(ctx context.Context, providers []domain.Provider) error
	Invalidate(ctx context.Context) error
}

type redisProviderCache struct {
	client redis.Cmdable
	ttl    time.Duration
}

// NewRedisProviderCache 创建基于 Redis 的 Provider 缓存
func NewRedisProviderCache(client redis.Cmdable) ProviderCache {
	return &redisProviderCache{
		client: client,
		ttl:    5 * time.Minute,
	}
}

func (c *redisProviderCache) key() string {
	return "cache:providers:all"
}

func (c *redisProviderCache) GetAll(ctx context.Context) ([]domain.Provider, bool) {
	val, err := c.client.Get(ctx, c.key()).Result()
	if err != nil {
		return nil, false
	}

	var providers []domain.Provider
	if err := json.Unmarshal([]byte(val), &providers); err != nil {
		return nil, false
	}

	return providers, true
}

func (c *redisProviderCache) SetAll(ctx context.Context, providers []domain.Provider) error {
	data, err := json.Marshal(providers)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, c.key(), data, c.ttl).Err()
}

func (c *redisProviderCache) Invalidate(ctx context.Context) error {
	return c.client.Del(ctx, c.key()).Err()
}
