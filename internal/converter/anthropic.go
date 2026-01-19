// Package converter 处理不同 API 格式之间的协议转换。
package converter

import (
	"encoding/json"
	"fmt"
	"time"

	"ai-gateway/internal/domain"
)

// AnthropicConverter 在 Anthropic API 格式和统一格式之间进行转换。
type AnthropicConverter struct{}

// NewAnthropicConverter 创建一个新的 Anthropic 转换器。
func NewAnthropicConverter() *AnthropicConverter {
	return &AnthropicConverter{}
}

func (c *AnthropicConverter) FormatName() string { return "anthropic" }

// flexibleSystem 支持 string 或 content block 数组两种格式
type flexibleSystem struct {
	Text   string         // 当 system 为字符串时使用
	Blocks []contentBlock // 当 system 为数组时使用
}

// UnmarshalJSON 自定义解析，支持字符串或数组格式
func (f *flexibleSystem) UnmarshalJSON(data []byte) error {
	// 尝试解析为字符串
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		f.Text = str
		return nil
	}

	// 尝试解析为 content block 数组
	var blocks []contentBlock
	if err := json.Unmarshal(data, &blocks); err == nil {
		f.Blocks = blocks
		// 将 blocks 中的文本合并为一个字符串
		var combined string
		for _, block := range blocks {
			if block.Type == "text" && block.Text != "" {
				if combined != "" {
					combined += "\n"
				}
				combined += block.Text
			}
		}
		f.Text = combined
		return nil
	}

	return fmt.Errorf("system 字段必须是字符串或内容块数组")
}

// flexibleContent 支持 string、单个对象或 content block 数组等多种格式
type flexibleContent struct {
	Blocks []contentBlock
}

// UnmarshalJSON 自定义解析，支持多种格式
func (f *flexibleContent) UnmarshalJSON(data []byte) error {
	// 处理 null 值
	if string(data) == "null" {
		f.Blocks = []contentBlock{}
		return nil
	}

	// 尝试解析为字符串
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		f.Blocks = []contentBlock{{Type: "text", Text: str}}
		return nil
	}

	// 尝试解析为 content block 数组
	var blocks []contentBlock
	if err := json.Unmarshal(data, &blocks); err == nil {
		f.Blocks = blocks
		return nil
	}

	// 尝试解析为单个 content block 对象
	var block contentBlock
	if err := json.Unmarshal(data, &block); err == nil {
		f.Blocks = []contentBlock{block}
		return nil
	}

	return fmt.Errorf("content 字段必须是字符串、对象或内容块数组，得到: %s", string(data))
}

// claudeToolChoice 支持多种 tool_choice 格式
type claudeToolChoice struct {
	Type                   string `json:"type"`           // auto, any, tool, none
	Name                   string `json:"name,omitempty"` // 仅当 type 为 "tool" 时使用
	DisableParallelToolUse bool   `json:"disable_parallel_tool_use,omitempty"`
}

// claudeThinking 表示扩展思考配置
type claudeThinking struct {
	Type         string `json:"type"`                    // "enabled" 或 "disabled"
	BudgetTokens int    `json:"budget_tokens,omitempty"` // 思考 token 预算
}

// Anthropic 请求类型
type anthropicRequest struct {
	Model         string            `json:"model"`
	Messages      []claudeMessage   `json:"messages"`
	System        *flexibleSystem   `json:"system,omitempty"`
	MaxTokens     int               `json:"max_tokens"`
	Stream        bool              `json:"stream,omitempty"`
	Temperature   *float64          `json:"temperature,omitempty"`
	TopP          *float64          `json:"top_p,omitempty"`
	TopK          *int              `json:"top_k,omitempty"`
	StopSequences []string          `json:"stop_sequences,omitempty"`
	Tools         []claudeTool      `json:"tools,omitempty"`
	ToolChoice    *claudeToolChoice `json:"tool_choice,omitempty"`
	Thinking      *claudeThinking   `json:"thinking,omitempty"`
}

type claudeMessage struct {
	Role    string          `json:"role"`
	Content flexibleContent `json:"content"`
}

// flexibleToolResultContent 支持 tool_result 中 content 字段的多种格式
type flexibleToolResultContent struct {
	Text string // 合并后的文本内容
}

// UnmarshalJSON 自定义解析 tool_result 的 content 字段
func (f *flexibleToolResultContent) UnmarshalJSON(data []byte) error {
	// 处理 null 值
	if string(data) == "null" {
		f.Text = ""
		return nil
	}

	// 尝试解析为字符串
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		f.Text = str
		return nil
	}

	// 尝试解析为 content block 数组
	var blocks []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if err := json.Unmarshal(data, &blocks); err == nil {
		var combined string
		for _, block := range blocks {
			if block.Type == "text" && block.Text != "" {
				if combined != "" {
					combined += "\n"
				}
				combined += block.Text
			}
		}
		f.Text = combined
		return nil
	}

	// 尝试解析为单个对象
	var block struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if err := json.Unmarshal(data, &block); err == nil {
		f.Text = block.Text
		return nil
	}

	return nil // 忽略无法解析的格式
}

