// Package providers 定义 LLM 提供商接口。
package providers

import (
	"context"
	"errors"

	"ai-gateway/internal/domain"
)

// Provider 是所有 LLM 提供商必须实现的接口。
type Provider interface {
	// Name 返回提供商名称（例如 "openai"、"anthropic"）
	Name() string

	// Chat 发送聊天补全请求并返回响应。
	// 用于非流式请求。
	Chat(ctx context.Context, req *domain.ChatRequest) (*domain.ChatResponse, error)

	// ChatStream 发送流式聊天补全请求。
	// 返回一个发射流增量（stream deltas）的通道。
	ChatStream(ctx context.Context, req *domain.ChatRequest) (<-chan domain.StreamDelta, error)

	// ListModels 返回可用模型列表。
	ListModels(ctx context.Context) ([]string, error)

	// SupportsStreaming 如果提供商支持流式传输，则返回 true。
	SupportsStreaming() bool

	// SupportsTools 如果提供商支持工具/函数调用，则返回 true。
	SupportsTools() bool

	// SupportsVision 如果提供商支持视觉/图像，则返回 true。
	SupportsVision() bool
}

// ErrProviderUnavailable 当没有可用提供商时返回。
var ErrProviderUnavailable = errors.New("no provider available")
