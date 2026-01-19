-- 001: Create providers table
CREATE TABLE IF NOT EXISTS providers (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(64) NOT NULL UNIQUE COMMENT 'Provider unique identifier, e.g., siliconflow, openai-official',
    type VARCHAR(32) NOT NULL COMMENT 'Provider type: openai, anthropic',
    api_key VARCHAR(512) NOT NULL COMMENT 'Encrypted API key',
    base_url VARCHAR(256) NOT NULL COMMENT 'API base URL',
    timeout_ms INT UNSIGNED DEFAULT 60000 COMMENT 'Request timeout in milliseconds',
    is_default BOOLEAN DEFAULT FALSE COMMENT 'Default provider for this type',
    enabled BOOLEAN DEFAULT TRUE COMMENT 'Whether this provider is enabled',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_type (type),
    INDEX idx_enabled (enabled)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Provider configurations';
