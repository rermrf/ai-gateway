# Go 微服务项目结构模板

> 基于 notification-platform 项目总结的企业级 Go 微服务项目结构规范
>
> **技术栈：Gin + Zap + gRPC + GORM + Wire**

## 目录结构总览

```
project-name/
├── api/                          # API 定义层
│   └── proto/                    # Protocol Buffer 定义
│       ├── gen/                  # 生成的代码（自动生成，勿手动修改）
│       │   └── {service}/v1/     # 按服务和版本组织
│       │       ├── xxx.pb.go
│       │       ├── xxx_grpc.pb.go
│       │       └── xxx.pb.validate.go
│       └── {service}/v1/         # Proto 源文件
│           └── xxx.proto
│
├── cmd/                          # 应用入口
│   ├── {app-name}/               # 主应用
│   │   ├── main.go               # 程序入口
│   │   └── ioc/                  # Wire 依赖注入
│   │       ├── wire.go           # Wire 定义
│   │       └── wire_gen.go       # Wire 生成（自动生成）
│   └── admin/                    # 管理后台（可选）
│       └── main.go
│
├── config/                       # 配置文件
│   └── config.yaml               # 主配置文件
│
├── internal/                     # 内部代码（不对外暴露）
│   ├── api/                      # 接口层
│   │   ├── grpc/                 # gRPC 服务实现
│   │   │   ├── server.go         # gRPC 服务器
│   │   │   ├── interceptor/      # gRPC 拦截器
│   │   │   │   ├── jwt/          # JWT 认证
│   │   │   │   ├── log/          # 日志
│   │   │   │   ├── metrics/      # 指标
│   │   │   │   ├── tracing/      # 链路追踪
│   │   │   │   ├── limit/        # 限流
│   │   │   │   ├── timeout/      # 超时
│   │   │   │   ├── circuitbreaker/ # 熔断
│   │   │   │   ├── degrade/      # 降级
│   │   │   │   └── idempotent/   # 幂等
│   │   │   └── integration/      # 集成测试
│   │   └── http/                 # HTTP 服务实现 (Gin)
│   │       ├── server.go         # Gin 服务器
│   │       ├── router.go         # 路由定义
│   │       ├── handler/          # HTTP 处理器
│   │       │   └── {module}.go
│   │       └── middleware/       # Gin 中间件
│   │           ├── jwt.go
│   │           ├── log.go
│   │           ├── recovery.go
│   │           ├── cors.go
│   │           └── ratelimit.go
│   │
│   ├── domain/                   # 领域层（核心业务模型）
│   │   ├── {entity}.go           # 领域实体
│   │   └── {value_object}.go     # 值对象
│   │
│   ├── service/                  # 业务逻辑层
│   │   └── {module}/             # 按业务模块组织
│   │       ├── {module}.go       # 服务接口 + 实现
│   │       ├── types.go          # 类型定义（可选）
│   │       └── mocks/            # Mock 文件（自动生成）
│   │           └── {module}.mock.go
│   │
│   ├── repository/               # 数据仓储层
│   │   ├── {entity}.go           # Repository 接口 + 实现
│   │   ├── dao/                  # 数据访问对象
│   │   │   ├── {entity}.go       # DAO 接口 + 实现 + 数据库模型
│   │   │   └── sharding/         # 分库分表实现（可选）
│   │   └── cache/                # 缓存层
│   │       ├── {entity}.go       # 缓存接口 + 实现
│   │       ├── redis/            # Redis 缓存实现
│   │       └── local/            # 本地缓存实现
│   │
│   ├── event/                    # 事件驱动
│   │   └── {event-name}/         # 按事件类型组织
│   │       ├── event.go          # 事件定义
│   │       ├── producer.go       # 生产者
│   │       └── consumer.go       # 消费者
│   │
│   ├── pkg/                      # 内部通用包
│   │   ├── logger/               # Zap 日志封装
│   │   │   └── logger.go
│   │   ├── {util}/               # 工具包
│   │   │   ├── {util}.go
│   │   │   └── {util}_test.go
│   │   └── ...
│   │
│   ├── ioc/                      # 依赖注入配置
│   │   ├── db.go                 # 数据库初始化
│   │   ├── redis.go              # Redis 初始化
│   │   ├── logger.go             # Zap 日志初始化
│   │   ├── grpc.go               # gRPC 初始化
│   │   ├── gin.go                # Gin 初始化
│   │   ├── cron.go               # 定时任务初始化
│   │   └── ...
│   │
│   ├── errs/                     # 错误定义
│   │   └── error.go              # 所有业务错误
│   │
│   └── test/                     # 测试相关
│       ├── integration/          # 集成测试
│       └── ioc/                  # 测试用 IOC
│
├── scripts/                      # 脚本文件
│   ├── setup.sh                  # 环境初始化
│   ├── mysql/                    # 数据库脚本
│   ├── lint/                     # Lint 配置
│   └── ...
│
├── docs/                         # 文档
│   └── ...
│
├── go.mod                        # Go 模块定义
├── go.sum                        # 依赖锁定
├── Makefile                      # 构建命令
├── .golangci.yaml                # Lint 配置
└── README.md                     # 项目说明
```

