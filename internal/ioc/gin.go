// Package ioc 提供依赖注入初始化。
package ioc

import (
	"time"

	"ai-gateway/config"
	httpapi "ai-gateway/internal/api/http"
	"ai-gateway/internal/api/http/handler"
	"ai-gateway/internal/pkg/logger"
	"ai-gateway/internal/pkg/ratelimit"
	"ai-gateway/internal/repository"
	"ai-gateway/internal/repository/cache"
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
func InitGinServer(cfg *config.Config, l logger.Logger) *httpapi.Server {
	// 初始化数据库
	db, err := InitDB(cfg, l)
	if err != nil {
		l.Error("failed to initialize database", logger.Error(err))
		panic(err)
	}

	// 初始化 Redis
	rdb, err := InitRedis(cfg, l)
	if err != nil {
		l.Error("failed to initialize redis", logger.Error(err))
		panic(err)
	}

	// 初始化限流器 (1000 请求/分钟)
	// 如果 Redis 未初始化 (rdb == nil)，限流器将为 nil，中间件会自动处理
	var limiter ratelimit.Limiter
	if rdb != nil {
		limiter = ratelimit.NewRedisSlidingWindowLimiter(rdb, time.Minute, 1000)
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

	// 初始化缓存
	// 初始化缓存
	var apiKeyCache cache.APIKeyCache
	var providerCache cache.ProviderCache
	var routingRuleCache cache.RoutingRuleCache
	var modelRateCache cache.ModelRateCache

	if rdb != nil {
		apiKeyCache = cache.NewRedisAPIKeyCache(rdb)
		providerCache = cache.NewRedisProviderCache(rdb)
		routingRuleCache = cache.NewRedisRoutingRuleCache(rdb)
		modelRateCache = cache.NewRedisModelRateCache(rdb)
	}

	// 初始化仓库 (Repositories)
	providerRepo := repository.NewProviderRepository(providerDAO, providerCache)
	routingRuleRepo := repository.NewRoutingRuleRepository(routingRuleDAO, routingRuleCache)
	apiKeyRepo := repository.NewAPIKeyRepository(apiKeyDAO, apiKeyCache)
	loadBalanceRepo := repository.NewLoadBalanceRepository(loadBalanceDAO)
	userRepo := repository.NewUserRepository(userDAO)
	usageLogRepo := repository.NewUsageLogRepository(usageLogDAO)
	walletRepo := repository.NewWalletRepository(walletDAO)
	modelRateRepo := repository.NewModelRateRepository(modelRateDAO, modelRateCache)

	// 初始化认证服务
	jwtSecret := cfg.Auth.JWTSecret
	if jwtSecret == "" {
		jwtSecret = "ai-gateway-default-secret-change-in-production"
	}
	authService := auth.NewAuthService(jwtSecret, 24*time.Hour)

	// 初始化 API Key 服务
	apiKeySvc := apikey.NewService(apiKeyRepo, l)

	// 初始化 ModelRate 服务 (UserSvc 之前可能不需要，但 Wallet 需要)
	modelRateSvc := modelrate.NewService(modelRateRepo, l)

	// 初始化 Wallet 服务 (依赖 ModelRateSvc)
	walletSvc := wallet.NewService(walletRepo, modelRateSvc, l)

	// 初始化用户服务
	userSvc := user.NewService(userRepo, usageLogRepo, l)

	// 初始化使用统计服务 (依赖 WalletSvc)
	usageSvc := usage.NewService(usageLogRepo, walletSvc, l)

	// 初始化管理服务 (Admin Services)
	providerSvc := provider.NewService(providerRepo, l)
	routingRuleSvc := routingrule.NewService(routingRuleRepo, l)
	loadBalanceSvc := loadbalance.NewService(loadBalanceRepo, l)

	// 使用仓库初始化网关服务
	gw := gateway.NewGatewayService(
		providerRepo,
		routingRuleRepo,
		loadBalanceRepo,
		l,
	)

	// 初始化处理器
	openaiHandler := handler.NewOpenAIHandler(gw, walletSvc, usageSvc, apiKeySvc, modelRateSvc, l)
	anthropicHandler := handler.NewAnthropicHandler(gw, walletSvc, usageSvc, l)
	authHandler := handler.NewAuthHandler(userSvc, authService, l)
	userHandler := handler.NewUserHandler(userSvc, apiKeySvc, walletSvc, gw, modelRateSvc, l)
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
		l,
	)

	healthHandler := handler.NewHealthHandler(db, rdb, l)

	// 创建并返回带有身份验证配置的服务器
	return httpapi.NewServer(
		openaiHandler,
		anthropicHandler,
		adminHandler,
		authHandler,
		userHandler,
		healthHandler,
		authService,
		apiKeySvc,
		limiter,
		cfg.Auth,
		l,
	)
}
