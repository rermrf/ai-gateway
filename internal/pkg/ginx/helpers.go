package ginx

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"ai-gateway/internal/errs"
)

// OK 返回统一成功响应。
// 如果调用方提前设置了状态码（例如 c.Status(201)），这里会沿用。
func OK(c *gin.Context, data any) {
	status := c.Writer.Status()
	if status == 0 {
		status = http.StatusOK
	}
	if status < 200 || status >= 600 {
		status = http.StatusOK
	}
	c.JSON(status, Result{Code: int(errs.CodeSuccess), Msg: "ok", Data: data})
}

// Fail 直接按错误码返回失败响应。
func Fail(c *gin.Context, code errs.ErrorCode, msg string) {
	appErr := errs.New(code, msg)
	c.JSON(appErr.HTTPStatus(), Result{Code: int(appErr.Code), Msg: appErr.Message})
}

// FromErr 将 error 映射为统一失败响应。
// - 如果 err 不是 *errs.AppError，则回退为 500。
func FromErr(c *gin.Context, err error) {
	if err == nil {
		OK(c, nil)
		return
	}

	var appErr *errs.AppError
	if !errors.As(err, &appErr) {
		appErr = errs.Wrap(errs.CodeInternalError, "服务器内部错误", err)
	}
	// 失败响应 Data 为空，避免对外暴露内部细节
	c.JSON(appErr.HTTPStatus(), Result{Code: int(appErr.Code), Msg: appErr.Message})
}
