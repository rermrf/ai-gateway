// Package handler 提供 AI 网关的 HTTP 请求处理器。
package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"ai-gateway/internal/domain"
	"ai-gateway/internal/service/auth"
	"ai-gateway/internal/service/user"
)

// AuthHandler 处理认证相关的 API 请求。
type AuthHandler struct {
	userSvc     user.Service
	authService *auth.AuthService
	logger      *zap.Logger
}

// NewAuthHandler 创建一个新的 AuthHandler。
func NewAuthHandler(
	userSvc user.Service,
	authService *auth.AuthService,
	logger *zap.Logger,
) *AuthHandler {
	return &AuthHandler{
		userSvc:     userSvc,
		authService: authService,
		logger:      logger.Named("handler.auth"),
	}
}

// RegisterRequest 注册请求。
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// Register 用户注册。
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	u, err := h.userSvc.Register(c.Request.Context(), req.Username, req.Email, req.Password)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "注册成功，请等待管理员审核",
		"data": gin.H{
			"id":       u.ID,
			"username": u.Username,
			"email":    u.Email,
			"role":     u.Role.String(),
		},
	})
}

// LoginRequest 登录请求。
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应。
type LoginResponse struct {
	Token    string `json:"token"`
	UserID   int64  `json:"userId"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

// Login 用户登录。
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 查找用户
	u, err := h.userSvc.GetByUsername(c.Request.Context(), req.Username)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
			return
		}
		h.handleError(c, err)
		return
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}

	// 检查用户状态
	if !u.CanLogin() {
		if u.Status == domain.UserStatusPending {
			c.JSON(http.StatusForbidden, gin.H{"error": "账号审核中，请联系管理员"})
			return
		}
		c.JSON(http.StatusForbidden, gin.H{"error": "用户已被禁用"})
		return
	}

	// 生成 JWT
	token, err := h.authService.GenerateToken(u.ID, u.Username, string(u.Role))
	if err != nil {
		h.logger.Error("failed to generate token", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "登录失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": LoginResponse{
			Token:    token,
			UserID:   u.ID,
			Username: u.Username,
			Role:     u.Role.String(),
		},
	})
}

// handleError 统一错误处理。
func (h *AuthHandler) handleError(c *gin.Context, err error) {
	h.logger.Warn("request failed", zap.Error(err))

	switch err {
	case user.ErrUserAlreadyExists:
		c.JSON(http.StatusConflict, gin.H{"error": "用户名已存在"})
	case user.ErrEmailAlreadyExists:
		c.JSON(http.StatusConflict, gin.H{"error": "邮箱已被注册"})
	case user.ErrUserNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器内部错误"})
	}
}

// toUserResponse 辅助函数用于转换用户响应（如果需要）。
func (h *AuthHandler) toUserResponse(u *domain.User) map[string]interface{} {
	return map[string]interface{}{
		"id":        u.ID,
		"username":  u.Username,
		"email":     u.Email,
		"role":      u.Role.String(),
		"status":    u.Status.String(),
		"createdAt": u.CreatedAt.UnixMilli(),
	}
}
