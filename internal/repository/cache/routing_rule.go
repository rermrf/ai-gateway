package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"

	"ai-gateway/internal/domain"
)

// RoutingRuleCache 定义 RoutingRule 缓存接口
type RoutingRuleCache interface {
	GetAll(ctx context.Context) ([]domain.RoutingRule, bool)
	SetAll(ctx context.Context, rules []domain.RoutingRule) error
	Invalidate(ctx context.Context) error
}

type redisRoutingRuleCache struct {
	client redis.Cmdable
	ttl    time.Duration
}

// NewRedisRoutingRuleCache 创建基于 Redis 的 RoutingRule 缓存
func NewRedisRoutingRuleCache(client redis.Cmdable) RoutingRuleCache {
	return &redisRoutingRuleCache{
		client: client,
		ttl:    5 * time.Minute,
	}
}

func (c *redisRoutingRuleCache) key() string {
	return "cache:routing_rules:all"
}

func (c *redisRoutingRuleCache) GetAll(ctx context.Context) ([]domain.RoutingRule, bool) {
	val, err := c.client.Get(ctx, c.key()).Result()
	if err != nil {
		return nil, false
	}

	var rules []domain.RoutingRule
	if err := json.Unmarshal([]byte(val), &rules); err != nil {
		return nil, false
	}

	return rules, true
}

func (c *redisRoutingRuleCache) SetAll(ctx context.Context, rules []domain.RoutingRule) error {
	data, err := json.Marshal(rules)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, c.key(), data, c.ttl).Err()
}

func (c *redisRoutingRuleCache) Invalidate(ctx context.Context) error {
	return c.client.Del(ctx, c.key()).Err()
}
