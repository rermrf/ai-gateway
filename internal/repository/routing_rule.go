// Package repository 定义数据访问的存储库接口。
package repository

import (
	"context"

	"ai-gateway/internal/repository/dao"
)

// RoutingRuleRepository 定义 RoutingRule 实体的存储库接口。
type RoutingRuleRepository interface {
	Create(ctx context.Context, rule *dao.RoutingRule) error
	Update(ctx context.Context, rule *dao.RoutingRule) error
	Delete(ctx context.Context, id int64) error
	GetByID(ctx context.Context, id int64) (*dao.RoutingRule, error)
	List(ctx context.Context) ([]dao.RoutingRule, error)
	FindByPattern(ctx context.Context, pattern string) (*dao.RoutingRule, error)
	FindPrefixRules(ctx context.Context) ([]dao.RoutingRule, error)
}

// routingRuleRepository 是 RoutingRuleRepository 的默认实现。
type routingRuleRepository struct {
	dao dao.RoutingRuleDAO
}

// NewRoutingRuleRepository 创建一个新的 RoutingRuleRepository。
func NewRoutingRuleRepository(routingRuleDAO dao.RoutingRuleDAO) RoutingRuleRepository {
	return &routingRuleRepository{dao: routingRuleDAO}
}

func (r *routingRuleRepository) Create(ctx context.Context, rule *dao.RoutingRule) error {
	return r.dao.Create(ctx, rule)
}

func (r *routingRuleRepository) Update(ctx context.Context, rule *dao.RoutingRule) error {
	return r.dao.Update(ctx, rule)
}

func (r *routingRuleRepository) Delete(ctx context.Context, id int64) error {
	return r.dao.Delete(ctx, id)
}

func (r *routingRuleRepository) GetByID(ctx context.Context, id int64) (*dao.RoutingRule, error) {
	return r.dao.GetByID(ctx, id)
}

func (r *routingRuleRepository) List(ctx context.Context) ([]dao.RoutingRule, error) {
	return r.dao.List(ctx)
}

func (r *routingRuleRepository) FindByPattern(ctx context.Context, pattern string) (*dao.RoutingRule, error) {
	return r.dao.FindByPattern(ctx, pattern)
}

func (r *routingRuleRepository) FindPrefixRules(ctx context.Context) ([]dao.RoutingRule, error) {
	return r.dao.FindPrefixRules(ctx)
}