---

## 分层架构规范

### 层次依赖关系

```
┌─────────────────────────────────────────────────────────────┐
│                        API Layer                             │
│         (grpc/server.go, http/server.go + handler/)         │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                      Service Layer                           │
│                  (service/{module}/*.go)                    │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    Repository Layer                          │
│                  (repository/{entity}.go)                   │
└─────────────────────────────────────────────────────────────┘
                              │
              ┌───────────────┴───────────────┐
              ▼                               ▼
┌─────────────────────────┐     ┌─────────────────────────────┐
│       DAO Layer         │     │        Cache Layer          │
│ (repository/dao/*.go)   │     │   (repository/cache/*.go)   │
└─────────────────────────┘     └─────────────────────────────┘
              │                               │
              ▼                               ▼
        ┌──────────┐                  ┌──────────────┐
        │  MySQL   │                  │ Redis/Local  │
        └──────────┘                  └──────────────┘

┌─────────────────────────────────────────────────────────────┐
│                      Domain Layer                            │
│                     (domain/*.go)                           │
│          被所有层引用，不依赖任何其他层                        │
└─────────────────────────────────────────────────────────────┘
```

### 依赖规则

1. **API 层** → 依赖 Service 层
2. **Service 层** → 依赖 Repository 层 + Domain 层
3. **Repository 层** → 依赖 DAO 层 + Cache 层 + Domain 层
4. **DAO 层** → 依赖 Domain 层（仅类型引用）
5. **Domain 层** → 不依赖任何层（纯业务模型）

---

## 代码规范

### 1. Zap 日志初始化

```go
// 文件: internal/pkg/logger/logger.go

package logger

import (
    "os"

    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

var Logger *zap.Logger

// InitLogger 初始化 Zap 日志
func InitLogger(level string, jsonFormat bool) *zap.Logger {
    var zapLevel zapcore.Level
    switch level {
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

    Logger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
    return Logger
}

// L 获取全局 Logger
func L() *zap.Logger {
    if Logger == nil {
        Logger, _ = zap.NewDevelopment()
    }
    return Logger
}

// S 获取 SugaredLogger
func S() *zap.SugaredLogger {
    return L().Sugar()
}
```

```go
// 文件: internal/ioc/logger.go

package ioc

import (
    "your-project/internal/pkg/logger"

    "go.uber.org/zap"
)

// InitLogger 初始化日志（供 Wire 使用）
func InitLogger() *zap.Logger {
    return logger.InitLogger("info", true)
}
```

### 2. 接口定义规范

```go
// 文件: internal/service/{module}/{module}.go

package {module}

import (
    "context"

    "go.uber.org/zap"

    "your-project/internal/domain"
    "your-project/internal/repository"
)

// Service 服务接口定义
// 使用 mockgen 生成 mock
//
//go:generate mockgen -source=./{module}.go -destination=./mocks/{module}.mock.go -package={module}mocks -typed Service
type Service interface {
    // Create 创建资源
    Create(ctx context.Context, entity domain.Entity) (domain.Entity, error)
    // GetByID 根据ID获取
    GetByID(ctx context.Context, id uint64) (domain.Entity, error)
    // List 列表查询
    List(ctx context.Context, offset, limit int) ([]domain.Entity, error)
    // Update 更新资源
    Update(ctx context.Context, entity domain.Entity) error
    // Delete 删除资源
    Delete(ctx context.Context, id uint64) error
}

// service 服务实现（小写，不导出）
type service struct {
    repo   repository.Repository
    logger *zap.Logger
}

// NewService 创建服务实例（工厂函数）
func NewService(repo repository.Repository, logger *zap.Logger) Service {
    return &service{
        repo:   repo,
        logger: logger.Named("service.{module}"),
    }
}

// Create 创建资源
func (s *service) Create(ctx context.Context, entity domain.Entity) (domain.Entity, error) {
    s.logger.Info("creating entity",
        zap.String("name", entity.Name),
    )

    result, err := s.repo.Create(ctx, entity)
    if err != nil {
        s.logger.Error("failed to create entity",
            zap.Error(err),
            zap.String("name", entity.Name),
        )
        return domain.Entity{}, err
    }

    s.logger.Info("entity created",
        zap.Uint64("id", result.ID),
    )
    return result, nil
}

// GetByID 根据ID获取
func (s *service) GetByID(ctx context.Context, id uint64) (domain.Entity, error) {
    s.logger.Debug("getting entity by id",
        zap.Uint64("id", id),
    )
    return s.repo.GetByID(ctx, id)
}
```

### 3. Gin HTTP 服务器规范

