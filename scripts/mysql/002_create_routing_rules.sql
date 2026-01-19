-- 002: Create routing_rules table
CREATE TABLE IF NOT EXISTS routing_rules (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    rule_type ENUM('exact', 'prefix', 'wildcard') NOT NULL COMMENT 'Rule matching type',
    pattern VARCHAR(128) NOT NULL COMMENT 'Model pattern to match, e.g., gpt-4o, deepseek-, gpt-*',
    provider_name VARCHAR(64) NOT NULL COMMENT 'Target provider name',
    actual_model VARCHAR(128) DEFAULT NULL COMMENT 'Optional: actual model name to use',
    priority INT DEFAULT 0 COMMENT 'Priority for rule matching, higher = higher priority',
    enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_rule_type (rule_type),
    INDEX idx_pattern (pattern),
    INDEX idx_enabled_priority (enabled, priority DESC),
    FOREIGN KEY (provider_name) REFERENCES providers(name) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Model routing rules';
