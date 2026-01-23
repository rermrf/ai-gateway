// Package middleware 提供 AI 网关的 HTTP 中间件。
package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"

	"ai-gateway/internal/pkg/logger"
)

// Recovery 返回 panic 恢复中间件。
func Recovery(l logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				l.Error("panic recovered",
					logger.Any("error", err),
					logger.String("stack", string(debug.Stack())),
					logger.String("path", c.Request.URL.Path),
					logger.String("method", c.Request.Method),
				)

				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": gin.H{
						"message": "Internal server error",
						"type":    "server_error",
					},
				})
			}
		}()
		c.Next()
	}
}
