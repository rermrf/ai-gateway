package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"ai-gateway/internal/pkg/logger"
)

type HealthHandler struct {
	db     *gorm.DB
	redis  redis.Cmdable
	logger logger.Logger
}

func NewHealthHandler(db *gorm.DB, redis redis.Cmdable, l logger.Logger) *HealthHandler {
	return &HealthHandler{
		db:     db,
		redis:  redis,
		logger: l,
	}
}

// LivenessCheck godoc
// @Summary SVG 存活检查
// @Description 检查服务是否存活
// @Tags Health
// @Success 200 {object} map[string]string
// @Router /health/live [get]
func (h *HealthHandler) LivenessCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	})
}

// ReadinessCheck godoc
// @Summary 服务就绪检查
// @Description 检查服务及其依赖（DB, Redis）是否就绪
// @Tags Health
// @Success 200 {object} map[string]interface{}
// @Router /health/ready [get]
func (h *HealthHandler) ReadinessCheck(c *gin.Context) {
	status := gin.H{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
		"components": gin.H{
			"database": h.checkDB(c.Request.Context()),
			"redis":    h.checkRedis(c.Request.Context()),
		},
	}

	// 如果任何组件不健康，返回 503
	components := status["components"].(gin.H)
	if components["database"].(gin.H)["status"] != "ok" || components["redis"].(gin.H)["status"] != "ok" {
		status["status"] = "degraded"
		c.JSON(http.StatusServiceUnavailable, status)
		return
	}

	c.JSON(http.StatusOK, status)
}

func (h *HealthHandler) checkDB(ctx context.Context) gin.H {
	start := time.Now()
	sqlDB, err := h.db.DB()
	if err != nil {
		return gin.H{"status": "error", "error": err.Error()}
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		return gin.H{"status": "error", "error": err.Error()}
	}

	return gin.H{
		"status":     "ok",
		"latency_ms": time.Since(start).Milliseconds(),
	}
}

func (h *HealthHandler) checkRedis(ctx context.Context) gin.H {
	if h.redis == nil {
		return gin.H{"status": "disabled"}
	}

	start := time.Now()
	if err := h.redis.Ping(ctx).Err(); err != nil {
		return gin.H{"status": "error", "error": err.Error()}
	}

	return gin.H{
		"status":     "ok",
		"latency_ms": time.Since(start).Milliseconds(),
	}
}
