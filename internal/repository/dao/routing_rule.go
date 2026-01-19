// Package dao 提供数据访问对象 (DAO) 接口和模型。
package dao

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// RoutingRule 是路由规则的数据库模型。
type RoutingRule struct {
	ID           int64     `gorm:"primaryKey;autoIncrement"`
	RuleType     string    `gorm:"size:16;not null;index"` // exact, prefix, wildcard
	Pattern      string    `gorm:"size:128;not null;index"`
	ProviderName string    `gorm:"size:64;not null"`
	ActualModel  string    `gorm:"size:128"`
	Priority     int       `gorm:"default:0;index"`
	Enabled      bool      `gorm:"default:true;index"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"`
}

// TableName 返回 RoutingRule 的表名。
func (RoutingRule) TableName() string {
	return "routing_rules"
}

// RoutingRuleDAO 定义 RoutingRule 的数据访问操作。
type RoutingRuleDAO interface {
	Create(ctx context.Context, rule *RoutingRule) error
	Update(ctx context.Context, rule *RoutingRule) error
	Delete(ctx context.Context, id int64) error
	GetByID(ctx context.Context, id int64) (*RoutingRule, error)
	List(ctx context.Context) ([]RoutingRule, error)
	FindByPattern(ctx context.Context, pattern string) (*RoutingRule, error)
	FindPrefixRules(ctx context.Context) ([]RoutingRule, error)
}

// GormRoutingRuleDAO 是 RoutingRuleDAO 的 GORM 实现。
type GormRoutingRuleDAO struct {
	db *gorm.DB
}

// NewGormRoutingRuleDAO 创建一个新的基于 GORM 的 RoutingRuleDAO。
func NewGormRoutingRuleDAO(db *gorm.DB) RoutingRuleDAO {
	return &GormRoutingRuleDAO{db: db}
}

func (d *GormRoutingRuleDAO) Create(ctx context.Context, r *RoutingRule) error {
	return d.db.WithContext(ctx).Create(r).Error
}

func (d *GormRoutingRuleDAO) Update(ctx context.Context, r *RoutingRule) error {
	return d.db.WithContext(ctx).Save(r).Error
}

func (d *GormRoutingRuleDAO) Delete(ctx context.Context, id int64) error {
	return d.db.WithContext(ctx).Delete(&RoutingRule{}, id).Error
}

func (d *GormRoutingRuleDAO) GetByID(ctx context.Context, id int64) (*RoutingRule, error) {
	var r RoutingRule
	err := d.db.WithContext(ctx).First(&r, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &r, err
}

func (d *GormRoutingRuleDAO) List(ctx context.Context) ([]RoutingRule, error) {
	var rules []RoutingRule
	err := d.db.WithContext(ctx).
		Where("enabled = ?", true).
		Order("priority DESC").
		Find(&rules).Error
	return rules, err
}

func (d *GormRoutingRuleDAO) FindByPattern(ctx context.Context, pattern string) (*RoutingRule, error) {
	var r RoutingRule
	err := d.db.WithContext(ctx).
		Where("pattern = ? AND rule_type = ? AND enabled = ?", pattern, "exact", true).
		First(&r).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &r, err
}

func (d *GormRoutingRuleDAO) FindPrefixRules(ctx context.Context) ([]RoutingRule, error) {
	var rules []RoutingRule
	err := d.db.WithContext(ctx).
		Where("rule_type = ? AND enabled = ?", "prefix", true).
		Order("priority DESC, LENGTH(pattern) DESC").
		Find(&rules).Error
	return rules, err
}

var _ RoutingRuleDAO = (*GormRoutingRuleDAO)(nil)
