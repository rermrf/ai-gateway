// Package ioc 提供依赖注入初始化。
package ioc

import (
	"go.uber.org/zap"

	"ai-gateway/config"
	httpapi "ai-gateway/internal/api/http"
	"ai-gateway/internal/api/http/handler"
	"ai-gateway/internal/repository"
	"ai-gateway/internal/repository/dao"
	"ai-gateway/internal/service/gateway"
)

// InitGinServer 初始化带有所有依赖项的 Gin HTTP 服务器。
func InitGinServer(cfg *config.Config, logger *zap.Logger) *httpapi.Server {
	// 初始化数据库
	db, err := InitDB(cfg, logger)
	if err != nil {
		logger.Fatal("failed to initialize database", zap.Error(err))
	}

	// 初始化 DAOs
	providerDAO := dao.NewGormProviderDAO(db)
	routingRuleDAO := dao.NewGormRoutingRuleDAO(db)
	apiKeyDAO := dao.NewGormAPIKeyDAO(db)
	loadBalanceDAO := dao.NewGormLoadBalanceDAO(db)
	userDAO := dao.NewGormUserDAO(db)
	tenantDAO := dao.NewGormTenantDAO(db)

	// 初始化仓库 (Repositories)
	providerRepo := repository.NewProviderRepository(providerDAO)
	routingRuleRepo := repository.NewRoutingRuleRepository(routingRuleDAO)
	apiKeyRepo := repository.NewAPIKeyRepository(apiKeyDAO)
	loadBalanceRepo := repository.NewLoadBalanceRepository(loadBalanceDAO)
	userRepo := repository.NewUserRepository(userDAO)
	tenantRepo := repository.NewTenantRepository(tenantDAO)

	// 使用仓库初始化网关服务
	gw := gateway.NewGatewayService(
		providerRepo,
		routingRuleRepo,
		loadBalanceRepo,
		logger,
	)

	// 初始化处理器
	openaiHandler := handler.NewOpenAIHandler(gw, logger)
	anthropicHandler := handler.NewAnthropicHandler(gw, logger)
	adminHandler := handler.NewAdminHandler(
		providerRepo,
		routingRuleRepo,
		loadBalanceRepo,
		apiKeyRepo,
		userRepo,
		tenantRepo,
		logger,
	)

	// 创建并返回带有身份验证配置的服务器
	return httpapi.NewServer(openaiHandler, anthropicHandler, adminHandler, cfg.Auth, logger)
}
