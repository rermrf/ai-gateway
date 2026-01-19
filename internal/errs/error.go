// Package errs 定义 AI 网关的所有错误类型。
package errs

import "errors"

// 通用错误
var (
	ErrInvalidRequest     = errors.New("invalid request")
	ErrInvalidModel       = errors.New("invalid model")
	ErrModelNotFound      = errors.New("model not found")
	ErrProviderNotFound   = errors.New("provider not found")
	ErrAuthRequired       = errors.New("authentication required")
	ErrInvalidAPIKey      = errors.New("invalid API key")
	ErrRateLimited        = errors.New("rate limited")
	ErrContextCanceled    = errors.New("context canceled")
	ErrStreamClosed       = errors.New("stream closed")
)

// 提供商错误
var (
	ErrProviderTimeout    = errors.New("provider timeout")
	ErrProviderError      = errors.New("provider error")
	ErrProviderOverloaded = errors.New("provider overloaded")
)

// 转换错误
var (
	ErrUnsupportedFeature = errors.New("unsupported feature")
	ErrConversionFailed   = errors.New("conversion failed")
)
