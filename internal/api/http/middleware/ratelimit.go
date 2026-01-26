// Package middleware 提供 AI 网关的 HTTP 中间件。
package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"ai-gateway/internal/pkg/logger"
	"ai-gateway/internal/pkg/ratelimit"
)

// RateLimiter 返回基于 Redis 的限流中间件。
// 该中间件基于 IP 进行限流。如果 Redis 不可用，则默认放行（Fail Open）。
func RateLimiter(limiter ratelimit.Limiter, l logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		if limiter == nil {
			c.Next()
			return
		}

		// 使用客户端 IP 作为限流 Key
		// key 格式: ratelimit:ip:<ip>
		key := "ratelimit:ip:" + c.ClientIP()

		// 调用限流器
		// 注意：Limiter 接口返回 bool, err
		// limit=true 表示被限流（超过阈值）
		limited, err := limiter.Limit(c.Request.Context(), key)
		if err != nil {
			// Redis 错误，记录日志但放行请求（Fail Open 策略）
			l.Warn("rate limiter failed",
				logger.Error(err),
				logger.String("key", key),
			)
			c.Next()
			return
		}

		if limited {
			// 触发限流
			l.Warn("request rate limited",
				logger.String("ip", c.ClientIP()),
				logger.String("path", c.Request.URL.Path),
			)
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": gin.H{
					"message": "Too many requests. Please try again later.",
					"type":    "rate_limit_error",
				},
			})
			return
		}

		c.Next()
	}
}
