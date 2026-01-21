// Package repository 定义数据访问的存储库接口。
package repository

import (
	"context"

	"ai-gateway/internal/domain"
	"ai-gateway/internal/repository/dao"
)

// UserRepository 定义用户的存储库接口。
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id int64) error
	GetByID(ctx context.Context, id int64) (*domain.User, error)
	GetByUsername(ctx context.Context, username string) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	List(ctx context.Context) ([]domain.User, error)
}

// userRepository 是 UserRepository 的默认实现。
type userRepository struct {
	dao dao.UserDAO
}

// NewUserRepository 创建一个新的 UserRepository。
func NewUserRepository(userDAO dao.UserDAO) UserRepository {
	return &userRepository{dao: userDAO}
}

// toDAO 将 domain.User 转换为 dao.User。
func (r *userRepository) toDAO(user *domain.User) *dao.User {
	return &dao.User{
		ID:           user.ID,
		Username:     user.Username,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		Role:         dao.UserRole(user.Role),
		Status:       dao.UserStatus(user.Status),
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}
}

// toDomain 将 dao.User 转换为 domain.User。
func (r *userRepository) toDomain(user *dao.User) *domain.User {
	if user == nil {
		return nil
	}
	return &domain.User{
		ID:           user.ID,
		Username:     user.Username,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		Role:         domain.UserRole(user.Role),
		Status:       domain.UserStatus(user.Status),
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	daoUser := r.toDAO(user)
	if err := r.dao.Create(ctx, daoUser); err != nil {
		return err
	}
	user.ID = daoUser.ID
	user.CreatedAt = daoUser.CreatedAt
	user.UpdatedAt = daoUser.UpdatedAt
	return nil
}

func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	return r.dao.Update(ctx, r.toDAO(user))
}

func (r *userRepository) Delete(ctx context.Context, id int64) error {
	return r.dao.Delete(ctx, id)
}

func (r *userRepository) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	daoUser, err := r.dao.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return r.toDomain(daoUser), nil
}

func (r *userRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	daoUser, err := r.dao.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	return r.toDomain(daoUser), nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	daoUser, err := r.dao.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return r.toDomain(daoUser), nil
}

func (r *userRepository) List(ctx context.Context) ([]domain.User, error) {
	daoUsers, err := r.dao.List(ctx)
	if err != nil {
		return nil, err
	}
	users := make([]domain.User, len(daoUsers))
	for i, u := range daoUsers {
		users[i] = *r.toDomain(&u)
	}
	return users, nil
}
