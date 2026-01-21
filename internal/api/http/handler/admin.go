// Package handler 提供 AI 网关的 HTTP 请求处理器。
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"ai-gateway/internal/domain"
	"ai-gateway/internal/service/apikey"
	"ai-gateway/internal/service/gateway"
	"ai-gateway/internal/service/loadbalance"
	"ai-gateway/internal/service/provider"
	"ai-gateway/internal/service/routingrule"
	"ai-gateway/internal/service/usage"
	"ai-gateway/internal/service/user"
)

// AdminHandler 处理管理后台 API 请求。
type AdminHandler struct {
	providerSvc    provider.Service
	routingRuleSvc routingrule.Service
	loadBalanceSvc loadbalance.Service
	apiKeySvc      apikey.Service
	userSvc        user.Service
	usageSvc       usage.Service
	gatewaySvc     gateway.GatewayService
	logger         *zap.Logger
}

// NewAdminHandler 创建一个新的 AdminHandler。
func NewAdminHandler(
	providerSvc provider.Service,
	routingRuleSvc routingrule.Service,
	loadBalanceSvc loadbalance.Service,
	apiKeySvc apikey.Service,
	userSvc user.Service,
	usageSvc usage.Service,
	gatewaySvc gateway.GatewayService,
	logger *zap.Logger,
) *AdminHandler {
	return &AdminHandler{
		providerSvc:    providerSvc,
		routingRuleSvc: routingRuleSvc,
		loadBalanceSvc: loadBalanceSvc,
		apiKeySvc:      apiKeySvc,
		userSvc:        userSvc,
		usageSvc:       usageSvc,
		gatewaySvc:     gatewaySvc,
		logger:         logger.Named("handler.admin"),
	}
}

// --- Provider 管理 API ---

// ListProviders 获取所有提供商列表。
func (h *AdminHandler) ListProviders(c *gin.Context) {
	providers, err := h.providerSvc.List(c.Request.Context())
	if err != nil {
		h.logger.Error("failed to list providers", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list providers"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": providers})
}

// GetProvider 获取单个提供商详情。
func (h *AdminHandler) GetProvider(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	provider, err := h.providerSvc.GetByID(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("failed to get provider", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get provider"})
		return
	}
	if provider == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "provider not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": provider})
}

// CreateProviderRequest 创建提供商的请求体。
type CreateProviderRequest struct {
	Name      string `json:"name" binding:"required"`
	Type      string `json:"type" binding:"required"` // openai, anthropic
	APIKey    string `json:"apiKey" binding:"required"`
	BaseURL   string `json:"baseURL" binding:"required"`
	TimeoutMs int    `json:"timeoutMs"`
	IsDefault bool   `json:"isDefault"`
	Enabled   bool   `json:"enabled"`
}

// CreateProvider 创建新的提供商。
func (h *AdminHandler) CreateProvider(c *gin.Context) {
	var req CreateProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	provider := &domain.Provider{
		Name:      req.Name,
		Type:      req.Type,
		APIKey:    req.APIKey,
		BaseURL:   req.BaseURL,
		TimeoutMs: req.TimeoutMs,
		IsDefault: req.IsDefault,
		Enabled:   req.Enabled,
	}

	if err := h.providerSvc.Create(c.Request.Context(), provider); err != nil {
		h.logger.Error("failed to create provider", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create provider"})
		return
	}
	// Reload gateway configuration
	if err := h.gatewaySvc.Reload(c.Request.Context()); err != nil {
		h.logger.Warn("failed to reload gateway configuration", zap.Error(err))
	}
	c.JSON(http.StatusCreated, gin.H{"data": provider})
}

// UpdateProvider 更新提供商。
func (h *AdminHandler) UpdateProvider(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	provider, err := h.providerSvc.GetByID(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("failed to get provider", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get provider"})
		return
	}
	if provider == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "provider not found"})
		return
	}

	var req CreateProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	provider.Name = req.Name
	provider.Type = req.Type
	provider.APIKey = req.APIKey
	provider.BaseURL = req.BaseURL
	provider.TimeoutMs = req.TimeoutMs
	provider.IsDefault = req.IsDefault
	provider.Enabled = req.Enabled

	if err := h.providerSvc.Update(c.Request.Context(), provider); err != nil {
		h.logger.Error("failed to update provider", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update provider"})
		return
	}
	// Reload gateway configuration
	if err := h.gatewaySvc.Reload(c.Request.Context()); err != nil {
		h.logger.Warn("failed to reload gateway configuration", zap.Error(err))
	}
	c.JSON(http.StatusOK, gin.H{"data": provider})
}

