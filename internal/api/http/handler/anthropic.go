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
	"ai-gateway/internal/pkg/logger"
	"ai-gateway/internal/domain"
	gatewaysvc "ai-gateway/internal/service/gateway"
	"ai-gateway/internal/service/usage"
	"ai-gateway/internal/service/wallet"
)

// AnthropicHandler 处理 Anthropic 兼容的 API 请求。
type AnthropicHandler struct {
	gw        gatewaysvc.GatewayService
	walletSvc wallet.Service
	usageSvc  usage.Service
	converter *converter.AnthropicConverter
	logger    logger.Logger
}

// NewAnthropicHandler 创建一个新的 Anthropic 处理器。
func NewAnthropicHandler(gw gatewaysvc.GatewayService, walletSvc wallet.Service, usageSvc usage.Service, l logger.Logger) *AnthropicHandler {
	return &AnthropicHandler{
		gw:        gw,
		walletSvc: walletSvc,
		usageSvc:  usageSvc,
		converter: converter.NewAnthropicConverter(),
		logger:    l.With(logger.String("handler", "anthropic")),
	}
}

// Messages 处理 POST /v1/messages
func (h *AnthropicHandler) Messages(c *gin.Context) {
	// 1. 鉴权 (JWT Metadata)
	userID := ctxGetInt64(c, "user_id") // Use correct key from APIKeyAuth

	// 检查余额
	if userID > 0 {
		hasBalance, err := h.walletSvc.HasBalance(c.Request.Context(), userID)
		if err != nil {
			h.logger.Error("failed to check balance", logger.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"type": "error",
				"error": gin.H{
					"type":    "api_error",
					"message": "Failed to check balance",
				},
			})
			return
		}
		if !hasBalance {
			c.JSON(http.StatusPaymentRequired, gin.H{
				"type": "error",
				"error": gin.H{
					"type":    "invalid_request_error",
					"message": "Insufficient balance. Please top up your wallet.",
				},
			})
			return
		}
	}

	start := time.Now()
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.logger.Error("failed to read request body", logger.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"type": "error",
			"error": gin.H{
				"type":    "invalid_request_error",
				"message": "无法读取请求体",
			},
		})
		return
	}

	req, err := h.converter.DecodeRequest(body)
	if err != nil {
		h.logger.Error("failed to decode request", logger.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"type": "error",
			"error": gin.H{
				"type":    "invalid_request_error",
				"message": err.Error(),
			},
		})
		return
	}

	if req.Stream {
		h.handleStream(c, req, start)
	} else {
		h.handleNonStream(c, req, start)
	}
}

func (h *AnthropicHandler) handleNonStream(c *gin.Context, req *domain.ChatRequest, start time.Time) {
	resp, err := h.gw.Chat(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("chat request failed", logger.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"type": "error",
			"error": gin.H{
				"type":    "api_error",
				"message": err.Error(),
			},
		})
		return
	}

	// 记录使用情况
	latency := int(time.Since(start).Milliseconds())
	h.logUsage(c, req.Model, resp.Provider, resp.Usage.PromptTokens, resp.Usage.CompletionTokens, resp.Usage.TotalTokens, http.StatusOK, latency)

	// Encode Response
	respBody, err := h.converter.EncodeResponse(resp)
	if err != nil {
		h.logger.Error("failed to encode response", logger.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"type": "error",
			"error": gin.H{
				"type":    "api_error",
				"message": "无法编码响应",
			},
		})
		return
	}

	c.Data(http.StatusOK, "application/json", respBody)
}

