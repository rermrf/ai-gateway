// Package dao 提供数据访问对象 (DAO) 接口和模型。
package dao

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// Provider 是提供商配置的数据库模型。
type Provider struct {
	ID        int64     `gorm:"primaryKey;autoIncrement"`
	Name      string    `gorm:"uniqueIndex;size:64;not null"`
	Type      string    `gorm:"size:32;not null"` // openai, anthropic
	APIKey    string    `gorm:"size:512;not null"`
	BaseURL   string    `gorm:"size:256;not null"`
	Models    string    `gorm:"type:text;serializer:json"` // JSON encoded list of models
	TimeoutMs int       `gorm:"default:60000"`
	IsDefault bool      `gorm:"default:false"`
	Enabled   bool      `gorm:"default:true;index"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

// TableName 返回 Provider 的表名。
func (Provider) TableName() string {
	return "providers"
}

// ProviderDAO 定义 Provider 的数据访问操作。
type ProviderDAO interface {
	Create(ctx context.Context, provider *Provider) error
	Update(ctx context.Context, provider *Provider) error
	Delete(ctx context.Context, id int64) error
	GetByID(ctx context.Context, id int64) (*Provider, error)
	GetByName(ctx context.Context, name string) (*Provider, error)
	List(ctx context.Context) ([]Provider, error)
	GetDefaultByType(ctx context.Context, providerType string) (*Provider, error)
}

// GormProviderDAO 是 ProviderDAO 的 GORM 实现。
type GormProviderDAO struct {
	db *gorm.DB
}

// NewGormProviderDAO 创建一个新的基于 GORM 的 ProviderDAO。
func NewGormProviderDAO(db *gorm.DB) ProviderDAO {
	return &GormProviderDAO{db: db}
}

func (d *GormProviderDAO) Create(ctx context.Context, p *Provider) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if p.IsDefault {
			// Unset other defaults for this type
			if err := tx.Model(&Provider{}).Where("type = ? AND is_default = ?", p.Type, true).Update("is_default", false).Error; err != nil {
				return err
			}
		}
		return tx.Create(p).Error
	})
}

func (d *GormProviderDAO) Update(ctx context.Context, p *Provider) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if p.IsDefault {
			// Unset other defaults for this type
			// We exclude the current ID to avoid redundant update, though not strictly necessary as Save will overwrite
			if err := tx.Model(&Provider{}).Where("type = ? AND is_default = ? AND id != ?", p.Type, true, p.ID).Update("is_default", false).Error; err != nil {
				return err
			}
		}
		return tx.Save(p).Error
	})
}

func (d *GormProviderDAO) Delete(ctx context.Context, id int64) error {
	return d.db.WithContext(ctx).Delete(&Provider{}, id).Error
}

func (d *GormProviderDAO) GetByID(ctx context.Context, id int64) (*Provider, error) {
	var p Provider
	err := d.db.WithContext(ctx).First(&p, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &p, err
}

func (d *GormProviderDAO) GetByName(ctx context.Context, name string) (*Provider, error) {
	var p Provider
	err := d.db.WithContext(ctx).Where("name = ?", name).First(&p).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &p, err
}

func (d *GormProviderDAO) List(ctx context.Context) ([]Provider, error) {
	var providers []Provider
	err := d.db.WithContext(ctx).Where("enabled = ?", true).Find(&providers).Error
	return providers, err
}

func (d *GormProviderDAO) GetDefaultByType(ctx context.Context, providerType string) (*Provider, error) {
	var p Provider
	err := d.db.WithContext(ctx).
		Where("type = ? AND is_default = ? AND enabled = ?", providerType, true, true).
		First(&p).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &p, err
}

var _ ProviderDAO = (*GormProviderDAO)(nil)
