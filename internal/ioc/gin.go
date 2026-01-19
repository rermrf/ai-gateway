// Package ioc provides dependency injection initialization.
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

// InitGinServer initializes the Gin HTTP server with all dependencies.
func InitGinServer(cfg *config.Config, logger *zap.Logger) *httpapi.Server {
	// Initialize database (required now)
	db, err := InitDB(cfg, logger)
	if err != nil {
		logger.Fatal("failed to initialize database", zap.Error(err))
	}
	if db == nil {
		logger.Fatal("MySQL must be enabled - please configure mysql.enabled=true in config.yaml")
	}

	// Initialize DAOs
	providerDAO := dao.NewGormProviderDAO(db)
	routingRuleDAO := dao.NewGormRoutingRuleDAO(db)
	apiKeyDAO := dao.NewGormAPIKeyDAO(db)
	loadBalanceDAO := dao.NewGormLoadBalanceDAO(db)

	// Initialize repositories
	providerRepo := repository.NewProviderRepository(providerDAO)
	routingRuleRepo := repository.NewRoutingRuleRepository(routingRuleDAO)
	apiKeyRepo := repository.NewAPIKeyRepository(apiKeyDAO)
	loadBalanceRepo := repository.NewLoadBalanceRepository(loadBalanceDAO)

	// Initialize gateway service with repositories
	gw := gateway.NewGatewayService(
		providerRepo,
		routingRuleRepo,
		loadBalanceRepo,
		logger,
	)

	// Initialize handlers
	openaiHandler := handler.NewOpenAIHandler(gw, logger)
	anthropicHandler := handler.NewAnthropicHandler(gw, logger)

	// TODO: Initialize admin handler when implemented
	_ = apiKeyRepo

	// Create and return server with auth config
	return httpapi.NewServer(openaiHandler, anthropicHandler, nil, cfg.Auth, logger)
}