```go
// 文件: internal/api/http/server.go

package http

import (
    "context"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "go.uber.org/zap"

    "your-project/internal/api/http/handler"
    "your-project/internal/api/http/middleware"
)

// Server HTTP 服务器
type Server struct {
    engine *gin.Engine
    server *http.Server
    logger *zap.Logger
}

// NewServer 创建 HTTP 服务器
func NewServer(
    entityHandler *handler.EntityHandler,
    logger *zap.Logger,
) *Server {
    // 设置 Gin 模式
    gin.SetMode(gin.ReleaseMode)

    engine := gin.New()

    // 注册全局中间件
    engine.Use(
        middleware.Recovery(logger),
        middleware.Logger(logger),
        middleware.Cors(),
        middleware.RequestID(),
    )

    // 注册路由
    registerRoutes(engine, entityHandler, logger)

    return &Server{
        engine: engine,
        logger: logger.Named("http.server"),
    }
}

// registerRoutes 注册路由
func registerRoutes(
    engine *gin.Engine,
    entityHandler *handler.EntityHandler,
    logger *zap.Logger,
) {
    // 健康检查
    engine.GET("/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"status": "ok"})
    })

    // API v1
    v1 := engine.Group("/api/v1")
    {
        // 需要认证的路由
        authGroup := v1.Group("")
        authGroup.Use(middleware.JWTAuth(logger))
        {
            // Entity 相关路由
            entities := authGroup.Group("/entities")
            {
                entities.POST("", entityHandler.Create)
                entities.GET("/:id", entityHandler.GetByID)
                entities.GET("", entityHandler.List)
                entities.PUT("/:id", entityHandler.Update)
                entities.DELETE("/:id", entityHandler.Delete)
            }
        }
    }
}

// Start 启动服务器
func (s *Server) Start(addr string) error {
    s.server = &http.Server{
        Addr:         addr,
        Handler:      s.engine,
        ReadTimeout:  10 * time.Second,
        WriteTimeout: 10 * time.Second,
    }

    s.logger.Info("starting http server",
        zap.String("addr", addr),
    )

    return s.server.ListenAndServe()
}

// Shutdown 优雅关闭
func (s *Server) Shutdown(ctx context.Context) error {
    s.logger.Info("shutting down http server")
    return s.server.Shutdown(ctx)
}

// Engine 获取 Gin 引擎（用于测试）
func (s *Server) Engine() *gin.Engine {
    return s.engine
}
```

### 4. Gin Handler 规范

```go
// 文件: internal/api/http/handler/entity.go

package handler

import (
    "net/http"
    "strconv"

    "github.com/gin-gonic/gin"
    "go.uber.org/zap"

    "your-project/internal/domain"
    "your-project/internal/errs"
    "your-project/internal/service/entity"
)

// EntityHandler 实体处理器
type EntityHandler struct {
    svc    entity.Service
    logger *zap.Logger
}

// NewEntityHandler 创建实体处理器
func NewEntityHandler(svc entity.Service, logger *zap.Logger) *EntityHandler {
    return &EntityHandler{
        svc:    svc,
        logger: logger.Named("handler.entity"),
    }
}

// CreateRequest 创建请求
type CreateRequest struct {
    Name   string `json:"name" binding:"required,min=1,max=256"`
    Status string `json:"status" binding:"omitempty,oneof=ACTIVE INACTIVE"`
}

// EntityResponse 实体响应
type EntityResponse struct {
    ID        uint64 `json:"id"`
    Name      string `json:"name"`
    Status    string `json:"status"`
    CreatedAt int64  `json:"createdAt"`
    UpdatedAt int64  `json:"updatedAt"`
}

// Create 创建实体
// @Summary 创建实体
// @Tags Entity
// @Accept json
// @Produce json
// @Param request body CreateRequest true "创建请求"
// @Success 200 {object} EntityResponse
// @Router /api/v1/entities [post]
func (h *EntityHandler) Create(c *gin.Context) {
    var req CreateRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        h.logger.Warn("invalid request",
            zap.Error(err),
        )
        c.JSON(http.StatusBadRequest, gin.H{
            "code":    400,
            "message": "参数错误: " + err.Error(),
        })
        return
    }

    entity := domain.Entity{
        Name:   req.Name,
        Status: domain.EntityStatus(req.Status),
    }

    result, err := h.svc.Create(c.Request.Context(), entity)
    if err != nil {
        h.handleError(c, err)
        return
    }

    c.JSON(http.StatusOK, h.toResponse(result))
}

// GetByID 根据ID获取
// @Summary 根据ID获取实体
// @Tags Entity
// @Produce json
// @Param id path int true "实体ID"
// @Success 200 {object} EntityResponse
// @Router /api/v1/entities/{id} [get]
func (h *EntityHandler) GetByID(c *gin.Context) {
    idStr := c.Param("id")
    id, err := strconv.ParseUint(idStr, 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "code":    400,
            "message": "无效的ID",
        })
        return
    }

    result, err := h.svc.GetByID(c.Request.Context(), id)
    if err != nil {
        h.handleError(c, err)
        return
    }

    c.JSON(http.StatusOK, h.toResponse(result))
}

// List 列表查询
// @Summary 列表查询
// @Tags Entity
// @Produce json
// @Param offset query int false "偏移量" default(0)
// @Param limit query int false "限制数" default(20)
// @Success 200 {array} EntityResponse
// @Router /api/v1/entities [get]
func (h *EntityHandler) List(c *gin.Context) {
    offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

    if limit > 100 {
        limit = 100
    }

    results, err := h.svc.List(c.Request.Context(), offset, limit)
    if err != nil {
        h.handleError(c, err)
        return
    }

    responses := make([]EntityResponse, len(results))
    for i, r := range results {
        responses[i] = h.toResponse(r)
    }

    c.JSON(http.StatusOK, gin.H{
        "data":   responses,
        "total":  len(responses),
        "offset": offset,
        "limit":  limit,
    })
}

// Update 更新实体
func (h *EntityHandler) Update(c *gin.Context) {
    idStr := c.Param("id")
    id, err := strconv.ParseUint(idStr, 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "code":    400,
            "message": "无效的ID",
        })
        return
    }

    var req CreateRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "code":    400,
            "message": "参数错误: " + err.Error(),
        })
        return
    }

    entity := domain.Entity{
        ID:     id,
        Name:   req.Name,
        Status: domain.EntityStatus(req.Status),
    }

    if err := h.svc.Update(c.Request.Context(), entity); err != nil {
        h.handleError(c, err)
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "code":    0,
        "message": "更新成功",
    })
}

// Delete 删除实体
func (h *EntityHandler) Delete(c *gin.Context) {
    idStr := c.Param("id")
    id, err := strconv.ParseUint(idStr, 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "code":    400,
            "message": "无效的ID",
        })
        return
    }

    if err := h.svc.Delete(c.Request.Context(), id); err != nil {
        h.handleError(c, err)
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "code":    0,
        "message": "删除成功",
    })
}

// toResponse 转换为响应
func (h *EntityHandler) toResponse(entity domain.Entity) EntityResponse {
    return EntityResponse{
        ID:        entity.ID,
        Name:      entity.Name,
        Status:    entity.Status.String(),
        CreatedAt: entity.CreatedAt.UnixMilli(),
        UpdatedAt: entity.UpdatedAt.UnixMilli(),
    }
}

// handleError 统一错误处理
func (h *EntityHandler) handleError(c *gin.Context, err error) {
    h.logger.Error("request failed",
        zap.Error(err),
        zap.String("path", c.Request.URL.Path),
    )

    switch {
    case errors.Is(err, errs.ErrEntityNotFound):
        c.JSON(http.StatusNotFound, gin.H{
            "code":    404,
            "message": err.Error(),
        })
    case errors.Is(err, errs.ErrInvalidParameter):
        c.JSON(http.StatusBadRequest, gin.H{
            "code":    400,
            "message": err.Error(),
        })
    case errors.Is(err, errs.ErrEntityDuplicate):
        c.JSON(http.StatusConflict, gin.H{
            "code":    409,
            "message": err.Error(),
        })
    default:
        c.JSON(http.StatusInternalServerError, gin.H{
            "code":    500,
            "message": "服务器内部错误",
        })
    }
}
```