func (h *AnthropicHandler) handleStream(c *gin.Context, req *domain.ChatRequest, start time.Time) {
	deltaCh, _, err := h.gw.ChatStream(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("stream request failed", logger.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"type": "error",
			"error": gin.H{
				"type":    "api_error",
				"message": err.Error(),
			},
		})
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Transfer-Encoding", "chunked")

	// 跟踪使用情况
	var inputTokens, outputTokens int

	// 发送 message_start 事件
	messageID := converter.GenerateID()
	startEvent := fmt.Sprintf(`{"type":"message_start","message":{"id":"%s","type":"message","role":"assistant","content":[],"model":"%s","stop_reason":null,"stop_sequence":null,"usage":{"input_tokens":0,"output_tokens":0}}}`, messageID, req.Model)
	fmt.Fprintf(c.Writer, "event: message_start\ndata: %s\n\n", startEvent)
	c.Writer.(http.Flusher).Flush()

	contentIndex := 0

	c.Stream(func(w io.Writer) bool {
		select {
		case delta, ok := <-deltaCh:
			if !ok {
				// 流结束，记录使用情况
				// 对于 Anthropic，流结束可能不包含完整 usage，除非在 message_delta 中捕获
				latency := int(time.Since(start).Milliseconds())
				h.logUsage(c, req.Model, "anthropic", inputTokens, outputTokens, inputTokens+outputTokens, http.StatusOK, latency)
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

			switch delta.Type {
			case "content":
				if contentIndex == 0 {
					// 发送 content_block_start
					startBlock := fmt.Sprintf(`{"type":"content_block_start","index":%d,"content_block":{"type":"text","text":""}}`, contentIndex)
					fmt.Fprintf(w, "event: content_block_start\ndata: %s\n\n", startBlock)
				}

				if delta.Content != nil && delta.Content.Text != "" {
					chunk, _ := h.converter.EncodeStreamDelta(&delta)
					fmt.Fprintf(w, "event: content_block_delta\ndata: %s\n\n", chunk)
				}

			case "thinking":
				chunk, _ := h.converter.EncodeStreamDelta(&delta)
				fmt.Fprintf(w, "event: content_block_delta\ndata: %s\n\n", chunk)

			case "tool_use":
				contentIndex++
				// 发送 tool_use content_block_start
				if delta.Content != nil {
					startBlock := fmt.Sprintf(`{"type":"content_block_start","index":%d,"content_block":{"type":"tool_use","id":"%s","name":"%s","input":{}}}`,
						contentIndex, delta.Content.ToolID, delta.Content.ToolName)
					fmt.Fprintf(w, "event: content_block_start\ndata: %s\n\n", startBlock)

					chunk, _ := h.converter.EncodeStreamDelta(&delta)
					fmt.Fprintf(w, "event: content_block_delta\ndata: %s\n\n", chunk)
				}

			case "done":
				// 发送 content_block_stop
				stopBlock := fmt.Sprintf(`{"type":"content_block_stop","index":%d}`, contentIndex)
				fmt.Fprintf(w, "event: content_block_stop\ndata: %s\n\n", stopBlock)

				// 发送 message_delta
				chunk, _ := h.converter.EncodeStreamDelta(&delta)
				fmt.Fprintf(w, "event: message_delta\ndata: %s\n\n", chunk)

				// 发送 message_stop
				fmt.Fprintf(w, "event: message_stop\ndata: {\"type\":\"message_stop\"}\n\n")

				if f, ok := w.(http.Flusher); ok {
					f.Flush()
				}
				return false
			}

			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
			return true

		case <-c.Request.Context().Done():
			// 客户端断开连接，记录部分使用情况
			latency := int(time.Since(start).Milliseconds())
			h.logUsage(c, req.Model, "anthropic", inputTokens, outputTokens, inputTokens+outputTokens, 499, latency)
			return false
		}
	})
}

func (h *AnthropicHandler) logUsage(c *gin.Context, model, provider string, inputTokens, outputTokens, totalTokens, statusCode, latency int) {
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
	}

	// 异步记录
	go func() {
		if err := h.usageSvc.LogRequest(context.Background(), log); err != nil {
			h.logger.Error("failed to log usage", logger.Error(err))
		}
	}()
}

func ctxGetInt64(c *gin.Context, key string) int64 {
	val, exists := c.Get(key)
	if !exists {
		return 0
	}
	if id, ok := val.(int64); ok {
		return id
	}
	return 0
}
