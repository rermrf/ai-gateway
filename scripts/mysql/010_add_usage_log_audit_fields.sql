-- Add audit fields to usage_logs
ALTER TABLE usage_logs ADD COLUMN client_ip VARCHAR(45) DEFAULT NULL COMMENT '客户端 IP';
ALTER TABLE usage_logs ADD COLUMN user_agent VARCHAR(512) DEFAULT NULL COMMENT 'User-Agent';
ALTER TABLE usage_logs ADD COLUMN request_id VARCHAR(64) DEFAULT NULL COMMENT '请求追踪 ID';

ALTER TABLE usage_logs ADD INDEX idx_client_ip (client_ip);
