// Package ioc 提供依赖注入初始化。
package ioc

import (
	"ai-gateway/config"
	"ai-gateway/internal/pkg/logger"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// InitLogger 初始化日志记录器。
func InitLogger(cfg *config.Config) logger.Logger {
	jsonFormat := cfg.Log.Format == "json"

	var zapLevel zapcore.Level
	switch cfg.Log.Level {
	case "debug":
		zapLevel = zapcore.DebugLevel
	case "info":
		zapLevel = zapcore.InfoLevel
	case "warn":
		zapLevel = zapcore.WarnLevel
	case "error":
		zapLevel = zapcore.ErrorLevel
	default:
		zapLevel = zapcore.InfoLevel
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	var encoder zapcore.Encoder
	if jsonFormat {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(os.Stdout),
		zapLevel,
	)

	l := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	return logger.NewZapLogger(l)

}