### 5. Gin 中间件规范

```go
// 文件: internal/api/http/middleware/logger.go

package middleware

import (
    "time"

    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
)

// Logger 日志中间件
func Logger(logger *zap.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        path := c.Request.URL.Path
        query := c.Request.URL.RawQuery

        c.Next()

        latency := time.Since(start)
        statusCode := c.Writer.Status()

        fields := []zap.Field{
            zap.Int("status", statusCode),
            zap.String("method", c.Request.Method),
            zap.String("path", path),
            zap.String("query", query),
            zap.String("ip", c.ClientIP()),
            zap.Duration("latency", latency),
            zap.String("user-agent", c.Request.UserAgent()),
        }

        if requestID := c.GetString("request_id"); requestID != "" {
            fields = append(fields, zap.String("request_id", requestID))
        }

        if len(c.Errors) > 0 {
            fields = append(fields, zap.String("errors", c.Errors.String()))
        }

        if statusCode >= 500 {
            logger.Error("server error", fields...)
        } else if statusCode >= 400 {
            logger.Warn("client error", fields...)
        } else {
            logger.Info("request completed", fields...)
        }
    }
}
```

```go
// 文件: internal/api/http/middleware/recovery.go

package middleware

import (
    "net/http"
    "runtime/debug"

    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
)

// Recovery 恢复中间件
func Recovery(logger *zap.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        defer func() {
            if err := recover(); err != nil {
                logger.Error("panic recovered",
                    zap.Any("error", err),
                    zap.String("stack", string(debug.Stack())),
                    zap.String("path", c.Request.URL.Path),
                    zap.String("method", c.Request.Method),
                )

                c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
                    "code":    500,
                    "message": "服务器内部错误",
                })
            }
        }()
        c.Next()
    }
}
```

```go
// 文件: internal/api/http/middleware/jwt.go

package middleware

import (
    "net/http"
    "strings"

    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v5"
    "go.uber.org/zap"
)

// JWTAuth JWT 认证中间件
func JWTAuth(logger *zap.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
                "code":    401,
                "message": "缺少认证信息",
            })
            return
        }

        parts := strings.SplitN(authHeader, " ", 2)
        if len(parts) != 2 || parts[0] != "Bearer" {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
                "code":    401,
                "message": "认证格式错误",
            })
            return
        }

        tokenString := parts[1]

        // 解析 Token（这里使用简化版本，实际应验证签名）
        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            // 这里返回用于验证的密钥
            return []byte("your-secret-key"), nil
        })

        if err != nil || !token.Valid {
            logger.Warn("invalid token",
                zap.Error(err),
            )
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
                "code":    401,
                "message": "无效的认证信息",
            })
            return
        }

        // 将用户信息存入上下文
        if claims, ok := token.Claims.(jwt.MapClaims); ok {
            c.Set("user_id", claims["sub"])
            c.Set("biz_id", claims["biz_id"])
        }

        c.Next()
    }
}
```

