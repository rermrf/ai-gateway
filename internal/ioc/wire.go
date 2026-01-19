//go:build wireinject
// +build wireinject

// Package ioc 提供基于 Wire 的依赖注入。
package ioc

import (
	"github.com/google/wire"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"ai-gateway/config"
	httpapi "ai-gateway/internal/api/http"
	"ai-gateway/internal/api/http/handler"
	"ai-gateway/internal/repository"
	"ai-gateway/internal/repository/dao"
	"ai-gateway/internal/service/gateway"
)

// InitGinServerWithWire 使用 Wire 初始化 Gin HTTP 服务器。
func InitGinServerWithWire(cfg *config.Config, logger *zap.Logger) *httpapi.Server {
	wire.Build(
		// 数据库
		provideDB, // 提供 *gorm.DB

		// DAOs
		dao.NewGormProviderDAO,
		dao.NewGormRoutingRuleDAO,
		dao.NewGormLoadBalanceDAO,

		// 仓库 (Repositories)
		repository.NewProviderRepository,
		repository.NewRoutingRuleRepository,
		repository.NewLoadBalanceRepository,

		// 网关服务 (Gateway service)
		gateway.NewGatewayService,

		// HTTP 处理器 (HTTP handlers)
		handler.NewOpenAIHandler,
		handler.NewAnthropicHandler,
		provideAdminHandler,

		// 验证配置 (Auth config)
		provideAuthConfig,

		// HTTP 服务器 (HTTP server)
		httpapi.NewServer,
	)
	return nil
}

// provideDB 处理来自 InitDB 的错误。
func provideDB(cfg *config.Config, logger *zap.Logger) *gorm.DB {
	db, err := InitDB(cfg, logger)
	if err != nil {
		logger.Fatal("failed to initialize database", zap.Error(err))
	}
	return db
}

// provideAuthConfig 从 Config 中提取 AuthConfig。
func provideAuthConfig(cfg *config.Config) config.AuthConfig {
	return cfg.Auth
}

// provideAdminHandler 提供 AdminHandler（将在以后实现）。
func provideAdminHandler() *handler.AdminHandler {
	return nil
}