// DeleteProvider 删除提供商。
func (h *AdminHandler) DeleteProvider(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.providerSvc.Delete(c.Request.Context(), id); err != nil {
		h.logger.Error("failed to delete provider", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete provider"})
		return
	}
	// Reload gateway configuration
	if err := h.gatewaySvc.Reload(c.Request.Context()); err != nil {
		h.logger.Warn("failed to reload gateway configuration", zap.Error(err))
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// --- 路由规则管理 API ---

// ListRoutingRules 获取所有路由规则。
func (h *AdminHandler) ListRoutingRules(c *gin.Context) {
	rules, err := h.routingRuleSvc.List(c.Request.Context())
	if err != nil {
		h.logger.Error("failed to list routing rules", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list routing rules"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rules})
}

// CreateRoutingRuleRequest 创建路由规则的请求体。
type CreateRoutingRuleRequest struct {
	RuleType     string `json:"ruleType" binding:"required"` // exact, prefix, wildcard
	Pattern      string `json:"pattern" binding:"required"`
	ProviderName string `json:"providerName" binding:"required"`
	ActualModel  string `json:"actualModel"`
	Priority     int    `json:"priority"`
	Enabled      bool   `json:"enabled"`
}

// CreateRoutingRule 创建新的路由规则。
func (h *AdminHandler) CreateRoutingRule(c *gin.Context) {
	var req CreateRoutingRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rule := &domain.RoutingRule{
		RuleType:     req.RuleType,
		Pattern:      req.Pattern,
		ProviderName: req.ProviderName,
		ActualModel:  req.ActualModel,
		Priority:     req.Priority,
		Enabled:      req.Enabled,
	}

	if err := h.routingRuleSvc.Create(c.Request.Context(), rule); err != nil {
		h.logger.Error("failed to create routing rule", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create routing rule"})
		return
	}
	// Reload gateway configuration
	if err := h.gatewaySvc.Reload(c.Request.Context()); err != nil {
		h.logger.Warn("failed to reload gateway configuration", zap.Error(err))
	}
	c.JSON(http.StatusCreated, gin.H{"data": rule})
}

// UpdateRoutingRule 更新路由规则。
func (h *AdminHandler) UpdateRoutingRule(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req CreateRoutingRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rule := &domain.RoutingRule{
		ID:           id,
		RuleType:     req.RuleType,
		Pattern:      req.Pattern,
		ProviderName: req.ProviderName,
		ActualModel:  req.ActualModel,
		Priority:     req.Priority,
		Enabled:      req.Enabled,
	}

	if err := h.routingRuleSvc.Update(c.Request.Context(), rule); err != nil {
		h.logger.Error("failed to update routing rule", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update routing rule"})
		return
	}
	// Reload gateway configuration
	if err := h.gatewaySvc.Reload(c.Request.Context()); err != nil {
		h.logger.Warn("failed to reload gateway configuration", zap.Error(err))
	}
	c.JSON(http.StatusOK, gin.H{"data": rule})
}

// DeleteRoutingRule 删除路由规则。
func (h *AdminHandler) DeleteRoutingRule(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.routingRuleSvc.Delete(c.Request.Context(), id); err != nil {
		h.logger.Error("failed to delete routing rule", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete routing rule"})
		return
	}
	// Reload gateway configuration
	if err := h.gatewaySvc.Reload(c.Request.Context()); err != nil {
		h.logger.Warn("failed to reload gateway configuration", zap.Error(err))
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// --- 负载均衡组管理 API ---

// ListLoadBalanceGroups 获取所有负载均衡组。
func (h *AdminHandler) ListLoadBalanceGroups(c *gin.Context) {
	groups, err := h.loadBalanceSvc.ListGroups(c.Request.Context())
	if err != nil {
		h.logger.Error("failed to list load balance groups", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list load balance groups"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": groups})
}

// CreateLoadBalanceGroupRequest 创建负载均衡组的请求体。
type CreateLoadBalanceGroupRequest struct {
	Name         string `json:"name" binding:"required"`
	ModelPattern string `json:"modelPattern" binding:"required"`
	Strategy     string `json:"strategy" binding:"required"` // round-robin, random, failover, weighted
	Enabled      bool   `json:"enabled"`
}

// CreateLoadBalanceGroup 创建新的负载均衡组。
func (h *AdminHandler) CreateLoadBalanceGroup(c *gin.Context) {
	var req CreateLoadBalanceGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	group := &domain.LoadBalanceGroup{
		Name:         req.Name,
		ModelPattern: req.ModelPattern,
		Strategy:     req.Strategy,
		Enabled:      req.Enabled,
	}

	if err := h.loadBalanceSvc.CreateGroup(c.Request.Context(), group); err != nil {
		h.logger.Error("failed to create load balance group", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create load balance group"})
		return
	}
	// Reload gateway configuration
	if err := h.gatewaySvc.Reload(c.Request.Context()); err != nil {
		h.logger.Warn("failed to reload gateway configuration", zap.Error(err))
	}
	c.JSON(http.StatusCreated, gin.H{"data": group})
}

// UpdateLoadBalanceGroup 更新负载均衡组。
func (h *AdminHandler) UpdateLoadBalanceGroup(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	group, err := h.loadBalanceSvc.GetGroupByID(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("failed to get load balance group", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get load balance group"})
		return
	}
	if group == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "load balance group not found"})
		return
	}

	var req CreateLoadBalanceGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	group.Name = req.Name
	group.ModelPattern = req.ModelPattern
	group.Strategy = req.Strategy
	group.Enabled = req.Enabled

	if err := h.loadBalanceSvc.UpdateGroup(c.Request.Context(), group); err != nil {
		h.logger.Error("failed to update load balance group", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update load balance group"})
		return
	}
	// Reload gateway configuration
	if err := h.gatewaySvc.Reload(c.Request.Context()); err != nil {
		h.logger.Warn("failed to reload gateway configuration", zap.Error(err))
	}
	c.JSON(http.StatusOK, gin.H{"data": group})
}

// DeleteLoadBalanceGroup 删除负载均衡组。
func (h *AdminHandler) DeleteLoadBalanceGroup(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.loadBalanceSvc.DeleteGroup(c.Request.Context(), id); err != nil {
		h.logger.Error("failed to delete load balance group", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete load balance group"})
		return
	}
	// Reload gateway configuration
	if err := h.gatewaySvc.Reload(c.Request.Context()); err != nil {
		h.logger.Warn("failed to reload gateway configuration", zap.Error(err))
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// --- API Key 管理 API ---

// ListAPIKeys 获取所有 API 密钥。
func (h *AdminHandler) ListAPIKeys(c *gin.Context) {
	keys, err := h.apiKeySvc.List(c.Request.Context())
	if err != nil {
		h.logger.Error("failed to list api keys", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list api keys"})
		return
	}

	// Service 已经脱敏，不需要再次脱敏
	c.JSON(http.StatusOK, gin.H{"data": keys})
}

// DeleteAPIKey 删除 API 密钥。
func (h *AdminHandler) DeleteAPIKey(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.apiKeySvc.DeleteByID(c.Request.Context(), id); err != nil {
		h.logger.Error("failed to delete api key", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete api key"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// --- 用户管理 API ---

// ListUsers 获取所有用户。
func (h *AdminHandler) ListUsers(c *gin.Context) {
	users, err := h.userSvc.List(c.Request.Context())
	if err != nil {
		h.logger.Error("failed to list users", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list users"})
		return
	}

	responses := make([]map[string]interface{}, len(users))
	for i, u := range users {
		responses[i] = h.toUserResponse(&u)
	}
	c.JSON(http.StatusOK, gin.H{"data": responses})
}

// GetUser 获取单个用户详情。
func (h *AdminHandler) GetUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	u, err := h.userSvc.GetByID(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("failed to get user", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": h.toUserResponse(u)})
}

// UpdateUserRequest 更新用户的请求体。
type UpdateUserRequest struct {
	Email  string `json:"email"`
	Role   string `json:"role"`   // user, admin
	Status string `json:"status"` // active, disabled
}

// UpdateUser 更新用户信息。
func (h *AdminHandler) UpdateUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	u, err := h.userSvc.UpdateUser(c.Request.Context(), id, domain.UserRole(req.Role), domain.UserStatus(req.Status))
	if err != nil {
		h.logger.Error("failed to update user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": h.toUserResponse(u)})
}

// DeleteUser 删除用户。
func (h *AdminHandler) DeleteUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	// 禁止删除管理员用户（ID 1）
	if id == 1 {
		c.JSON(http.StatusForbidden, gin.H{"error": "cannot delete admin user"})
		return
	}

	if err := h.userSvc.Delete(c.Request.Context(), id); err != nil {
		h.logger.Error("failed to delete user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete user"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// --- 仪表盘统计 API ---

// DashboardStats 获取仪表盘统计数据。
func (h *AdminHandler) DashboardStats(c *gin.Context) {
	// 获取用户数量
	users, _ := h.userSvc.List(c.Request.Context())
	userCount := len(users)

	// 获取 API Key 数量
	keys, _ := h.apiKeySvc.List(c.Request.Context())
	keyCount := len(keys)

	// 获取全局使用统计
	stats, err := h.usageSvc.GetGlobalStats(c.Request.Context())
	if err != nil {
		h.logger.Error("failed to get global stats", zap.Error(err))
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"userCount":   userCount,
			"apiKeyCount": keyCount,
			"usage":       stats,
		},
	})
}

// toUserResponse 将 domain.User 转换为响应格式。
func (h *AdminHandler) toUserResponse(u *domain.User) map[string]interface{} {
	return map[string]interface{}{
		"id":        u.ID,
		"username":  u.Username,
		"email":     u.Email,
		"role":      u.Role.String(),
		"status":    u.Status.String(),
		"createdAt": u.CreatedAt.UnixMilli(),
		"updatedAt": u.UpdatedAt.UnixMilli(),
	}
}

// GetGlobalUsage 获取全局使用统计（管理员）。
func (h *AdminHandler) GetGlobalUsage(c *gin.Context) {
	stats, err := h.usageSvc.GetGlobalStats(c.Request.Context())
	if err != nil {
		h.logger.Error("failed to get global usage", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get global usage"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": stats})
}
