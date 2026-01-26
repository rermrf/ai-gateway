// Package middleware 提供 HTTP 中间件。
package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"ai-gateway/internal/errs"
	"ai-gateway/internal/service/auth"
)

// ContextKey 用于在 Context 中存储值的键类型。
type ContextKey string

const (
	// UserIDKey 用户 ID 在 Context 中的键。
	UserIDKey ContextKey = "userId"
	// UsernameKey 用户名在 Context 中的键。
	UsernameKey ContextKey = "username"
	// UserRoleKey 用户角色在 Context 中的键。
	UserRoleKey ContextKey = "userRole"
)

// JWTAuth 创建 JWT 认证中间件。
func JWTAuth(authService *auth.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未提供认证信息"})
			c.Abort()
			return
		}

		// 解析 Bearer Token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "认证格式错误"})
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims, err := authService.ValidateToken(tokenString)
		if err != nil {
			if errors.Is(err, errs.ErrTokenExpired) {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Token 已过期"})
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的 Token"})
			}
			c.Abort()
			return
		}

		// 将用户信息存储到 Context
		c.Set(string(UserIDKey), claims.UserID)
		c.Set(string(UsernameKey), claims.Username)
		c.Set(string(UserRoleKey), claims.Role)

		c.Next()
	}
}

// RequireAdmin 要求管理员权限的中间件。
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get(string(UserRoleKey))
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
			c.Abort()
			return
		}

		if role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetUserID 从 Context 获取当前用户 ID。
func GetUserID(c *gin.Context) int64 {
	if id, exists := c.Get(string(UserIDKey)); exists {
		return id.(int64)
	}
	return 0
}

// GetUsername 从 Context 获取当前用户名。
func GetUsername(c *gin.Context) string {
	if username, exists := c.Get(string(UsernameKey)); exists {
		return username.(string)
	}
	return ""
}

// GetUserRole 从 Context 获取当前用户角色。
func GetUserRole(c *gin.Context) string {
	if role, exists := c.Get(string(UserRoleKey)); exists {
		return role.(string)
	}
	return ""
}
