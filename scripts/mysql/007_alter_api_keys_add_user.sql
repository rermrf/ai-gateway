-- 007: 为 api_keys 表添加用户关联
-- 每个 API Key 必须关联到一个用户

-- 添加 user_id 列，默认关联到系统用户(id=1)
ALTER TABLE api_keys 
ADD COLUMN IF NOT EXISTS user_id BIGINT UNSIGNED DEFAULT 1 AFTER id;

-- 添加外键约束（删除用户时级联删除其 API Key）
-- 注意：如果已存在约束，需要先删除
-- ALTER TABLE api_keys DROP FOREIGN KEY IF EXISTS fk_api_keys_user;
ALTER TABLE api_keys 
ADD CONSTRAINT fk_api_keys_user 
FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- 添加索引以优化按用户查询
ALTER TABLE api_keys ADD INDEX IF NOT EXISTS idx_user_id (user_id);
