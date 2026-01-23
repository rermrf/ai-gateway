// Package middleware 提供 AI 网关的 HTTP 中间件。
package middleware

import (
	"time"

	"github.com/gin-gonic/gin"


	"ai-gateway/internal/pkg/logger"
)

// Logger 返回基于 Zap 的日志中间件。
func Logger(l logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()

		fields := []logger.Field{
			logger.Int("status", statusCode),
			logger.String("method", c.Request.Method),
			logger.String("path", path),
			logger.String("query", query),
			logger.String("ip", c.ClientIP()),
			logger.Duration("latency", latency),
			logger.String("user-agent", c.Request.UserAgent()),
		}

		if requestID := c.GetString("request_id"); requestID != "" {
			fields = append(fields, logger.String("request_id", requestID))
		}

		if len(c.Errors) > 0 {
			fields = append(fields, logger.String("errors", c.Errors.String()))
		}

		if statusCode >= 500 {
			l.Error("server error", fields...)
		} else if statusCode >= 400 {
			l.Warn("client error", fields...)
		} else {
			l.Info("request completed", fields...)
		}
	}
}