type contentBlock struct {
	Type      string                     `json:"type"`
	Text      string                     `json:"text,omitempty"`
	Source    *imageSource               `json:"source,omitempty"`
	ID        string                     `json:"id,omitempty"`
	Name      string                     `json:"name,omitempty"`
	Input     map[string]any             `json:"input,omitempty"`
	ToolUseID string                     `json:"tool_use_id,omitempty"`
	Content   *flexibleToolResultContent `json:"content,omitempty"` // 用于 tool_result，支持字符串或数组
	IsError   bool                       `json:"is_error,omitempty"`
	Thinking  string                     `json:"thinking,omitempty"`
}

type imageSource struct {
	Type      string `json:"type"`       // "base64" 或 "url"
	MediaType string `json:"media_type"` // 仅 base64 时使用
	Data      string `json:"data"`       // 仅 base64 时使用
	URL       string `json:"url"`        // 仅 url 时使用
}

type claudeTool struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	InputSchema map[string]any `json:"input_schema"`
}

// Anthropic 响应类型
type anthropicResponse struct {
	ID           string         `json:"id"`
	Type         string         `json:"type"`
	Role         string         `json:"role"`
	Content      []contentBlock `json:"content"`
	Model        string         `json:"model"`
	StopReason   string         `json:"stop_reason,omitempty"`
	StopSequence string         `json:"stop_sequence,omitempty"`
	Usage        *claudeUsage   `json:"usage,omitempty"`
}

type claudeUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// 流式事件类型
type streamEvent struct {
	Type         string             `json:"type"`
	Message      *anthropicResponse `json:"message,omitempty"`
	Index        int                `json:"index,omitempty"`
	ContentBlock *contentBlock      `json:"content_block,omitempty"`
	Delta        *streamDelta       `json:"delta,omitempty"`
	Usage        *claudeUsage       `json:"usage,omitempty"`
}

type streamDelta struct {
	Type        string `json:"type"`
	Text        string `json:"text,omitempty"`
	PartialJSON string `json:"partial_json,omitempty"`
	Thinking    string `json:"thinking,omitempty"`
	StopReason  string `json:"stop_reason,omitempty"`
}

// DecodeRequest 将 Anthropic API 请求转换为统一格式。
func (c *AnthropicConverter) DecodeRequest(data []byte) (*domain.ChatRequest, error) {
	var req anthropicRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return nil, fmt.Errorf("unmarshal anthropic request: %w", err)
	}

	// 提取 system 文本
	var systemText string
	if req.System != nil {
		systemText = req.System.Text
	}

	unified := &domain.ChatRequest{
		Model:         req.Model,
		System:        systemText,
		Stream:        req.Stream,
		MaxTokens:     req.MaxTokens,
		Temperature:   req.Temperature,
		TopP:          req.TopP,
		TopK:          req.TopK,
		StopSequences: req.StopSequences,
	}

	// 转换 tool_choice
	if req.ToolChoice != nil {
		unified.ToolChoice = c.decodeToolChoice(req.ToolChoice)
	}

	// 转换 thinking
	if req.Thinking != nil {
		unified.Thinking = &domain.ThinkingConfig{
			Type:         req.Thinking.Type,
			BudgetTokens: req.Thinking.BudgetTokens,
		}
	}

	// 转换消息
	for _, m := range req.Messages {
		unified.Messages = append(unified.Messages, c.decodeMessage(m))
	}

	// 转换工具
	for _, t := range req.Tools {
		unified.Tools = append(unified.Tools, domain.ToolDefinition{
			Name:        t.Name,
			Description: t.Description,
			InputSchema: t.InputSchema,
		})
	}

	return unified, nil
}

func (c *AnthropicConverter) decodeMessage(m claudeMessage) domain.Message {
	msg := domain.Message{
		Role: domain.Role(m.Role),
	}

	for _, block := range m.Content.Blocks {
		switch block.Type {
		case "text":
			msg.Content = append(msg.Content, domain.ContentPart{
				Type: domain.ContentTypeText,
				Text: block.Text,
			})
		case "image":
			if block.Source != nil {
				part := domain.ContentPart{
					Type: domain.ContentTypeImage,
				}
				if block.Source.Type == "url" {
					// URL 格式图像
					part.URL = block.Source.URL
				} else {
					// base64 格式图像
					part.MediaType = block.Source.MediaType
					part.Data = block.Source.Data
				}
				msg.Content = append(msg.Content, part)
			}
		case "tool_use":
			msg.Content = append(msg.Content, domain.ContentPart{
				Type:      domain.ContentTypeToolUse,
				ToolID:    block.ID,
				ToolName:  block.Name,
				ToolInput: block.Input,
			})
		case "tool_result":
			var toolResultText string
			if block.Content != nil {
				toolResultText = block.Content.Text
			}
			msg.Content = append(msg.Content, domain.ContentPart{
				Type:      domain.ContentTypeToolResult,
				ToolUseID: block.ToolUseID,
				Text:      toolResultText,
				IsError:   block.IsError,
			})
		case "thinking":
			msg.Content = append(msg.Content, domain.ContentPart{
				Type:     domain.ContentTypeThinking,
				Thinking: block.Thinking,
			})
		}
	}

	return msg
}

