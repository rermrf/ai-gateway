# AI Gateway

一个通用的 AI 网关服务，提供标准的 OpenAI 和 Anthropic 兼容接口，实现不同 LLM 提供商之间的协议转换。

## 功能特性

- ✅ **双向协议兼容**
  - 使用 OpenAI SDK 访问 Claude 模型
  - 使用 Claude SDK 访问 OpenAI 模型
- ✅ **流式响应支持** (Server-Sent Events)
- ✅ **工具/函数调用** (Tool Calling)
- ✅ **多模态支持** (图片/视觉)
- ✅ **思考/推理模式** (Thinking)
- ✅ **灵活的模型路由**

## 快速开始

### 1. 安装依赖

```bash
make setup
# 或
go mod tidy
```

### 2. 配置 API Keys

通过环境变量设置：

```bash
# API Keys
export OPENAI_API_KEY="sk-xxx"
export ANTHROPIC_API_KEY="sk-ant-xxx"

# 可选：自定义 Base URL（用于第三方代理或自建服务）
export OPENAI_BASE_URL="https://your-openai-proxy.com/v1"
export ANTHROPIC_BASE_URL="https://your-anthropic-proxy.com"
```

或者编辑 `config/config.yaml` 文件。

### 3. 运行服务

```bash
make run
# 或
go run cmd/server/main.go
```

服务默认运行在 `http://localhost:8080`。

## API 使用

### OpenAI 兼容接口

```bash
# 使用 OpenAI 格式访问 GPT 模型
curl http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'

# 使用 OpenAI 格式访问 Claude 模型
curl http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-3-5-sonnet-20241022",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

### Anthropic 兼容接口

```bash
# 使用 Claude 格式访问 Claude 模型
curl http://localhost:8080/v1/messages \
  -H "x-api-key: your-api-key" \
  -H "anthropic-version: 2023-06-01" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-3-5-sonnet-20241022",
    "max_tokens": 1024,
    "messages": [{"role": "user", "content": "Hello!"}]
  }'

# 使用 Claude 格式访问 GPT 模型
curl http://localhost:8080/v1/messages \
  -H "x-api-key: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o",
    "max_tokens": 1024,
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

### 流式响应

在请求中添加 `"stream": true` 即可获得流式响应：

```bash
curl http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o",
    "stream": true,
    "messages": [{"role": "user", "content": "Tell me a story"}]
  }'
```

## 项目结构

```
ai-gateway/
├── cmd/
│   └── server/           # 应用入口
│       └── main.go
├── config/               # 配置
│   ├── config.go         # 配置结构定义
│   └── config.yaml       # 配置文件
├── docs/                 # 文档
│   ├── design.md         # 设计文档
│   └── requirements.md   # 需求分析
├── internal/
│   ├── api/http/         # HTTP 层
│   │   ├── handler/      # 请求处理器
│   │   ├── middleware/   # 中间件
│   │   └── server.go     # HTTP 服务器
│   ├── converter/        # 协议转换器
│   ├── domain/           # 领域模型
│   ├── errs/             # 错误定义
│   ├── ioc/              # 依赖注入
│   ├── pkg/              # 内部通用包
│   ├── providers/        # LLM 适配器
│   │   ├── openai/
│   │   └── anthropic/
│   └── service/          # 业务逻辑
│       └── gateway/
├── go.mod
├── Makefile
└── README.md
```

## 模型路由

网关会根据模型名称自动检测使用哪个提供商：
- `gpt-*`, `o1-*` → OpenAI
- `claude-*` → Anthropic

也可以在 `config/config.yaml` 中配置自定义路由：

```yaml
models:
  routing:
    "my-custom-model":
      provider: "openai"
      actualModel: "gpt-4-turbo"
```

## 开发

```bash
# 格式化代码
make fmt

# 运行测试
make test

# 运行 linter
make lint

# 构建
make build
```

## License

MIT
