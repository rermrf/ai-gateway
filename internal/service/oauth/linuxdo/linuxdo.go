// Package linuxdo 提供 LinuxDo OAuth2 认证服务。
package linuxdo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"ai-gateway/internal/pkg/logger"
)

const (
	authorizeURL = "https://connect.linux.do/oauth2/authorize"
	tokenURL     = "https://connect.linux.do/oauth2/token"
	userInfoURL  = "https://connect.linux.do/api/user"
)

// UserInfo LinuxDo 用户信息。
type UserInfo struct {
	ID             int64  `json:"id"`
	Username       string `json:"username"`
	Name           string `json:"name"`
	AvatarTemplate string `json:"avatar_template"`
	Active         bool   `json:"active"`
	TrustLevel     int    `json:"trust_level"`
	Silenced       bool   `json:"silenced"`
}

// TokenResponse OAuth2 Token 响应。
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// Service LinuxDo OAuth 服务接口。
//
//go:generate mockgen -source=./linuxdo.go -destination=./mocks/linuxdo.mock.go -package=linuxdomocks -typed Service
type Service interface {
	GetAuthorizeURL(state string) string
	ExchangeToken(ctx context.Context, code string) (*TokenResponse, error)
	GetUserInfo(ctx context.Context, accessToken string) (*UserInfo, error)
}

type service struct {
	clientID     string
	clientSecret string
	redirectURL  string
	httpClient   *http.Client
	logger       logger.Logger
}

// NewService 创建 LinuxDo OAuth 服务。
func NewService(clientID, clientSecret, redirectURL string, l logger.Logger) Service {
	return &service{
		clientID:     clientID,
		clientSecret: clientSecret,
		redirectURL:  redirectURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: l.With(logger.String("service", "oauth.linuxdo")),
	}
}

// GetAuthorizeURL 生成 OAuth 授权跳转 URL。
func (s *service) GetAuthorizeURL(state string) string {
	params := url.Values{}
	params.Set("client_id", s.clientID)
	params.Set("redirect_uri", s.redirectURL)
	params.Set("response_type", "code")
	params.Set("state", state)
	return fmt.Sprintf("%s?%s", authorizeURL, params.Encode())
}

// ExchangeToken 使用授权码换取访问令牌。
func (s *service) ExchangeToken(ctx context.Context, code string) (*TokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", s.redirectURL)
	data.Set("client_id", s.clientID)
	data.Set("client_secret", s.clientSecret)

	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		s.logger.Error("failed to create token request", logger.Error(err))
		return nil, fmt.Errorf("create request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.logger.Error("failed to exchange token", logger.Error(err))
		return nil, fmt.Errorf("token exchange request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		s.logger.Error("token exchange failed",
			logger.Int("status", resp.StatusCode),
			logger.String("body", string(body)),
		)
		return nil, fmt.Errorf("token exchange failed: status %d, body: %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		s.logger.Error("failed to decode token response", logger.Error(err))
		return nil, fmt.Errorf("decode token response failed: %w", err)
	}

	s.logger.Debug("token exchange successful")
	return &tokenResp, nil
}

// GetUserInfo 获取用户信息。
func (s *service) GetUserInfo(ctx context.Context, accessToken string) (*UserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", userInfoURL, nil)
	if err != nil {
		s.logger.Error("failed to create user info request", logger.Error(err))
		return nil, fmt.Errorf("create request failed: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.logger.Error("failed to get user info", logger.Error(err))
		return nil, fmt.Errorf("user info request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		s.logger.Error("get user info failed",
			logger.Int("status", resp.StatusCode),
			logger.String("body", string(body)),
		)
		return nil, fmt.Errorf("get user info failed: status %d, body: %s", resp.StatusCode, string(body))
	}

	var userInfo UserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		s.logger.Error("failed to decode user info response", logger.Error(err))
		return nil, fmt.Errorf("decode user info failed: %w", err)
	}

	s.logger.Debug("get user info successful",
		logger.Int64("userId", userInfo.ID),
		logger.String("username", userInfo.Username),
	)
	return &userInfo, nil
}
