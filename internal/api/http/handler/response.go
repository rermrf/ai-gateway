// Package handler 提供 AI 网关的 HTTP 请求处理器。
package handler

import (
	"errors"

	"github.com/gin-gonic/gin"

	"ai-gateway/internal/errs"
)

func toAppError(err error, fallbackCode errs.ErrorCode, fallbackMsg string) *errs.AppError {
	if err == nil {
		return errs.New(errs.CodeSuccess, "")
	}
	var appErr *errs.AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	return errs.Wrap(fallbackCode, fallbackMsg, err)
}

// writeOpenAIError 用于 /v1/* (OpenAI 兼容) 的统一错误返回。
func writeOpenAIError(c *gin.Context, err error) {
	appErr := toAppError(err, errs.CodeInternalError, "Internal server error")
	c.JSON(appErr.HTTPStatus(), gin.H{
		"error": gin.H{
			"message": appErr.Message,
			"type":    appErr.APIErrorType(),
		},
	})
}

// writeAnthropicError 用于 /v1/messages (Anthropic 兼容) 的统一错误返回。
func writeAnthropicError(c *gin.Context, err error) {
	appErr := toAppError(err, errs.CodeInternalError, "Internal server error")
	c.JSON(appErr.HTTPStatus(), gin.H{
		"type": "error",
		"error": gin.H{
			"type":    appErr.APIErrorType(),
			"message": appErr.Message,
		},
	})
}
