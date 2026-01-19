# AI Gateway 需求分析与模块详述 (Requirements & Module Analysis)

基于 `design.md` 的架构设计，本文档详细拆解 AI 网关的功能需求、核心模块职责以及实施路径。

## 1. 核心功能需求 (Functional Requirements)

系统的核心价值在于**协议的“任意门”**：任何客户端（OpenAI/Claude）都可以访问任何服务端（OpenAI/Claude），且感知不到差异。

### 1.1 协议互通矩阵
| 客户端协议 (Client SDK) | 目标模型 (Provider) | 需求描述 | 关键难点 |
| :--- | :--- | :--- | :--- |
| **OpenAI** | **OpenAI** | 透明透传 | 鉴权转发，流式响应保持 |
| **OpenAI** | **Claude** | **转换核心** | `messages` -> `system`+`messages`转换；SSE流格式重写 |
| **Claude** | **OpenAI** | **转换核心** | `system`参数降级为message；`max_tokens`等参数映射 |
| **Claude** | **Claude** | 透明透传 | 鉴权转发，流式响应保持 |

### 1.2 高级特性支持
*   **流式响应 (Streaming)**: 必须支持 Server-Sent Events (SSE) 的接收、解析、转换和重发，保持低延迟。
*   **工具调用 (Tool/Function Calling)**:
    *   OpenAI `tools` (JSON Schema) <-> Claude `tools` (Input Schema) 的双向转换。
    *   处理多轮对话中的 Tool Execution Result 格式对齐。
*   **文件与视觉 (Multimodal)**:
    *   支持 Base64 图片数据的互通。
    *   支持 URL 图片的自动下载（如果目标模型只支持 Base64）。
    *   处理 PDF/文档内容的提取（若目标模型不支持直接文件上传，需在网关层做简单的 parsing 或报错，建议初期依赖模型原生能力）。
*   **思考/推理 (Reasoning)**:
    *   透传 DeepSeek/o1 的推理字段。
    *   在不同协议间标准化“正在思考”的状态输出。

## 2. 模块细化与职责 (Module Breakdown)

系统应采用模块化设计，以 Go 语言为例进行划分。

### 2.1 接入模块 (Ingress Layer)
负责处理 HTTP 请求，不含业务逻辑。
*   **路由 (Router)**:
    *   `POST /v1/chat/completions` (OpenAI 格式入口)
    *   `POST /v1/messages` (Claude 格式入口)
    *   `GET /v1/models` (模型列表)
*   **鉴权 (Auth Middleware)**:
    *   提取 `Authorization: Bearer` 或 `x-api-key`。
    *   验证 Key 的有效性（本地静态配置或数据库）。
    *   确定租户/用户身份（用于限流或计费）。
*   **上下文构建 (Context Builder)**: 生成 Request Scope 的 Context，传递 RequestID。

### 2.2 转换核心 (Converter Core) - *最复杂部分*
位于 `/pkg/converter`。
*   **Schema 定义**:
    *   定义 `UnifiedRequest` (超集)：包含 System, Messages (Content Parts), Tools, Stream, Configs。
    *   定义 `UnifiedResponse` (超集)：包含 Content, ToolCalls, TokenUsage, FinishReason。
*   **Decoder (Unmarshaler)**:
    *   `OpenAIRequestDecoder`: Parse JSON -> `UnifiedRequest`.
    *   `ClaudeRequestDecoder`: Parse JSON -> `UnifiedRequest`.
*   **Encoder (Marshaler)**:
    *   `OpenAIResponseEncoder`: `UnifiedResponse` -> JSON/SSE.
    *   `ClaudeResponseEncoder`: `UnifiedResponse` -> JSON/SSE.

### 2.3 适配器模块 (Provider Adapters)
位于 `/pkg/providers`。每个 Provider 实现统一接口 `GenerateStream(ctx, req *UnifiedRequest) (<-chan UnifiedResponse, error)`。
*   **OpenAI Provider**: 封装官方/第三方 OpenAI Client。
*   **Anthropic Provider**: 封装 Anthropic Client。
*   **Ollama Provider**: 适配 Ollama 本地 API。

### 2.4 配置与路由策略 (Config & Strategy)
*   **模型路由表 (Model Map)**:
    *   `model: "gpt-4"` -> Route to OpenAI Provider
    *   `model: "claude-3-opus"` -> Route to Anthropic Provider
    *   `model: "claude-3-via-openai"` -> Route to Anthropic Provider (Client sees specific model name)
    *   *Alias 支持*: 用户请求 `gpt-4o`，实际路由到 Azure OpenAI `gpt-4o-eastus`.

## 3. 实现步骤详解 (Implementation Plan)

### Phase 1: 基础框架 (Skeleton)
1.  初始化 Go Mod 项目。
2.  定义 `UnifiedRequest` 和 `UnifiedResponse` 结构体 (Go struct)。
3.  搭建 Gin/Echo/Fiber HTTP Server。
4.  实现 `v1/chat/completions` 接口，Mock 返回 "Hello World"。

### Phase 2: OpenAI 原生透传 (Direct Pass-through)
1.  实现 `OpenAIProvider`。
2.  在 Handler 中解析标准 OpenAI 请求。
3.  调用 Provider 并将结果原样写回 ResponseWriter。
4.  **验证**: 使用 `curl` 或 Python `openai` 库访问网关。

### Phase 3: Claude 到 OpenAI 的转换 (Claude Client -> OpenAI Model)
1.  实现 `ClaudeRequestDecoder`: 将 `/v1/messages` 的 body 转为 `UnifiedRequest`。
2.  确保 `UnifiedRequest` 能正确映射到 OpenAI 的参数 (Role: user/assistant)。
3.  实现 `ClaudeResponseEncoder`: 将 OpenAI 的 SSE 流转换为 Anthropic 的 Event Stream (`message_start`, `content_block_delta` 等)。

### Phase 4: OpenAI 到 Claude 的转换 (OpenAI Client -> Claude Model)
1.  实现 `AnthropicProvider`。
2.  实现 `OpenAIResponseEncoder` 的反向逻辑：将 Claude 的流式 Delta 封装为 OpenAI `chat.completion.chunk`。
3.  **难点**: 处理 Claude 的 `prefill` (Assistant 消息预填)，这在 OpenAI 协议中不直接支持，需要通过组合 Message 实现。

### Phase 5: 工具调用 (Tool Calling)
1.  完善 `UnifiedRequest` 中的 `Tools` 定义。
2.  编写 JSON Schema 转换器 (OpenAI JSON Schema <-> Anthropic JSON Schema)。
3.  测试多步调用：Model 请求调用工具 -> Client 执行工具 -> Client 提交结果 -> Model 生成最终答案。

## 4. 技术选型 (Tech Stack)

*   **Language**: Go 1.22+ (高性能，并发处理 SSE 优势明显)
*   **HTTP Framework**: Gin 或标准库 `net/http` (保持轻量)
*   **JSON Lib**: `encoding/json` 或 `sonic` (如果追求极致性能)
*   **Validation**: `go-playground/validator`
*   **Logging**: `zap` or `slog`

## 5. 目录结构建议

```
ai-gateway/
├── cmd/
│   └── server/          # main.go
├── config/              # 配置文件定义
├── docs/                # 文档
├── pkg/
│   ├── api/             # HTTP Handlers (OpenAI/Claude handler)
│   ├── converter/       # 协议转换逻辑 (Encoders/Decoders)
│   ├── core/            # 统一数据模型 (Unified Schema)
│   ├── providers/       # 下游适配器 (OpenAI, Anthropic)
│   └── utils/
└── go.mod
```
