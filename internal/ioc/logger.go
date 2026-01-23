// Package ioc 提供依赖注入初始化。
package ioc

import (
	"ai-gateway/config"
	"ai-gateway/internal/pkg/logger"


)

// InitLogger 初始化日志记录器。
func InitLogger(cfg *config.Config) logger.Logger {
	jsonFormat := cfg.Log.Format == "json"
	zapL := logger.InitLogger(cfg.Log.Level, jsonFormat)
	return logger.NewZapLogger(zapL)
}