```go
// 文件: internal/api/http/middleware/cors.go

package middleware

import (
    "github.com/gin-gonic/gin"
)

// Cors 跨域中间件
func Cors() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Header("Access-Control-Allow-Origin", "*")
        c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, X-Request-ID")
        c.Header("Access-Control-Expose-Headers", "Content-Length, X-Request-ID")
        c.Header("Access-Control-Max-Age", "86400")

        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }

        c.Next()
    }
}
```

```go
// 文件: internal/api/http/middleware/request_id.go

package middleware

import (
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
)

// RequestID 请求ID中间件
func RequestID() gin.HandlerFunc {
    return func(c *gin.Context) {
        requestID := c.GetHeader("X-Request-ID")
        if requestID == "" {
            requestID = uuid.New().String()
        }

        c.Set("request_id", requestID)
        c.Header("X-Request-ID", requestID)

        c.Next()
    }
}
```

### 6. Repository 层规范

```go
// 文件: internal/repository/{entity}.go

package repository

import (
    "context"

    "go.uber.org/zap"

    "your-project/internal/domain"
    "your-project/internal/repository/dao"
    "your-project/internal/repository/cache"
)

// EntityRepository 仓储接口
type EntityRepository interface {
    Create(ctx context.Context, entity domain.Entity) (domain.Entity, error)
    GetByID(ctx context.Context, id uint64) (domain.Entity, error)
    // ... 其他方法
}

// entityRepository 仓储实现
type entityRepository struct {
    dao    dao.EntityDAO
    cache  cache.EntityCache
    logger *zap.Logger
}

// NewEntityRepository 创建仓储实例
func NewEntityRepository(d dao.EntityDAO, c cache.EntityCache, logger *zap.Logger) EntityRepository {
    return &entityRepository{
        dao:    d,
        cache:  c,
        logger: logger.Named("repository.entity"),
    }
}

// toEntity 将 domain 对象转换为 DAO 实体
func (r *entityRepository) toEntity(entity domain.Entity) dao.Entity {
    return dao.Entity{
        ID:    entity.ID,
        Name:  entity.Name,
        // ... 字段映射
    }
}

// toDomain 将 DAO 实体转换为 domain 对象
func (r *entityRepository) toDomain(entity dao.Entity) domain.Entity {
    return domain.Entity{
        ID:   entity.ID,
        Name: entity.Name,
        // ... 字段映射
    }
}

// Create 创建实体
func (r *entityRepository) Create(ctx context.Context, entity domain.Entity) (domain.Entity, error) {
    daoEntity := r.toEntity(entity)
    result, err := r.dao.Create(ctx, daoEntity)
    if err != nil {
        r.logger.Error("failed to create entity",
            zap.Error(err),
        )
        return domain.Entity{}, err
    }
    return r.toDomain(result), nil
}
```

### 7. DAO 层规范

```go
// 文件: internal/repository/dao/{entity}.go

package dao

import (
    "context"
    "errors"
    "time"

    "go.uber.org/zap"
    "gorm.io/gorm"

    "your-project/internal/errs"
)

// EntityDAO 数据访问接口
type EntityDAO interface {
    Create(ctx context.Context, entity Entity) (Entity, error)
    GetByID(ctx context.Context, id uint64) (Entity, error)
    // ... 其他方法
}

// Entity 数据库模型（与表结构对应）
type Entity struct {
    ID      uint64 `gorm:"primaryKey;comment:'主键ID'"`
    Name    string `gorm:"type:VARCHAR(256);NOT NULL;comment:'名称'"`
    Status  string `gorm:"type:ENUM('ACTIVE','INACTIVE');DEFAULT:'ACTIVE';comment:'状态'"`
    Version int    `gorm:"type:INT;NOT NULL;DEFAULT:1;comment:'版本号'"`
    Ctime   int64  `gorm:"comment:'创建时间'"`
    Utime   int64  `gorm:"comment:'更新时间'"`
}

// TableName 指定表名
func (Entity) TableName() string {
    return "entity"
}

// entityDAO DAO 实现
type entityDAO struct {
    db     *gorm.DB
    logger *zap.Logger
}

// NewEntityDAO 创建 DAO 实例
func NewEntityDAO(db *gorm.DB, logger *zap.Logger) EntityDAO {
    return &entityDAO{
        db:     db,
        logger: logger.Named("dao.entity"),
    }
}

// Create 创建记录
func (d *entityDAO) Create(ctx context.Context, entity Entity) (Entity, error) {
    now := time.Now().UnixMilli()
    entity.Ctime, entity.Utime = now, now
    entity.Version = 1

    err := d.db.WithContext(ctx).Create(&entity).Error
    if err != nil {
        d.logger.Error("failed to create entity",
            zap.Error(err),
        )
    }
    return entity, err
}

// GetByID 根据ID获取
func (d *entityDAO) GetByID(ctx context.Context, id uint64) (Entity, error) {
    var entity Entity
    err := d.db.WithContext(ctx).First(&entity, id).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return Entity{}, errs.ErrEntityNotFound
        }
        d.logger.Error("failed to get entity",
            zap.Error(err),
            zap.Uint64("id", id),
        )
        return Entity{}, err
    }
    return entity, nil
}
```

