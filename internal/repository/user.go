// Package repository 定义数据访问的存储库接口。
package repository

import (
	"context"

	"ai-gateway/internal/repository/dao"
)

// UserRepository 定义用户的存储库接口。
type UserRepository interface {
	Create(ctx context.Context, user *dao.User) error
	Update(ctx context.Context, user *dao.User) error
	Delete(ctx context.Context, id int64) error
	GetByID(ctx context.Context, id int64) (*dao.User, error)
	GetByTenantAndUsername(ctx context.Context, tenantID int64, username string) (*dao.User, error)
	GetByTenantAndEmail(ctx context.Context, tenantID int64, email string) (*dao.User, error)
	List(ctx context.Context) ([]dao.User, error)
	ListByTenantID(ctx context.Context, tenantID int64) ([]dao.User, error)
}

// userRepository 是 UserRepository 的默认实现。
type userRepository struct {
	dao dao.UserDAO
}

// NewUserRepository 创建一个新的 UserRepository。
func NewUserRepository(userDAO dao.UserDAO) UserRepository {
	return &userRepository{dao: userDAO}
}

func (r *userRepository) Create(ctx context.Context, user *dao.User) error {
	return r.dao.Create(ctx, user)
}

func (r *userRepository) Update(ctx context.Context, user *dao.User) error {
	return r.dao.Update(ctx, user)
}

func (r *userRepository) Delete(ctx context.Context, id int64) error {
	return r.dao.Delete(ctx, id)
}

func (r *userRepository) GetByID(ctx context.Context, id int64) (*dao.User, error) {
	return r.dao.GetByID(ctx, id)
}

func (r *userRepository) GetByTenantAndUsername(ctx context.Context, tenantID int64, username string) (*dao.User, error) {
	return r.dao.GetByTenantAndUsername(ctx, tenantID, username)
}

func (r *userRepository) GetByTenantAndEmail(ctx context.Context, tenantID int64, email string) (*dao.User, error) {
	return r.dao.GetByTenantAndEmail(ctx, tenantID, email)
}

func (r *userRepository) List(ctx context.Context) ([]dao.User, error) {
	return r.dao.List(ctx)
}

func (r *userRepository) ListByTenantID(ctx context.Context, tenantID int64) ([]dao.User, error) {
	return r.dao.ListByTenantID(ctx, tenantID)
}
