// Package handler provides HTTP handlers for the AI Gateway.
package handler

import (
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"ai-gateway/internal/converter"
	"ai-gateway/internal/domain"
	gatewaysvc "ai-gateway/internal/service/gateway"
)

// AnthropicHandler handles Anthropic-compatible API requests.
type AnthropicHandler struct {
	gw        gatewaysvc.GatewayService
	converter *converter.AnthropicConverter
	logger    *zap.Logger
}

// NewAnthropicHandler creates a new Anthropic handler.
func NewAnthropicHandler(gw gatewaysvc.GatewayService, logger *zap.Logger) *AnthropicHandler {
	return &AnthropicHandler{
		gw:        gw,
		converter: converter.NewAnthropicConverter(),
		logger:    logger.Named("handler.anthropic"),
	}
}

// Messages handles POST /v1/messages
func (h *AnthropicHandler) Messages(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.logger.Error("failed to read request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"type": "error",
			"error": gin.H{
				"type":    "invalid_request_error",
				"message": "Failed to read request body",
			},
		})
		return
	}

	req, err := h.converter.DecodeRequest(body)
	if err != nil {
		h.logger.Error("failed to decode request", zap.Error(err))
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
		h.handleStream(c, req)
	} else {
		h.handleNonStream(c, req)
	}
}

func (h *AnthropicHandler) handleNonStream(c *gin.Context, req *domain.ChatRequest) {
	resp, err := h.gw.Chat(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("chat request failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"type": "error",
			"error": gin.H{
				"type":    "api_error",
				"message": err.Error(),
			},
		})
		return
	}

	respBody, err := h.converter.EncodeResponse(resp)
	if err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"type": "error",
			"error": gin.H{
				"type":    "api_error",
				"message": "Failed to encode response",
			},
		})
		return
	}

	c.Data(http.StatusOK, "application/json", respBody)
}

func (h *AnthropicHandler) handleStream(c *gin.Context, req *domain.ChatRequest) {
	deltaCh, err := h.gw.ChatStream(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("stream request failed", zap.Error(err))
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

	// Send message_start event
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
					// Send content_block_start
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
				// Send tool_use content_block_start
				if delta.Content != nil {
					startBlock := fmt.Sprintf(`{"type":"content_block_start","index":%d,"content_block":{"type":"tool_use","id":"%s","name":"%s","input":{}}}`,
						contentIndex, delta.Content.ToolID, delta.Content.ToolName)
					fmt.Fprintf(w, "event: content_block_start\ndata: %s\n\n", startBlock)

					chunk, _ := h.converter.EncodeStreamDelta(&delta)
					fmt.Fprintf(w, "event: content_block_delta\ndata: %s\n\n", chunk)
				}

			case "done":
				// Send content_block_stop
				stopBlock := fmt.Sprintf(`{"type":"content_block_stop","index":%d}`, contentIndex)
				fmt.Fprintf(w, "event: content_block_stop\ndata: %s\n\n", stopBlock)

				// Send message_delta
				chunk, _ := h.converter.EncodeStreamDelta(&delta)
				fmt.Fprintf(w, "event: message_delta\ndata: %s\n\n", chunk)

				// Send message_stop
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
