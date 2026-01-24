package retry

import (
	"context"
	"math/rand"
	"time"
)

// Config 重试配置
type Config struct {
	MaxAttempts  int
	InitialDelay time.Duration
	MaxDelay     time.Duration
	Multiplier   float64
	Jitter       float64 // 0.0 - 1.0, 随机抖动因子
}

// DefaultConfig 默认配置
var DefaultConfig = Config{
	MaxAttempts:  3,
	InitialDelay: 100 * time.Millisecond,
	MaxDelay:     2 * time.Second,
	Multiplier:   2.0,
	Jitter:       0.2,
}

// Do 执行重试操作
func Do(ctx context.Context, cfg Config, fn func() error) error {
	var err error
	for i := 0; i < cfg.MaxAttempts; i++ {
		if err = fn(); err == nil {
			return nil
		}

		// 如果上下文已取消，直接返回上下文错误
		if ctx.Err() != nil {
			return ctx.Err()
		}

		// 检查错误是否可重试 (这里假设所有错误都重试，实际可以过滤)
		// 如果是不可重试错误，return err

		if i == cfg.MaxAttempts-1 {
			break
		}

		delay := calculateDelay(cfg, i)

		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
	return err
}

func calculateDelay(cfg Config, attempt int) time.Duration {
	delay := float64(cfg.InitialDelay) * pow(cfg.Multiplier, attempt)

	// Apply jitter
	if cfg.Jitter > 0 {
		jitter := (rand.Float64()*2 - 1) * cfg.Jitter // -Jitter to +Jitter
		delay = delay * (1 + jitter)
	}

	if delay > float64(cfg.MaxDelay) {
		delay = float64(cfg.MaxDelay)
	}

	return time.Duration(delay)
}

func pow(base float64, exp int) float64 {
	result := 1.0
	for i := 0; i < exp; i++ {
		result *= base
	}
	return result
}
