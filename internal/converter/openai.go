// Package converter 处理不同 API 格式之间的协议转换。
package converter

import (
	"encoding/json"
	"fmt"
	"time"

	"ai-gateway/internal/domain"
)

// OpenAIConverter 在 OpenAI API 格式和统一格式之间进行转换。
type OpenAIConverter struct{}

// NewOpenAIConverter 创建一个新的 OpenAI 转换器。
func NewOpenAIConverter() *OpenAIConverter {
	return &OpenAIConverter{}
}

func (c *OpenAIConverter) FormatName() string { return "openai" }

// oaiToolChoice 支持多种 tool_choice 格式
// OpenAI tool_choice 可以是：
// - 字符串: "none", "auto", "required"
// - 对象: {"type": "function", "function": {"name": "..."}}
type oaiToolChoice struct {
	IsString     bool   // 内部标记，表示是否为字符串格式
	StringValue  string // 当为字符串时的值
	Type         string `json:"type,omitempty"` // "function"
	FunctionSpec *struct {
		Name string `json:"name"`
	} `json:"function,omitempty"`
}

// UnmarshalJSON 自定义解析，支持字符串或对象格式
func (tc *oaiToolChoice) UnmarshalJSON(data []byte) error {
	// 尝试解析为字符串
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		tc.IsString = true
		tc.StringValue = str
		return nil
	}

	// 尝试解析为对象
	type Alias oaiToolChoice
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(tc),
	}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	tc.IsString = false
	return nil
}

// oaiResponseFormat 支持多种 response_format 格式
type oaiResponseFormat struct {
	Type       string             `json:"type"` // text, json_object, json_schema
	JSONSchema *oaiJSONSchemaSpec `json:"json_schema,omitempty"`
}

type oaiJSONSchemaSpec struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Schema      map[string]any `json:"schema"`
	Strict      bool           `json:"strict,omitempty"`
}

// OpenAI 请求类型
type openAIRequest struct {
	Model             string             `json:"model"`
	Messages          []oaiMessage       `json:"messages"`
	Stream            bool               `json:"stream,omitempty"`
	MaxTokens         int                `json:"max_tokens,omitempty"`
	Temperature       *float64           `json:"temperature,omitempty"`
	TopP              *float64           `json:"top_p,omitempty"`
	Stop              []string           `json:"stop,omitempty"`
	Tools             []oaiTool          `json:"tools,omitempty"`
	ToolChoice        *oaiToolChoice     `json:"tool_choice,omitempty"`
	ResponseFormat    *oaiResponseFormat `json:"response_format,omitempty"`
	PresencePenalty   *float64           `json:"presence_penalty,omitempty"`
	FrequencyPenalty  *float64           `json:"frequency_penalty,omitempty"`
	ParallelToolCalls *bool              `json:"parallel_tool_calls,omitempty"`
}

type oaiMessage struct {
	Role             string        `json:"role"`
	Content          interface{}   `json:"content"`                     // 字符串或 []contentPart
	ReasoningContent string        `json:"reasoning_content,omitempty"` // 思考内容
	Name             string        `json:"name,omitempty"`
	ToolCalls        []oaiToolCall `json:"tool_calls,omitempty"`
	ToolCallID       string        `json:"tool_call_id,omitempty"`
}

type oaiContentPart struct {
	Type     string       `json:"type"`
	Text     string       `json:"text,omitempty"`
	ImageURL *oaiImageURL `json:"image_url,omitempty"`
}

type oaiImageURL struct {
	URL    string `json:"url"`
	Detail string `json:"detail,omitempty"`
}

type oaiTool struct {
	Type     string      `json:"type"`
	Function oaiFunction `json:"function"`
}

type oaiFunction struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Parameters  map[string]any `json:"parameters,omitempty"`
}

type oaiToolCall struct {
	ID       string          `json:"id"`
	Type     string          `json:"type"`
	Function oaiFunctionCall `json:"function"`
}

type oaiFunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// OpenAI 响应类型
type openAIResponse struct {
	ID      string      `json:"id"`
	Object  string      `json:"object"`
	Created int64       `json:"created"`
	Model   string      `json:"model"`
	Choices []oaiChoice `json:"choices"`
	Usage   *oaiUsage   `json:"usage,omitempty"`
}

type oaiChoice struct {
	Index        int        `json:"index"`
	Message      oaiMessage `json:"message"`
	FinishReason string     `json:"finish_reason"`
}

type oaiUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// 流式传输类型
type openAIStreamChunk struct {
	ID      string            `json:"id"`
	Object  string            `json:"object"`
	Created int64             `json:"created"`
	Model   string            `json:"model"`
	Choices []oaiStreamChoice `json:"choices"`
	Usage   *oaiUsage         `json:"usage,omitempty"`
}

type oaiStreamChoice struct {
	Index        int        `json:"index"`
	Delta        oaiMessage `json:"delta"`
	FinishReason string     `json:"finish_reason,omitempty"`
}

// DecodeRequest 将 OpenAI API 请求转换为统一格式。
func (c *OpenAIConverter) DecodeRequest(data []byte) (*domain.ChatRequest, error) {
	var req openAIRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return nil, fmt.Errorf("unmarshal openai request: %w", err)
	}

	unified := &domain.ChatRequest{
		Model:            req.Model,
		Stream:           req.Stream,
		MaxTokens:        req.MaxTokens,
		Temperature:      req.Temperature,
		TopP:             req.TopP,
		StopSequences:    req.Stop,
		PresencePenalty:  req.PresencePenalty,
		FrequencyPenalty: req.FrequencyPenalty,
	}

	// 转换消息
	for _, m := range req.Messages {
		msg := c.decodeMessage(m)
		// 提取系统消息
		if msg.Role == domain.RoleSystem {
			unified.System = msg.GetTextContent()
		}
		unified.Messages = append(unified.Messages, msg)
	}

	// 转换工具
	for _, t := range req.Tools {
		unified.Tools = append(unified.Tools, domain.ToolDefinition{
			Name:        t.Function.Name,
			Description: t.Function.Description,
			InputSchema: t.Function.Parameters,
		})
	}

	// 转换 tool_choice
	if req.ToolChoice != nil {
		unified.ToolChoice = c.decodeToolChoice(req.ToolChoice, req.ParallelToolCalls)
	}

	// 转换 response_format
	if req.ResponseFormat != nil {
		unified.ResponseFormat = c.decodeResponseFormat(req.ResponseFormat)
	}

	return unified, nil
}

func (c *OpenAIConverter) decodeMessage(m oaiMessage) domain.Message {
	msg := domain.Message{
		Role: domain.Role(m.Role),
		Name: m.Name,
	}

	// 处理工具结果
	if m.ToolCallID != "" {
		msg.Role = domain.RoleTool
		if text, ok := m.Content.(string); ok {
			msg.Content = append(msg.Content, domain.ContentPart{
				Type:      domain.ContentTypeToolResult,
				ToolUseID: m.ToolCallID,
				Text:      text,
			})
		}
		return msg
	}

	// 处理内容
	switch content := m.Content.(type) {
	case string:
		msg.Content = append(msg.Content, domain.ContentPart{
			Type: domain.ContentTypeText,
			Text: content,
		})
	case []interface{}:
		for _, part := range content {
			if partMap, ok := part.(map[string]interface{}); ok {
				c.decodeContentPart(partMap, &msg)
			}
		}
	}

	// 处理工具调用
	for _, tc := range m.ToolCalls {
		var input map[string]any
		json.Unmarshal([]byte(tc.Function.Arguments), &input)
		msg.Content = append(msg.Content, domain.ContentPart{
			Type:      domain.ContentTypeToolUse,
			ToolID:    tc.ID,
			ToolName:  tc.Function.Name,
			ToolInput: input,
		})
	}

	return msg
}

func (c *OpenAIConverter) decodeContentPart(part map[string]interface{}, msg *domain.Message) {
	partType, _ := part["type"].(string)
	switch partType {
	case "text":
		if text, ok := part["text"].(string); ok {
			msg.Content = append(msg.Content, domain.ContentPart{
				Type: domain.ContentTypeText,
				Text: text,
			})
		}
	case "image_url":
		if imgURL, ok := part["image_url"].(map[string]interface{}); ok {
			url, _ := imgURL["url"].(string)
			msg.Content = append(msg.Content, domain.ContentPart{
				Type: domain.ContentTypeImage,
				URL:  url,
			})
		}
	}
}

