//go:build wireinject
// +build wireinject

// Package ioc provides Wire-based dependency injection.
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

// InitGinServerWithWire initializes the Gin HTTP server using Wire.
func InitGinServerWithWire(cfg *config.Config, logger *zap.Logger) *httpapi.Server {
	wire.Build(
		// Database
		provideDB, // Provides *gorm.DB

		// DAOs
		dao.NewGormProviderDAO,
		dao.NewGormRoutingRuleDAO,
		dao.NewGormLoadBalanceDAO,

		// Repositories
		repository.NewProviderRepository,
		repository.NewRoutingRuleRepository,
		repository.NewLoadBalanceRepository,

		// Gateway service
		gateway.NewGatewayService,

		// HTTP handlers
		handler.NewOpenAIHandler,
		handler.NewAnthropicHandler,
		provideAdminHandler,

		// Auth config
		provideAuthConfig,

		// HTTP server
		httpapi.NewServer,
	)
	return nil
}

// provideDB handles the error from InitDB.
func provideDB(cfg *config.Config, logger *zap.Logger) *gorm.DB {
	db, err := InitDB(cfg, logger)
	if err != nil {
		logger.Fatal("failed to initialize database", zap.Error(err))
	}
	if db == nil {
		logger.Fatal("MySQL must be enabled")
	}
	return db
}

// provideAuthConfig extracts AuthConfig from Config.
func provideAuthConfig(cfg *config.Config) config.AuthConfig {
	return cfg.Auth
}

// provideAdminHandler provides nil AdminHandler (to be implemented later).
func provideAdminHandler() *handler.AdminHandler {
	return nil
}
