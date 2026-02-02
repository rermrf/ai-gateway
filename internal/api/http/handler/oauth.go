// Package handler 提供 AI 网关的 HTTP 请求处理器。
package handler

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"

	"ai-gateway/internal/errs"
	"ai-gateway/internal/pkg/ginx"
	"ai-gateway/internal/pkg/logger"
	"ai-gateway/internal/service/auth"
	"ai-gateway/internal/service/oauth/linuxdo"
	"ai-gateway/internal/service/user"
)

// OAuthHandler 处理 OAuth 相关的 API 请求。
type OAuthHandler struct {
	linuxDoSvc  linuxdo.Service
	userSvc     user.Service
	authService *auth.AuthService
	logger      logger.Logger
	enabled     bool
}

// NewOAuthHandler 创建 OAuthHandler。
func NewOAuthHandler(
	linuxDoSvc linuxdo.Service,
	userSvc user.Service,
	authService *auth.AuthService,
	enabled bool,
	l logger.Logger,
) *OAuthHandler {
	return &OAuthHandler{
		linuxDoSvc:  linuxDoSvc,
		userSvc:     userSvc,
		authService: authService,
		enabled:     enabled,
		logger:      l.With(logger.String("handler", "oauth")),
	}
}

// LinuxDoLogin 发起 LinuxDo OAuth 登录。
// @Summary 发起 LinuxDo OAuth 登录
// @Tags OAuth
// @Success 302 {string} string "重定向到 LinuxDo 授权页"
// @Router /api/oauth/linuxdo/login [get]
func (h *OAuthHandler) LinuxDoLogin(c *gin.Context) {
	if !h.enabled {
		h.logger.Warn("LinuxDo login attempted but not enabled")
		ginx.Fail(c, errs.CodeForbidden, "LinuxDo 登录未启用")
		return
	}

	// 生成随机 state 防止 CSRF
	state, err := generateState()
	if err != nil {
		h.logger.Error("failed to generate state", logger.Error(err))
		ginx.Fail(c, errs.CodeInternalError, "登录初始化失败")
		return
	}

	// 将 state 存储到 cookie 中（简单实现，生产环境建议使用 Redis）
	c.SetCookie("oauth_state", state, 300, "/", "", false, true)

	authorizeURL := h.linuxDoSvc.GetAuthorizeURL(state)
	h.logger.Info("redirecting to LinuxDo authorization",
		logger.String("state", state[:8]+"..."), // 只记录前8位
	)
	c.Redirect(302, authorizeURL)
}

// LinuxDoCallback 处理 LinuxDo OAuth 回调。
// @Summary 处理 LinuxDo OAuth 回调
// @Tags OAuth
// @Param code query string true "授权码"
// @Param state query string true "状态码"
// @Success 200 {object} LoginResponse
// @Router /api/oauth/linuxdo/callback [get]
func (h *OAuthHandler) LinuxDoCallback(c *gin.Context) {
	if !h.enabled {
		ginx.Fail(c, errs.CodeForbidden, "LinuxDo 登录未启用")
		return
	}

	code := c.Query("code")
	if code == "" {
		h.logger.Warn("missing authorization code in callback")
		ginx.Fail(c, errs.CodeInvalidParameter, "缺少授权码")
		return
	}

	// 验证 state 防止 CSRF
	stateParam := c.Query("state")
	stateCookie, err := c.Cookie("oauth_state")
	if err != nil || stateParam == "" || stateParam != stateCookie {
		h.logger.Warn("invalid oauth state",
			logger.String("param", stateParam),
			logger.Bool("cookieExists", err == nil),
		)
		ginx.Fail(c, errs.CodeInvalidParameter, "无效的请求状态")
		return
	}

	// 清除 state cookie
	c.SetCookie("oauth_state", "", -1, "/", "", false, true)

	// 换取 token
	tokenResp, err := h.linuxDoSvc.ExchangeToken(c.Request.Context(), code)
	if err != nil {
		h.logger.Error("failed to exchange token", logger.Error(err))
		ginx.Fail(c, errs.CodeInternalError, "登录失败，请重试")
		return
	}

	// 获取用户信息
	userInfo, err := h.linuxDoSvc.GetUserInfo(c.Request.Context(), tokenResp.AccessToken)
	if err != nil {
		h.logger.Error("failed to get user info", logger.Error(err))
		ginx.Fail(c, errs.CodeInternalError, "获取用户信息失败")
		return
	}

	// 检查用户状态
	if !userInfo.Active {
		h.logger.Warn("inactive LinuxDo user attempted login",
			logger.Int64("linuxDoId", userInfo.ID),
		)
		ginx.Fail(c, errs.CodeForbidden, "LinuxDo 账号未激活")
		return
	}
	if userInfo.Silenced {
		h.logger.Warn("silenced LinuxDo user attempted login",
			logger.Int64("linuxDoId", userInfo.ID),
		)
		ginx.Fail(c, errs.CodeForbidden, "LinuxDo 账号已被禁言")
		return
	}

	// 生成头像 URL
	avatarURL := ""
	if userInfo.AvatarTemplate != "" {
		avatarURL = "https://linux.do" + strings.Replace(userInfo.AvatarTemplate, "{size}", "120", 1)
	}

	// 查找或创建用户
	email := fmt.Sprintf("%d@linuxdo.oauth", userInfo.ID) // 生成占位邮箱
	u, err := h.userSvc.FindOrCreateByLinuxDo(
		c.Request.Context(),
		userInfo.ID,
		userInfo.Username,
		email,
		avatarURL,
	)
	if err != nil {
		h.logger.Error("failed to find or create user",
			logger.Error(err),
			logger.Int64("linuxDoId", userInfo.ID),
		)
		ginx.Fail(c, errs.CodeInternalError, "用户创建失败")
		return
	}

	// 生成 JWT
	token, err := h.authService.GenerateToken(u.ID, u.Username, string(u.Role))
	if err != nil {
		h.logger.Error("failed to generate token", logger.Error(err))
		ginx.Fail(c, errs.CodeInternalError, "登录失败")
		return
	}

	h.logger.Info("user logged in via LinuxDo",
		logger.Int64("userId", u.ID),
		logger.String("username", u.Username),
		logger.Int64("linuxDoId", userInfo.ID),
	)

	// 返回登录成功
	ginx.OK(c, LoginResponse{
		Token:    token,
		UserID:   u.ID,
		Username: u.Username,
		Role:     u.Role.String(),
	})
}

// generateState 生成随机 state 字符串用于 CSRF 防护。
func generateState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return hex.EncodeToString(b), nil
}
