// Package repository 定义数据访问的存储库接口。
package repository

import (
	"context"

	"ai-gateway/internal/domain"
	"ai-gateway/internal/repository/dao"
)

// RoutingRuleRepository 定义路由规则实体的存储库接口。
type RoutingRuleRepository interface {
	Create(ctx context.Context, rule *domain.RoutingRule) error
	Update(ctx context.Context, rule *domain.RoutingRule) error
	Delete(ctx context.Context, id int64) error
	GetByID(ctx context.Context, id int64) (*domain.RoutingRule, error)
	List(ctx context.Context) ([]domain.RoutingRule, error)
	FindByPattern(ctx context.Context, pattern string) (*domain.RoutingRule, error)
	FindPrefixRules(ctx context.Context) ([]domain.RoutingRule, error)
}

// routingRuleRepository 是 RoutingRuleRepository 的默认实现。
type routingRuleRepository struct {
	dao dao.RoutingRuleDAO
}

// NewRoutingRuleRepository 创建一个新的 RoutingRuleRepository。
func NewRoutingRuleRepository(routingRuleDAO dao.RoutingRuleDAO) RoutingRuleRepository {
	return &routingRuleRepository{dao: routingRuleDAO}
}

// toDAO 将 domain.RoutingRule 转换为 dao.RoutingRule
func (r *routingRuleRepository) toDAO(rule *domain.RoutingRule) *dao.RoutingRule {
	return &dao.RoutingRule{
		ID:           rule.ID,
		RuleType:     rule.RuleType,
		Pattern:      rule.Pattern,
		ProviderName: rule.ProviderName,
		ActualModel:  rule.ActualModel,
		Priority:     rule.Priority,
		Enabled:      rule.Enabled,
		CreatedAt:    rule.CreatedAt,
		UpdatedAt:    rule.UpdatedAt,
	}
}

// toDomain 将 dao.RoutingRule 转换为 domain.RoutingRule
func (r *routingRuleRepository) toDomain(rule *dao.RoutingRule) *domain.RoutingRule {
	if rule == nil {
		return nil
	}
	return &domain.RoutingRule{
		ID:           rule.ID,
		RuleType:     rule.RuleType,
		Pattern:      rule.Pattern,
		ProviderName: rule.ProviderName,
		ActualModel:  rule.ActualModel,
		Priority:     rule.Priority,
		Enabled:      rule.Enabled,
		CreatedAt:    rule.CreatedAt,
		UpdatedAt:    rule.UpdatedAt,
	}
}

func (r *routingRuleRepository) Create(ctx context.Context, rule *domain.RoutingRule) error {
	daoRule := r.toDAO(rule)
	if err := r.dao.Create(ctx, daoRule); err != nil {
		return err
	}
	rule.ID = daoRule.ID
	rule.CreatedAt = daoRule.CreatedAt
	rule.UpdatedAt = daoRule.UpdatedAt
	return nil
}

func (r *routingRuleRepository) Update(ctx context.Context, rule *domain.RoutingRule) error {
	return r.dao.Update(ctx, r.toDAO(rule))
}

func (r *routingRuleRepository) Delete(ctx context.Context, id int64) error {
	return r.dao.Delete(ctx, id)
}

func (r *routingRuleRepository) GetByID(ctx context.Context, id int64) (*domain.RoutingRule, error) {
	daoRule, err := r.dao.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return r.toDomain(daoRule), nil
}

func (r *routingRuleRepository) List(ctx context.Context) ([]domain.RoutingRule, error) {
	daoRules, err := r.dao.List(ctx)
	if err != nil {
		return nil, err
	}

	rules := make([]domain.RoutingRule, len(daoRules))
	for i, rule := range daoRules {
		rules[i] = *r.toDomain(&rule)
	}
	return rules, nil
}

func (r *routingRuleRepository) FindByPattern(ctx context.Context, pattern string) (*domain.RoutingRule, error) {
	daoRule, err := r.dao.FindByPattern(ctx, pattern)
	if err != nil {
		return nil, err
	}
	return r.toDomain(daoRule), nil
}

func (r *routingRuleRepository) FindPrefixRules(ctx context.Context) ([]domain.RoutingRule, error) {
	daoRules, err := r.dao.FindPrefixRules(ctx)
	if err != nil {
		return nil, err
	}

	rules := make([]domain.RoutingRule, len(daoRules))
	for i, rule := range daoRules {
		rules[i] = *r.toDomain(&rule)
	}
	return rules, nil
}
