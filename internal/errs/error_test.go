package errs

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppError_Error(t *testing.T) {
	err := New(CodeInvalidParameter, "invalid parameter")
	assert.Equal(t, "[100002] invalid parameter", err.Error())

	wrapped := Wrap(CodeProviderError, "provider failed", errors.New("timeout"))
	assert.Equal(t, "[600004] provider failed: timeout", wrapped.Error())
}

func TestAppError_HTTPStatus(t *testing.T) {
	tests := []struct {
		name string
		err  *AppError
		want int
	}{
		{
			name: "Success",
			err:  New(CodeSuccess, "success"),
			want: http.StatusOK,
		},
		{
			name: "InvalidRequest",
			err:  New(CodeInvalidRequest, "invalid"),
			want: http.StatusBadRequest,
		},
		{
			name: "AuthError",
			err:  New(CodeInvalidToken, "invalid token"),
			want: http.StatusUnauthorized,
		},
		{
			name: "Forbidden",
			err:  New(CodeUserDisabled, "disabled"),
			want: http.StatusForbidden,
		},
		{
			name: "NotFound",
			err:  New(CodeUserNotFound, "not found"),
			want: http.StatusNotFound,
		},
		{
			name: "Server Error",
			err:  New(CodeInternalError, "internal error"),
			want: http.StatusInternalServerError,
		},
		{
			name: "Unknown Error Code",
			err:  New(ErrorCode(999999), "unknown"),
			want: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.err.HTTPStatus())
		})
	}
}

func TestAppError_Is(t *testing.T) {
	err1 := New(CodeUserNotFound, "user not found")
	err2 := New(CodeUserNotFound, "another user not found")
	err3 := New(CodeInvalidParameter, "invalid param")

	assert.True(t, err1.Is(err2))
	assert.False(t, err1.Is(err3))
	assert.False(t, err1.Is(errors.New("standard error")))
}

func TestHelpers(t *testing.T) {
	t.Run("IsNotFound", func(t *testing.T) {
		assert.True(t, IsNotFound(New(CodeNotFound, "")))
		assert.True(t, IsNotFound(New(CodeUserNotFound, "")))
		assert.True(t, IsNotFound(New(CodeAPIKeyNotFound, "")))
		assert.True(t, IsNotFound(New(CodeModelNotFound, "")))
		assert.False(t, IsNotFound(New(CodeInternalError, "")))
		assert.False(t, IsNotFound(nil))
	})

	t.Run("IsAuthError", func(t *testing.T) {
		assert.True(t, IsAuthError(New(CodeUnauthorized, "")))
		assert.True(t, IsAuthError(New(CodeInvalidToken, "")))
		assert.True(t, IsAuthError(New(CodeTokenExpired, "")))
		// CodeAPIKeyInvalid 是 4XX 范围 (API Key 错误)，不是 2XX 范围 (认证错误)
		assert.False(t, IsAuthError(New(CodeAPIKeyInvalid, "")))
		assert.False(t, IsAuthError(New(CodeInternalError, "")))
	})

	t.Run("IsUserError", func(t *testing.T) {
		assert.False(t, IsUserError(New(CodeInvalidRequest, "")))
		assert.False(t, IsUserError(New(CodeInvalidParameter, "")))
		assert.True(t, IsUserError(New(CodeUserAlreadyExists, "")))
		assert.False(t, IsUserError(New(CodeInternalError, "")))
	})

	t.Run("GetCode", func(t *testing.T) {
		assert.Equal(t, CodeUserNotFound, GetCode(New(CodeUserNotFound, "")))
		assert.Equal(t, CodeInternalError, GetCode(errors.New("std error")))
		assert.Equal(t, CodeSuccess, GetCode(nil))
	})
}
