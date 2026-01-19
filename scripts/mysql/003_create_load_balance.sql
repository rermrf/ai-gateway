-- 003: Create load balance tables
CREATE TABLE IF NOT EXISTS load_balance_groups (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(64) NOT NULL UNIQUE COMMENT 'Group unique identifier',
    model_pattern VARCHAR(128) NOT NULL COMMENT 'Model pattern this group applies to',
    strategy ENUM('round-robin', 'random', 'failover', 'weighted') NOT NULL DEFAULT 'round-robin',
    enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_model_pattern (model_pattern),
    INDEX idx_enabled (enabled)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Load balance groups';

CREATE TABLE IF NOT EXISTS load_balance_members (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    group_id BIGINT UNSIGNED NOT NULL,
    provider_name VARCHAR(64) NOT NULL,
    weight INT UNSIGNED DEFAULT 1 COMMENT 'Weight for weighted strategy',
    priority INT DEFAULT 0 COMMENT 'Priority for failover strategy, lower = higher priority',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_group_id (group_id),
    FOREIGN KEY (group_id) REFERENCES load_balance_groups(id) ON DELETE CASCADE,
    FOREIGN KEY (provider_name) REFERENCES providers(name) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Load balance group members';
