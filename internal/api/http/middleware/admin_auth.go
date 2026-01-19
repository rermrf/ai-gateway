// Package middleware 为 AI 网关提供 HTTP 中间件。
package middleware

import (
	"crypto/subtle"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// AdminAuth 返回 Admin API 认证中间件。
// 支持 Basic Auth 和 Bearer Token 两种认证方式。
func AdminAuth() gin.HandlerFunc {
	// 从环境变量获取管理员凭证
	adminUser := os.Getenv("ADMIN_USER")
	adminPass := os.Getenv("ADMIN_PASS")
	adminToken := os.Getenv("ADMIN_TOKEN")

	// 如果没有配置任何凭证，使用默认值（仅用于开发环境）
	if adminUser == "" && adminToken == "" {
		adminUser = "admin"
		adminPass = "admin"
	}

	return func(c *gin.Context) {
		// 尝试 Bearer Token 认证
		authHeader := c.GetHeader("Authorization")
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			token := authHeader[7:]
			if adminToken != "" && subtle.ConstantTimeCompare([]byte(token), []byte(adminToken)) == 1 {
				c.Next()
				return
			}
		}

		// 尝试 Basic Auth 认证
		user, pass, hasAuth := c.Request.BasicAuth()
		if hasAuth {
			userMatch := subtle.ConstantTimeCompare([]byte(user), []byte(adminUser)) == 1
			passMatch := subtle.ConstantTimeCompare([]byte(pass), []byte(adminPass)) == 1
			if userMatch && passMatch {
				c.Next()
				return
			}
		}

		// 认证失败
		c.Header("WWW-Authenticate", `Basic realm="AI Gateway Admin"`)
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized",
		})
	}
}
