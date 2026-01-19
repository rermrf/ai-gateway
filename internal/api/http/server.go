// Package http 为 AI 网关提供 HTTP 服务器。
package http

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"ai-gateway/config"
	"ai-gateway/internal/api/http/handler"
	"ai-gateway/internal/api/http/middleware"
)

// Server 是 AI 网关的 HTTP 服务器。
type Server struct {
	engine *gin.Engine
	server *http.Server
	logger *zap.Logger
}

// NewServer 创建一个新的 HTTP 服务器。
func NewServer(
	openaiHandler *handler.OpenAIHandler,
	anthropicHandler *handler.AnthropicHandler,
	adminHandler *handler.AdminHandler,
	authCfg config.AuthConfig,
	logger *zap.Logger,
) *Server {
	gin.SetMode(gin.ReleaseMode)

	engine := gin.New()

	// 全局中间件
	engine.Use(
		middleware.Recovery(logger),
		middleware.Logger(logger),
		middleware.Cors(),
		middleware.RequestID(),
	)

	// 注册路由
	registerRoutes(engine, openaiHandler, anthropicHandler, adminHandler, authCfg)

	return &Server{
		engine: engine,
		logger: logger.Named("http.server"),
	}
}

func registerRoutes(
	engine *gin.Engine,
	openaiHandler *handler.OpenAIHandler,
	anthropicHandler *handler.AnthropicHandler,
	adminHandler *handler.AdminHandler,
	authCfg config.AuthConfig,
) {
	// 健康检查（不需要身份验证）
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// OpenAI 兼容 API
	v1 := engine.Group("/v1")
	v1.Use(middleware.Auth(authCfg))
	{
		v1.POST("/chat/completions", openaiHandler.ChatCompletions)
		v1.GET("/models", openaiHandler.ListModels)
	}

	// Anthropic 兼容 API（为简单起见使用相同的 /v1 前缀）
	v1.POST("/messages", anthropicHandler.Messages)

	// Admin API 路由组
	adminGroup := engine.Group("/api/admin")
	adminGroup.Use(middleware.AdminAuth())
	{
		// Provider 管理
		adminGroup.GET("/providers", adminHandler.ListProviders)
		adminGroup.POST("/providers", adminHandler.CreateProvider)
		adminGroup.GET("/providers/:id", adminHandler.GetProvider)
		adminGroup.PUT("/providers/:id", adminHandler.UpdateProvider)
		adminGroup.DELETE("/providers/:id", adminHandler.DeleteProvider)

		// 路由规则管理
		adminGroup.GET("/routing-rules", adminHandler.ListRoutingRules)
		adminGroup.POST("/routing-rules", adminHandler.CreateRoutingRule)
		adminGroup.PUT("/routing-rules/:id", adminHandler.UpdateRoutingRule)
		adminGroup.DELETE("/routing-rules/:id", adminHandler.DeleteRoutingRule)

		// 负载均衡管理
		adminGroup.GET("/load-balance-groups", adminHandler.ListLoadBalanceGroups)
		adminGroup.POST("/load-balance-groups", adminHandler.CreateLoadBalanceGroup)
		adminGroup.DELETE("/load-balance-groups/:id", adminHandler.DeleteLoadBalanceGroup)

		// API Key 管理
		adminGroup.GET("/api-keys", adminHandler.ListAPIKeys)
		adminGroup.POST("/api-keys", adminHandler.CreateAPIKey)
		adminGroup.DELETE("/api-keys/:id", adminHandler.DeleteAPIKey)

		// 仪表盘统计
		adminGroup.GET("/dashboard/stats", adminHandler.DashboardStats)
	}

	// 静态文件服务（生产模式下托管前端）
	engine.Static("/admin", "./web/admin/dist")
	
	// SPA fallback: 非 API 路径的 404 返回 index.html
	engine.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		// 如果是 /admin 开头的路径，返回 SPA 入口
		if len(path) >= 6 && path[:6] == "/admin" {
			c.File("./web/admin/dist/index.html")
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
	})
}

// Start 启动 HTTP 服务器。
func (s *Server) Start(addr string) error {
	s.server = &http.Server{
		Addr:         addr,
		Handler:      s.engine,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 120 * time.Second, // 针对流式传输的长超时时间
		IdleTimeout:  60 * time.Second,
	}

	s.logger.Info("starting http server", zap.String("addr", addr))
	return s.server.ListenAndServe()
}

// Shutdown 优雅地关闭服务器。
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("shutting down http server")
	return s.server.Shutdown(ctx)
}

// Engine returns the Gin engine (for testing).
func (s *Server) Engine() *gin.Engine {
	return s.engine
}
