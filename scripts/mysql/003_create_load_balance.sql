-- 003: 创建负载均衡表
CREATE TABLE IF NOT EXISTS load_balance_groups (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(64) NOT NULL UNIQUE COMMENT '负载均衡组唯一标识符',
    model_pattern VARCHAR(128) NOT NULL COMMENT '此组适用的模型模式',
    strategy ENUM('round-robin', 'random', 'failover', 'weighted') NOT NULL DEFAULT 'round-robin' COMMENT '负载均衡策略',
    enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_model_pattern (model_pattern),
    INDEX idx_enabled (enabled)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='负载均衡组';

CREATE TABLE IF NOT EXISTS load_balance_members (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    group_id BIGINT UNSIGNED NOT NULL COMMENT '负载均衡组 ID',
    provider_name VARCHAR(64) NOT NULL COMMENT '供应商名称',
    weight INT UNSIGNED DEFAULT 1 COMMENT '权重策略的权重',
    priority INT DEFAULT 0 COMMENT '故障转移策略的优先级，数值越小优先级越高',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_group_id (group_id),
    FOREIGN KEY (group_id) REFERENCES load_balance_groups(id) ON DELETE CASCADE,
    FOREIGN KEY (provider_name) REFERENCES providers(name) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='负载均衡组成员';
