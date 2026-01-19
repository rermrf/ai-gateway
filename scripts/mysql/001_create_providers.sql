-- 001: 创建 providers 表
CREATE TABLE IF NOT EXISTS providers (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(64) NOT NULL UNIQUE COMMENT '供应商唯一标识符，例如 siliconflow, openai-official',
    type VARCHAR(32) NOT NULL COMMENT '供应商类型：openai, anthropic',
    api_key VARCHAR(512) NOT NULL COMMENT '加密的 API 密钥',
    base_url VARCHAR(256) NOT NULL COMMENT 'API 基础 URL',
    timeout_ms INT UNSIGNED DEFAULT 60000 COMMENT '请求超时时间（毫秒）',
    is_default BOOLEAN DEFAULT FALSE COMMENT '此类供应商的默认供应商',
    enabled BOOLEAN DEFAULT TRUE COMMENT '此供应商是否已启用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_type (type),
    INDEX idx_enabled (enabled)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='供应商配置';
