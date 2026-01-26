package apikey

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"ai-gateway/internal/domain"
	"ai-gateway/internal/errs"
	"ai-gateway/internal/pkg/logger"
	"ai-gateway/internal/repository/mocks"
)

func TestService_Create(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockAPIKeyRepository(ctrl)
	svc := NewService(mockRepo, logger.NewNopLogger())
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		mockRepo.EXPECT().Create(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, k *domain.APIKey) error {
			assert.NotEmpty(t, k.Key)
			assert.NotEmpty(t, k.KeyHash)
			assert.Equal(t, "test-key", k.Name)
			assert.Equal(t, int64(1), k.UserID)
			return nil
		})

		apiKey, fullKey, err := svc.Create(ctx, 1, "test-key", nil, nil, nil)
		assert.NoError(t, err)
		assert.NotNil(t, apiKey)
		assert.NotEmpty(t, fullKey)
	})
}

func TestService_ValidateAPIKey(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockAPIKeyRepository(ctrl)
	svc := NewService(mockRepo, logger.NewNopLogger())
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		key := &domain.APIKey{
			ID:      1,
			Enabled: true,
		}
		mockRepo.EXPECT().GetByKey(ctx, "valid-key").Return(key, nil)

		result, err := svc.ValidateAPIKey(ctx, "valid-key")
		assert.NoError(t, err)
		assert.Equal(t, key, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		mockRepo.EXPECT().GetByKey(ctx, "invalid-key").Return(nil, nil)

		result, err := svc.ValidateAPIKey(ctx, "invalid-key")
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.True(t, errors.Is(err, errs.ErrAPIKeyInvalid))
	})

	t.Run("Disabled", func(t *testing.T) {
		key := &domain.APIKey{
			ID:      1,
			Enabled: false,
		}
		mockRepo.EXPECT().GetByKey(ctx, "disabled-key").Return(key, nil)

		result, err := svc.ValidateAPIKey(ctx, "disabled-key")
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.True(t, errors.Is(err, errs.ErrAPIKeyDisabled))
	})

	t.Run("Expired", func(t *testing.T) {
		yesterday := time.Now().Add(-24 * time.Hour)
		key := &domain.APIKey{
			ID:        1,
			Enabled:   true,
			ExpiresAt: &yesterday,
		}
		mockRepo.EXPECT().GetByKey(ctx, "expired-key").Return(key, nil)

		result, err := svc.ValidateAPIKey(ctx, "expired-key")
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.True(t, errors.Is(err, errs.ErrAPIKeyExpired))
	})

	t.Run("QuotaExceeded", func(t *testing.T) {
		quota := 100.0
		key := &domain.APIKey{
			ID:         1,
			Enabled:    true,
			Quota:      &quota,
			UsedAmount: 150.0, // 已超过配额
		}
		mockRepo.EXPECT().GetByKey(ctx, "over-quota-key").Return(key, nil)

		result, err := svc.ValidateAPIKey(ctx, "over-quota-key")
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.True(t, errors.Is(err, errs.ErrAPIKeyQuotaExceeded))
	})
}

func TestService_Delete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockAPIKeyRepository(ctrl)
	svc := NewService(mockRepo, logger.NewNopLogger())
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		key := &domain.APIKey{ID: 1, UserID: 100}
		mockRepo.EXPECT().GetByID(ctx, int64(1)).Return(key, nil)
		mockRepo.EXPECT().Delete(ctx, int64(1)).Return(nil)

		err := svc.Delete(ctx, 100, 1)
		assert.NoError(t, err)
	})

	t.Run("NotOwned", func(t *testing.T) {
		key := &domain.APIKey{ID: 1, UserID: 200} // Owned by user 200
		mockRepo.EXPECT().GetByID(ctx, int64(1)).Return(key, nil)

		err := svc.Delete(ctx, 100, 1) // User 100 tries to delete
		assert.Error(t, err)
		assert.True(t, errors.Is(err, errs.ErrAPIKeyNotOwned))
	})

	t.Run("NotFound", func(t *testing.T) {
		mockRepo.EXPECT().GetByID(ctx, int64(1)).Return(nil, nil)

		err := svc.Delete(ctx, 100, 1)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, errs.ErrAPIKeyNotFound))
	})
}
