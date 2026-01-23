// Package apikey 提供 API Key 相关的业务逻辑服务。
package apikey

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"go.uber.org/zap"

	"ai-gateway/internal/domain"
	"ai-gateway/internal/repository"
)

var (
	// ErrInvalidAPIKey API key 无效
	ErrInvalidAPIKey = errors.New("invalid API key")
	// ErrAPIKeyDisabled API key 已禁用
	ErrAPIKeyDisabled = errors.New("API key is disabled")
	// ErrAPIKeyExpired API key 已过期
	ErrAPIKeyExpired = errors.New("API key has expired")
	// ErrAPIKeyNotFound API Key 不存在
	ErrAPIKeyNotFound = errors.New("API Key 不存在")
	// ErrAPIKeyNotOwned 无权操作此 API Key
	ErrAPIKeyNotOwned = errors.New("无权操作此 API Key")
)

// Service API Key 服务接口。
//
//go:generate mockgen -source=./apikey.go -destination=./mocks/apikey.mock.go -package=apikeymocks -typed Service
type Service interface {
	// ValidateAPIKey 验证 API key 并返回用户信息
	ValidateAPIKey(ctx context.Context, key string) (*domain.APIKey, error)
	// RecordUsage 记录 API key 使用（异步）
	RecordUsage(ctx context.Context, apiKeyID int64) error

	// --- 用户级 API Key 管理 ---
	// ListByUserID 获取指定用户的 API Key 列表
	ListByUserID(ctx context.Context, userID int64) ([]domain.APIKey, error)
	// Create 创建 API Key（返回完整密钥）
	Create(ctx context.Context, userID int64, name string) (*domain.APIKey, string, error)
	// Delete 删除用户的 API Key（需验证所有权）
	Delete(ctx context.Context, userID int64, keyID int64) error

	// --- 管理员级 API Key 管理 ---
	// List 获取所有 API Key（管理员）
	List(ctx context.Context) ([]domain.APIKey, error)
	// DeleteByID 直接删除 API Key（管理员，无需验证所有权）
	DeleteByID(ctx context.Context, keyID int64) error
}

// service API Key 服务实现（小写，不导出）。
type service struct {
	repo   repository.APIKeyRepository
	logger *zap.Logger
}

// NewService 创建 API Key 服务实例。
func NewService(repo repository.APIKeyRepository, logger *zap.Logger) Service {
	return &service{
		repo:   repo,
		logger: logger.Named("service.apikey"),
	}
}

// ValidateAPIKey 验证 API key 并返回用户信息。
func (s *service) ValidateAPIKey(ctx context.Context, key string) (*domain.APIKey, error) {
	s.logger.Debug("validating API key",
		zap.String("key_prefix", maskAPIKey(key)),
	)

	// 从数据库查询 API key
	apiKey, err := s.repo.GetByKey(ctx, key)
	if err != nil {
		s.logger.Error("failed to query API key",
			zap.Error(err),
			zap.String("key_prefix", maskAPIKey(key)),
		)
		return nil, err
	}

	// API key 不存在
	if apiKey == nil {
		s.logger.Warn("API key not found",
			zap.String("key_prefix", maskAPIKey(key)),
		)
		return nil, ErrInvalidAPIKey
	}

	// 检查是否启用
	if !apiKey.Enabled {
		s.logger.Warn("API key is disabled",
			zap.Int64("key_id", apiKey.ID),
			zap.String("key_prefix", maskAPIKey(key)),
		)
		return nil, ErrAPIKeyDisabled
	}

	// 检查是否过期
	if apiKey.ExpiresAt != nil && apiKey.ExpiresAt.Before(time.Now()) {
		s.logger.Warn("API key has expired",
			zap.Int64("key_id", apiKey.ID),
			zap.String("key_prefix", maskAPIKey(key)),
			zap.Time("expires_at", *apiKey.ExpiresAt),
		)
		return nil, ErrAPIKeyExpired
	}

	s.logger.Info("API key validated successfully",
		zap.Int64("key_id", apiKey.ID),
		zap.Int64("user_id", apiKey.UserID),
		zap.String("key_name", apiKey.Name),
	)

	return apiKey, nil
}

// RecordUsage 记录 API key 使用时间。
func (s *service) RecordUsage(ctx context.Context, apiKeyID int64) error {
	s.logger.Debug("recording API key usage",
		zap.Int64("key_id", apiKeyID),
	)

	if err := s.repo.UpdateLastUsed(ctx, apiKeyID); err != nil {
		s.logger.Error("failed to update last used time",
			zap.Error(err),
			zap.Int64("key_id", apiKeyID),
		)
		return err
	}

	return nil
}

// maskAPIKey 遮盖 API key 的大部分内容，仅保留前后几个字符用于日志记录。
func maskAPIKey(key string) string {
	if len(key) <= 8 {
		return "***"
	}
	return key[:4] + "..." + key[len(key)-4:]
}

// --- 用户级 API Key 管理实现 ---

// ListByUserID 获取指定用户的 API Key 列表。
func (s *service) ListByUserID(ctx context.Context, userID int64) ([]domain.APIKey, error) {
	keys, err := s.repo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 脱敏
	for i := range keys {
		keys[i].Key = keys[i].MaskKey()
	}
	return keys, nil
}

// Create 创建 API Key。
func (s *service) Create(ctx context.Context, userID int64, name string) (*domain.APIKey, string, error) {
	// 生成随机 Key
	bytes := make([]byte, 32)
	rand.Read(bytes)
	key := "sk-" + hex.EncodeToString(bytes)

	// 生成 Hash
	hash := sha256.Sum256([]byte(key))
	keyHash := hex.EncodeToString(hash[:])

	apiKey := &domain.APIKey{
		UserID:  userID,
		Key:     key,
		KeyHash: keyHash,
		Name:    name,
		Enabled: true,
	}

	if err := s.repo.Create(ctx, apiKey); err != nil {
		s.logger.Error("failed to create api key", zap.Error(err))
		return nil, "", err
	}

	s.logger.Info("api key created", zap.Int64("userId", userID), zap.String("name", name))
	return apiKey, key, nil
}

// Delete 删除用户的 API Key（需验证所有权）。
func (s *service) Delete(ctx context.Context, userID int64, keyID int64) error {
	apiKey, err := s.repo.GetByID(ctx, keyID)
	if err != nil {
		return err
	}
	if apiKey == nil {
		return ErrAPIKeyNotFound
	}
	if apiKey.UserID != userID {
		return ErrAPIKeyNotOwned
	}

	return s.repo.Delete(ctx, keyID)
}

// --- 管理员级 API Key 管理实现 ---

// List 获取所有 API Key（管理员）。
func (s *service) List(ctx context.Context) ([]domain.APIKey, error) {
	keys, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	// 脱敏
	for i := range keys {
		keys[i].Key = keys[i].MaskKey()
	}
	return keys, nil
}

// DeleteByID 直接删除 API Key（管理员，无需验证所有权）。
func (s *service) DeleteByID(ctx context.Context, keyID int64) error {
	return s.repo.Delete(ctx, keyID)
}
