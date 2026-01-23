// Package handler 提供 AI 网关的 HTTP 请求处理器。
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"ai-gateway/internal/errs"
	"ai-gateway/internal/pkg/logger"
	"ai-gateway/internal/service/apikey"
	"ai-gateway/internal/service/user"
	"ai-gateway/internal/service/wallet"
)

// ErrorHandler 提供统一的 HTTP 错误处理。
type ErrorHandler struct {
	logger logger.Logger
}

// NewErrorHandler 创建错误处理器。
func NewErrorHandler(l logger.Logger) *ErrorHandler {
	return &ErrorHandler{logger: l}
}

// HandleError 根据错误类型返回适当的 HTTP 响应。
// 支持多种错误类型的统一处理。
func (h *ErrorHandler) HandleError(c *gin.Context, err error) {
	h.logger.Warn("request failed",
		logger.Error(err),
		logger.String("path", c.Request.URL.Path),
		logger.String("method", c.Request.Method),
	)

	switch err {
	// 用户服务错误
	case user.ErrUserNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case user.ErrUserAlreadyExists, user.ErrEmailAlreadyExists:
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case user.ErrInvalidPassword:
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case user.ErrUserDisabled:
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})

	// API Key 服务错误
	case apikey.ErrAPIKeyNotFound, apikey.ErrInvalidAPIKey:
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case apikey.ErrAPIKeyNotOwned:
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case apikey.ErrAPIKeyDisabled, apikey.ErrAPIKeyExpired:
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})

	// 钱包服务错误
	case wallet.ErrWalletNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case wallet.ErrInsufficientBalance:
		c.JSON(http.StatusPaymentRequired, gin.H{"error": err.Error()})

	// 通用错误
	case errs.ErrInvalidRequest:
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errs.ErrAuthRequired, errs.ErrInvalidAPIKey:
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
	case errs.ErrProviderNotFound, errs.ErrModelNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errs.ErrRateLimited:
		c.JSON(http.StatusTooManyRequests, gin.H{"error": err.Error()})

	// 默认内部错误
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器内部错误"})
	}
}

// HandleAPIError 处理 OpenAI/Anthropic 兼容格式的 API 错误。
func (h *ErrorHandler) HandleAPIError(c *gin.Context, err error, errorType string) {
	h.logger.Warn("API request failed",
		logger.Error(err),
		logger.String("path", c.Request.URL.Path),
	)

	statusCode := http.StatusInternalServerError
	errType := errorType
	if errType == "" {
		errType = "api_error"
	}

	switch err {
	case errs.ErrInvalidRequest:
		statusCode = http.StatusBadRequest
		errType = "invalid_request_error"
	case errs.ErrAuthRequired, errs.ErrInvalidAPIKey:
		statusCode = http.StatusUnauthorized
		errType = "authentication_error"
	case errs.ErrProviderNotFound, errs.ErrModelNotFound:
		statusCode = http.StatusNotFound
		errType = "not_found_error"
	case errs.ErrRateLimited:
		statusCode = http.StatusTooManyRequests
		errType = "rate_limit_error"
	case wallet.ErrInsufficientBalance:
		statusCode = http.StatusPaymentRequired
		errType = "invalid_request_error"
	}

	c.JSON(statusCode, gin.H{
		"error": gin.H{
			"message": err.Error(),
			"type":    errType,
		},
	})
}
