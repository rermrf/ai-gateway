# AI Gateway

一个企业级的 AI 网关服务，提供 OpenAI 和 Anthropic 兼容接口，实现不同 LLM 提供商之间的协议双向转换，并提供完整的用户管理、成本控制和使用统计功能。

[![Go Version](https://img.shields.io/badge/Go-1.24.3-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

## ✨ 核心特性

### 🔄 协议转换
- **OpenAI ↔ Anthropic 双向兼容**
  - 使用 OpenAI SDK 访问 Claude 模型
  - 使用 Anthropic SDK 访问 GPT 模型
- **流式响应支持** (Server-Sent Events)
- **完整功能映射**
  - 工具/函数调用 (Tool/Function Calling)
  - 多模态支持 (图片/视觉输入)
  - JSON 模式输出 (Structured Output)
  - 扩展思考模式 (Extended Thinking - Anthropic)

### 🎯 智能路由
- **多级路由策略**
  - 精确匹配：直接指定模型到提供商的映射
  - 前缀匹配：支持通配符规则（如 `gpt-*` → OpenAI）
  - 负载均衡：多提供商自动分发
  - 自动检测：根据模型名自动选择提供商
- **灵活的负载均衡**
  - 轮询 (Round Robin)
  - 随机 (Random)
  - 加权轮询 (Weighted Round Robin)
  - 最少连接 (Least Connections)

### 💰 成本管理
- **钱包系统**
  - 用户余额管理
  - 充值/扣费记录
  - 交易历史查询
- **灵活的费率配置**
  - 按模型分别设置输入/输出价格
  - 支持不同用户不同费率
- **详细的使用统计**
  - Token 消耗记录
  - 按用户/模型/时间维度统计
  - 成本分析

### 👥 用户管理
- **完整的认证系统**
  - 用户注册/登录
  - JWT 身份验证
  - 角色权限控制（管理员/普通用户）
- **API Key 管理**
  - 自助创建/删除 API Key
  - Key 权限控制
  - 使用记录追踪

### 🎨 管理后台
- **React + TypeScript 前端**
  - Dashboard 概览
  - 提供商管理
  - 路由规则配置
  - 负载均衡设置
  - 用户管理
  - 费率配置
  - 使用统计报表

## 🚀 快速开始

### 前置要求

- Go 1.24.3+
- MySQL 8.0+
- Node.js 18+ (如需构建前端)

### 1. 克隆项目

```bash
git clone https://github.com/yourusername/ai-gateway.git
cd ai-gateway
```

**⚡ 快速启动流程：**

```bash
# 1. 初始化数据库
mysql -u root -p -e "CREATE DATABASE ai_gateway CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
mysql -u root -p ai_gateway < scripts/migrations/001_init.sql

# 2. 配置环境变量
cp .env.example .env
# 编辑 .env 文件，设置 DB_PASSWORD 和 JWT_SECRET

# 3. 启动服务
./scripts/start-with-env.sh
```

服务将在 `http://localhost:8081` 启动。访问管理后台进行配置。

---

**详细步骤：**

### 2. 初始化数据库

```bash
# 创建数据库
mysql -u root -p -e "CREATE DATABASE ai_gateway CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

# 运行迁移脚本
mysql -u root -p ai_gateway < scripts/migrations/001_init.sql
```

### 3. 配置服务

#### 🔒 方式一: 使用环境变量（推荐）

为了安全性，建议使用环境变量来配置敏感信息，如数据库密码和 JWT 密钥。

```bash
# 1. 创建环境变量配置文件
cp .env.example .env

# 2. 编辑 .env 文件，填入真��值
# 至少需要设置：
#   DB_PASSWORD=你的数据库密码
#   JWT_SECRET=你的JWT密钥（建议使用: openssl rand -base64 32 生成）
```

**支持的环境变量:**
- `DB_HOST`: 数据库主机（默认: localhost）
- `DB_PORT`: 数据库端口（默认: 3306）
- `DB_USER`: 数据库用户（默认: root）
- `DB_PASSWORD`: 数据库密码 ⚠️ **必填**
- `DB_NAME`: 数据库名称（默认: ai_gateway）
- `JWT_SECRET`: JWT 密钥 ⚠️ **必填**

> 💡 **安全提示**: 
> - `.env` 文件已在 `.gitignore` 中，不会被提交到版本控制
> - 生产环境建议使用密钥管理服务（AWS Secrets Manager, HashiCorp Vault 等）
> - 确保 `.env` 文件权限: `chmod 600 .env`

#### 方式二: 直接修改配置文件（仅用于开发测试）

编辑 `config/config.yaml`：

```yaml
# HTTP 服务器
http:
  addr: ":8081"
  readTimeout: 30s
  writeTimeout: 120s

# MySQL 数据库
mysql:
  host: "localhost"
  port: 3306
  user: "root"
  password: ""  # 留空，使用环境变量 DB_PASSWORD
  database: "ai_gateway"
  charset: "utf8mb4"
  maxIdle: 10
  maxOpen: 100

# 身份验证
auth:
  enabled: true
  jwtSecret: ""  # 留空，使用环境变量 JWT_SECRET
```

### 4. 配置提供商

通过 Admin API 或直接插入数据库：

```sql
-- 添加 OpenAI 提供商
INSERT INTO providers (name, type, api_key, base_url, is_default, enabled)
VALUES ('openai-main', 'openai', 'sk-your-openai-key', 'https://api.openai.com/v1', 1, 1);

-- 添加 Anthropic 提供商
INSERT INTO providers (name, type, api_key, base_url, enabled)
VALUES ('anthropic-main', 'anthropic', 'sk-ant-your-key', 'https://api.anthropic.com', 1);

-- 添加路由规则（可选，默认会自动检测）
INSERT INTO routing_rules (rule_type, pattern, provider_name, priority, enabled)
VALUES ('prefix', 'gpt-', 'openai-main', 10, 1),
       ('prefix', 'claude-', 'anthropic-main', 10, 1);
```

### 5. 启动服务

#### 🚀 推荐: 使用启动脚本（自动加载环境变量）

```bash
# 启动脚本会自动加载 .env 文件并验证必需的环境变量
./scripts/start-with-env.sh
```

#### 其他启动方式

```bash
# 方式一：手动设置环境变量后运行
export DB_PASSWORD="your_password"
export JWT_SECRET="your_jwt_secret"
go run cmd/server/main.go --config=./config/config.yaml

# 方式二：使用 Makefile
make run

# 方式三：构建后运行
make build
./bin/ai-gateway --config=./config/config.yaml
```

服务将在 `http://localhost:8081` 启动。

### 6. 创建用户和 API Key

```bash
# 注册用户
curl -X POST http://localhost:8081/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "your_username",
    "email": "your@email.com",
    "password": "your_password"
  }'

# 登录获取 JWT
curl -X POST http://localhost:8081/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "your_username",
    "password": "your_password"
  }'

# 创建 API Key
curl -X POST http://localhost:8081/api/keys \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "my-app-key"}'
```

## 📖 API 使用

### OpenAI 兼容接口

```bash
# 访问 GPT 模型
curl http://localhost:8081/v1/chat/completions \
  -H "Authorization: Bearer sk-your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'

# 使用 OpenAI 格式访问 Claude 模型
curl http://localhost:8081/v1/chat/completions \
  -H "Authorization: Bearer sk-your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-3-5-sonnet-20241022",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'

# 流式响应
curl http://localhost:8081/v1/chat/completions \
  -H "Authorization: Bearer sk-your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o",
    "stream": true,
    "messages": [{"role": "user", "content": "Tell me a story"}]
  }'
```

### Anthropic 兼容接口

```bash
# 访问 Claude 模型
curl http://localhost:8081/v1/messages \
  -H "x-api-key: sk-your-api-key" \
  -H "anthropic-version: 2023-06-01" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-3-5-sonnet-20241022",
    "max_tokens": 1024,
    "messages": [{"role": "user", "content": "Hello!"}]
  }'

# 使用 Anthropic 格式访问 GPT 模型
curl http://localhost:8081/v1/messages \
  -H "x-api-key: sk-your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o",
    "max_tokens": 1024,
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

### 高级功能

#### 工具调用 (Function Calling)

```bash
curl http://localhost:8081/v1/chat/completions \
  -H "Authorization: Bearer sk-your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o",
    "messages": [{"role": "user", "content": "What is the weather in Beijing?"}],
    "tools": [{
      "type": "function",
      "function": {
        "name": "get_weather",
        "description": "Get current weather",
        "parameters": {
          "type": "object",
          "properties": {
            "location": {"type": "string"}
          },
          "required": ["location"]
        }
      }
    }]
  }'
```

#### JSON 模式输出

```bash
curl http://localhost:8081/v1/chat/completions \
  -H "Authorization: Bearer sk-your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o",
    "messages": [{"role": "user", "content": "List 3 colors"}],
    "response_format": {"type": "json_object"}
  }'
