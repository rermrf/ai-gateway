// Package domain 定义领域模型和业务实体。
package domain

import "time"

// Provider 提供商领域实体。
type Provider struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"` // openai, anthropic
	APIKey    string    `json:"apiKey"`
	BaseURL   string    `json:"baseURL"`
	TimeoutMs int       `json:"timeoutMs"`
	IsDefault bool      `json:"isDefault"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// RoutingRule 路由规则领域实体。
type RoutingRule struct {
	ID           int64     `json:"id"`
	RuleType     string    `json:"ruleType"` // exact, prefix, wildcard
	Pattern      string    `json:"pattern"`
	ProviderName string    `json:"providerName"`
	ActualModel  string    `json:"actualModel"`
	Priority     int       `json:"priority"`
	Enabled      bool      `json:"enabled"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// LoadBalanceGroup 负载均衡组领域实体。
type LoadBalanceGroup struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	ModelPattern string    `json:"modelPattern"`
	Strategy     string    `json:"strategy"` // round-robin, random, failover, weighted
	Enabled      bool      `json:"enabled"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// LoadBalanceMember 负载均衡成员领域实体。
type LoadBalanceMember struct {
	ID           int64     `json:"id"`
	GroupID      int64     `json:"groupId"`
	ProviderName string    `json:"providerName"`
	Weight       int       `json:"weight"`
	Priority     int       `json:"priority"`
	CreatedAt    time.Time `json:"createdAt"`
}
