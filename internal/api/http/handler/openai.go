// Package handler 为 AI 网关提供 HTTP 处理器。
package handler

import (
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"ai-gateway/internal/converter"
	"ai-gateway/internal/domain"
	"ai-gateway/internal/errs"
	"ai-gateway/internal/pkg/logger"
	"ai-gateway/internal/service/chat"
	gatewaysvc "ai-gateway/internal/service/gateway"
)

// OpenAIHandler 处理 OpenAI 兼容的 API 请求。
type OpenAIHandler struct {
	gw        gatewaysvc.GatewayService
	chatSvc   chat.Service
	converter *converter.OpenAIConverter
	logger    logger.Logger
}

// NewOpenAIHandler 创建一个新的 OpenAI 处理器。
func NewOpenAIHandler(
	gatewayService gatewaysvc.GatewayService,
	chatSvc chat.Service,
	l logger.Logger,
) *OpenAIHandler {
	return &OpenAIHandler{
		gw:        gatewayService,
		chatSvc:   chatSvc,
		converter: converter.NewOpenAIConverter(),
		logger:    l.With(logger.String("handler", "openai")),
	}
}

// ChatCompletions 处理 POST /v1/chat/completions
func (h *OpenAIHandler) ChatCompletions(c *gin.Context) {
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
		h.handleStream(c, req)
	} else {
		h.handleNonStream(c, req)
	}
}

func (h *OpenAIHandler) handleNonStream(c *gin.Context, req *domain.ChatRequest) {
	meta := chat.RequestMeta{
		UserID:    ctxGetInt64(c, "user_id"),
		APIKeyID:  ctxGetInt64Ptr(c, "api_key_id"),
		RequestID: c.GetString("request_id"),
		ClientIP:  c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	}

	resp, err := h.chatSvc.Chat(c.Request.Context(), req, meta)
	if err != nil {
		h.logger.Error("chat request failed", logger.Error(err))
		writeOpenAIError(c, err)
		return
	}

	respBody, err := h.converter.EncodeResponse(resp)
	if err != nil {
		h.logger.Error("failed to encode response", logger.Error(err))
		writeOpenAIError(c, errs.Wrap(errs.CodeInternalError, "Failed to encode response", err))
		return
	}

	c.Data(http.StatusOK, "application/json", respBody)
}

func (h *OpenAIHandler) handleStream(c *gin.Context, req *domain.ChatRequest) {
	meta := chat.RequestMeta{
		UserID:    ctxGetInt64(c, "user_id"),
		APIKeyID:  ctxGetInt64Ptr(c, "api_key_id"),
		RequestID: c.GetString("request_id"),
		ClientIP:  c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	}

	deltaCh, _, err := h.chatSvc.ChatStream(c.Request.Context(), req, meta)
	if err != nil {
		h.logger.Error("stream request failed", logger.Error(err))
		writeOpenAIError(c, err)
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
