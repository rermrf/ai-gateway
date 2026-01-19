// Package handler 为 AI 网关提供 HTTP 处理器。
package handler

import (
	"bufio"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"ai-gateway/internal/converter"
	"ai-gateway/internal/domain"
	gatewaysvc "ai-gateway/internal/service/gateway"
)

// OpenAIHandler 处理 OpenAI 兼容的 API 请求。
type OpenAIHandler struct {
	gw        gatewaysvc.GatewayService
	converter *converter.OpenAIConverter
	logger    *zap.Logger
}

// NewOpenAIHandler 创建一个新的 OpenAI 处理器。
func NewOpenAIHandler(gw gatewaysvc.GatewayService, logger *zap.Logger) *OpenAIHandler {
	return &OpenAIHandler{
		gw:        gw,
		converter: converter.NewOpenAIConverter(),
		logger:    logger.Named("handler.openai"),
	}
}

// ChatCompletions 处理 POST /v1/chat/completions
func (h *OpenAIHandler) ChatCompletions(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.logger.Error("failed to read request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"message": "Failed to read request body",
				"type":    "invalid_request_error",
			},
		})
		return
	}

	req, err := h.converter.DecodeRequest(body)
	if err != nil {
		h.logger.Error("failed to decode request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"message": err.Error(),
				"type":    "invalid_request_error",
			},
		})
		return
	}

	if req.Stream {
		h.handleStream(c, req)
	} else {
		h.handleNonStream(c, req)
	}
}

func (h *OpenAIHandler) handleNonStream(c *gin.Context, req *domain.ChatRequest) {
	resp, err := h.gw.Chat(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("chat request failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"message": err.Error(),
				"type":    "api_error",
			},
		})
		return
	}

	respBody, err := h.converter.EncodeResponse(resp)
	if err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"message": "Failed to encode response",
				"type":    "api_error",
			},
		})
		return
	}

	c.Data(http.StatusOK, "application/json", respBody)
}

func (h *OpenAIHandler) handleStream(c *gin.Context, req *domain.ChatRequest) {
	deltaCh, err := h.gw.ChatStream(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("stream request failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"message": err.Error(),
				"type":    "api_error",
			},
		})
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Transfer-Encoding", "chunked")

	c.Stream(func(w io.Writer) bool {
		select {
		case delta, ok := <-deltaCh:
			if !ok {
				// 流已结束
				fmt.Fprintf(w, "data: [DONE]\n\n")
				return false
			}

			chunk, err := h.converter.EncodeStreamDelta(&delta)
			if err != nil {
				h.logger.Warn("failed to encode delta", zap.Error(err))
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
				return false
			}

			return true

		case <-c.Request.Context().Done():
			return false
		}
	})
}

// ListModels 处理 GET /v1/models
func (h *OpenAIHandler) ListModels(c *gin.Context) {
	models, err := h.gw.ListModels(c.Request.Context())
	if err != nil {
		h.logger.Error("failed to list models", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"message": err.Error(),
				"type":    "api_error",
			},
		})
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

// Ensure io.Writer interface for SSE streaming
var _ io.Writer = (*bufio.Writer)(nil)
