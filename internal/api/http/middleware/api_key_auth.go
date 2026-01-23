// Package middleware 为 AI Gateway 提供 HTTP 中间件。
package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"


	"ai-gateway/internal/pkg/logger"
	"ai-gateway/internal/service/apikey"
)

// APIKeyAuth 创建基于数据库的 API Key 认证中间件。
// 此中间件从数据库验证 API keys，并记录使用统计。
func APIKeyAuth(apiKeyService apikey.Service, l logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var key string

		// 尝试从 Authorization header 提取（OpenAI 格式：Authorization: Bearer xxx）
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && parts[0] == "Bearer" {
				key = parts[1]
			}
		}

		// 尝试从 x-api-key header 提取（Anthropic 格式）
		if key == "" {
			key = c.GetHeader("x-api-key")
		}

		// 检查是否提供了 API key
		if key == "" {
			l.Warn("missing API key",
				logger.String("path", c.Request.URL.Path),
				logger.String("ip", c.ClientIP()),
			)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"message": "Missing API key. Please provide via Authorization header (Bearer) or x-api-key header.",
					"type":    "authentication_error",
				},
			})
			return
		}

		// 验证 API key
		apiKey, err := apiKeyService.ValidateAPIKey(c.Request.Context(), key)
		if err != nil {
			l.Warn("API key validation failed",
				logger.Error(err),
				logger.String("path", c.Request.URL.Path),
				logger.String("ip", c.ClientIP()),
			)

			var message string
			switch err {
			case apikey.ErrInvalidAPIKey:
				message = "Invalid API key."
			case apikey.ErrAPIKeyDisabled:
				message = "API key is disabled."
			case apikey.ErrAPIKeyExpired:
				message = "API key has expired."
			default:
				message = "Authentication failed."
			}

			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"message": message,
					"type":    "authentication_error",
				},
			})
			return
		}

		// 将 API key 信息存储到上下文中，供后续处理器使用
		c.Set("api_key", key)
		c.Set("api_key_id", apiKey.ID)
		c.Set("user_id", apiKey.UserID)
		c.Set("api_key_name", apiKey.Name)

		// 异步记录使用情况（不阻塞请求）
		go func() {
			// 使用新的 context，因为原始请求可能已完成
			if err := apiKeyService.RecordUsage(context.Background(), apiKey.ID); err != nil {
				l.Error("failed to record API key usage",
					logger.Error(err),
					logger.Int64("key_id", apiKey.ID),
				)
			}
		}()

		c.Next()
	}
}
