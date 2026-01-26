// AI 网关 - 主入口点
package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ai-gateway/cmd/server/ioc"
	"ai-gateway/config"
	"ai-gateway/internal/pkg/logger"
)

func main() {
	// 解析命令行标志
	configPath := flag.String("config", "", "配置文件路径")
	flag.Parse()

	// 加载配置
	var cfg *config.Config
	var err error

	if *configPath != "" {
		cfg, err = config.Load(*configPath)
		if err != nil {
			panic("failed to load config: " + err.Error())
		}
	} else {
		// 使用默认配置
		cfg = config.DefaultConfig()

		// 从环境变量配置默认 provider
		// 如果没有配置文件，可以通过环境变量快速启动
		if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
			baseURL := os.Getenv("OPENAI_BASE_URL")
			if baseURL == "" {
				baseURL = "https://api.openai.com/v1"
			}
			cfg.Providers = append(cfg.Providers, config.ProviderConfig{
				Name:    "openai",
				Type:    "openai",
				APIKey:  apiKey,
				BaseURL: baseURL,
				Default: true,
			})
		}

		if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
			baseURL := os.Getenv("ANTHROPIC_BASE_URL")
			if baseURL == "" {
				baseURL = "https://api.anthropic.com"
			}
			cfg.Providers = append(cfg.Providers, config.ProviderConfig{
				Name:    "anthropic",
				Type:    "anthropic",
				APIKey:  apiKey,
				BaseURL: baseURL,
				Default: true,
			})
		}

		// HTTP 服务器配置
		if addr := os.Getenv("HTTP_ADDR"); addr != "" {
			cfg.HTTP.Addr = addr
		}
	}

	app, err := ioc.InitApp(cfg)
	if err != nil {
		panic("failed to init app: " + err.Error())
	}

	l := app.Logger

	l.Info("starting ai-gateway",
		logger.String("addr", cfg.HTTP.Addr),
		logger.Int("providers_count", len(cfg.Providers)),
	)

	server := app.HTTPServer

	// 在协程中启动服务器
	go func() {
		if err := server.Start(cfg.HTTP.Addr); err != nil && err != http.ErrServerClosed {
			l.Error("http server failed", logger.Error(err))
			os.Exit(1)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	l.Info("shutting down server...")

	// 带有超时的优雅停机
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		l.Error("server shutdown failed", logger.Error(err))
	}

	l.Info("server exited")
}
