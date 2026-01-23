package dao

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// ModelRate 模型费率数据库模型
type ModelRate struct {
	ID              int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	ModelPattern    string    `gorm:"size:128;not null;uniqueIndex" json:"modelPattern"`
	PromptPrice     float64   `gorm:"type:decimal(20,8);default:0" json:"promptPrice"`
	CompletionPrice float64   `gorm:"type:decimal(20,8);default:0" json:"completionPrice"`
	Enabled         bool      `gorm:"default:true" json:"enabled"`
	CreatedAt       time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (ModelRate) TableName() string {
	return "model_rates"
}

// ModelRateDAO 模型费率 DAO 接口
type ModelRateDAO interface {
	Create(ctx context.Context, rate *ModelRate) error
	Update(ctx context.Context, rate *ModelRate) error
	Delete(ctx context.Context, id int64) error
	GetByID(ctx context.Context, id int64) (*ModelRate, error)
	List(ctx context.Context) ([]ModelRate, error)
	// FindMatch 查找匹配的费率（用于 Service 层逻辑，或者直接在 SQL 做）
	// 但通常我们在内存匹配，或者 SQL LIKE。这里提供获取所有 enabled 的方法
	GetAllEnabled(ctx context.Context) ([]ModelRate, error)
}

type GormModelRateDAO struct {
	db *gorm.DB
}

func NewGormModelRateDAO(db *gorm.DB) ModelRateDAO {
	return &GormModelRateDAO{db: db}
}

func (d *GormModelRateDAO) Create(ctx context.Context, rate *ModelRate) error {
	return d.db.WithContext(ctx).Create(rate).Error
}

func (d *GormModelRateDAO) Update(ctx context.Context, rate *ModelRate) error {
	return d.db.WithContext(ctx).Save(rate).Error
}

func (d *GormModelRateDAO) Delete(ctx context.Context, id int64) error {
	return d.db.WithContext(ctx).Delete(&ModelRate{}, id).Error
}

func (d *GormModelRateDAO) GetByID(ctx context.Context, id int64) (*ModelRate, error) {
	var rate ModelRate
	err := d.db.WithContext(ctx).First(&rate, id).Error
	if err != nil {
		return nil, err
	}
	return &rate, nil
}

func (d *GormModelRateDAO) List(ctx context.Context) ([]ModelRate, error) {
	var rates []ModelRate
	err := d.db.WithContext(ctx).Find(&rates).Error
	return rates, err
}

func (d *GormModelRateDAO) GetAllEnabled(ctx context.Context) ([]ModelRate, error) {
	var rates []ModelRate
	err := d.db.WithContext(ctx).Where("enabled = ?", true).Find(&rates).Error
	return rates, err
}
