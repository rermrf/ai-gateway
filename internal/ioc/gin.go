// Package ioc 提供依赖注入初始化。
package ioc

import (
	"time"

	"go.uber.org/zap"

	"ai-gateway/config"
	httpapi "ai-gateway/internal/api/http"
	"ai-gateway/internal/api/http/handler"
	"ai-gateway/internal/repository"
	"ai-gateway/internal/repository/dao"
	"ai-gateway/internal/service/apikey"
	"ai-gateway/internal/service/auth"
	"ai-gateway/internal/service/gateway"
	"ai-gateway/internal/service/loadbalance"
	"ai-gateway/internal/service/modelrate"
	"ai-gateway/internal/service/provider"
	"ai-gateway/internal/service/routingrule"
	"ai-gateway/internal/service/usage"
	"ai-gateway/internal/service/user"
	"ai-gateway/internal/service/wallet"
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
	usageLogDAO := dao.NewGormUsageLogDAO(db)
	walletDAO := dao.NewGormWalletDAO(db)
	modelRateDAO := dao.NewGormModelRateDAO(db)

	// 初始化仓库 (Repositories)
	providerRepo := repository.NewProviderRepository(providerDAO)
	routingRuleRepo := repository.NewRoutingRuleRepository(routingRuleDAO)
	apiKeyRepo := repository.NewAPIKeyRepository(apiKeyDAO)
	loadBalanceRepo := repository.NewLoadBalanceRepository(loadBalanceDAO)
	userRepo := repository.NewUserRepository(userDAO)
	usageLogRepo := repository.NewUsageLogRepository(usageLogDAO)
	walletRepo := repository.NewWalletRepository(walletDAO)
	modelRateRepo := repository.NewModelRateRepository(modelRateDAO)

	// 初始化认证服务
	jwtSecret := cfg.Auth.JWTSecret
	if jwtSecret == "" {
		jwtSecret = "ai-gateway-default-secret-change-in-production"
	}
	authService := auth.NewAuthService(jwtSecret, 24*time.Hour)

	// 初始化 API Key 服务
	apiKeySvc := apikey.NewService(apiKeyRepo, logger)

	// 初始化 ModelRate 服务 (UserSvc 之前可能不需要，但 Wallet 需要)
	modelRateSvc := modelrate.NewService(modelRateRepo, logger)

	// 初始化 Wallet 服务 (依赖 ModelRateSvc)
	walletSvc := wallet.NewService(walletRepo, modelRateSvc, logger)

	// 初始化用户服务
	userSvc := user.NewService(userRepo, usageLogRepo, logger)

	// 初始化使用统计服务 (依赖 WalletSvc)
	usageSvc := usage.NewService(usageLogRepo, walletSvc, logger)

	// 初始化管理服务 (Admin Services)
	providerSvc := provider.NewService(providerRepo, logger)
	routingRuleSvc := routingrule.NewService(routingRuleRepo, logger)
	loadBalanceSvc := loadbalance.NewService(loadBalanceRepo, logger)

	// 使用仓库初始化网关服务
	gw := gateway.NewGatewayService(
		providerRepo,
		routingRuleRepo,
		loadBalanceRepo,
		logger,
	)

	// 初始化处理器
	openaiHandler := handler.NewOpenAIHandler(gw, walletSvc, usageSvc, logger)
	anthropicHandler := handler.NewAnthropicHandler(gw, walletSvc, usageSvc, logger)
	authHandler := handler.NewAuthHandler(userSvc, authService, logger)
	userHandler := handler.NewUserHandler(userSvc, apiKeySvc, walletSvc, gw, modelRateSvc, logger)
	adminHandler := handler.NewAdminHandler(
		providerSvc,
		routingRuleSvc,
		loadBalanceSvc,
		apiKeySvc,
		userSvc,
		usageSvc,
		gw,
		modelRateSvc,
		walletSvc,
		logger,
	)

	// 创建并返回带有身份验证配置的服务器
	return httpapi.NewServer(
		openaiHandler,
		anthropicHandler,
		adminHandler,
		authHandler,
		userHandler,
		authService,
		apiKeySvc,
		cfg.Auth,
		logger,
	)
}
