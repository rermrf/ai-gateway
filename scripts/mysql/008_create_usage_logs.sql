-- 008: 创建使用记录表
-- 记录 API 调用和 Token 使用量

CREATE TABLE IF NOT EXISTS usage_logs (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL COMMENT '用户 ID',
    api_key_id BIGINT UNSIGNED COMMENT 'API Key ID',
    
    model VARCHAR(64) COMMENT '模型名称',
    provider VARCHAR(32) COMMENT '提供商',
    input_tokens INT DEFAULT 0 COMMENT '输入 Token 数',
    output_tokens INT DEFAULT 0 COMMENT '输出 Token 数',
    latency_ms INT COMMENT '响应延迟 (毫秒)',
    status_code INT COMMENT 'HTTP 状态码',
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_user_created (user_id, created_at),
    INDEX idx_api_key (api_key_id),
    INDEX idx_created_at (created_at),
    
    CONSTRAINT fk_usage_logs_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='使用记录表';
