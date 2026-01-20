// Package handler 提供 AI 网关的 HTTP 请求处理器。
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"ai-gateway/internal/repository"
	"ai-gateway/internal/repository/dao"
)

// AdminHandler 处理管理后台 API 请求。
type AdminHandler struct {
	providerRepo    repository.ProviderRepository
	routingRuleRepo repository.RoutingRuleRepository
	loadBalanceRepo repository.LoadBalanceRepository
	apiKeyRepo      repository.APIKeyRepository
	userRepo        repository.UserRepository
	tenantRepo      repository.TenantRepository
	logger          *zap.Logger
}

// NewAdminHandler 创建一个新的 AdminHandler。
func NewAdminHandler(
	providerRepo repository.ProviderRepository,
	routingRuleRepo repository.RoutingRuleRepository,
	loadBalanceRepo repository.LoadBalanceRepository,
	apiKeyRepo repository.APIKeyRepository,
	userRepo repository.UserRepository,
	tenantRepo repository.TenantRepository,
	logger *zap.Logger,
) *AdminHandler {
	return &AdminHandler{
		providerRepo:    providerRepo,
		routingRuleRepo: routingRuleRepo,
		loadBalanceRepo: loadBalanceRepo,
		apiKeyRepo:      apiKeyRepo,
		userRepo:        userRepo,
		tenantRepo:      tenantRepo,
		logger:          logger.Named("admin.handler"),
	}
}

// --- Provider 管理 API ---

