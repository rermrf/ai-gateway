-- 插入示例供应商
INSERT INTO providers (name, type, api_key, base_url, timeout_ms, is_default, enabled) VALUES
('siliconflow', 'openai', 'sk-your-siliconflow-key', 'https://api.siliconflow.cn/v1', 60000, TRUE, TRUE),
('claude-proxy', 'anthropic', 'sk-your-claude-key', 'https://icat.pp.ua', 60000, TRUE, TRUE);

-- 插入示例路由规则
INSERT INTO routing_rules (rule_type, pattern, provider_name, actual_model, priority, enabled) VALUES
-- DeepSeek 模型的前缀路由
('prefix', 'deepseek', 'siliconflow', NULL, 10, TRUE),
-- Qwen 模型的前缀路由
('prefix', 'qwen', 'siliconflow', NULL, 10, TRUE),
-- Claude 模型的前缀路由
('prefix', 'claude', 'claude-proxy', NULL, 10, TRUE);

-- 可选：插入示例负载均衡组
-- INSERT INTO load_balance_groups (name, model_pattern, strategy, enabled) VALUES
-- ('gpt-4o-balancer', 'gpt-4o', 'weighted', TRUE);

-- INSERT INTO load_balance_members (group_id, provider_name, weight, priority) VALUES
-- (1, 'siliconflow', 2, 0),
-- (1, 'openrouter', 1, 0);