// decodeToolChoice 将 OpenAI tool_choice 转换为统一格式
func (c *OpenAIConverter) decodeToolChoice(tc *oaiToolChoice, parallelToolCalls *bool) *domain.ToolChoice {
	if tc == nil {
		return nil
	}

	choice := &domain.ToolChoice{}

	// 如果禁用并行工具调用
	if parallelToolCalls != nil && !*parallelToolCalls {
		choice.DisableParallelToolUse = true
	}

	if tc.IsString {
		switch tc.StringValue {
		case "none":
			choice.Type = domain.ToolChoiceNone
		case "auto":
			choice.Type = domain.ToolChoiceAuto
		case "required":
			choice.Type = domain.ToolChoiceAny
		default:
			choice.Type = domain.ToolChoiceAuto
		}
	} else {
		// 对象格式，指定特定函数
		choice.Type = domain.ToolChoiceTool
		if tc.FunctionSpec != nil {
			choice.Name = tc.FunctionSpec.Name
		}
	}

	return choice
}

// decodeResponseFormat 将 OpenAI response_format 转换为统一格式
func (c *OpenAIConverter) decodeResponseFormat(rf *oaiResponseFormat) *domain.ResponseFormat {
	if rf == nil {
		return nil
	}

	format := &domain.ResponseFormat{}

	switch rf.Type {
	case "text":
		format.Type = domain.ResponseFormatText
	case "json_object":
		format.Type = domain.ResponseFormatJSONObject
	case "json_schema":
		format.Type = domain.ResponseFormatJSONSchema
		if rf.JSONSchema != nil {
			format.JSONSchema = &domain.JSONSchemaConfig{
				Name:        rf.JSONSchema.Name,
				Description: rf.JSONSchema.Description,
				Schema:      rf.JSONSchema.Schema,
				Strict:      rf.JSONSchema.Strict,
			}
		}
	default:
		format.Type = domain.ResponseFormatText
	}

	return format
}

// EncodeResponse 将统一响应转换为 OpenAI API 格式。
func (c *OpenAIConverter) EncodeResponse(resp *domain.ChatResponse) ([]byte, error) {
	oaiResp := openAIResponse{
		ID:      resp.ID,
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   resp.Model,
		Choices: []oaiChoice{
			{
				Index:        0,
				FinishReason: string(resp.FinishReason),
			},
		},
	}

	// 构建消息内容
	var textContent string
	var toolCalls []oaiToolCall
	for _, part := range resp.Content {
		switch part.Type {
		case domain.ContentTypeText:
			textContent += part.Text
		case domain.ContentTypeToolUse:
			args, _ := json.Marshal(part.ToolInput)
			toolCalls = append(toolCalls, oaiToolCall{
				ID:   part.ToolID,
				Type: "function",
				Function: oaiFunctionCall{
					Name:      part.ToolName,
					Arguments: string(args),
				},
			})
		}
	}

	oaiResp.Choices[0].Message = oaiMessage{
		Role:      "assistant",
		Content:   textContent,
		ToolCalls: toolCalls,
	}

	if resp.Usage != nil {
		oaiResp.Usage = &oaiUsage{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		}
	}

	return json.Marshal(oaiResp)
}

// EncodeStreamDelta 将流式增量转换为 OpenAI SSE 格式。
func (c *OpenAIConverter) EncodeStreamDelta(delta *domain.StreamDelta) ([]byte, error) {
	chunk := openAIStreamChunk{
		ID:      fmt.Sprintf("chatcmpl-%d", time.Now().UnixNano()),
		Object:  "chat.completion.chunk",
		Created: time.Now().Unix(),
		Choices: []oaiStreamChoice{
			{Index: 0},
		},
	}

	switch delta.Type {
	case "content":
		if delta.Content != nil && delta.Content.Text != "" {
			chunk.Choices[0].Delta = oaiMessage{
				Content: delta.Content.Text,
			}
		}
	case "thinking":
		if delta.Content != nil && delta.Content.Thinking != "" {
			chunk.Choices[0].Delta = oaiMessage{
				ReasoningContent: delta.Content.Thinking,
			}
		}
	case "tool_use":
		if delta.Content != nil {
			args, _ := json.Marshal(delta.Content.ToolInput)
			chunk.Choices[0].Delta = oaiMessage{
				ToolCalls: []oaiToolCall{
					{
						ID:   delta.Content.ToolID,
						Type: "function",
						Function: oaiFunctionCall{
							Name:      delta.Content.ToolName,
							Arguments: string(args),
						},
					},
				},
			}
		}
	case "done":
		chunk.Choices[0].FinishReason = string(delta.FinishReason)
		if delta.Usage != nil {
			chunk.Usage = &oaiUsage{
				PromptTokens:     delta.Usage.PromptTokens,
				CompletionTokens: delta.Usage.CompletionTokens,
				TotalTokens:      delta.Usage.TotalTokens,
			}
		}
	}

	return json.Marshal(chunk)
}
