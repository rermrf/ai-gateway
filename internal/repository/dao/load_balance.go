// Package dao 提供数据访问对象 (DAO) 接口和模型。
package dao

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// LoadBalanceGroup 是负载均衡组的数据库模型。
type LoadBalanceGroup struct {
	ID           int64     `gorm:"primaryKey;autoIncrement"`
	Name         string    `gorm:"uniqueIndex;size:64;not null"`
	ModelPattern string    `gorm:"size:128;not null;index"`
	Strategy     string    `gorm:"size:32;not null"` // round-robin, random, failover, weighted
	Enabled      bool      `gorm:"default:true;index"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"`
}

// TableName 返回 LoadBalanceGroup 的表名。
func (LoadBalanceGroup) TableName() string {
	return "load_balance_groups"
}

// LoadBalanceMember 是负载均衡成员的数据库模型。
type LoadBalanceMember struct {
	ID           int64     `gorm:"primaryKey;autoIncrement"`
	GroupID      int64     `gorm:"not null;index"`
	ProviderName string    `gorm:"size:64;not null"`
	Weight       int       `gorm:"default:1"`
	Priority     int       `gorm:"default:0"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
}

// TableName 返回 LoadBalanceMember 的表名。
func (LoadBalanceMember) TableName() string {
	return "load_balance_members"
}

// LoadBalanceDAO 定义负载均衡的数据访问操作。
type LoadBalanceDAO interface {
	CreateGroup(ctx context.Context, group *LoadBalanceGroup) error
	UpdateGroup(ctx context.Context, group *LoadBalanceGroup) error
	DeleteGroup(ctx context.Context, id int64) error
	GetGroupByID(ctx context.Context, id int64) (*LoadBalanceGroup, error)
	GetGroupByPattern(ctx context.Context, pattern string) (*LoadBalanceGroup, error)
	ListGroups(ctx context.Context) ([]LoadBalanceGroup, error)
	
	AddMember(ctx context.Context, member *LoadBalanceMember) error
	RemoveMember(ctx context.Context, id int64) error
	GetMembers(ctx context.Context, groupID int64) ([]LoadBalanceMember, error)
}

// GormLoadBalanceDAO 是 LoadBalanceDAO 的 GORM 实现。
type GormLoadBalanceDAO struct {
	db *gorm.DB
}

// NewGormLoadBalanceDAO 创建一个新的基于 GORM 的 LoadBalanceDAO。
func NewGormLoadBalanceDAO(db *gorm.DB) LoadBalanceDAO {
	return &GormLoadBalanceDAO{db: db}
}

func (d *GormLoadBalanceDAO) CreateGroup(ctx context.Context, g *LoadBalanceGroup) error {
	return d.db.WithContext(ctx).Create(g).Error
}

func (d *GormLoadBalanceDAO) UpdateGroup(ctx context.Context, g *LoadBalanceGroup) error {
	return d.db.WithContext(ctx).Save(g).Error
}

func (d *GormLoadBalanceDAO) DeleteGroup(ctx context.Context, id int64) error {
	d.db.WithContext(ctx).Where("group_id = ?", id).Delete(&LoadBalanceMember{})
	return d.db.WithContext(ctx).Delete(&LoadBalanceGroup{}, id).Error
}

func (d *GormLoadBalanceDAO) GetGroupByID(ctx context.Context, id int64) (*LoadBalanceGroup, error) {
	var g LoadBalanceGroup
	err := d.db.WithContext(ctx).First(&g, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &g, err
}

func (d *GormLoadBalanceDAO) GetGroupByPattern(ctx context.Context, pattern string) (*LoadBalanceGroup, error) {
	var g LoadBalanceGroup
	err := d.db.WithContext(ctx).
		Where("model_pattern = ? AND enabled = ?", pattern, true).
		First(&g).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &g, err
}

func (d *GormLoadBalanceDAO) ListGroups(ctx context.Context) ([]LoadBalanceGroup, error) {
	var groups []LoadBalanceGroup
	err := d.db.WithContext(ctx).Where("enabled = ?", true).Find(&groups).Error
	return groups, err
}

func (d *GormLoadBalanceDAO) AddMember(ctx context.Context, m *LoadBalanceMember) error {
	return d.db.WithContext(ctx).Create(m).Error
}

func (d *GormLoadBalanceDAO) RemoveMember(ctx context.Context, id int64) error {
	return d.db.WithContext(ctx).Delete(&LoadBalanceMember{}, id).Error
}

func (d *GormLoadBalanceDAO) GetMembers(ctx context.Context, groupID int64) ([]LoadBalanceMember, error) {
	var members []LoadBalanceMember
	err := d.db.WithContext(ctx).
		Where("group_id = ?", groupID).
		Order("priority ASC").
		Find(&members).Error
	return members, err
}

var _ LoadBalanceDAO = (*GormLoadBalanceDAO)(nil)
