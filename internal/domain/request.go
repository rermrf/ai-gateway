// Package domain 定义 AI 网关的核心领域模型。
package domain

// ToolChoiceType 表示工具选择的类型。
type ToolChoiceType string

const (
	// ToolChoiceAuto 让模型自动决定是否使用工具
	ToolChoiceAuto ToolChoiceType = "auto"
	// ToolChoiceNone 禁止模型使用任何工具
	ToolChoiceNone ToolChoiceType = "none"
	// ToolChoiceAny 强制模型使用任意一个工具（Anthropic）/ required（OpenAI）
	ToolChoiceAny ToolChoiceType = "any"
	// ToolChoiceTool 强制模型使用指定的工具
	ToolChoiceTool ToolChoiceType = "tool"
)

// ToolChoice 表示工具选择配置。
type ToolChoice struct {
	// Type 是工具选择类型：auto, none, any, tool
	Type ToolChoiceType `json:"type"`
	// Name 是指定工具的名称（仅当 Type 为 "tool" 时使用）
	Name string `json:"name,omitempty"`
	// DisableParallelToolUse 禁用并行工具调用（Anthropic 特有）
	DisableParallelToolUse bool `json:"disable_parallel_tool_use,omitempty"`
}

// ToolDefinition 表示可以由模型调用的工具/函数。
type ToolDefinition struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	InputSchema map[string]any `json:"input_schema"` // JSON Schema
}

// ResponseFormatType 表示响应格式类型。
type ResponseFormatType string

const (
	// ResponseFormatText 返回普通文本（默认）
	ResponseFormatText ResponseFormatType = "text"
	// ResponseFormatJSONObject 返回有效的 JSON 对象
	ResponseFormatJSONObject ResponseFormatType = "json_object"
	// ResponseFormatJSONSchema 返回符合指定 JSON Schema 的对象
	ResponseFormatJSONSchema ResponseFormatType = "json_schema"
)

// ResponseFormat 表示响应格式配置。
type ResponseFormat struct {
	// Type 是响应格式类型：text, json_object, json_schema
	Type ResponseFormatType `json:"type"`
	// JSONSchema 仅当 Type 为 "json_schema" 时使用
	JSONSchema *JSONSchemaConfig `json:"json_schema,omitempty"`
}

// JSONSchemaConfig 表示 JSON Schema 配置。
type JSONSchemaConfig struct {
	// Name 是 schema 的名称
	Name string `json:"name"`
	// Description 是 schema 的描述
	Description string `json:"description,omitempty"`
	// Schema 是 JSON Schema 定义
	Schema map[string]any `json:"schema"`
	// Strict 是否严格模式
	Strict bool `json:"strict,omitempty"`
}

// ThinkingConfig 表示扩展思考配置（Anthropic 特有）。
type ThinkingConfig struct {
	// Type 是思考类型：enabled 或 disabled
	Type string `json:"type"`
	// BudgetTokens 是思考的 token 预算（仅当 Type 为 "enabled" 时使用）
	// 必须 >= 1024 且 < max_tokens
	BudgetTokens int `json:"budget_tokens,omitempty"`
}

// ChatRequest 表示统一的聊天补全请求。
// 这是所有协议与之相互转换的内部表示。
type ChatRequest struct {
	// 模型标识符（例如 "gpt-4"、"claude-3-opus"）
	Model string `json:"model"`

	// 系统提示词（为了 Claude 兼容性而分离）
	System string `json:"system,omitempty"`

	// 会话消息
	Messages []Message `json:"messages"`

	// 可用工具
	Tools []ToolDefinition `json:"tools,omitempty"`

	// 工具选择策略
	ToolChoice *ToolChoice `json:"tool_choice,omitempty"`

	// 是否流式传输响应
	Stream bool `json:"stream"`

	// 生成参数
	MaxTokens        int      `json:"max_tokens,omitempty"`
	Temperature      *float64 `json:"temperature,omitempty"`
	TopP             *float64 `json:"top_p,omitempty"`
	TopK             *int     `json:"top_k,omitempty"`
	StopSequences    []string `json:"stop,omitempty"`
	PresencePenalty  *float64 `json:"presence_penalty,omitempty"`
	FrequencyPenalty *float64 `json:"frequency_penalty,omitempty"`

	// 响应格式（JSON 模式）
	ResponseFormat *ResponseFormat `json:"response_format,omitempty"`

	// 扩展思考配置（Anthropic 特有）
	Thinking *ThinkingConfig `json:"thinking,omitempty"`

	// 提供商特定的元数据（用于透传）
	Metadata map[string]any `json:"metadata,omitempty"`
}

// FinishReason 表示模型停止生成的原因。
type FinishReason string

const (
	FinishReasonStop      FinishReason = "stop"
	FinishReasonLength    FinishReason = "length"
	FinishReasonToolCalls FinishReason = "tool_calls"
	FinishReasonError     FinishReason = "error"
)

// TokenUsage 表示令牌消耗统计数据。
type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ChatResponse 表示统一的聊天补全响应。
type ChatResponse struct {
	// 此响应的唯一 ID
	ID string `json:"id"`

	// 生成响应的模型标识符
	Model string `json:"model"`

	// 响应内容
	Content []ContentPart `json:"content"`

	// 生成停止的原因
	FinishReason FinishReason `json:"finish_reason,omitempty"`

	// 令牌使用统计数据
	Usage *TokenUsage `json:"usage,omitempty"`
}

// StreamDelta 表示流式响应中的单个分块。
type StreamDelta struct {
	// 增量类型
	Type string `json:"type"` // "content", "tool_use", "thinking", "done"

	// 内容增量
	Content *ContentPart `json:"content,omitempty"`

	// 最终分块
	FinishReason FinishReason `json:"finish_reason,omitempty"`
	Usage        *TokenUsage  `json:"usage,omitempty"`
}
