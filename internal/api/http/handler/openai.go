// Package handler 为 AI 网关提供 HTTP 处理器。
package handler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"ai-gateway/internal/converter"
	"ai-gateway/internal/domain"
	"ai-gateway/internal/errs"
	"ai-gateway/internal/pkg/logger"
	"ai-gateway/internal/service/apikey"
	gatewaysvc "ai-gateway/internal/service/gateway"
	"ai-gateway/internal/service/modelrate"
	"ai-gateway/internal/service/usage"
	"ai-gateway/internal/service/wallet"
)

// OpenAIHandler 处理 OpenAI 兼容的 API 请求。
type OpenAIHandler struct {
	gw           gatewaysvc.GatewayService
	walletSvc    wallet.Service
	usageSvc     usage.Service
	apiKeySvc    apikey.Service
	modelRateSvc modelrate.Service
	converter    *converter.OpenAIConverter
	logger       logger.Logger
}

// NewOpenAIHandler 创建一个新的 OpenAI 处理器。
func NewOpenAIHandler(
	gatewayService gatewaysvc.GatewayService,
	walletSvc wallet.Service,
	usageSvc usage.Service,
	apiKeySvc apikey.Service,
	modelRateSvc modelrate.Service,
	l logger.Logger,
) *OpenAIHandler {
	return &OpenAIHandler{
		gw:           gatewayService,
		walletSvc:    walletSvc,
		usageSvc:     usageSvc,
		apiKeySvc:    apiKeySvc,
		modelRateSvc: modelRateSvc,
		converter:    converter.NewOpenAIConverter(),
		logger:       l.With(logger.String("handler", "openai")),
	}
}

// ChatCompletions 处理 POST /v1/chat/completions
func (h *OpenAIHandler) ChatCompletions(c *gin.Context) {
	// 1. Check Balance via Pre-flight
	// Get user_id from context (set by APIKeyAuth)
	var userID int64
	if val, exists := c.Get("user_id"); exists {
		if id, ok := val.(int64); ok {
			userID = id
		}
	}

	if userID > 0 {
		hasBalance, err := h.walletSvc.HasBalance(c.Request.Context(), userID)
		if err != nil {
			h.logger.Error("failed to check balance", logger.Error(err))
			writeOpenAIError(c, errs.Wrap(errs.CodeInternalError, "Failed to check balance", err))
			return
		}
		if !hasBalance {
			writeOpenAIError(c, errs.Wrap(errs.CodeInsufficientBalance, "Insufficient balance. Please top up your wallet.", nil))
			return
		}
	}

	start := time.Now()
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.logger.Error("failed to read request body", logger.Error(err))
		writeOpenAIError(c, errs.Wrap(errs.CodeInvalidRequest, "Failed to read request body", err))
		return
	}

	req, err := h.converter.DecodeRequest(body)
	if err != nil {
		h.logger.Error("failed to decode request", logger.Error(err))
		writeOpenAIError(c, errs.New(errs.CodeInvalidRequest, err.Error()))
		return
	}

	if req.Stream {
		h.handleStream(c, req, start)
	} else {
		h.handleNonStream(c, req, start)
	}
}

func (h *OpenAIHandler) handleNonStream(c *gin.Context, req *domain.ChatRequest, start time.Time) {
	resp, err := h.gw.Chat(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("chat request failed", logger.Error(err))
		writeOpenAIError(c, err)
		return
	}

	// 记录使用情况
	latency := int(time.Since(start).Milliseconds())
	h.logUsage(c, req.Model, resp.Provider, resp.Usage.PromptTokens, resp.Usage.CompletionTokens, resp.Usage.TotalTokens, http.StatusOK, latency)

	respBody, err := h.converter.EncodeResponse(resp)
	if err != nil {
		h.logger.Error("failed to encode response", logger.Error(err))
		writeOpenAIError(c, errs.Wrap(errs.CodeInternalError, "Failed to encode response", err))
		return
	}

	c.Data(http.StatusOK, "application/json", respBody)
}

