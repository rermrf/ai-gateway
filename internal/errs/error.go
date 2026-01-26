// Package errs 定义 AI 网关的所有错误类型。
// 使用统一的错误码体系，避免错误定义散落在各处。
package errs

import (
	"errors"
	"fmt"
	"net/http"
)

// ErrorCode 错误码类型
type ErrorCode int

// 错误码定义
// 格式：XYYZZZ
// X: 1=通用, 2=认证, 3=用户, 4=API Key, 5=钱包, 6=提供商, 7=转换
// YY: 子模块
// ZZZ: 具体错误
const (
	// 通用错误 (1XXYYY)
	CodeSuccess          ErrorCode = 0
	CodeInvalidRequest   ErrorCode = 100001
	CodeInvalidParameter ErrorCode = 100002
	CodeNotFound         ErrorCode = 100003
	CodeContextCanceled  ErrorCode = 100004
	CodeInternalError    ErrorCode = 100500

	// 认证错误 (2XXYYY)
	CodeUnauthorized       ErrorCode = 200001
	CodeForbidden          ErrorCode = 200002
	CodeAuthRequired       ErrorCode = 200003
	CodeInvalidCredentials ErrorCode = 200004
	CodeInvalidToken       ErrorCode = 200005
	CodeTokenExpired       ErrorCode = 200006

	// 用户错误 (3XXYYY)
	CodeUserNotFound       ErrorCode = 300001
	CodeUserAlreadyExists  ErrorCode = 300002
	CodeEmailAlreadyExists ErrorCode = 300003
	CodeInvalidPassword    ErrorCode = 300004
	CodeUserDisabled       ErrorCode = 300005

	// API Key 错误 (4XXYYY)
	CodeAPIKeyNotFound      ErrorCode = 400001
	CodeAPIKeyInvalid       ErrorCode = 400002
	CodeAPIKeyDisabled      ErrorCode = 400003
	CodeAPIKeyExpired       ErrorCode = 400004
	CodeAPIKeyNotOwned      ErrorCode = 400005
	CodeAPIKeyQuotaExceeded ErrorCode = 400006

	// 钱包错误 (5XXYYY)
	CodeWalletNotFound      ErrorCode = 500001
	CodeInsufficientBalance ErrorCode = 500002

	// 提供商错误 (6XXYYY)
	CodeProviderNotFound    ErrorCode = 600001
	CodeProviderUnavailable ErrorCode = 600002
	CodeProviderTimeout     ErrorCode = 600003
	CodeProviderError       ErrorCode = 600004
	CodeProviderOverloaded  ErrorCode = 600005
	CodeModelNotFound       ErrorCode = 600006
	CodeInvalidModel        ErrorCode = 600007

	// 限流错误
	CodeRateLimited  ErrorCode = 600100
	CodeStreamClosed ErrorCode = 600101

	// 转换错误 (7XXYYY)
	CodeUnsupportedFeature ErrorCode = 700001
	CodeConversionFailed   ErrorCode = 700002
)

// AppError 统一的应用错误类型
type AppError struct {
	Code    ErrorCode // 错误码
	Message string    // 用户可读的错误信息
	Err     error     // 原始错误（可选）
}

// Error 实现 error 接口
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// Unwrap 支持 errors.Unwrap
func (e *AppError) Unwrap() error {
	return e.Err
}

// Is 支持 errors.Is 比较
func (e *AppError) Is(target error) bool {
	if t, ok := target.(*AppError); ok {
		return e.Code == t.Code
	}
	return false
}

// HTTPStatus 返回对应的 HTTP 状态码
func (e *AppError) HTTPStatus() int {
	switch {
	case e.Code == CodeSuccess:
		return http.StatusOK
	case e.Code >= 100000 && e.Code < 200000:
		// 通用错误
		switch e.Code {
		case CodeNotFound:
			return http.StatusNotFound
		case CodeInvalidRequest, CodeInvalidParameter:
			return http.StatusBadRequest
		default:
			return http.StatusInternalServerError
		}
	case e.Code >= 200000 && e.Code < 300000:
		// 认证错误
		switch e.Code {
		case CodeForbidden:
			return http.StatusForbidden
		default:
			return http.StatusUnauthorized
		}
	case e.Code >= 300000 && e.Code < 400000:
		// 用户错误
		switch e.Code {
		case CodeUserNotFound:
			return http.StatusNotFound
		case CodeUserAlreadyExists, CodeEmailAlreadyExists:
			return http.StatusConflict
		case CodeUserDisabled:
			return http.StatusForbidden
		default:
			return http.StatusBadRequest
		}
	case e.Code >= 400000 && e.Code < 500000:
		// API Key 错误
		switch e.Code {
		case CodeAPIKeyNotFound:
			return http.StatusNotFound
		case CodeAPIKeyNotOwned:
			return http.StatusForbidden
		case CodeAPIKeyDisabled, CodeAPIKeyExpired, CodeAPIKeyInvalid:
			return http.StatusUnauthorized
		default:
			return http.StatusBadRequest
		}
	case e.Code >= 500000 && e.Code < 600000:
		// 钱包错误
		switch e.Code {
		case CodeWalletNotFound:
			return http.StatusNotFound
		case CodeInsufficientBalance:
			return http.StatusPaymentRequired
		default:
			return http.StatusBadRequest
		}
	case e.Code >= 600000 && e.Code < 700000:
		// 提供商错误
		switch e.Code {
		case CodeProviderNotFound, CodeModelNotFound:
			return http.StatusNotFound
		case CodeRateLimited:
			return http.StatusTooManyRequests
		case CodeProviderTimeout:
			return http.StatusGatewayTimeout
		case CodeProviderOverloaded:
			return http.StatusServiceUnavailable
		default:
			return http.StatusBadGateway
		}
	default:
		return http.StatusInternalServerError
	}
}