```

## 🏗️ 架构设计

### 分层架构

```
┌─────────────────────────────────────┐
│      Web Admin (React + TS)        │  前端管理界面
└─────────────────────────────────────┘
              ↓ HTTP
┌─────────────────────────────────────┐
│      API Layer (HTTP Handler)       │  HTTP 请求处理
│  - OpenAI Handler                   │  - 请求验证
│  - Anthropic Handler                │  - 参数解析
│  - Admin Handler                    │  - 响应格式化
└─────────────────────────────────────┘
              ↓
┌─────────────────────────────────────┐
│       Converter Layer               │  协议转换
│  - OpenAI ↔ Domain                  │
│  - Anthropic ↔ Domain               │
└─────────────────────────────────────┘
              ↓
┌─────────────────────────────────────┐
│       Service Layer                 │  业务逻辑
│  - Gateway Service (路由/转发)      │
│  - User Service (用户管理)          │
│  - Auth Service (认证授权)          │
│  - Wallet Service (钱包管理)        │
│  - Usage Service (使用统计)         │
└─────────────────────────────────────┘
              ↓
┌─────────────────────────────────────┐
│      Repository Layer               │  数据访问
│  - DAO ↔ Domain 转换                │
│  - 数据库 CRUD                       │
└─────────────────────────────────────┘
              ↓
