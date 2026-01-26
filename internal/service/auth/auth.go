// Package auth 提供认证相关服务。
package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"ai-gateway/internal/errs"
)

// Claims JWT claims 结构。
type Claims struct {
	UserID   int64  `json:"userId"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// AuthService 提供认证相关功能。
type AuthService struct {
	jwtSecret     []byte
	tokenDuration time.Duration
}

// NewAuthService 创建一个新的 AuthService。
func NewAuthService(jwtSecret string, tokenDuration time.Duration) *AuthService {
	return &AuthService{
		jwtSecret:     []byte(jwtSecret),
		tokenDuration: tokenDuration,
	}
}

// HashPassword 使用 bcrypt 加密密码。
func (s *AuthService) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword 验证密码是否匹配。
func (s *AuthService) CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateToken 生成 JWT Token。
func (s *AuthService) GenerateToken(userID int64, username, role string) (string, error) {
	claims := &Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.tokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "ai-gateway",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

// ValidateToken 验证 JWT Token 并返回 Claims。
func (s *AuthService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return s.jwtSecret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, errs.ErrTokenExpired
		}
		return nil, errs.ErrInvalidToken
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errs.ErrInvalidToken
}
