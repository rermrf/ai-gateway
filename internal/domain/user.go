// Package domain 定义领域模型和业务实体。
package domain

import (
	"time"
)

// UserRole 用户角色枚举。
type UserRole string

const (
	UserRoleUser  UserRole = "user"
	UserRoleAdmin UserRole = "admin"
)

func (r UserRole) String() string {
	return string(r)
}

// IsAdmin 判断是否为管理员。
func (r UserRole) IsAdmin() bool {
	return r == UserRoleAdmin
}

// UserStatus 用户状态枚举。
type UserStatus string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusDisabled UserStatus = "disabled"
)

func (s UserStatus) String() string {
	return string(s)
}

// IsActive 判断是否为激活状态。
func (s UserStatus) IsActive() bool {
	return s == UserStatusActive
}

// User 用户领域实体。
type User struct {
	ID           int64      `json:"id"`
	Username     string     `json:"username"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"` // 不序列化密码
	Role         UserRole   `json:"role"`
	Status       UserStatus `json:"status"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}

// IsAdmin 判断用户是否为管理员。
func (u *User) IsAdmin() bool {
	return u.Role.IsAdmin()
}

// CanLogin 判断用户是否可以登录。
func (u *User) CanLogin() bool {
	return u.Status.IsActive()
}