### 8. Domain 层规范

```go
// 文件: internal/domain/{entity}.go

package domain

import (
    "fmt"
    "time"

    "your-project/internal/errs"
)

// EntityStatus 实体状态（枚举）
type EntityStatus string

const (
    EntityStatusActive   EntityStatus = "ACTIVE"
    EntityStatusInactive EntityStatus = "INACTIVE"
)

func (s EntityStatus) String() string {
    return string(s)
}

// Entity 领域实体
type Entity struct {
    ID        uint64       `json:"id"`
    Name      string       `json:"name"`
    Status    EntityStatus `json:"status"`
    Version   int          `json:"version"`
    CreatedAt time.Time    `json:"createdAt"`
    UpdatedAt time.Time    `json:"updatedAt"`
}

// Validate 验证实体
func (e *Entity) Validate() error {
    if e.Name == "" {
        return fmt.Errorf("%w: Name不能为空", errs.ErrInvalidParameter)
    }
    return nil
}

// IsActive 判断是否激活
func (e *Entity) IsActive() bool {
    return e.Status == EntityStatusActive
}
```

### 9. 错误定义规范

```go
// 文件: internal/errs/error.go

package errs

import "errors"

// 通用错误
var (
    ErrInvalidParameter = errors.New("无效参数")
    ErrDatabaseError    = errors.New("数据库错误")
)

// 实体相关错误
var (
    ErrEntityNotFound     = errors.New("实体不存在")
    ErrEntityDuplicate    = errors.New("实体已存在")
    ErrVersionMismatch    = errors.New("版本号不匹配")
)

// 业务相关错误
var (
    ErrNoQuota          = errors.New("配额不足")
    ErrRateLimited      = errors.New("请求过于频繁")
    ErrChannelDisabled  = errors.New("渠道已禁用")
)
```

### 10. gRPC Server 规范

```go
// 文件: internal/api/grpc/server.go

package grpc

import (
    "context"
    "errors"

    "go.uber.org/zap"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"

    pb "your-project/api/proto/gen/service/v1"
    "your-project/internal/domain"
    "your-project/internal/errs"
    "your-project/internal/service"
)

// Server gRPC 服务器
type Server struct {
    pb.UnimplementedServiceServer

    svc    service.Service
    logger *zap.Logger
}

// NewServer 创建 gRPC 服务器
func NewServer(svc service.Service, logger *zap.Logger) *Server {
    return &Server{
        svc:    svc,
        logger: logger.Named("grpc.server"),
    }
}

// CreateEntity 创建实体
func (s *Server) CreateEntity(ctx context.Context, req *pb.CreateEntityRequest) (*pb.CreateEntityResponse, error) {
    // 1. 参数验证
    if req == nil {
        return nil, status.Errorf(codes.InvalidArgument, "请求不能为空")
    }

    s.logger.Info("creating entity",
        zap.String("name", req.Name),
    )

    // 2. 转换为领域对象
    entity := s.toDomain(req)

    // 3. 调用服务
    result, err := s.svc.Create(ctx, entity)
    if err != nil {
        s.logger.Error("failed to create entity",
            zap.Error(err),
        )
        // 4. 错误处理
        return nil, s.handleError(err)
    }

    s.logger.Info("entity created",
        zap.Uint64("id", result.ID),
    )

    // 5. 返回响应
    return s.toResponse(result), nil
}

// toDomain 转换为领域对象
func (s *Server) toDomain(req *pb.CreateEntityRequest) domain.Entity {
    return domain.Entity{
        Name: req.Name,
    }
}

// toResponse 转换为响应
func (s *Server) toResponse(entity domain.Entity) *pb.CreateEntityResponse {
    return &pb.CreateEntityResponse{
        Id:   entity.ID,
        Name: entity.Name,
    }
}

// handleError 错误处理
func (s *Server) handleError(err error) error {
    // 根据错误类型返回不同的 gRPC status
    switch {
    case errors.Is(err, errs.ErrInvalidParameter):
        return status.Errorf(codes.InvalidArgument, "%v", err)
    case errors.Is(err, errs.ErrEntityNotFound):
        return status.Errorf(codes.NotFound, "%v", err)
    default:
        return status.Errorf(codes.Internal, "%v", err)
    }
}
```

### 11. IOC 依赖注入规范