┌─────────────────────────────────────┐
│      Providers Layer                │  LLM 提供商
│  - OpenAI Provider                  │
│  - Anthropic Provider               │
└─────────────────────────────────────┘
```

### 核心组件

- **Domain Layer**: 领域模型，协议无关的统一数据结构
- **Converter**: 协议转换器，处理不同 API 格式之间的转换
- **Gateway Service**: 核心路由逻辑，负责选择提供商和转发请求
- **Repository**: 数据访问层，封装数据库操作
- **Provider**: 提供商适配器，封装与 LLM 服务的交互

## 📁 项目结构

```
ai-gateway/
├── cmd/
│   └── server/              # 应用入口
│       └── main.go
├── config/                  # 配置
│   ├── config.go            # 配置结构定义
│   └── config.yaml          # 配置文件
├── docs/                    # 文档
│   ├── design.md            # 设计文档
│   ├── requirements.md      # 需求分析
│   └── template.md          # 代码模板
├── internal/
│   ├── api/http/            # HTTP 层
│   │   ├── handler/         # 请求处理器
│   │   │   ├── openai.go    # OpenAI 接口
│   │   │   ├── anthropic.go # Anthropic 接口
│   │   │   ├── admin.go     # 管理接口
│   │   │   ├── auth.go      # 认证接口
│   │   │   └── user.go      # 用户接口
│   │   ├── middleware/      # 中间件
│   │   │   ├── auth.go      # 认证中间件
│   │   │   ├── cors.go      # CORS
│   │   │   └── logger.go    # 日志
│   │   └── server.go        # HTTP 服务器
│   ├── converter/           # 协议转换器
│   │   ├── openai.go
│   │   └── anthropic.go
│   ├── domain/              # 领域模型
│   │   ├── request.go       # 统一请求模型
│   │   ├── message.go       # 消息模型
│   │   ├── user.go          # 用户模型
│   │   ├── api_key.go       # API Key
│   │   ├── wallet.go        # 钱包
│   │   └── usage.go         # 使用记录
│   ├── errs/                # 错误定义
│   ├── ioc/                 # 依赖注入 (Wire)
│   ├── pkg/                 # 内部通用包
│   │   ├── loadbalancer/    # 负载均衡
│   │   └── hash/            # 哈希工具
│   ├── providers/           # LLM 提供商适配器
│   │   ├── provider.go      # Provider 接口
│   │   ├── openai/
│   │   └── anthropic/
│   ├── repository/          # 数据访问层
│   │   ├── dao/             # DAO 模型
│   │   ├── user.go
│   │   ├── api_key.go
│   │   ├── wallet.go
│   │   ├── provider.go
│   │   ├── routing_rule.go
│   │   └── load_balance.go
│   └── service/             # 业务逻辑层
│       ├── gateway/         # 网关服务
│       ├── user/            # 用户服务
│       ├── auth/            # 认证服务
│       ├── wallet/          # 钱包服务
│       └── usage/           # 使用统计
├── scripts/
│   └── migrations/          # 数据库迁移脚本
├── web/
│   └── admin/               # 管理后台前端
│       ├── src/
│       └── package.json
├── examples/                # 示例代码
├── go.mod
├── Makefile
└── README.md
```

## 🗄️ 数据库设计

### 核心表

- **users**: 用户表
- **api_keys**: API 密钥表
- **wallets**: 钱包余额表
- **wallet_transactions**: 钱包交易记录
- **usage_logs**: 使用记录表
- **providers**: LLM 提供商配置
- **routing_rules**: 路由规则
- **load_balance_groups**: 负载均衡组
- **load_balance_members**: 负载均衡成员
- **model_rates**: 模型费率配置

详细的表结构请参考 `scripts/migrations/001_init.sql`。

## 🔧 开发

### 依赖管理

```bash
# 安装依赖
make setup
# 或
go mod tidy
```

### 代码格式化

```bash
make fmt
```

### 代码检查

```bash
make lint
```

### 构建

```bash
# 构建二进制
make build

