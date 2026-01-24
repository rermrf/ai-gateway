package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"

	"ai-gateway/internal/domain"
)

// ModelRateCache 定义 ModelRate 缓存接口
type ModelRateCache interface {
	GetAllEnabled(ctx context.Context) ([]domain.ModelRate, bool)
	SetAllEnabled(ctx context.Context, rates []domain.ModelRate) error
	Invalidate(ctx context.Context) error
}

type redisModelRateCache struct {
	client redis.Cmdable
	ttl    time.Duration
}

// NewRedisModelRateCache 创建基于 Redis 的 ModelRate 缓存
func NewRedisModelRateCache(client redis.Cmdable) ModelRateCache {
	return &redisModelRateCache{
		client: client,
		ttl:    10 * time.Minute,
	}
}

func (c *redisModelRateCache) key() string {
	return "cache:model_rates:enabled"
}

func (c *redisModelRateCache) GetAllEnabled(ctx context.Context) ([]domain.ModelRate, bool) {
	val, err := c.client.Get(ctx, c.key()).Result()
	if err != nil {
		return nil, false
	}

	var rates []domain.ModelRate
	if err := json.Unmarshal([]byte(val), &rates); err != nil {
		return nil, false
	}

	return rates, true
}

func (c *redisModelRateCache) SetAllEnabled(ctx context.Context, rates []domain.ModelRate) error {
	data, err := json.Marshal(rates)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, c.key(), data, c.ttl).Err()
}

func (c *redisModelRateCache) Invalidate(ctx context.Context) error {
	return c.client.Del(ctx, c.key()).Err()
}
