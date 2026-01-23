// Package dao 提供数据访问对象 (DAO) 接口和模型。
package dao

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// UserRole 定义用户角色。
type UserRole string

const (
	UserRoleUser  UserRole = "user"  // 普通用户
	UserRoleAdmin UserRole = "admin" // 管理员
)

// UserStatus 定义用户状态。
type UserStatus string

const (
	UserStatusActive   UserStatus = "active"   // 激活
	UserStatusPending  UserStatus = "pending"  // 待审核
	UserStatusDisabled UserStatus = "disabled" // 禁用
)

// User 是用户的数据库模型。
type User struct {
	ID           int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	Username     string     `gorm:"uniqueIndex;size:64;not null" json:"username"`
	Email        string     `gorm:"uniqueIndex;size:128;not null" json:"email"`
	PasswordHash string     `gorm:"size:256;not null" json:"-"`
	Role         UserRole   `gorm:"type:enum('user','admin');default:'user'" json:"role"`
	Status       UserStatus `gorm:"type:enum('active','pending','disabled');default:'pending';index" json:"status"`
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
	GetByUsername(ctx context.Context, username string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	List(ctx context.Context) ([]User, error)
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

func (d *GormUserDAO) GetByUsername(ctx context.Context, username string) (*User, error) {
	var user User
	err := d.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &user, err
}

func (d *GormUserDAO) GetByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	err := d.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
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

var _ UserDAO = (*GormUserDAO)(nil)
