// Package ioc 提供依赖注入初始化。
package ioc

import (
	"ai-gateway/config"
	"ai-gateway/internal/pkg/logger"

	"go.uber.org/zap"
)

// InitLogger 初始化日志记录器。
func InitLogger(cfg *config.Config) *zap.Logger {
	jsonFormat := cfg.Log.Format == "json"
	return logger.InitLogger(cfg.Log.Level, jsonFormat)
}
