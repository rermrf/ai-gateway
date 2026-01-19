// Package converter handles protocol conversion between different API formats.
package converter

import (
	"ai-gateway/internal/domain"
)

// Converter defines the interface for protocol converters.
type Converter interface {
	// DecodeRequest converts an API-specific request to the unified format.
	DecodeRequest(data []byte) (*domain.ChatRequest, error)

	// EncodeResponse converts a unified response to the API-specific format.
	EncodeResponse(resp *domain.ChatResponse) ([]byte, error)

	// EncodeStreamDelta converts a stream delta to the API-specific format.
	EncodeStreamDelta(delta *domain.StreamDelta) ([]byte, error)

	// FormatName returns the format name (e.g., "openai", "anthropic").
	FormatName() string
}
