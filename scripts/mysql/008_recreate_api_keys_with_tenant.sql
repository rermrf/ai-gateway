-- 008: 重构 API Keys 表支持多租户
-- API Key 属于租户，可选关联用户

-- 先删除旧表（如果需要保留数据，请先备份）
DROP TABLE IF EXISTS api_keys;

CREATE TABLE api_keys (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    tenant_id BIGINT UNSIGNED NOT NULL COMMENT '所属租户',
    user_id BIGINT UNSIGNED NULL COMMENT '关联用户 (NULL 表示租户级共享 Key)',
    `key` VARCHAR(128) NOT NULL UNIQUE COMMENT 'API 密钥',
    name VARCHAR(64) NOT NULL COMMENT '密钥名称/描述',
    
    -- 权限与限制
    allowed_models JSON COMMENT '允许的模型列表, NULL 表示全部允许',
    rate_limit_rpm INT DEFAULT 60 COMMENT '每分钟请求限制',
    
    enabled BOOLEAN DEFAULT TRUE COMMENT '是否启用',
    expires_at TIMESTAMP NULL DEFAULT NULL COMMENT '过期时间',
    last_used_at TIMESTAMP NULL COMMENT '最后使用时间',
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_key (`key`),
    INDEX idx_tenant_id (tenant_id),
    INDEX idx_user_id (user_id),
    INDEX idx_enabled (enabled),
    
    CONSTRAINT fk_api_keys_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    CONSTRAINT fk_api_keys_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='API 密钥表（多租户）';
