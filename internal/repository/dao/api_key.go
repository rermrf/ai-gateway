// Package dao 提供数据访问对象 (DAO) 接口和模型。
package dao

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// APIKey 是网关 API 密钥的数据库模型。
type APIKey struct {
	ID         int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID     int64      `gorm:"index;not null" json:"userId"`
	Key        string     `gorm:"uniqueIndex;size:128;not null" json:"key"` // 原始 Key (通常不存，但此处为了展示暂存)
	KeyHash    string     `gorm:"uniqueIndex;size:128;not null" json:"-"`   // Key 的哈希值，用于验证
	Name       string     `gorm:"size:64;not null" json:"name"`
	Enabled    bool       `gorm:"default:true;index" json:"enabled"`
	ExpiresAt  *time.Time `gorm:"default:null" json:"expiresAt,omitempty"`
	LastUsedAt *time.Time `gorm:"default:null" json:"lastUsedAt,omitempty"`
	CreatedAt  time.Time  `gorm:"autoCreateTime" json:"createdAt"`
}

// TableName 返回 APIKey 的表名。
func (APIKey) TableName() string {
	return "api_keys"
}

// APIKeyDAO 定义 API 密钥的数据访问操作。
type APIKeyDAO interface {
	Create(ctx context.Context, key *APIKey) error
	Update(ctx context.Context, key *APIKey) error
	Delete(ctx context.Context, id int64) error
	GetByID(ctx context.Context, id int64) (*APIKey, error)
	GetByKey(ctx context.Context, key string) (*APIKey, error)
	List(ctx context.Context) ([]APIKey, error)
	ListByUserID(ctx context.Context, userID int64) ([]APIKey, error)
	UpdateLastUsed(ctx context.Context, id int64) error
}

// GormAPIKeyDAO 是 APIKeyDAO 的 GORM 实现。
type GormAPIKeyDAO struct {
	db *gorm.DB
}

// NewGormAPIKeyDAO 创建一个新的基于 GORM 的 APIKeyDAO。
func NewGormAPIKeyDAO(db *gorm.DB) APIKeyDAO {
	return &GormAPIKeyDAO{db: db}
}

func (d *GormAPIKeyDAO) Create(ctx context.Context, k *APIKey) error {
	return d.db.WithContext(ctx).Create(k).Error
}

func (d *GormAPIKeyDAO) Update(ctx context.Context, k *APIKey) error {
	return d.db.WithContext(ctx).Save(k).Error
}

func (d *GormAPIKeyDAO) Delete(ctx context.Context, id int64) error {
	return d.db.WithContext(ctx).Delete(&APIKey{}, id).Error
}

func (d *GormAPIKeyDAO) GetByKey(ctx context.Context, key string) (*APIKey, error) {
	var k APIKey
	err := d.db.WithContext(ctx).Where("`key` = ?", key).First(&k).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &k, err
}

func (d *GormAPIKeyDAO) List(ctx context.Context) ([]APIKey, error) {
	var keys []APIKey
	err := d.db.WithContext(ctx).Find(&keys).Error
	return keys, err
}

func (d *GormAPIKeyDAO) GetByID(ctx context.Context, id int64) (*APIKey, error) {
	var k APIKey
	err := d.db.WithContext(ctx).First(&k, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &k, err
}

func (d *GormAPIKeyDAO) ListByUserID(ctx context.Context, userID int64) ([]APIKey, error) {
	var keys []APIKey
	err := d.db.WithContext(ctx).Where("user_id = ?", userID).Find(&keys).Error
	return keys, err
}

func (d *GormAPIKeyDAO) UpdateLastUsed(ctx context.Context, id int64) error {
	return d.db.WithContext(ctx).Model(&APIKey{}).Where("id = ?", id).Update("last_used_at", time.Now()).Error
}

var _ APIKeyDAO = (*GormAPIKeyDAO)(nil)
