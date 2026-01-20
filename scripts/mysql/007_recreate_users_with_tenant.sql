-- 007: 重构用户表支持多租户
-- 用户属于租户，在租户内有角色

-- 先删除旧表（如果需要保留数据，请先备份）
DROP TABLE IF EXISTS users;

CREATE TABLE users (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    tenant_id BIGINT UNSIGNED NOT NULL COMMENT '所属租户',
    username VARCHAR(64) NOT NULL COMMENT '用户名',
    email VARCHAR(128) NOT NULL COMMENT '邮箱',
    password_hash VARCHAR(256) COMMENT '密码哈希',
    role ENUM('owner', 'admin', 'member') DEFAULT 'member' COMMENT '租户内角色',
    status ENUM('active', 'disabled') DEFAULT 'active' COMMENT '用户状态',
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    -- 租户内用户名和邮箱唯一
    UNIQUE KEY uk_tenant_username (tenant_id, username),
    UNIQUE KEY uk_tenant_email (tenant_id, email),
    INDEX idx_tenant_id (tenant_id),
    INDEX idx_status (status),
    
    CONSTRAINT fk_users_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表（多租户）';

-- 创建平台管理员用户
INSERT INTO users (id, tenant_id, username, email, role, status)
VALUES (1, 1, 'admin', 'admin@platform.local', 'owner', 'active')
ON DUPLICATE KEY UPDATE username = 'admin';