// ListProviders 获取所有提供商列表。
func (h *AdminHandler) ListProviders(c *gin.Context) {
	providers, err := h.providerRepo.List(c.Request.Context())
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

	provider, err := h.providerRepo.GetByID(c.Request.Context(), id)
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
	Type      string `json:"type" binding:"required"`
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

	provider := &dao.Provider{
		Name:      req.Name,
		Type:      req.Type,
		APIKey:    req.APIKey,
		BaseURL:   req.BaseURL,
		TimeoutMs: req.TimeoutMs,
		IsDefault: req.IsDefault,
		Enabled:   req.Enabled,
	}

	if err := h.providerRepo.Create(c.Request.Context(), provider); err != nil {
		h.logger.Error("failed to create provider", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create provider"})
		return
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

	var req CreateProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	provider := &dao.Provider{
		ID:        id,
		Name:      req.Name,
		Type:      req.Type,
		APIKey:    req.APIKey,
		BaseURL:   req.BaseURL,
		TimeoutMs: req.TimeoutMs,
		IsDefault: req.IsDefault,
		Enabled:   req.Enabled,
	}

	if err := h.providerRepo.Update(c.Request.Context(), provider); err != nil {
		h.logger.Error("failed to update provider", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update provider"})
		return
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

	if err := h.providerRepo.Delete(c.Request.Context(), id); err != nil {
		h.logger.Error("failed to delete provider", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete provider"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// --- 路由规则管理 API ---

// ListRoutingRules 获取所有路由规则。
func (h *AdminHandler) ListRoutingRules(c *gin.Context) {
	rules, err := h.routingRuleRepo.List(c.Request.Context())
	if err != nil {
		h.logger.Error("failed to list routing rules", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list routing rules"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rules})
}

// CreateRoutingRuleRequest 创建路由规则的请求体。
type CreateRoutingRuleRequest struct {
	RuleType     string `json:"ruleType" binding:"required"`
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

	rule := &dao.RoutingRule{
		RuleType:     req.RuleType,
		Pattern:      req.Pattern,
		ProviderName: req.ProviderName,
		ActualModel:  req.ActualModel,
		Priority:     req.Priority,
		Enabled:      req.Enabled,
	}

	if err := h.routingRuleRepo.Create(c.Request.Context(), rule); err != nil {
		h.logger.Error("failed to create routing rule", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create routing rule"})
		return
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

	rule := &dao.RoutingRule{
		ID:           id,
		RuleType:     req.RuleType,
		Pattern:      req.Pattern,
		ProviderName: req.ProviderName,
		ActualModel:  req.ActualModel,
		Priority:     req.Priority,
		Enabled:      req.Enabled,
	}

	if err := h.routingRuleRepo.Update(c.Request.Context(), rule); err != nil {
		h.logger.Error("failed to update routing rule", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update routing rule"})
		return
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

	if err := h.routingRuleRepo.Delete(c.Request.Context(), id); err != nil {
		h.logger.Error("failed to delete routing rule", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete routing rule"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// --- API Key 管理 API ---

// ListAPIKeys 获取所有 API 密钥。
func (h *AdminHandler) ListAPIKeys(c *gin.Context) {
	keys, err := h.apiKeyRepo.List(c.Request.Context())
	if err != nil {
		h.logger.Error("failed to list api keys", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list api keys"})
		return
	}
	// 脱敏显示
	for i := range keys {
		if len(keys[i].Key) > 8 {
			keys[i].Key = keys[i].Key[:4] + "****" + keys[i].Key[len(keys[i].Key)-4:]
		}
	}
	c.JSON(http.StatusOK, gin.H{"data": keys})
}

// CreateAPIKeyRequest 创建 API 密钥的请求体。
type CreateAPIKeyRequest struct {
	TenantID  int64  `json:"tenantId" binding:"required"`
	Name      string `json:"name" binding:"required"`
	UserID    *int64 `json:"userId"`    // 关联用户 (NULL 表示租户级共享 Key)
	ExpiresAt string `json:"expiresAt"` // ISO 8601 格式
}

// CreateAPIKey 创建新的 API 密钥。
func (h *AdminHandler) CreateAPIKey(c *gin.Context) {
	var req CreateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证租户是否存在
	tenant, err := h.tenantRepo.GetByID(c.Request.Context(), req.TenantID)
	if err != nil {
		h.logger.Error("failed to get tenant", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to validate tenant"})
		return
	}
	if tenant == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant not found"})
		return
	}

	// 如果指定了用户，验证用户存在
	if req.UserID != nil {
		user, err := h.userRepo.GetByID(c.Request.Context(), *req.UserID)
		if err != nil {
			h.logger.Error("failed to get user", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to validate user"})
			return
		}
		if user == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user not found"})
			return
		}
	}

	// 生成随机密钥
	key := generateAPIKey()

	apiKey := &dao.APIKey{
		TenantID: req.TenantID,
		UserID:   req.UserID,
		Key:      key,
		Name:     req.Name,
		Enabled:  true,
	}

	if err := h.apiKeyRepo.Create(c.Request.Context(), apiKey); err != nil {
		h.logger.Error("failed to create api key", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create api key"})
		return
	}

	// 返回完整密钥（仅此一次）
	c.JSON(http.StatusCreated, gin.H{
		"data": apiKey,
		"key":  key, // 仅在创建时显示完整密钥
	})
}

// DeleteAPIKey 删除 API 密钥。
func (h *AdminHandler) DeleteAPIKey(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.apiKeyRepo.Delete(c.Request.Context(), id); err != nil {
		h.logger.Error("failed to delete api key", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete api key"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// --- 用户管理 API ---

// ListUsers 获取所有用户列表。
func (h *AdminHandler) ListUsers(c *gin.Context) {
	users, err := h.userRepo.List(c.Request.Context())
	if err != nil {
		h.logger.Error("failed to list users", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list users"})
		return
	}
	// 脱敏：不返回密码哈希
	for i := range users {
		users[i].PasswordHash = ""
	}
	c.JSON(http.StatusOK, gin.H{"data": users})
}

// GetUser 获取单个用户详情。
func (h *AdminHandler) GetUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	user, err := h.userRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("failed to get user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user"})
		return
	}
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	user.PasswordHash = "" // 脱敏
	c.JSON(http.StatusOK, gin.H{"data": user})
}

// CreateUserRequest 创建用户的请求体。
type CreateUserRequest struct {
	TenantID int64  `json:"tenantId" binding:"required"`
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password"`
	Role     string `json:"role"` // owner, admin, member
}

// CreateUser 创建新用户。
func (h *AdminHandler) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查用户名是否已存在
	existing, _ := h.userRepo.GetByTenantAndUsername(c.Request.Context(), req.TenantID, req.Username)
	if existing != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "username already exists in this tenant"})
		return
	}

	role := dao.UserRole(req.Role)
	if role == "" {
		role = dao.UserRoleMember
	}

	user := &dao.User{
		TenantID: req.TenantID,
		Username: req.Username,
		Email:    req.Email,
		Role:     role,
		Status:   dao.UserStatusActive,
	}

	// 简单密码处理（生产环境应使用 bcrypt）
	if req.Password != "" {
		user.PasswordHash = req.Password // TODO: 使用 bcrypt 加密
	}

	if err := h.userRepo.Create(c.Request.Context(), user); err != nil {
		h.logger.Error("failed to create user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": user})
}

// UpdateUserRequest 更新用户的请求体。
type UpdateUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`   // owner, admin, member
	Status   string `json:"status"` // active, disabled
}

// UpdateUser 更新用户信息。
func (h *AdminHandler) UpdateUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	// 获取现有用户
	user, err := h.userRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("failed to get user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user"})
		return
	}
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 更新字段
	if req.Username != "" {
		user.Username = req.Username
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Role != "" {
		user.Role = dao.UserRole(req.Role)
	}
	if req.Status != "" {
		user.Status = dao.UserStatus(req.Status)
	}

	if err := h.userRepo.Update(c.Request.Context(), user); err != nil {
		h.logger.Error("failed to update user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": user})
}

// DeleteUser 删除用户。
func (h *AdminHandler) DeleteUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	// 禁止删除系统用户
	if id == 1 {
		c.JSON(http.StatusForbidden, gin.H{"error": "cannot delete system user"})
		return
	}

	if err := h.userRepo.Delete(c.Request.Context(), id); err != nil {
		h.logger.Error("failed to delete user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete user"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// GetUserAPIKeys 获取用户的所有 API 密钥。
func (h *AdminHandler) GetUserAPIKeys(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	// 验证用户存在
	user, err := h.userRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("failed to get user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user"})
		return
	}
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	keys, err := h.apiKeyRepo.ListByUserID(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("failed to list user api keys", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list user api keys"})
		return
	}

	// 脱敏显示
	for i := range keys {
		if len(keys[i].Key) > 8 {
			keys[i].Key = keys[i].Key[:4] + "****" + keys[i].Key[len(keys[i].Key)-4:]
		}
	}
	c.JSON(http.StatusOK, gin.H{"data": keys})
}

// --- 负载均衡管理 API ---

// ListLoadBalanceGroups 获取所有负载均衡组。
func (h *AdminHandler) ListLoadBalanceGroups(c *gin.Context) {
	groups, err := h.loadBalanceRepo.ListGroups(c.Request.Context())
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
	Strategy     string `json:"strategy" binding:"required"`
	Enabled      bool   `json:"enabled"`
}

// CreateLoadBalanceGroup 创建新的负载均衡组。
func (h *AdminHandler) CreateLoadBalanceGroup(c *gin.Context) {
	var req CreateLoadBalanceGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	group := &dao.LoadBalanceGroup{
		Name:         req.Name,
		ModelPattern: req.ModelPattern,
		Strategy:     req.Strategy,
		Enabled:      req.Enabled,
	}

	if err := h.loadBalanceRepo.CreateGroup(c.Request.Context(), group); err != nil {
		h.logger.Error("failed to create load balance group", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create load balance group"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": group})
}

// DeleteLoadBalanceGroup 删除负载均衡组。
func (h *AdminHandler) DeleteLoadBalanceGroup(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.loadBalanceRepo.DeleteGroup(c.Request.Context(), id); err != nil {
		h.logger.Error("failed to delete load balance group", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete load balance group"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// --- 仪表盘 API ---

// DashboardStats 获取仪表盘统计数据。
func (h *AdminHandler) DashboardStats(c *gin.Context) {
	ctx := c.Request.Context()

	// 获取各项统计数据
	providers, _ := h.providerRepo.List(ctx)
	rules, _ := h.routingRuleRepo.List(ctx)
	groups, _ := h.loadBalanceRepo.ListGroups(ctx)
	keys, _ := h.apiKeyRepo.List(ctx)

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"providerCount":    len(providers),
			"routingRuleCount": len(rules),
			"loadBalanceCount": len(groups),
			"apiKeyCount":      len(keys),
		},
	})
}

// --- 辅助函数 ---

func generateAPIKey() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const keyLen = 32
	key := make([]byte, keyLen)
	for i := range key {
		key[i] = charset[i%len(charset)]
	}
	return "gw-" + string(key)
}
