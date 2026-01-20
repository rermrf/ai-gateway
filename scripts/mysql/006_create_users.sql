-- 006: 创建 users 表
-- 用户表，支持个人用户和部门

CREATE TABLE IF NOT EXISTS users (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(64) NOT NULL UNIQUE COMMENT '用户名/部门名',
    email VARCHAR(128) COMMENT '邮箱地址',
    password_hash VARCHAR(256) COMMENT '密码哈希（可选，用于管理后台登录）',
    user_type ENUM('personal', 'department', 'system') DEFAULT 'personal' COMMENT '用户类型：个人/部门/系统',
    department_name VARCHAR(64) COMMENT '所属部门名称（用于个人用户）',
    status ENUM('active', 'disabled') DEFAULT 'active' COMMENT '账户状态',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_username (username),
    INDEX idx_status (status),
    INDEX idx_user_type (user_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户/部门表';

-- 创建系统默认用户，用于兼容现有无归属的 API Key
INSERT INTO users (id, username, user_type, status) 
VALUES (1, 'system', 'system', 'active')
ON DUPLICATE KEY UPDATE username = 'system';
