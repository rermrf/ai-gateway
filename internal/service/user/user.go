// Package user 提供用户相关业务逻辑服务。
package user

import (
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"

	"ai-gateway/internal/domain"
	"ai-gateway/internal/pkg/logger"
	"ai-gateway/internal/repository"
)

// 用户服务错误定义
var (
	ErrUserNotFound       = errors.New("用户不存在")
	ErrUserAlreadyExists  = errors.New("用户已存在")
	ErrEmailAlreadyExists = errors.New("邮箱已被注册")
	ErrInvalidPassword    = errors.New("密码错误")
	ErrUserDisabled       = errors.New("用户已被禁用")
)

// Service 用户服务接口。
//
//go:generate mockgen -source=./user.go -destination=./mocks/user.mock.go -package=usermocks -typed Service
type Service interface {
	// 用户管理
	Register(ctx context.Context, username, email, password string) (*domain.User, error)
	GetByID(ctx context.Context, id int64) (*domain.User, error)
	GetByUsername(ctx context.Context, username string) (*domain.User, error)
	UpdateProfile(ctx context.Context, userID int64, email string) (*domain.User, error)
	ChangePassword(ctx context.Context, userID int64, oldPassword, newPassword string) error
	List(ctx context.Context) ([]domain.User, error)
	UpdateUser(ctx context.Context, userID int64, role domain.UserRole, status domain.UserStatus) (*domain.User, error)
	Delete(ctx context.Context, userID int64) error

	// 使用统计
	GetUsageStats(ctx context.Context, userID int64) (*domain.UsageStats, error)
	GetDailyUsage(ctx context.Context, userID int64, days int) ([]domain.DailyUsage, error)
}

// service 用户服务实现。
type service struct {
	userRepo     repository.UserRepository
	usageLogRepo repository.UsageLogRepository
	logger       logger.Logger
}

// NewService 创建用户服务实例。
func NewService(
	userRepo repository.UserRepository,
	usageLogRepo repository.UsageLogRepository,
	l logger.Logger,
) Service {
	return &service{
		userRepo:     userRepo,
		usageLogRepo: usageLogRepo,
		logger:       l.With(logger.String("service", "user")),
	}
}

// Register 用户注册。
func (s *service) Register(ctx context.Context, username, email, password string) (*domain.User, error) {
	s.logger.Info("registering user", logger.String("username", username))

	// 检查用户名是否已存在
	existing, _ := s.userRepo.GetByUsername(ctx, username)
	if existing != nil {
		return nil, ErrUserAlreadyExists
	}

	// 检查邮箱是否已存在
	existing, _ = s.userRepo.GetByEmail(ctx, email)
	if existing != nil {
		return nil, ErrEmailAlreadyExists
	}

	// 加密密码
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("failed to hash password", logger.Error(err))
		return nil, err
	}

	user := &domain.User{
		Username:     username,
		Email:        email,
		PasswordHash: string(hash),
		Role:         domain.UserRoleUser,
		Status:       domain.UserStatusPending,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		s.logger.Error("failed to create user", logger.Error(err))
		return nil, err
	}

	s.logger.Info("user registered", logger.Int64("userId", user.ID))
	return user, nil
}

// GetByID 根据 ID 获取用户。
func (s *service) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// GetByUsername 根据用户名获取用户。
func (s *service) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// UpdateProfile 更新用户资料。
func (s *service) UpdateProfile(ctx context.Context, userID int64, email string) (*domain.User, error) {
	user, err := s.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if email != "" && email != user.Email {
		// 检查邮箱是否被占用
		existing, _ := s.userRepo.GetByEmail(ctx, email)
		if existing != nil && existing.ID != userID {
			return nil, ErrEmailAlreadyExists
		}
		user.Email = email
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.Error("failed to update user", logger.Error(err))
		return nil, err
	}
	return user, nil
}

// ChangePassword 修改密码。
func (s *service) ChangePassword(ctx context.Context, userID int64, oldPassword, newPassword string) error {
	user, err := s.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	// 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return ErrInvalidPassword
	}

	// 加密新密码
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.PasswordHash = string(hash)
	return s.userRepo.Update(ctx, user)
}

// List 获取所有用户。
func (s *service) List(ctx context.Context) ([]domain.User, error) {
	return s.userRepo.List(ctx)
}

// UpdateUser 更新用户（管理员）。
func (s *service) UpdateUser(ctx context.Context, userID int64, role domain.UserRole, status domain.UserStatus) (*domain.User, error) {
	user, err := s.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if role != "" {
		user.Role = role
	}
	if status != "" {
		user.Status = status
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

// Delete 删除用户。
func (s *service) Delete(ctx context.Context, userID int64) error {
	return s.userRepo.Delete(ctx, userID)
}

// GetUsageStats 获取使用统计。
func (s *service) GetUsageStats(ctx context.Context, userID int64) (*domain.UsageStats, error) {
	return s.usageLogRepo.GetStatsByUserID(ctx, userID)
}

// GetDailyUsage 获取每日使用统计。
func (s *service) GetDailyUsage(ctx context.Context, userID int64, days int) ([]domain.DailyUsage, error) {
	return s.usageLogRepo.GetDailyUsageByUserID(ctx, userID, days)
}
