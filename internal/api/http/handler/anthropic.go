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
)

// AnthropicHandler 处理 Anthropic 兼容的 API 请求。
type AnthropicHandler struct {
	chatSvc   chat.Service
	converter *converter.AnthropicConverter
	logger    logger.Logger
}

// NewAnthropicHandler 创建一个新的 Anthropic 处理器。
func NewAnthropicHandler(chatSvc chat.Service, l logger.Logger) *AnthropicHandler {
	return &AnthropicHandler{
		chatSvc:   chatSvc,
		converter: converter.NewAnthropicConverter(),
		logger:    l.With(logger.String("handler", "anthropic")),
	}
}

// Messages 处理 POST /v1/messages
func (h *AnthropicHandler) Messages(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.logger.Error("failed to read request body", logger.Error(err))
		writeAnthropicError(c, errs.Wrap(errs.CodeInvalidRequest, "无法读取请求体", err))
		return
	}

	req, err := h.converter.DecodeRequest(body)
	if err != nil {
		h.logger.Error("failed to decode request", logger.Error(err))
		writeAnthropicError(c, errs.New(errs.CodeInvalidRequest, err.Error()))
		return
	}

	if req.Stream {
		h.handleStream(c, req)
	} else {
		h.handleNonStream(c, req)
	}
}

func (h *AnthropicHandler) handleNonStream(c *gin.Context, req *domain.ChatRequest) {
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
		writeAnthropicError(c, err)
		return
	}

	// Encode Response
	respBody, err := h.converter.EncodeResponse(resp)
	if err != nil {
		h.logger.Error("failed to encode response", logger.Error(err))
		writeAnthropicError(c, errs.Wrap(errs.CodeInternalError, "无法编码响应", err))
		return
	}

	c.Data(http.StatusOK, "application/json", respBody)
}

func (h *AnthropicHandler) handleStream(c *gin.Context, req *domain.ChatRequest) {
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
		writeAnthropicError(c, err)
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Transfer-Encoding", "chunked")

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
				return false
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
			return false
		}
	})
}