// APIErrorType 返回 OpenAI/Anthropic 兼容的错误类型
func (e *AppError) APIErrorType() string {
	switch {
	case e.Code >= 100000 && e.Code < 200000:
		return "invalid_request_error"
	case e.Code >= 200000 && e.Code < 300000:
		return "authentication_error"
	case e.Code >= 300000 && e.Code < 400000:
		return "invalid_request_error"
	case e.Code >= 400000 && e.Code < 500000:
		return "authentication_error"
	case e.Code >= 500000 && e.Code < 600000:
		return "invalid_request_error"
	case e.Code >= 600000 && e.Code < 700000:
		if e.Code == CodeRateLimited {
			return "rate_limit_error"
		}
		return "api_error"
	default:
		return "api_error"
	}
}

// New 创建新的 AppError
func New(code ErrorCode, message string) *AppError {
	return &AppError{Code: code, Message: message}
}

// Wrap 包装原始错误创建 AppError
func Wrap(code ErrorCode, message string, err error) *AppError {
	return &AppError{Code: code, Message: message, Err: err}
}

// ============================================
// 预定义的错误实例（向后兼容）
// ============================================

// 通用错误
var (
	ErrInvalidRequest   = New(CodeInvalidRequest, "invalid request")
	ErrInvalidParameter = New(CodeInvalidParameter, "invalid parameter")
	ErrNotFound         = New(CodeNotFound, "record not found")
	ErrContextCanceled  = New(CodeContextCanceled, "context canceled")
)

// 认证错误
var (
	ErrUnauthorized       = New(CodeUnauthorized, "unauthorized")
	ErrForbidden          = New(CodeForbidden, "forbidden")
	ErrAuthRequired       = New(CodeAuthRequired, "authentication required")
	ErrInvalidCredentials = New(CodeInvalidCredentials, "invalid credentials")
	ErrInvalidToken       = New(CodeInvalidToken, "invalid token")
	ErrTokenExpired       = New(CodeTokenExpired, "token expired")
)

// 用户错误
var (
	ErrUserNotFound       = New(CodeUserNotFound, "用户不存在")
	ErrUserAlreadyExists  = New(CodeUserAlreadyExists, "用户已存在")
	ErrEmailAlreadyExists = New(CodeEmailAlreadyExists, "邮箱已被注册")
	ErrInvalidPassword    = New(CodeInvalidPassword, "密码错误")
	ErrUserDisabled       = New(CodeUserDisabled, "用户已被禁用")
)

// API Key 错误
var (
	ErrAPIKeyNotFound      = New(CodeAPIKeyNotFound, "API Key 不存在")
	ErrAPIKeyInvalid       = New(CodeAPIKeyInvalid, "invalid API key")
	ErrAPIKeyDisabled      = New(CodeAPIKeyDisabled, "API key is disabled")
	ErrAPIKeyExpired       = New(CodeAPIKeyExpired, "API key has expired")
	ErrAPIKeyNotOwned      = New(CodeAPIKeyNotOwned, "无权操作此 API Key")
	ErrAPIKeyQuotaExceeded = New(CodeAPIKeyQuotaExceeded, "API Key 额度不足")
)

// 钱包错误
var (
	ErrWalletNotFound      = New(CodeWalletNotFound, "wallet not found")
	ErrInsufficientBalance = New(CodeInsufficientBalance, "insufficient balance")
)

// 提供商错误
var (
	ErrProviderNotFound    = New(CodeProviderNotFound, "provider not found")
	ErrProviderUnavailable = New(CodeProviderUnavailable, "no provider available")
	ErrProviderTimeout     = New(CodeProviderTimeout, "provider timeout")
	ErrProviderError       = New(CodeProviderError, "provider error")
	ErrProviderOverloaded  = New(CodeProviderOverloaded, "provider overloaded")
	ErrModelNotFound       = New(CodeModelNotFound, "model not found")
	ErrInvalidModel        = New(CodeInvalidModel, "invalid model")
	ErrRateLimited         = New(CodeRateLimited, "rate limited")
	ErrStreamClosed        = New(CodeStreamClosed, "stream closed")
)

// 转换错误
var (
	ErrUnsupportedFeature = New(CodeUnsupportedFeature, "unsupported feature")
	ErrConversionFailed   = New(CodeConversionFailed, "conversion failed")
)

// ============================================
// 辅助函数
// ============================================

// IsNotFound 检查是否为"未找到"类型错误
func IsNotFound(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		switch appErr.Code {
		case CodeNotFound, CodeUserNotFound, CodeAPIKeyNotFound,
			CodeWalletNotFound, CodeProviderNotFound, CodeModelNotFound:
			return true
		}
	}
	return false
}

// IsAuthError 检查是否为认证类型错误
func IsAuthError(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code >= 200000 && appErr.Code < 300000
	}
	return false
}

// IsUserError 检查是否为用户类型错误
func IsUserError(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code >= 300000 && appErr.Code < 400000
	}
	return false
}

// GetCode 从错误中获取错误码
func GetCode(err error) ErrorCode {
	if err == nil {
		return CodeSuccess
	}
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code
	}
	return CodeInternalError
}
