-- 006: 创建租户表
-- 租户是多租户架构的顶级隔离单元

CREATE TABLE IF NOT EXISTS tenants (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(64) NOT NULL UNIQUE COMMENT '租户名称',
    slug VARCHAR(32) NOT NULL UNIQUE COMMENT '租户标识符 (URL friendly)',
    status ENUM('active', 'suspended', 'disabled') DEFAULT 'active' COMMENT '租户状态',
    plan ENUM('free', 'pro', 'enterprise') DEFAULT 'free' COMMENT '订阅计划',
    
    -- 配额设置
    quota_tokens_monthly BIGINT DEFAULT 1000000 COMMENT '每月 Token 配额',
    quota_requests_daily INT DEFAULT 1000 COMMENT '每日请求配额',
    
    -- 元数据
    settings JSON COMMENT '租户自定义设置',
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_slug (slug),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='租户表';

-- 创建默认租户（平台级）
INSERT INTO tenants (id, name, slug, status, plan, quota_tokens_monthly, quota_requests_daily)
VALUES (1, 'Platform Default', 'default', 'active', 'enterprise', -1, -1)
ON DUPLICATE KEY UPDATE name = 'Platform Default';
