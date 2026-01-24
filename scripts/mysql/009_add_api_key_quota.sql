-- Add quota and used_amount to api_keys
ALTER TABLE api_keys ADD COLUMN quota DECIMAL(15,6) DEFAULT NULL COMMENT '额度限制(tokens/price，null=无限)';
ALTER TABLE api_keys ADD COLUMN used_amount DECIMAL(15,6) DEFAULT 0 COMMENT '已使用额度';
