// Package dao 提供数据访问对象 (DAO) 接口和模型。
package dao

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// UserRole 定义用户在租户内的角色。
type UserRole string

const (
	UserRoleOwner  UserRole = "owner"  // 租户所有者
	UserRoleAdmin  UserRole = "admin"  // 租户管理员
	UserRoleMember UserRole = "member" // 普通成员
)

// UserStatus 定义用户状态。
type UserStatus string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusDisabled UserStatus = "disabled"
)

// User 是用户的数据库模型（多租户）。
type User struct {
	ID           int64      `gorm:"primaryKey;autoIncrement"`
	TenantID     int64      `gorm:"index;not null" json:"tenantId"`
	Username     string     `gorm:"size:64;not null" json:"username"`
	Email        string     `gorm:"size:128;not null" json:"email"`
	PasswordHash string     `gorm:"size:256" json:"-"`
	Role         UserRole   `gorm:"type:enum('owner','admin','member');default:'member'" json:"role"`
	Status       UserStatus `gorm:"type:enum('active','disabled');default:'active';index" json:"status"`
	CreatedAt    time.Time  `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt    time.Time  `gorm:"autoUpdateTime" json:"updatedAt"`
}

// TableName 返回 User 的表名。
func (User) TableName() string {
	return "users"
}

// UserDAO 定义用户的数据访问操作。
type UserDAO interface {
	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id int64) error
	GetByID(ctx context.Context, id int64) (*User, error)
	GetByTenantAndUsername(ctx context.Context, tenantID int64, username string) (*User, error)
	GetByTenantAndEmail(ctx context.Context, tenantID int64, email string) (*User, error)
	List(ctx context.Context) ([]User, error)
	ListByTenantID(ctx context.Context, tenantID int64) ([]User, error)
}

// GormUserDAO 是 UserDAO 的 GORM 实现。
type GormUserDAO struct {
	db *gorm.DB
}

// NewGormUserDAO 创建一个新的基于 GORM 的 UserDAO。
func NewGormUserDAO(db *gorm.DB) UserDAO {
	return &GormUserDAO{db: db}
}

func (d *GormUserDAO) Create(ctx context.Context, user *User) error {
	return d.db.WithContext(ctx).Create(user).Error
}

func (d *GormUserDAO) Update(ctx context.Context, user *User) error {
	return d.db.WithContext(ctx).Save(user).Error
}

func (d *GormUserDAO) Delete(ctx context.Context, id int64) error {
	return d.db.WithContext(ctx).Delete(&User{}, id).Error
}

func (d *GormUserDAO) GetByID(ctx context.Context, id int64) (*User, error) {
	var user User
	err := d.db.WithContext(ctx).First(&user, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &user, err
}

func (d *GormUserDAO) GetByTenantAndUsername(ctx context.Context, tenantID int64, username string) (*User, error) {
	var user User
	err := d.db.WithContext(ctx).Where("tenant_id = ? AND username = ?", tenantID, username).First(&user).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &user, err
}

func (d *GormUserDAO) GetByTenantAndEmail(ctx context.Context, tenantID int64, email string) (*User, error) {
	var user User
	err := d.db.WithContext(ctx).Where("tenant_id = ? AND email = ?", tenantID, email).First(&user).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &user, err
}

func (d *GormUserDAO) List(ctx context.Context) ([]User, error) {
	var users []User
	err := d.db.WithContext(ctx).Find(&users).Error
	return users, err
}

func (d *GormUserDAO) ListByTenantID(ctx context.Context, tenantID int64) ([]User, error) {
	var users []User
	err := d.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Find(&users).Error
	return users, err
}

var _ UserDAO = (*GormUserDAO)(nil)
