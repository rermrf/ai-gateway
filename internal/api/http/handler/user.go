// Package handler 提供 AI 网关的 HTTP 请求处理器。
package handler

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"ai-gateway/internal/api/http/middleware"
	"ai-gateway/internal/domain"
	"ai-gateway/internal/errs"
	"ai-gateway/internal/pkg/ginx"
	"ai-gateway/internal/pkg/logger"
	"ai-gateway/internal/service/apikey"
	"ai-gateway/internal/service/gateway"
	"ai-gateway/internal/service/modelrate"
	"ai-gateway/internal/service/user"
	"ai-gateway/internal/service/wallet"
)

// UserHandler 处理用户自助服务 API 请求。
type UserHandler struct {
	svc          user.Service
	apiKeySvc    apikey.Service
	walletSvc    wallet.Service
	gw           gateway.GatewayService
	modelRateSvc modelrate.Service
	logger       logger.Logger
}

// NewUserHandler 创建一个新的 UserHandler。
func NewUserHandler(
	svc user.Service,
	apiKeySvc apikey.Service,
	walletSvc wallet.Service,
	gw gateway.GatewayService,
	modelRateSvc modelrate.Service,
	l logger.Logger,
) *UserHandler {
	return &UserHandler{
		svc:          svc,
		apiKeySvc:    apiKeySvc,
		walletSvc:    walletSvc,
		gw:           gw,
		modelRateSvc: modelRateSvc,
		logger:       l.With(logger.String("handler", "user")),
	}
}

// ... existing methods ...

// ListAvailableModels 获取可用模型列表。
func (h *UserHandler) ListAvailableModels(c *gin.Context) {
	models, err := h.gw.ListModels(c.Request.Context())
	if err != nil {
		h.handleError(c, err)
		return
	}
	ginx.OK(c, models)
}

// ModelWithPricing 模型及定价信息。
type ModelWithPricing struct {
	ModelName       string  `json:"modelName"`       // 模型名称
	PromptPrice     float64 `json:"promptPrice"`     // 输入价格（每 1M tokens）
	CompletionPrice float64 `json:"completionPrice"` // 输出价格（每 1M tokens）
}

// ListModelsWithPricing 获取带价格信息的模型列表。
func (h *UserHandler) ListModelsWithPricing(c *gin.Context) {
	// 1. 获取所有可用模型
	models, err := h.gw.ListModels(c.Request.Context())
	if err != nil {
		h.handleError(c, err)
		return
	}

	// 2. 为每个模型匹配价格
	result := make([]ModelWithPricing, 0, len(models))
	for _, model := range models {
		// 获取模型对应的费率
		promptPrice, completionPrice, err := h.modelRateSvc.GetRateForModel(c.Request.Context(), model)
		if err != nil {
			h.logger.Warn("failed to get rate for model",
				logger.String("model", model),
				logger.Error(err))
			// 继续处理，使用默认价格 0
		}

		result = append(result, ModelWithPricing{
			ModelName:       model,
			PromptPrice:     promptPrice,
			CompletionPrice: completionPrice,
		})
	}

	ginx.OK(c, result)
}

// ... existing methods ...

// GetProfile 获取当前用户信息。
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)
	u, err := h.svc.GetByID(c.Request.Context(), userID)
	if err != nil {
		h.handleError(c, err)
		return
	}
	ginx.OK(c, h.toResponse(u))
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
		ginx.Fail(c, errs.CodeInvalidParameter, err.Error())
		return
	}

	u, err := h.svc.UpdateProfile(c.Request.Context(), userID, req.Email)
	if err != nil {
		h.handleError(c, err)
		return
	}
	ginx.OK(c, h.toResponse(u))
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
		ginx.Fail(c, errs.CodeInvalidParameter, err.Error())
		return
	}

	if err := h.svc.ChangePassword(c.Request.Context(), userID, req.OldPassword, req.NewPassword); err != nil {
		h.handleError(c, err)
		return
	}
	ginx.OK(c, gin.H{"message": "密码修改成功"})
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
	ginx.OK(c, keys)
}

// CreateAPIKeyRequest 创建 API Key 请求。
type CreateAPIKeyRequest struct {
	Name      string     `json:"name" binding:"required"`
	Enabled   *bool      `json:"enabled"`
	Quota     *float64   `json:"quota"`
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
}

// CreateMyAPIKey 创建 API Key。
func (h *UserHandler) CreateMyAPIKey(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req CreateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ginx.Fail(c, errs.CodeInvalidParameter, err.Error())
		return
	}

	apiKey, fullKey, err := h.apiKeySvc.Create(c.Request.Context(), userID, req.Name, req.Enabled, req.Quota, req.ExpiresAt)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.Status(201)
	ginx.OK(c, gin.H{
		"key":  fullKey,
		"data": apiKey,
	})
}

// DeleteMyAPIKey 删除 API Key。
func (h *UserHandler) DeleteMyAPIKey(c *gin.Context) {
	userID := middleware.GetUserID(c)
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		ginx.Fail(c, errs.CodeInvalidParameter, "无效的 ID")
		return
	}

	if err := h.apiKeySvc.Delete(c.Request.Context(), userID, id); err != nil {
		h.handleError(c, err)
		return
	}
	ginx.OK(c, gin.H{"message": "删除成功"})
}

// --- 钱包管理 ---

// GetMyWallet 获取当前用户钱包信息。
func (h *UserHandler) GetMyWallet(c *gin.Context) {
	userID := middleware.GetUserID(c)
	wallet, err := h.walletSvc.GetBalance(c.Request.Context(), userID)
	if err != nil {
		h.handleError(c, err)
		return
	}
	// 如果钱包不存在，返回余额0
	if wallet == nil {
		wallet = &domain.Wallet{UserID: userID, Balance: 0}
	}
	ginx.OK(c, wallet)
}

// GetMyTransactions 获取当前用户交易记录。
func (h *UserHandler) GetMyTransactions(c *gin.Context) {
	userID := middleware.GetUserID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	txs, total, err := h.walletSvc.GetTransactions(c.Request.Context(), userID, page, size)
	if err != nil {
		h.handleError(c, err)
		return
	}
	ginx.OK(c, gin.H{
		"data":  txs,
		"total": total,
		"page":  page,
		"size":  size,
	})
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
	ginx.OK(c, stats)
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
	ginx.OK(c, usage)
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
	h.logger.Warn("request failed", logger.Error(err))
	ginx.FromErr(c, err)
}