```go
// 文件: internal/ioc/db.go

package ioc

import (
    "time"

    "go.uber.org/zap"
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
    gormlogger "gorm.io/gorm/logger"
)

// InitDB 初始化数据库
func InitDB(logger *zap.Logger) *gorm.DB {
    dsn := "root:password@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"

    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
        Logger: gormlogger.Default.LogMode(gormlogger.Info),
    })
    if err != nil {
        logger.Fatal("failed to connect database",
            zap.Error(err),
        )
    }

    sqlDB, err := db.DB()
    if err != nil {
        logger.Fatal("failed to get sql.DB",
            zap.Error(err),
        )
    }

    // 设置连接池
    sqlDB.SetMaxIdleConns(10)
    sqlDB.SetMaxOpenConns(100)
    sqlDB.SetConnMaxLifetime(time.Hour)

    logger.Info("database connected")

    return db
}
```

```go
// 文件: internal/ioc/gin.go

package ioc

import (
    "go.uber.org/zap"

    httpapi "your-project/internal/api/http"
    "your-project/internal/api/http/handler"
)

// InitGinServer 初始化 Gin 服务器
func InitGinServer(
    entityHandler *handler.EntityHandler,
    logger *zap.Logger,
) *httpapi.Server {
    return httpapi.NewServer(entityHandler, logger)
}
```

```go
// 文件: cmd/{app}/ioc/wire.go

//go:build wireinject

package ioc

import (
    "github.com/google/wire"

    httpapi "your-project/internal/api/http"
    "your-project/internal/api/http/handler"
    grpcapi "your-project/internal/api/grpc"
    "your-project/internal/ioc"
    "your-project/internal/repository"
    "your-project/internal/repository/dao"
    "your-project/internal/service/entity"
)

// App 应用
type App struct {
    GinServer  *httpapi.Server
    GrpcServer *grpcapi.Server
}

// InitApp 初始化应用
func InitApp() *App {
    wire.Build(
        // 基础设施
        ioc.InitLogger,
        ioc.InitDB,
        ioc.InitRedis,

        // DAO 层
        dao.NewEntityDAO,

        // Repository 层
        repository.NewEntityRepository,

        // Service 层
        entity.NewService,

        // Handler 层
        handler.NewEntityHandler,

        // API 层
        ioc.InitGinServer,
        grpcapi.NewServer,

        // App
        wire.Struct(new(App), "*"),
    )
    return nil
}
```

### 12. Main 入口规范

```go
// 文件: cmd/{app}/main.go

package main

import (
    "context"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "go.uber.org/zap"

    "your-project/cmd/{app}/ioc"
)

func main() {
    // 初始化应用
    app := ioc.InitApp()

    // 获取 logger
    logger := app.Logger

    // 启动 HTTP 服务器
    go func() {
        logger.Info("starting http server on :8080")
        if err := app.GinServer.Start(":8080"); err != nil && err != http.ErrServerClosed {
            logger.Fatal("http server failed",
                zap.Error(err),
            )
        }
    }()

    // 启动 gRPC 服务器（可选）
    go func() {
        logger.Info("starting grpc server on :9002")
        if err := app.GrpcServer.Start(":9002"); err != nil {
            logger.Fatal("grpc server failed",
                zap.Error(err),
            )
        }
    }()

    // 优雅关闭
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    logger.Info("shutting down servers...")

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    if err := app.GinServer.Shutdown(ctx); err != nil {
        logger.Error("http server shutdown failed",
            zap.Error(err),
        )
    }

    logger.Info("servers exited")
}
```

---

## 命名规范

### 文件命名

| 类型 | 命名规则 | 示例 |
|------|---------|------|
| 接口+实现 | `{name}.go` | `notification.go` |
| 测试文件 | `{name}_test.go` | `notification_test.go` |
| Mock 文件 | `{name}.mock.go` | `notification.mock.go` |
| 类型定义 | `types.go` | `types.go` |
| 常量定义 | `const.go` | `const.go` |

### 包命名

- 使用小写字母
- 避免使用下划线
- 简短且有意义

### 变量/函数命名

| 类型 | 规则 | 示例 |
|------|------|------|
| 导出函数 | 大驼峰 | `NewService` |
| 私有函数 | 小驼峰 | `toEntity` |
| 接口 | 大驼峰，通常以 er 结尾 | `Service`, `Repository`, `Sender` |
| 实现结构体 | 小写，不导出 | `service`, `repository` |
| 常量 | 大驼峰 | `DefaultTimeout` |
| 错误变量 | Err 开头 | `ErrNotFound` |

---

## Makefile 规范

```makefile
# 初始化项目环境
.PHONY: setup
setup:
	@sh ./scripts/setup.sh

# 格式化代码
.PHONY: fmt
fmt:
	@goimports -l -w $$(find . -type f -name '*.go' -not -path "./.idea/*")
	@gofumpt -l -w $$(find . -type f -name '*.go' -not -path "./.idea/*")

# 清理依赖
.PHONY: tidy
tidy:
	@go mod tidy -v

# 代码检查
.PHONY: lint
lint:
	@golangci-lint run -c ./scripts/lint/.golangci.yaml ./...

# 单元测试
.PHONY: ut
ut:
	@go test -race -shuffle=on -short -failfast -tags=unit ./...

# 集成测试
.PHONY: e2e
e2e:
	@docker compose -f scripts/test_docker_compose.yml up -d
	@go test -race -shuffle=on -failfast -tags=e2e ./...
	@docker compose -f scripts/test_docker_compose.yml down -v

# 生成 gRPC 代码
.PHONY: grpc
grpc:
	@buf format -w api/proto
	@buf lint api/proto
	@buf generate api/proto

# 生成 Go 代码 (mock, wire 等)
.PHONY: gen
gen:
	@go generate ./...

# 运行服务
.PHONY: run
run:
	@go run cmd/{app}/main.go --config=./config/config.yaml

# 构建
.PHONY: build
build:
	@go build -o bin/{app} cmd/{app}/main.go
```

