// Package dao 提供数据访问对象 (DAO) 接口和模型。
package dao

import (
	"context"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// TenantStatus 定义租户状态。
type TenantStatus string

const (
	TenantStatusActive    TenantStatus = "active"
	TenantStatusSuspended TenantStatus = "suspended"
	TenantStatusDisabled  TenantStatus = "disabled"
)

// TenantPlan 定义租户订阅计划。
type TenantPlan string

const (
	TenantPlanFree       TenantPlan = "free"
	TenantPlanPro        TenantPlan = "pro"
	TenantPlanEnterprise TenantPlan = "enterprise"
)

// Tenant 是租户的数据库模型。
type Tenant struct {
	ID                 int64        `gorm:"primaryKey;autoIncrement"`
	Name               string       `gorm:"uniqueIndex;size:64;not null"`
	Slug               string       `gorm:"uniqueIndex;size:32;not null"`
	Status             TenantStatus `gorm:"type:enum('active','suspended','disabled');default:'active'"`
	Plan               TenantPlan   `gorm:"type:enum('free','pro','enterprise');default:'free'"`
	QuotaTokensMonthly int64        `gorm:"default:1000000"`
	QuotaRequestsDaily int          `gorm:"default:1000"`
	Settings           JSON         `gorm:"type:json"`
	CreatedAt          time.Time    `gorm:"autoCreateTime"`
	UpdatedAt          time.Time    `gorm:"autoUpdateTime"`
}

// JSON 用于存储 JSON 数据。
type JSON json.RawMessage

// Scan 实现 sql.Scanner 接口。
func (j *JSON) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	*j = bytes
	return nil
}

// Value 实现 driver.Valuer 接口。
func (j JSON) Value() (interface{}, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return []byte(j), nil
}

// TableName 返回 Tenant 的表名。
func (Tenant) TableName() string {
	return "tenants"
}

// TenantDAO 定义租户的数据访问操作。
type TenantDAO interface {
	Create(ctx context.Context, tenant *Tenant) error
	Update(ctx context.Context, tenant *Tenant) error
	Delete(ctx context.Context, id int64) error
	GetByID(ctx context.Context, id int64) (*Tenant, error)
	GetBySlug(ctx context.Context, slug string) (*Tenant, error)
	List(ctx context.Context) ([]Tenant, error)
}

// GormTenantDAO 是 TenantDAO 的 GORM 实现。
type GormTenantDAO struct {
	db *gorm.DB
}

// NewGormTenantDAO 创建一个新的基于 GORM 的 TenantDAO。
func NewGormTenantDAO(db *gorm.DB) TenantDAO {
	return &GormTenantDAO{db: db}
}

func (d *GormTenantDAO) Create(ctx context.Context, tenant *Tenant) error {
	return d.db.WithContext(ctx).Create(tenant).Error
}

func (d *GormTenantDAO) Update(ctx context.Context, tenant *Tenant) error {
	return d.db.WithContext(ctx).Save(tenant).Error
}

func (d *GormTenantDAO) Delete(ctx context.Context, id int64) error {
	return d.db.WithContext(ctx).Delete(&Tenant{}, id).Error
}

func (d *GormTenantDAO) GetByID(ctx context.Context, id int64) (*Tenant, error) {
	var tenant Tenant
	err := d.db.WithContext(ctx).First(&tenant, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &tenant, err
}

func (d *GormTenantDAO) GetBySlug(ctx context.Context, slug string) (*Tenant, error) {
	var tenant Tenant
	err := d.db.WithContext(ctx).Where("slug = ?", slug).First(&tenant).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &tenant, err
}

func (d *GormTenantDAO) List(ctx context.Context) ([]Tenant, error) {
	var tenants []Tenant
	err := d.db.WithContext(ctx).Find(&tenants).Error
	return tenants, err
}

var _ TenantDAO = (*GormTenantDAO)(nil)
