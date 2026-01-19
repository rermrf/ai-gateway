# AI Gateway 技术设计文档

## 1. 项目概述 (Overview)

本项目旨在构建一个通用的 AI 网关（AI Gateway），通过标准化接口屏蔽不同 LLM 提供商（OpenAI, Anthropic, Google Gemini, Ollama 等）的协议差异。

**核心目标：**
*   **双向协议兼容**：
    *   支持使用 OpenAI SDK/API 访问 Claude 模型（及其他模型）。
    *   支持使用 Claude SDK/API 访问 OpenAI 模型（及其他模型）。
*   **全功能支持**：完整映射对话（Chat）、流式响应（Streaming）、工具调用（Tool Calling）、文件/视觉处理（Multimodal）、以及推理/思考过程（Thinking/Reasoning）。

## 2. 系统架构 (Architecture)

系统采用典型的 **六边形架构 (Hexagonal Architecture)** 或 **网关模式**。

### 2.1 核心组件

1.  **接入层 (Ingress / Listeners)**
    *   **OpenAI Compatible Handler**: 监听 `/v1/chat/completions`, `/v1/models` 等标准 OpenAI 路由。
    *   **Anthropic Compatible Handler**: 监听 `/v1/messages` 等标准 Anthropic 路由。
    
2.  **路由与转换层 (Router & Converter)**
    *   **Config Manager**: 管理模型映射配置（例如：用户请求 `model: cluade-3-5-sonnet` 时，路由到 Anthropic Adapter）。
    *   **Protocol Converter**: 核心转换引擎，负责 Request 和 Response 的双向转换。
    
3.  **适配器层 (Providers / Adapters)**
    *   **OpenAI Provider**: 连接真实 OpenAI 接口。
    *   **Anthropic Provider**: 连接真实 Anthropic 接口。
    *   **Generic Provider**: 连接其他符合 OpenAI 标准的服务（如 DeepSeek, Moonshot）。
    *   **Ollama/Local Provider**: 连接本地模型。

### 2.2 数据流向图

```mermaid
graph TD
    UserClient[User Client (OpenAI/Claude SDK)] -->|HTTP Request| Gateway
    
    subgraph Gateway ["AI Gateway Core"]
        Handler[Protocol Handler]
        Converter[Protocol Converter]
        Router[Model Router]
        
        Handler -->|Unified Schema| Router
        Router -->|Unified Schema| Converter
    end
    
    Converter -->|Target Protocol| ProviderOpenAI[OpenAI Adapter]
    Converter -->|Target Protocol| ProviderClaude[Anthropic Adapter]
    
    ProviderOpenAI -->|HTTP| RealOpenAI[OpenAI API]
    ProviderClaude -->|HTTP| RealClaude[Anthropic API]
```

## 3. 详细设计 (Detailed Design)

### 3.1 统一内部数据模型 (Unified Internal Schema)

为了实现 N-to-N 的转换，不应直接编写 OpenAI<->Claude 的硬编码转换，而应定义一个超集结构（Internal Schema）。

```go
// 伪代码示例：内部统一消息结构
type InternalMessage struct {
    Role    string  // system, user, assistant, tool
    Content []Part  // Text, Image, File, ToolResult
    Name    string  // For tool calls
}

type InternalRequest struct {
    Model       string
    Messages    []InternalMessage
    Tools       []ToolDefinition
    Stream      bool
    Temperature float64
    // ...
}
```

### 3.2 协议转换策略 (Protocol Conversion)

#### 3.2.1 文本与流式 (Text & Streaming)
*   **OpenAI -> Claude**: 需将 OpenAI 的 `messages` 列表转换为 Claude 的 `system` (顶层参数) + `messages`。OpenAI 的 SSE 格式 (`data: {...}`) 需转换为 Claude 的 Content Block 事件流。
*   **Claude -> OpenAI**: 需将 Claude 的 `system` 放入 `messages` 中作为首条 system 消息。Claude 的事件流需封装为 OpenAI Chunk 格式。

#### 3.2.2 工具调用 (Function/Tool Calling)
*   **定义转换**: OpenAI 使用 JSON Schema 定义工具；Claude 同样使用 JSON Schema，但结构略有不同（`input_schema` vs `parameters`）。需做字段映射。
*   **调用转换**: 
    *   OpenAI 返回 `tool_calls` 数组。
    *   Claude 返回 `content` 数组中的 `type: tool_use` 块。
    *   **挑战**: Claude 强制工具结果必须紧跟工具调用；OpenAI 相对宽松。中间层需维护会话状态一致性。

#### 3.2.3 文件与多媒体 (Files & Vision)
*   **OpenAI**: 使用 `image_url` (支持 base64 或 http url)。
*   **Claude**: 使用 `image` block (source type: base64)。
*   **转换**: 
    *   若收到 URL 且目标仅支持 Base64（如部分本地模型），网关需下载图片并转码。
    *   若收到 Base64 且目标支持 URL，直接透传或上传到临时存储（视目标能力而定）。

### 3.3 特殊功能支持

#### 3.3.1 思考/推理 (Reasoning/Thinking)
部分模型（如 DeepSeek-R1, OpenAI o1）会输出“思考过程”。
*   **OpenAI Protocol**: 通常在 `reasoning_content` 字段（非标准，但社区常用）或作为普通 content 的一部分。
*   **Claude Protocol**: 支持 `thinking` block（实验性）。
*   **设计**: 网关应识别思考内容，并根据客户端期望的协议进行封装。如果客户端是 OpenAI 格式，可放入 `usage` 扩展字段或特定的 `reasoning_content` 字段。

## 4. 接口规范 (API Specifications)

网关将同时暴露以下两个端口（或同一端口的不同路径）：

### 4.1 OpenAI 兼容接入点
*   `POST /v1/chat/completions`
*   `GET /v1/models`
*   Header: `Authorization: Bearer sk-...` (作为网关鉴权或透传 key)

### 4.2 Anthropic 兼容接入点
*   `POST /v1/messages`
*   Header: `x-api-key: ...`

## 5. 实施路线图 (Implementation Roadmap)

1.  **Phase 1: 骨架搭建**
    *   建立 HTTP Server。
    *   实现基本的 OpenAI -> OpenAI 透传，确保基础链路通畅。
2.  **Phase 2: 核心转换器**
    *   实现 OpenAI Request -> Claude Request 的转换。
    *   实现 Claude Response -> OpenAI Response 的流式转换。
3.  **Phase 3: 反向支持**
    *   实现 Claude 接口接入，访问 OpenAI 模型。
4.  **Phase 4: 高级特性**
    *   工具调用支持。
    *   多模态/文件支持。

