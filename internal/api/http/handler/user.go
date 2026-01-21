// Package handler 提供 AI 网关的 HTTP 请求处理器。
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"ai-gateway/internal/api/http/middleware"
	"ai-gateway/internal/domain"
	"ai-gateway/internal/service/apikey"
	"ai-gateway/internal/service/user"
)

// UserHandler 处理用户自助服务 API 请求。
type UserHandler struct {
	svc       user.Service
	apiKeySvc apikey.Service
	logger    *zap.Logger
}

// NewUserHandler 创建一个新的 UserHandler。
func NewUserHandler(svc user.Service, apiKeySvc apikey.Service, logger *zap.Logger) *UserHandler {
	return &UserHandler{
		svc:       svc,
		apiKeySvc: apiKeySvc,
		logger:    logger.Named("handler.user"),
	}
}

// GetProfile 获取当前用户信息。
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)
	u, err := h.svc.GetByID(c.Request.Context(), userID)
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": h.toResponse(u)})
}

// UpdateProfileRequest 更新个人信息请求。
type UpdateProfileRequest struct {
	Email string `json:"email" binding:"omitempty,email"`
}

// UpdateProfile 更新当前用户信息。
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	u, err := h.svc.UpdateProfile(c.Request.Context(), userID, req.Email)
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": h.toResponse(u)})
}

// ChangePasswordRequest 修改密码请求。
type ChangePasswordRequest struct {
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required,min=6"`
}

// ChangePassword 修改密码。
func (h *UserHandler) ChangePassword(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.ChangePassword(c.Request.Context(), userID, req.OldPassword, req.NewPassword); err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "密码修改成功"})
}

// --- API Key 管理 ---

// ListMyAPIKeys 获取当前用户的 API Key 列表。
func (h *UserHandler) ListMyAPIKeys(c *gin.Context) {
	userID := middleware.GetUserID(c)
	keys, err := h.apiKeySvc.ListByUserID(c.Request.Context(), userID)
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": keys})
}

// CreateAPIKeyRequest 创建 API Key 请求。
type CreateAPIKeyRequest struct {
	Name string `json:"name" binding:"required"`
}

// CreateMyAPIKey 创建 API Key。
func (h *UserHandler) CreateMyAPIKey(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req CreateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	apiKey, fullKey, err := h.apiKeySvc.Create(c.Request.Context(), userID, req.Name)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "API Key 创建成功，请妥善保存，此密钥只显示一次",
		"key":     fullKey,
		"data":    apiKey,
	})
}

// DeleteMyAPIKey 删除 API Key。
func (h *UserHandler) DeleteMyAPIKey(c *gin.Context) {
	userID := middleware.GetUserID(c)
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的 ID"})
		return
	}

	if err := h.apiKeySvc.Delete(c.Request.Context(), userID, id); err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// --- 使用统计 ---

// GetMyUsage 获取当前用户使用统计。
func (h *UserHandler) GetMyUsage(c *gin.Context) {
	userID := middleware.GetUserID(c)
	stats, err := h.svc.GetUsageStats(c.Request.Context(), userID)
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": stats})
}

// GetMyDailyUsage 获取当前用户每日使用详情。
func (h *UserHandler) GetMyDailyUsage(c *gin.Context) {
	userID := middleware.GetUserID(c)
	days := 30
	if d := c.Query("days"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil && parsed > 0 && parsed <= 90 {
			days = parsed
		}
	}

	usage, err := h.svc.GetDailyUsage(c.Request.Context(), userID, days)
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": usage})
}

// UserResponse 用户响应 DTO。
type UserResponse struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	Status    string `json:"status"`
	CreatedAt int64  `json:"createdAt"`
}

// toResponse 将 domain.User 转换为 UserResponse。
func (h *UserHandler) toResponse(u *domain.User) UserResponse {
	return UserResponse{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		Role:      u.Role.String(),
		Status:    u.Status.String(),
		CreatedAt: u.CreatedAt.UnixMilli(),
	}
}

// handleError 统一错误处理。
func (h *UserHandler) handleError(c *gin.Context, err error) {
	h.logger.Warn("request failed", zap.Error(err))

	switch err {
	case user.ErrUserNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case user.ErrUserAlreadyExists, user.ErrEmailAlreadyExists:
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case user.ErrInvalidPassword:
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case apikey.ErrAPIKeyNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case apikey.ErrAPIKeyNotOwned:
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器内部错误"})
	}
}
