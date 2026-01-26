package user

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"

	"ai-gateway/internal/domain"
	"ai-gateway/internal/errs"
	"ai-gateway/internal/pkg/logger"
	"ai-gateway/internal/repository/mocks"
)

func TestService_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockUsageRepo := mocks.NewMockUsageLogRepository(ctrl)
	svc := NewService(mockUserRepo, mockUsageRepo, logger.NewNopLogger())

	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		mockUserRepo.EXPECT().GetByUsername(ctx, "testuser").Return(nil, nil)
		mockUserRepo.EXPECT().GetByEmail(ctx, "test@example.com").Return(nil, nil)
		mockUserRepo.EXPECT().Create(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, u *domain.User) error {
			assert.Equal(t, "testuser", u.Username)
			assert.Equal(t, "test@example.com", u.Email)
			assert.NotEmpty(t, u.PasswordHash)
			u.ID = 1
			return nil
		})

		user, err := svc.Register(ctx, "testuser", "test@example.com", "password123")
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, int64(1), user.ID)
	})

	t.Run("UserAlreadyExists", func(t *testing.T) {
		mockUserRepo.EXPECT().GetByUsername(ctx, "testuser").Return(&domain.User{ID: 1}, nil)

		user, err := svc.Register(ctx, "testuser", "test@example.com", "password123")
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.True(t, errors.Is(err, errs.ErrUserAlreadyExists))
	})

	t.Run("EmailAlreadyExists", func(t *testing.T) {
		mockUserRepo.EXPECT().GetByUsername(ctx, "testuser").Return(nil, nil)
		mockUserRepo.EXPECT().GetByEmail(ctx, "test@example.com").Return(&domain.User{ID: 1}, nil)

		user, err := svc.Register(ctx, "testuser", "test@example.com", "password123")
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.True(t, errors.Is(err, errs.ErrEmailAlreadyExists))
	})
}

func TestService_GetByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	svc := NewService(mockUserRepo, nil, logger.NewNopLogger())
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		mockUserRepo.EXPECT().GetByID(ctx, int64(1)).Return(&domain.User{ID: 1, Username: "test"}, nil)

		user, err := svc.GetByID(ctx, 1)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), user.ID)
	})

	t.Run("NotFound", func(t *testing.T) {
		mockUserRepo.EXPECT().GetByID(ctx, int64(1)).Return(nil, nil)

		user, err := svc.GetByID(ctx, 1)
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.True(t, errors.Is(err, errs.ErrUserNotFound))
	})
}

func TestService_ChangePassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	svc := NewService(mockUserRepo, nil, logger.NewNopLogger())
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		// 每个子测试创建独立的用户对象，避免状态污染
		oldHash, _ := bcrypt.GenerateFromPassword([]byte("oldpass"), bcrypt.DefaultCost)
		user := &domain.User{
			ID:           1,
			PasswordHash: string(oldHash),
		}

		mockUserRepo.EXPECT().GetByID(ctx, int64(1)).Return(user, nil)
		mockUserRepo.EXPECT().Update(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, u *domain.User) error {
			err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte("newpass"))
			assert.NoError(t, err)
			return nil
		})

		err := svc.ChangePassword(ctx, 1, "oldpass", "newpass")
		assert.NoError(t, err)
	})

	t.Run("WrongOldPassword", func(t *testing.T) {
		// 独立的用户对象
		oldHash, _ := bcrypt.GenerateFromPassword([]byte("oldpass"), bcrypt.DefaultCost)
		user := &domain.User{
			ID:           1,
			PasswordHash: string(oldHash),
		}

		mockUserRepo.EXPECT().GetByID(ctx, int64(1)).Return(user, nil)

		err := svc.ChangePassword(ctx, 1, "wrongpass", "newpass")
		assert.Error(t, err)
		assert.True(t, errors.Is(err, errs.ErrInvalidPassword))
	})

	t.Run("UserNotFound", func(t *testing.T) {
		mockUserRepo.EXPECT().GetByID(ctx, int64(999)).Return(nil, nil)

		err := svc.ChangePassword(ctx, 999, "oldpass", "newpass")
		assert.Error(t, err)
		assert.True(t, errors.Is(err, errs.ErrUserNotFound))
	})
}

func TestService_UpdateProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	svc := NewService(mockUserRepo, nil, logger.NewNopLogger())
	ctx := context.Background()

	user := &domain.User{
		ID:    1,
		Email: "old@example.com",
	}

	t.Run("Success", func(t *testing.T) {
		mockUserRepo.EXPECT().GetByID(ctx, int64(1)).Return(user, nil)
		mockUserRepo.EXPECT().GetByEmail(ctx, "new@example.com").Return(nil, nil)
		mockUserRepo.EXPECT().Update(ctx, gomock.Any()).Return(nil)

		updated, err := svc.UpdateProfile(ctx, 1, "new@example.com")
		assert.NoError(t, err)
		assert.Equal(t, "new@example.com", updated.Email)
	})

	t.Run("EmailInfoUsed", func(t *testing.T) {
		mockUserRepo.EXPECT().GetByID(ctx, int64(1)).Return(user, nil)
		mockUserRepo.EXPECT().GetByEmail(ctx, "exist@example.com").Return(&domain.User{ID: 2}, nil)

		updated, err := svc.UpdateProfile(ctx, 1, "exist@example.com")
		assert.Error(t, err)
		assert.Nil(t, updated)
		assert.True(t, errors.Is(err, errs.ErrEmailAlreadyExists))
	})
}

func TestService_GetDailyUsage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsageRepo := mocks.NewMockUsageLogRepository(ctrl)
	svc := NewService(nil, mockUsageRepo, logger.NewNopLogger())
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		expected := []domain.DailyUsage{
			{Date: time.Now().Format("2006-01-02"), Requests: 10, InputTokens: 50, OutputTokens: 50},
		}
		mockUsageRepo.EXPECT().GetDailyUsageByUserID(ctx, int64(1), 30).Return(expected, nil)

		usage, err := svc.GetDailyUsage(ctx, 1, 30)
		assert.NoError(t, err)
		assert.Equal(t, expected, usage)
	})
}
