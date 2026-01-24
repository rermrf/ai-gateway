package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"ai-gateway/internal/domain"
)

// APIKeyCache 定义 API Key 的缓存接口
type APIKeyCache interface {
	Get(ctx context.Context, key string) (*domain.APIKey, error)
	Set(ctx context.Context, key *domain.APIKey) error
	Delete(ctx context.Context, key string) error
}

type redisAPIKeyCache struct {
	client redis.Cmdable
	ttl    time.Duration
}

// NewRedisAPIKeyCache 创建基于 Redis 的 API Key 缓存
func NewRedisAPIKeyCache(client redis.Cmdable) APIKeyCache {
	return &redisAPIKeyCache{
		client: client,
		ttl:    5 * time.Minute, // 默认缓存 5 分钟
	}
}

func (c *redisAPIKeyCache) cacheKey(key string) string {
	return fmt.Sprintf("cache:apikey:%s", key)
}

func (c *redisAPIKeyCache) Get(ctx context.Context, key string) (*domain.APIKey, error) {
	val, err := c.client.Get(ctx, c.cacheKey(key)).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Cache Miss
		}
		return nil, err
	}

	var apiKey domain.APIKey
	if err := json.Unmarshal([]byte(val), &apiKey); err != nil {
		return nil, err
	}

	return &apiKey, nil
}

func (c *redisAPIKeyCache) Set(ctx context.Context, apiKey *domain.APIKey) error {
	data, err := json.Marshal(apiKey)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, c.cacheKey(apiKey.Key), data, c.ttl).Err()
}

func (c *redisAPIKeyCache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, c.cacheKey(key)).Err()
}