---

## 配置文件规范

```yaml
# config/config.yaml

# 应用配置
app:
  name: "your-service"
  env: "development"  # development, staging, production

# 日志配置
log:
  level: "info"       # debug, info, warn, error
  format: "json"      # json, console

# MySQL 配置
mysql:
  dsn: "root:password@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
  maxIdleConns: 10
  maxOpenConns: 100
  connMaxLifetime: 3600

# Redis 配置
redis:
  addr: "localhost:6379"
  password: ""
  db: 0

# HTTP 服务器配置
http:
  addr: ":8080"
  readTimeout: 10
  writeTimeout: 10

# gRPC 服务器配置
grpc:
  addr: ":9002"

# JWT 配置
jwt:
  key: "your-secret-key"
  expireHours: 24

# 链路追踪
trace:
  enabled: true
  endpoint: "http://localhost:14268/api/traces"
  serviceName: "your-service"

# 定时任务
cron:
  someTask:
    spec: "0 0 * * *"  # 每天零点执行
```

---

## 快速开始

### 1. 创建新项目

```bash
mkdir my-service && cd my-service
go mod init github.com/your-org/my-service
```

### 2. 安装依赖

```bash
go get github.com/gin-gonic/gin
go get go.uber.org/zap
go get gorm.io/gorm
go get gorm.io/driver/mysql
go get github.com/google/wire/cmd/wire
go get google.golang.org/grpc
```

### 3. 创建目录结构

```bash
mkdir -p api/proto/{gen,service/v1}
mkdir -p cmd/{app}/ioc
mkdir -p config
mkdir -p internal/{api/{grpc,http/{handler,middleware}},domain,service,repository/{dao,cache},event,pkg/logger,ioc,errs,test}
mkdir -p scripts/{mysql,lint}
mkdir -p docs
```

### 4. 开发流程

```bash
# 1. 初始化日志
vim internal/pkg/logger/logger.go

# 2. 定义 Proto
vim api/proto/service/v1/service.proto

# 3. 生成 gRPC 代码
make grpc

# 4. 定义领域模型
vim internal/domain/entity.go

# 5. 实现 DAO 层
vim internal/repository/dao/entity.go

# 6. 实现 Repository 层
vim internal/repository/entity.go

# 7. 实现 Service 层
vim internal/service/entity/entity.go

# 8. 实现 Gin Handler
vim internal/api/http/handler/entity.go

# 9. 实现 Gin Server
vim internal/api/http/server.go

# 10. 实现 gRPC Server（可选）
vim internal/api/grpc/server.go

# 11. 配置依赖注入
vim cmd/{app}/ioc/wire.go

# 12. 生成 Wire 代码
make gen

# 13. 运行测试
make ut

# 14. 启动服务
make run
```

---

## 核心技术栈

| 类别 | 技术选型 |
|------|---------|
| Web 框架 | **Gin** |
| 日志 | **Zap** |
| RPC 框架 | gRPC |
| ORM | GORM |
| 缓存 | Redis |
| 消息队列 | Kafka |
| 配置管理 | Viper (可选) |
| 依赖注入 | Wire |
| 链路追踪 | OpenTelemetry + Jaeger |
| 监控指标 | Prometheus |
| 代码检查 | golangci-lint |
| Mock | mockgen |

---

## 最佳实践

### 1. 接口优先
- 先定义接口，再实现
- 便于 Mock 和单元测试

### 2. 依赖注入
- 使用 Wire 管理依赖
- 构造函数接收依赖，不要在函数内创建
- Logger 作为依赖注入，便于测试

### 3. 错误处理
- 使用 `errors.Is()` 判断错误类型
- 统一在 errs 包定义错误
- 错误信息要包含上下文

### 4. 日志规范 (Zap)
- 使用结构化日志 `zap.String()`, `zap.Int()` 等
- 每个组件使用 `logger.Named()` 区分
- 包含请求 ID 和追踪 ID
- 区分日志级别：Debug/Info/Warn/Error

### 5. Gin 最佳实践
- 使用中间件处理通用逻辑
- Handler 职责单一：参数校验 → 调用服务 → 返回响应
- 统一响应格式
- 使用 `c.Request.Context()` 传递上下文

### 6. 测试规范
- 单元测试使用 Mock
- 集成测试使用真实依赖
- 测试覆盖率 > 80%

### 7. 代码生成
- Mock、Wire、gRPC 代码自动生成
- 使用 `go generate` 统一管理
