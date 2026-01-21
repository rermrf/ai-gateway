-- 006: 创建用户表（简单版）
-- 用户系统：支持注册、登录、管理员角色

CREATE TABLE IF NOT EXISTS users (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(64) NOT NULL UNIQUE COMMENT '用户名',
    email VARCHAR(128) NOT NULL UNIQUE COMMENT '邮箱',
    password_hash VARCHAR(256) NOT NULL COMMENT '密码哈希 (bcrypt)',
    role ENUM('user', 'admin') DEFAULT 'user' COMMENT '角色',
    status ENUM('active', 'disabled') DEFAULT 'active' COMMENT '状态',
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_email (email),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- 创建默认管理员用户（密码: admin123，实际部署时需修改）
-- bcrypt hash of 'admin123'
INSERT INTO users (id, username, email, password_hash, role, status)
VALUES (1, 'admin', 'admin@localhost', '$2a$10$N9qo8uLOickgx2ZMRZoMye1Oddj1x7HOfLkckm7zJz7Yd5L5V5a.q', 'admin', 'active')
ON DUPLICATE KEY UPDATE username = 'admin';
