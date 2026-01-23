package domain

import "time"

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
