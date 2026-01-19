-- 004: 创建 api_keys 表
CREATE TABLE IF NOT EXISTS api_keys (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `key` VARCHAR(128) NOT NULL UNIQUE COMMENT '网关 API 密钥',
    name VARCHAR(64) NOT NULL COMMENT '密钥描述/名称',
    enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NULL DEFAULT NULL COMMENT '可选的过期时间',
    
    INDEX idx_key (key),
    INDEX idx_enabled (enabled)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用于身份验证的网关 API 密钥';