# 构建前端
cd web/admin
npm run build
```

### 依赖注入

项目使用 Google Wire 进行依赖注入，修改依赖后需要重新生成：

```bash
go generate ./internal/ioc/...
```

## 🚢 部署

### Docker 部署

```bash
# 构建镜像
docker build -t ai-gateway:latest .

# 运行
docker run -d \
  -p 8081:8081 \
  -e DB_HOST=mysql \
  -e DB_PASSWORD=yourpassword \
  -e JWT_SECRET=yoursecret \
  --name ai-gateway \
  ai-gateway:latest
```

### Docker Compose

```bash
docker-compose up -d
```

### 系统服务

```bash
# 复制服务文件
sudo cp scripts/systemd/ai-gateway.service /etc/systemd/system/

# 启动服务
sudo systemctl enable ai-gateway
sudo systemctl start ai-gateway
```

## 🛡️ 安全建议

1. **配置安全**
   - 不要在配置文件中使用明文密码
   - 使用环境变量或密钥管理服务
   - 定期轮换 JWT 密钥和 API Keys

2. **网络安全**
   - 使用 HTTPS (反向代理如 Nginx)
   - 配置防火墙规则
   - 启用速率限制

3. **数据库安全**
   - 使用专用数据库用户，限制权限
   - 启用 SSL 连接
   - 定期备份

4. **监控和审计**
   - 启用访问日志
   - 监控异常请求
   - 定期审查 API Key 使用情况

## 📊 性能优化

- **数据库连接池**: 根据负载调整 `maxIdle` 和 `maxOpen`
- **缓存**: 考虑使用 Redis 缓存热点数据（提供商配置、路由规则）
- **负载均衡**: 使用多个提供商实例分散请求
- **限流**: 实现请求限流和熔断机制

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 License

MIT License - 详见 [LICENSE](LICENSE) 文件

## 🔗 相关链接

- [OpenAI API 文档](https://platform.openai.com/docs/api-reference)
- [Anthropic API 文档](https://docs.anthropic.com/claude/reference)
- [项目设计文档](docs/design.md)

---

**如有问题或建议，欢迎提交 Issue！**

