// Package handler 提供 AI 网关的 HTTP 请求处理器。
package handler

import (
	"errors"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"ai-gateway/internal/domain"
	"ai-gateway/internal/errs"
	"ai-gateway/internal/pkg/ginx"
	"ai-gateway/internal/pkg/logger"
	"ai-gateway/internal/service/auth"
	"ai-gateway/internal/service/user"
)

// AuthHandler 处理认证相关的 API 请求。
type AuthHandler struct {
	userSvc     user.Service
	authService *auth.AuthService
	logger      logger.Logger
}

// NewAuthHandler 创建一个新的 AuthHandler。
func NewAuthHandler(
	userSvc user.Service,
	authService *auth.AuthService,
	l logger.Logger,
) *AuthHandler {
	return &AuthHandler{
		userSvc:     userSvc,
		authService: authService,
		logger:      l.With(logger.String("handler", "auth")),
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
		ginx.Fail(c, errs.CodeInvalidParameter, err.Error())
		return
	}

	u, err := h.userSvc.Register(c.Request.Context(), req.Username, req.Email, req.Password)
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.Status(201)
	ginx.OK(c, gin.H{
		"id":       u.ID,
		"username": u.Username,
		"email":    u.Email,
		"role":     u.Role.String(),
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
		ginx.Fail(c, errs.CodeInvalidParameter, err.Error())
		return
	}

	// 查找用户
	u, err := h.userSvc.GetByUsername(c.Request.Context(), req.Username)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) {
			ginx.Fail(c, errs.CodeInvalidCredentials, "用户名或密码错误")
			return
		}
		h.handleError(c, err)
		return
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)); err != nil {
		ginx.Fail(c, errs.CodeInvalidCredentials, "用户名或密码错误")
		return
	}

	// 检查用户状态
	if !u.CanLogin() {
		if u.Status == domain.UserStatusPending {
			ginx.Fail(c, errs.CodeForbidden, "账号审核中，请联系管理员")
			return
		}
		ginx.Fail(c, errs.CodeForbidden, "用户已被禁用")
		return
	}

	// 生成 JWT
	token, err := h.authService.GenerateToken(u.ID, u.Username, string(u.Role))
	if err != nil {
		h.logger.Error("failed to generate token", logger.Error(err))
		ginx.Fail(c, errs.CodeInternalError, "登录失败")
		return
	}
	ginx.OK(c, LoginResponse{
		Token:    token,
		UserID:   u.ID,
		Username: u.Username,
		Role:     u.Role.String(),
	})
}

// handleError 统一错误处理。
func (h *AuthHandler) handleError(c *gin.Context, err error) {
	h.logger.Warn("request failed", logger.Error(err))
	ginx.FromErr(c, err)
}
