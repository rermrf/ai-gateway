// Package ioc 提供依赖注入初始化。
package ioc

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"ai-gateway/config"
)

// InitRedis 初始化 Redis 客户端。
func InitRedis(cfg *config.Config, logger *zap.Logger) (redis.Cmdable, error) {
	// 如果未配置 Redis 地址，且未启用限流，则返回 nil
	if cfg.Redis.Addr == "" {
		if cfg.RateLimit.Enabled {
			return nil, fmt.Errorf("redis address is required when rate limit is enabled")
		}
		logger.Warn("redis not configured")
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
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	logger.Info("connected to redis", zap.String("addr", cfg.Redis.Addr), zap.Int("db", cfg.Redis.DB))
	return rdb, nil
}
