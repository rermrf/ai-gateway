//go:build wireinject

package ioc

import (
	"time"

	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

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
	"ai-gateway/internal/service/chat"
	"ai-gateway/internal/service/gateway"
	"ai-gateway/internal/service/loadbalance"
	"ai-gateway/internal/service/modelrate"
	"ai-gateway/internal/service/provider"
	"ai-gateway/internal/service/routingrule"
	"ai-gateway/internal/service/usage"
	"ai-gateway/internal/service/user"
	"ai-gateway/internal/service/wallet"

	baseioc "ai-gateway/internal/ioc"
)

//go:generate go run github.com/google/wire/cmd/wire@v0.7.0

// App 应用依赖集合。
type App struct {
	Logger     logger.Logger
	HTTPServer *httpapi.Server
}

// InitApp 初始化应用（Wire 生成实现位于 wire_gen.go）。
func InitApp(cfg *config.Config) (*App, error) {
	wire.Build(
		// 基础设施
		provideLogger,
		provideDB,
		provideRedis,
		provideLimiter,
		provideAuthService,
		provideAuthConfig,

		// 缓存
		provideAPIKeyCache,
		provideProviderCache,
		provideRoutingRuleCache,
		provideModelRateCache,

		// DAO
		dao.NewGormProviderDAO,
		dao.NewGormRoutingRuleDAO,
		dao.NewGormAPIKeyDAO,
		dao.NewGormLoadBalanceDAO,
		dao.NewGormUserDAO,
		dao.NewGormUsageLogDAO,
		dao.NewGormWalletDAO,
		dao.NewGormModelRateDAO,

		// Repository
		repository.NewProviderRepository,
		repository.NewRoutingRuleRepository,
		repository.NewAPIKeyRepository,
		repository.NewLoadBalanceRepository,
		repository.NewUserRepository,
		repository.NewUsageLogRepository,
		repository.NewWalletRepository,
		repository.NewModelRateRepository,

		// Service
		apikey.NewService,
		modelrate.NewService,
		wallet.NewService,
		user.NewService,
		usage.NewService,
		provider.NewService,
		routingrule.NewService,
		loadbalance.NewService,
		gateway.NewGatewayService,
		chat.NewService,

		// Handler
		handler.NewOpenAIHandler,
		handler.NewAnthropicHandler,
		handler.NewAuthHandler,
		handler.NewUserHandler,
		handler.NewAdminHandler,
		handler.NewHealthHandler,

		// HTTP server
		httpapi.NewServer,

		// App
		wire.Struct(new(App), "*"),
	)
	return nil, nil
}

func provideLogger(cfg *config.Config) logger.Logger {
	return baseioc.InitLogger(cfg)
}

func provideDB(cfg *config.Config, l logger.Logger) (*gorm.DB, error) {
	return baseioc.InitDB(cfg, l)
}

func provideRedis(cfg *config.Config, l logger.Logger) (redis.Cmdable, error) {
	return baseioc.InitRedis(cfg, l)
}

func provideLimiter(cfg *config.Config, rdb redis.Cmdable) ratelimit.Limiter {
	if !cfg.RateLimit.Enabled || rdb == nil {
		return nil
	}
	window := cfg.RateLimit.Window
	if window <= 0 {
		window = time.Minute
	}
	rate := cfg.RateLimit.Rate
	if rate <= 0 {
		rate = 100
	}
	return ratelimit.NewRedisSlidingWindowLimiter(rdb, window, rate)
}

func provideAuthConfig(cfg *config.Config) config.AuthConfig {
	return cfg.Auth
}

func provideAuthService(cfg *config.Config) *auth.AuthService {
	secret := cfg.Auth.JWTSecret
	if secret == "" {
		secret = "ai-gateway-default-secret-change-in-production"
	}
	return auth.NewAuthService(secret, 24*time.Hour)
}

func provideAPIKeyCache(rdb redis.Cmdable) cache.APIKeyCache {
	if rdb == nil {
		return nil
	}
	return cache.NewRedisAPIKeyCache(rdb)
}

func provideProviderCache(rdb redis.Cmdable) cache.ProviderCache {
	if rdb == nil {
		return nil
	}
	return cache.NewRedisProviderCache(rdb)
}

func provideRoutingRuleCache(rdb redis.Cmdable) cache.RoutingRuleCache {
	if rdb == nil {
		return nil
	}
	return cache.NewRedisRoutingRuleCache(rdb)
}

func provideModelRateCache(rdb redis.Cmdable) cache.ModelRateCache {
	if rdb == nil {
		return nil
	}
	return cache.NewRedisModelRateCache(rdb)
}
