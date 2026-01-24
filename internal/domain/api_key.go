// Package domain 定义领域模型和业务实体。
package domain

import (
	"time"
)

// APIKey API 密钥领域实体。
type APIKey struct {
	ID         int64      `json:"id"`
	UserID     int64      `json:"userId"`
	Key        string     `json:"key"`
	KeyHash    string     `json:"-"`
	Name       string     `json:"name"`
	Enabled    bool       `json:"enabled"`
	Quota      *float64   `json:"quota"` // 额度限制(nil=无限)
	UsedAmount float64    `json:"usedAmount"`
	ExpiresAt  *time.Time `json:"expiresAt,omitempty"`
	LastUsedAt *time.Time `json:"lastUsedAt,omitempty"`
	CreatedAt  time.Time  `json:"createdAt"`
}

// IsValid 判断 API Key 是否有效。
func (k *APIKey) IsValid() bool {
	if !k.Enabled {
		return false
	}
	if k.ExpiresAt != nil && k.ExpiresAt.Before(time.Now()) {
		return false
	}
	if k.IsQuotaExceeded() {
		return false
	}
	return true
}

// HasQuota 是否有额度限制。
func (k *APIKey) HasQuota() bool {
	return k.Quota != nil
}

// IsQuotaExceeded 是否超过额度限制。
func (k *APIKey) IsQuotaExceeded() bool {
	if !k.HasQuota() {
		return false
	}
	return k.UsedAmount >= *k.Quota
}

// MaskKey 返回脱敏后的 Key。
func (k *APIKey) MaskKey() string {
	if len(k.Key) <= 8 {
		return "****"
	}
	return k.Key[:4] + "****" + k.Key[len(k.Key)-4:]
}