// decodeToolChoice 将 Anthropic tool_choice 转换为统一格式
func (c *AnthropicConverter) decodeToolChoice(tc *claudeToolChoice) *domain.ToolChoice {
	if tc == nil {
		return nil
	}

	choice := &domain.ToolChoice{
		DisableParallelToolUse: tc.DisableParallelToolUse,
	}

	switch tc.Type {
	case "auto":
		choice.Type = domain.ToolChoiceAuto
	case "none":
		choice.Type = domain.ToolChoiceNone
	case "any":
		choice.Type = domain.ToolChoiceAny
	case "tool":
		choice.Type = domain.ToolChoiceTool
		choice.Name = tc.Name
	default:
		choice.Type = domain.ToolChoiceAuto
	}

	return choice
}

// EncodeResponse 将统一响应转换为 Anthropic API 格式。
func (c *AnthropicConverter) EncodeResponse(resp *domain.ChatResponse) ([]byte, error) {
	claudeResp := anthropicResponse{
		ID:         resp.ID,
		Type:       "message",
		Role:       "assistant",
		Model:      resp.Model,
		StopReason: c.mapFinishReason(resp.FinishReason),
	}

	for _, part := range resp.Content {
		switch part.Type {
		case domain.ContentTypeText:
			claudeResp.Content = append(claudeResp.Content, contentBlock{
				Type: "text",
				Text: part.Text,
			})
		case domain.ContentTypeToolUse:
			claudeResp.Content = append(claudeResp.Content, contentBlock{
				Type:  "tool_use",
				ID:    part.ToolID,
				Name:  part.ToolName,
				Input: part.ToolInput,
			})
		case domain.ContentTypeThinking:
			claudeResp.Content = append(claudeResp.Content, contentBlock{
				Type:     "thinking",
				Thinking: part.Thinking,
			})
		}
	}

	if resp.Usage != nil {
		claudeResp.Usage = &claudeUsage{
			InputTokens:  resp.Usage.PromptTokens,
			OutputTokens: resp.Usage.CompletionTokens,
		}
	}

	return json.Marshal(claudeResp)
}

func (c *AnthropicConverter) mapFinishReason(reason domain.FinishReason) string {
	switch reason {
	case domain.FinishReasonStop:
		return "end_turn"
	case domain.FinishReasonLength:
		return "max_tokens"
	case domain.FinishReasonToolCalls:
		return "tool_use"
	default:
		return "end_turn"
	}
}

// EncodeStreamDelta 将流式增量转换为 Anthropic SSE 格式。
func (c *AnthropicConverter) EncodeStreamDelta(delta *domain.StreamDelta) ([]byte, error) {
	var event streamEvent

	switch delta.Type {
	case "content":
		if delta.Content != nil && delta.Content.Text != "" {
			event = streamEvent{
				Type: "content_block_delta",
				Delta: &streamDelta{
					Type: "text_delta",
					Text: delta.Content.Text,
				},
			}
		}
	case "thinking":
		if delta.Content != nil {
			event = streamEvent{
				Type: "content_block_delta",
				Delta: &streamDelta{
					Type:     "thinking_delta",
					Thinking: delta.Content.Thinking,
				},
			}
		}
	case "tool_use":
		if delta.Content != nil {
			inputJSON, _ := json.Marshal(delta.Content.ToolInput)
			event = streamEvent{
				Type: "content_block_delta",
				Delta: &streamDelta{
					Type:        "input_json_delta",
					PartialJSON: string(inputJSON),
				},
			}
		}
	case "done":
		event = streamEvent{
			Type: "message_delta",
			Delta: &streamDelta{
				Type:       "message_delta",
				StopReason: c.mapFinishReason(delta.FinishReason),
			},
		}
		if delta.Usage != nil {
			event.Usage = &claudeUsage{
				InputTokens:  delta.Usage.PromptTokens,
				OutputTokens: delta.Usage.CompletionTokens,
			}
		}
	}

	return json.Marshal(event)
}

// GenerateID 生成唯一的消息 ID。
func GenerateID() string {
	return fmt.Sprintf("msg_%d", time.Now().UnixNano())
}
