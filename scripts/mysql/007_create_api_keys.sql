-- 007: 更新 API Keys 表
-- 关联用户，支持自助管理

-- 如果存在旧表则删除
DROP TABLE IF EXISTS api_keys;

CREATE TABLE api_keys (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL COMMENT '所属用户',
    `key` VARCHAR(128) NOT NULL UNIQUE COMMENT 'API 密钥',
    name VARCHAR(64) NOT NULL COMMENT '密钥名称/描述',
    
    enabled BOOLEAN DEFAULT TRUE COMMENT '是否启用',
    expires_at TIMESTAMP NULL DEFAULT NULL COMMENT '过期时间',
    last_used_at TIMESTAMP NULL COMMENT '最后使用时间',
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_key (`key`),
    INDEX idx_user_id (user_id),
    INDEX idx_enabled (enabled),
    
    CONSTRAINT fk_api_keys_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='API 密钥表';
