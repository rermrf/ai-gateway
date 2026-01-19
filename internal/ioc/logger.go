// Package ioc provides dependency injection initialization.
package ioc

import (
	"ai-gateway/config"
	"ai-gateway/internal/pkg/logger"

	"go.uber.org/zap"
)

// InitLogger initializes the logger.
func InitLogger(cfg *config.Config) *zap.Logger {
	jsonFormat := cfg.Log.Format == "json"
	return logger.InitLogger(cfg.Log.Level, jsonFormat)
}
