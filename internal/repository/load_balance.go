// Package repository 定义数据访问的存储库接口。
package repository

import (
	"context"

	"ai-gateway/internal/repository/dao"
)

// LoadBalanceRepository 定义负载均衡的存储库接口。
type LoadBalanceRepository interface {
	CreateGroup(ctx context.Context, group *dao.LoadBalanceGroup) error
	UpdateGroup(ctx context.Context, group *dao.LoadBalanceGroup) error
	DeleteGroup(ctx context.Context, id int64) error
	GetGroupByID(ctx context.Context, id int64) (*dao.LoadBalanceGroup, error)
	GetGroupByPattern(ctx context.Context, pattern string) (*dao.LoadBalanceGroup, error)
	ListGroups(ctx context.Context) ([]dao.LoadBalanceGroup, error)
	
	AddMember(ctx context.Context, member *dao.LoadBalanceMember) error
	RemoveMember(ctx context.Context, id int64) error
	GetMembers(ctx context.Context, groupID int64) ([]dao.LoadBalanceMember, error)
}

// loadBalanceRepository 是 LoadBalanceRepository 的默认实现。
type loadBalanceRepository struct {
	dao dao.LoadBalanceDAO
}

// NewLoadBalanceRepository 创建一个新的 LoadBalanceRepository。
func NewLoadBalanceRepository(loadBalanceDAO dao.LoadBalanceDAO) LoadBalanceRepository {
	return &loadBalanceRepository{dao: loadBalanceDAO}
}

func (r *loadBalanceRepository) CreateGroup(ctx context.Context, group *dao.LoadBalanceGroup) error {
	return r.dao.CreateGroup(ctx, group)
}

func (r *loadBalanceRepository) UpdateGroup(ctx context.Context, group *dao.LoadBalanceGroup) error {
	return r.dao.UpdateGroup(ctx, group)
}

func (r *loadBalanceRepository) DeleteGroup(ctx context.Context, id int64) error {
	return r.dao.DeleteGroup(ctx, id)
}

func (r *loadBalanceRepository) GetGroupByID(ctx context.Context, id int64) (*dao.LoadBalanceGroup, error) {
	return r.dao.GetGroupByID(ctx, id)
}

func (r *loadBalanceRepository) GetGroupByPattern(ctx context.Context, pattern string) (*dao.LoadBalanceGroup, error) {
	return r.dao.GetGroupByPattern(ctx, pattern)
}

func (r *loadBalanceRepository) ListGroups(ctx context.Context) ([]dao.LoadBalanceGroup, error) {
	return r.dao.ListGroups(ctx)
}

func (r *loadBalanceRepository) AddMember(ctx context.Context, member *dao.LoadBalanceMember) error {
	return r.dao.AddMember(ctx, member)
}

func (r *loadBalanceRepository) RemoveMember(ctx context.Context, id int64) error {
	return r.dao.RemoveMember(ctx, id)
}

func (r *loadBalanceRepository) GetMembers(ctx context.Context, groupID int64) ([]dao.LoadBalanceMember, error) {
	return r.dao.GetMembers(ctx, groupID)
}
