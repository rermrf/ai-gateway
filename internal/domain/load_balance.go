package domain

import "time"

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
