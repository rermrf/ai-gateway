// Package ioc 提供依赖注入初始化。
package ioc

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"ai-gateway/config"
	"ai-gateway/internal/pkg/logger"
)

// InitRedis 初始化 Redis 客户端。
func InitRedis(cfg *config.Config, l logger.Logger) (redis.Cmdable, error) {
	// 如果未配置 Redis 地址，且未启用限流，则返回 nil
	if cfg.Redis.Addr == "" {
		if cfg.RateLimit.Enabled {
			return nil, fmt.Errorf("redis address is required when rate limit is enabled")
		}
		l.Warn("redis not configured")
		return nil, nil
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:         cfg.Redis.Addr,
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		DialTimeout:  10 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		PoolSize:     10,
		MinIdleConns: 5,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		// Redis 目前主要用于限流与缓存：
		// - 限流启用时必须可用（fail fast）
		// - 限流关闭时允许降级为无 Redis（禁用缓存/限流）
		if cfg.RateLimit.Enabled {
			return nil, fmt.Errorf("failed to connect to redis: %w", err)
		}
		l.Warn("redis unavailable, continue without redis",
			logger.String("addr", cfg.Redis.Addr),
			logger.Error(err),
		)
		return nil, nil
	}

	l.Info("connected to redis", logger.String("addr", cfg.Redis.Addr), logger.Int("db", cfg.Redis.DB))
	return rdb, nil
}