func (h *OpenAIHandler) handleStream(c *gin.Context, req *domain.ChatRequest, start time.Time) {
	deltaCh, providerName, err := h.gw.ChatStream(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("stream request failed", logger.Error(err))
		writeOpenAIError(c, err)
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Transfer-Encoding", "chunked")

	// 跟踪使用情况
	var inputTokens, outputTokens int
	// 用于捕获提供商名称（如果我们在流Delta中得到它，或者我们只使用我们在开始时得到的那个）
	// 注意：handleStream 函数没有访问实际使用的提供商名称，因为它在 h.gw.ChatStream 内部解析。
	// 这是一个小的设计缺陷。现在我们只能记录请求的模型。
	// 或者我们等待流结束后，尝试从上下文中获取，或者让 GatewayService 返回它。
	// 暂时使用 req.Model 作为提供商（由于重写，它可能是实际模型）。
	// 更好的方法是让 ChatStream 返回 (chan, providerName, err)。
	// 为了快速修复，我们假设使用情况将在流结束时记录，这里我们先尽力而为。
	// 实际上，DeepSeek/OpenAI 流在最后一个块中发送 usage。
	// 我们需要解析它。

	c.Stream(func(w io.Writer) bool {
		select {
		case delta, ok := <-deltaCh:
			if !ok {
				// 流已结束
				fmt.Fprintf(w, "data: [DONE]\n\n")

				// 记录使用情况
				// 目前假设 outputTokens 是 delta 数量 * 1 (非常粗略) 或 0
				latency := int(time.Since(start).Milliseconds())
				h.logUsage(c, req.Model, providerName, inputTokens, outputTokens, inputTokens+outputTokens, http.StatusOK, latency)
				return false
			}

			// 更新 Output Tokens
			if delta.Content != nil {
				if delta.Content.Text != "" || delta.Content.Thinking != "" {
					outputTokens++
				}
			}

			// 如果 Delta 有 Usage 字段 (需要更新 Domain StreamDelta)
			if delta.Usage != nil {
				inputTokens = delta.Usage.PromptTokens
				outputTokens = delta.Usage.CompletionTokens
			}

			chunk, err := h.converter.EncodeStreamDelta(&delta)
			if err != nil {
				h.logger.Warn("failed to encode delta", logger.Error(err))
				return true
			}

			fmt.Fprintf(w, "data: %s\n\n", chunk)

			// 立即刷新
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}

			// Check if this was the final delta
			if delta.Type == "done" {
				fmt.Fprintf(w, "data: [DONE]\n\n")
				// 记录使用情况
				latency := int(time.Since(start).Milliseconds())
				h.logUsage(c, req.Model, providerName, inputTokens, outputTokens, inputTokens+outputTokens, http.StatusOK, latency)
				return false
			}

			return true

		case <-c.Request.Context().Done():
			// 客户端断开连接，记录部分使用情况
			latency := int(time.Since(start).Milliseconds())
			h.logUsage(c, req.Model, providerName, inputTokens, outputTokens, inputTokens+outputTokens, 499, latency)
			return false
		}
	})
}

func (h *OpenAIHandler) logUsage(c *gin.Context, model, provider string, inputTokens, outputTokens, totalTokens, statusCode, latency int) {
	// 提取 API Key 信息
	userID, _ := c.Get("user_id")
	apiKeyID, _ := c.Get("api_key_id")

	uid, ok := userID.(int64)
	if !ok {
		// 如果没有用户ID（例如未认证），则不记录或记录为匿名
		return
	}

	akID, ok := apiKeyID.(int64)
	var akIDPtr *int64
	if ok {
		akIDPtr = &akID
	}

	log := &domain.UsageLog{
		UserID:       uid,
		APIKeyID:     akIDPtr,
		Model:        model,
		Provider:     provider,
		InputTokens:  inputTokens,
		OutputTokens: outputTokens,
		LatencyMs:    latency,
		StatusCode:   statusCode,
		ClientIP:     c.ClientIP(),
		UserAgent:    c.Request.UserAgent(),
		RequestID:    c.GetString("request_id"), // Middleware should set this
	}

	// Try to get X-Request-ID if not in context
	if log.RequestID == "" {
		log.RequestID = c.GetHeader("X-Request-ID")
	}

	// 异步记录

	// 异步记录
	go func() {
		// 1. 记录 Usage Log (DB)
		if err := h.usageSvc.LogRequest(context.Background(), log); err != nil {
			h.logger.Error("failed to log usage", logger.Error(err))
		}

		// 2. 如果使用 API Key，增加已用额度
		if akIDPtr != nil {
			// 计算本次费用 (如果 Quota 是按金额) 或者 Token (如果 Quota 是按 Token)
			// 这里假设 Quota 是金额 (因为 decimal(15,6))
			// 需要 ModelRateService
			cost := 0.0
			promptPrice, completionPrice, err := h.modelRateSvc.GetRateForModel(context.Background(), model)
			if err == nil {
				cost = (float64(inputTokens)/1000000.0)*promptPrice + (float64(outputTokens)/1000000.0)*completionPrice
			} else {
				// 如果找不到费率，可能不需要计费，或者记录 0
				h.logger.Warn("failed to get model rate for api key usage", logger.String("model", model), logger.Error(err))
			}

			if cost > 0 {
				if err := h.apiKeySvc.IncrementUsage(context.Background(), *akIDPtr, cost); err != nil {
					h.logger.Error("failed to increment api key usage", logger.Error(err), logger.Int64("key_id", *akIDPtr))
				}
			}
		}
	}()
}

// ListModels 处理 GET /v1/models
func (h *OpenAIHandler) ListModels(c *gin.Context) {
	models, err := h.gw.ListModels(c.Request.Context())
	if err != nil {
		h.logger.Error("failed to list models", logger.Error(err))
		writeOpenAIError(c, err)
		return
	}

	data := make([]gin.H, len(models))
	for i, model := range models {
		data[i] = gin.H{
			"id":       model,
			"object":   "model",
			"created":  0,
			"owned_by": "ai-gateway",
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"object": "list",
		"data":   data,
	})
}
