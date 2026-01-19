// Package domain 定义 AI 网关的核心领域模型。
// 这些是内部使用的与协议无关的表示。
package domain

// Role 表示消息参与者的角色。
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

// ContentType 表示消息部分中的内容类型。
type ContentType string

const (
	ContentTypeText      ContentType = "text"
	ContentTypeImage     ContentType = "image"
	ContentTypeToolUse   ContentType = "tool_use"
	ContentTypeToolResult ContentType = "tool_result"
	ContentTypeThinking  ContentType = "thinking"
)

// ContentPart 表示消息内容的单个部分。
// 一条消息可以包含多个部分（文本、图像、工具调用等）。
type ContentPart struct {
	Type ContentType `json:"type"`

	// 文本内容
	Text string `json:"text,omitempty"`

	// 图像内容
	MediaType string `json:"media_type,omitempty"` // e.g., "image/png"
	Data      string `json:"data,omitempty"`       // base64 encoded
	URL       string `json:"url,omitempty"`        // image URL

	// 工具使用
	ToolID    string         `json:"tool_id,omitempty"`
	ToolName  string         `json:"tool_name,omitempty"`
	ToolInput map[string]any `json:"tool_input,omitempty"`

	// 工具结果
	ToolUseID string `json:"tool_use_id,omitempty"`
	IsError   bool   `json:"is_error,omitempty"`

	// 思考/推理内容
	Thinking string `json:"thinking,omitempty"`
}

// Message 表示会话中的单条消息。
type Message struct {
	Role    Role          `json:"role"`
	Content []ContentPart `json:"content"`
	Name    string        `json:"name,omitempty"` // For tool roles
}

// NewTextMessage 创建一条简单的文本消息。
func NewTextMessage(role Role, text string) Message {
	return Message{
		Role: role,
		Content: []ContentPart{
			{Type: ContentTypeText, Text: text},
		},
	}
}

// GetTextContent 从消息部分中提取所有文本内容。
func (m *Message) GetTextContent() string {
	var result string
	for _, part := range m.Content {
		if part.Type == ContentTypeText {
			result += part.Text
		}
	}
	return result
}
