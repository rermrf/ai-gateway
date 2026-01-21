// Package repository 定义数据访问的存储库接口。
package repository

import (
	"context"

	"ai-gateway/internal/domain"
	"ai-gateway/internal/repository/dao"
)

// LoadBalanceRepository 定义负载均衡的存储库接口。
type LoadBalanceRepository interface {
	CreateGroup(ctx context.Context, group *domain.LoadBalanceGroup) error
	UpdateGroup(ctx context.Context, group *domain.LoadBalanceGroup) error
	DeleteGroup(ctx context.Context, id int64) error
	GetGroupByID(ctx context.Context, id int64) (*domain.LoadBalanceGroup, error)
	GetGroupByPattern(ctx context.Context, pattern string) (*domain.LoadBalanceGroup, error)
	ListGroups(ctx context.Context) ([]domain.LoadBalanceGroup, error)

	AddMember(ctx context.Context, member *domain.LoadBalanceMember) error
	RemoveMember(ctx context.Context, id int64) error
	GetMembers(ctx context.Context, groupID int64) ([]domain.LoadBalanceMember, error)
}

// loadBalanceRepository 是 LoadBalanceRepository 的默认实现。
type loadBalanceRepository struct {
	dao dao.LoadBalanceDAO
}

// NewLoadBalanceRepository 创建一个新的 LoadBalanceRepository。
func NewLoadBalanceRepository(loadBalanceDAO dao.LoadBalanceDAO) LoadBalanceRepository {
	return &loadBalanceRepository{dao: loadBalanceDAO}
}

// toDAOGroup 将 domain.LoadBalanceGroup 转换为 dao.LoadBalanceGroup
func (r *loadBalanceRepository) toDAOGroup(g *domain.LoadBalanceGroup) *dao.LoadBalanceGroup {
	return &dao.LoadBalanceGroup{
		ID:           g.ID,
		Name:         g.Name,
		ModelPattern: g.ModelPattern,
		Strategy:     g.Strategy,
		Enabled:      g.Enabled,
		CreatedAt:    g.CreatedAt,
		UpdatedAt:    g.UpdatedAt,
	}
}

// toDomainGroup 将 dao.LoadBalanceGroup 转换为 domain.LoadBalanceGroup
func (r *loadBalanceRepository) toDomainGroup(g *dao.LoadBalanceGroup) *domain.LoadBalanceGroup {
	if g == nil {
		return nil
	}
	return &domain.LoadBalanceGroup{
		ID:           g.ID,
		Name:         g.Name,
		ModelPattern: g.ModelPattern,
		Strategy:     g.Strategy,
		Enabled:      g.Enabled,
		CreatedAt:    g.CreatedAt,
		UpdatedAt:    g.UpdatedAt,
	}
}

// toDAOMember 将 domain.LoadBalanceMember 转换为 dao.LoadBalanceMember
func (r *loadBalanceRepository) toDAOMember(m *domain.LoadBalanceMember) *dao.LoadBalanceMember {
	return &dao.LoadBalanceMember{
		ID:           m.ID,
		GroupID:      m.GroupID,
		ProviderName: m.ProviderName,
		Weight:       m.Weight,
		Priority:     m.Priority,
		CreatedAt:    m.CreatedAt,
	}
}

// toDomainMember 将 dao.LoadBalanceMember 转换为 domain.LoadBalanceMember
func (r *loadBalanceRepository) toDomainMember(m *dao.LoadBalanceMember) *domain.LoadBalanceMember {
	if m == nil {
		return nil
	}
	return &domain.LoadBalanceMember{
		ID:           m.ID,
		GroupID:      m.GroupID,
		ProviderName: m.ProviderName,
		Weight:       m.Weight,
		Priority:     m.Priority,
		CreatedAt:    m.CreatedAt,
	}
}

func (r *loadBalanceRepository) CreateGroup(ctx context.Context, g *domain.LoadBalanceGroup) error {
	daoGroup := r.toDAOGroup(g)
	if err := r.dao.CreateGroup(ctx, daoGroup); err != nil {
		return err
	}
	g.ID = daoGroup.ID
	g.CreatedAt = daoGroup.CreatedAt
	g.UpdatedAt = daoGroup.UpdatedAt
	return nil
}

func (r *loadBalanceRepository) UpdateGroup(ctx context.Context, g *domain.LoadBalanceGroup) error {
	return r.dao.UpdateGroup(ctx, r.toDAOGroup(g))
}

func (r *loadBalanceRepository) DeleteGroup(ctx context.Context, id int64) error {
	return r.dao.DeleteGroup(ctx, id)
}

func (r *loadBalanceRepository) GetGroupByID(ctx context.Context, id int64) (*domain.LoadBalanceGroup, error) {
	daoGroup, err := r.dao.GetGroupByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return r.toDomainGroup(daoGroup), nil
}

func (r *loadBalanceRepository) GetGroupByPattern(ctx context.Context, pattern string) (*domain.LoadBalanceGroup, error) {
	daoGroup, err := r.dao.GetGroupByPattern(ctx, pattern)
	if err != nil {
		return nil, err
	}
	return r.toDomainGroup(daoGroup), nil
}

func (r *loadBalanceRepository) ListGroups(ctx context.Context) ([]domain.LoadBalanceGroup, error) {
	daoGroups, err := r.dao.ListGroups(ctx)
	if err != nil {
		return nil, err
	}

	groups := make([]domain.LoadBalanceGroup, len(daoGroups))
	for i, g := range daoGroups {
		groups[i] = *r.toDomainGroup(&g)
	}
	return groups, nil
}

func (r *loadBalanceRepository) AddMember(ctx context.Context, m *domain.LoadBalanceMember) error {
	daoMember := r.toDAOMember(m)
	if err := r.dao.AddMember(ctx, daoMember); err != nil {
		return err
	}
	m.ID = daoMember.ID
	m.CreatedAt = daoMember.CreatedAt
	return nil
}

func (r *loadBalanceRepository) RemoveMember(ctx context.Context, id int64) error {
	return r.dao.RemoveMember(ctx, id)
}

func (r *loadBalanceRepository) GetMembers(ctx context.Context, groupID int64) ([]domain.LoadBalanceMember, error) {
	daoMembers, err := r.dao.GetMembers(ctx, groupID)
	if err != nil {
		return nil, err
	}

	members := make([]domain.LoadBalanceMember, len(daoMembers))
	for i, m := range daoMembers {
		members[i] = *r.toDomainMember(&m)
	}
	return members, nil
}
