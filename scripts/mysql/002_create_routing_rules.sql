-- 002: 创建 routing_rules 表
CREATE TABLE IF NOT EXISTS routing_rules (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    rule_type ENUM('exact', 'prefix', 'wildcard') NOT NULL COMMENT '规则匹配类型',
    pattern VARCHAR(128) NOT NULL COMMENT '要匹配的模型模式，例如 gpt-4o, deepseek-, gpt-*',
    provider_name VARCHAR(64) NOT NULL COMMENT '目标供应商名称',
    actual_model VARCHAR(128) DEFAULT NULL COMMENT '可选：要使用的实际模型名称',
    priority INT DEFAULT 0 COMMENT '规则匹配的优先级，数值越大优先级越高',
    enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_rule_type (rule_type),
    INDEX idx_pattern (pattern),
    INDEX idx_enabled_priority (enabled, priority DESC),
    FOREIGN KEY (provider_name) REFERENCES providers(name) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='模型路由规则';
