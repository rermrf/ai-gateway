// Package anthropic implements the Anthropic/Claude provider adapter.
package anthropic

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"ai-gateway/internal/pkg/logger"

	"ai-gateway/internal/domain"
	"ai-gateway/internal/errs"
)

// Provider implements the Provider interface for Anthropic API.
type Provider struct {
	apiKey  string
	baseURL string
	client  *http.Client
	logger  logger.Logger
}

// NewProvider creates a new Anthropic provider.
func NewProvider(apiKey, baseURL string, client *http.Client, l logger.Logger) *Provider {
	if baseURL == "" {
		baseURL = "https://api.anthropic.com"
	}
	if client == nil {
		client = http.DefaultClient
	}
	return &Provider{
		apiKey:  apiKey,
		baseURL: baseURL,
		client:  client,
		logger:  l.With(logger.String("provider", "anthropic")),
	}
}

func (p *Provider) Name() string { return "anthropic" }

func (p *Provider) SupportsStreaming() bool { return true }
func (p *Provider) SupportsTools() bool     { return true }
func (p *Provider) SupportsVision() bool    { return true }

// Anthropic API types
type messagesRequest struct {
	Model         string            `json:"model"`
	Messages      []claudeMessage   `json:"messages"`
	System        string            `json:"system,omitempty"`
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

type claudeThinking struct {
	Type         string `json:"type"`                    // "enabled" 或 "disabled"
	BudgetTokens int    `json:"budget_tokens,omitempty"` // 思考 token 预算
}

type claudeToolChoice struct {
	Type                   string `json:"type"`           // auto, any, tool, none
	Name                   string `json:"name,omitempty"` // 仅当 type 为 "tool" 时
	DisableParallelToolUse bool   `json:"disable_parallel_tool_use,omitempty"`
}

type claudeMessage struct {
	Role    string         `json:"role"`
	Content []contentBlock `json:"content"`
}

type contentBlock struct {
	Type      string         `json:"type"`
	Text      string         `json:"text,omitempty"`
	Source    *imageSource   `json:"source,omitempty"`
	ID        string         `json:"id,omitempty"`
	Name      string         `json:"name,omitempty"`
	Input     map[string]any `json:"input,omitempty"`
	ToolUseID string         `json:"tool_use_id,omitempty"`
	Content   string         `json:"content,omitempty"` // for tool_result
	IsError   bool           `json:"is_error,omitempty"`
	Thinking  string         `json:"thinking,omitempty"`
}

type imageSource struct {
	Type      string `json:"type"`                 // "base64" 或 "url"
	MediaType string `json:"media_type,omitempty"` // 仅 base64 时使用
	Data      string `json:"data,omitempty"`       // 仅 base64 时使用
	URL       string `json:"url,omitempty"`        // 仅 url 时使用
}

type claudeTool struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	InputSchema map[string]any `json:"input_schema"`
}

type messagesResponse struct {
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

// Streaming event types
type streamEvent struct {
	Type         string            `json:"type"`
	Message      *messagesResponse `json:"message,omitempty"`
	Index        int               `json:"index,omitempty"`
	ContentBlock *contentBlock     `json:"content_block,omitempty"`
	Delta        *streamDelta      `json:"delta,omitempty"`
	Usage        *claudeUsage      `json:"usage,omitempty"`
}

type streamDelta struct {
	Type        string `json:"type"`
	Text        string `json:"text,omitempty"`
	PartialJSON string `json:"partial_json,omitempty"`
	Thinking    string `json:"thinking,omitempty"`
	StopReason  string `json:"stop_reason,omitempty"`
}

// Chat sends a non-streaming chat request.
func (p *Provider) Chat(ctx context.Context, req *domain.ChatRequest) (*domain.ChatResponse, error) {
	claudeReq := p.toClaudeRequest(req)
	claudeReq.Stream = false

	body, err := json.Marshal(claudeReq)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/v1/messages", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	p.setHeaders(httpReq)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		p.logger.Error("Anthropic API error",
			logger.Int("status", resp.StatusCode),
			logger.String("body", string(respBody)),
		)
		return nil, fmt.Errorf("%w: status %d", errs.ErrProviderError, resp.StatusCode)
	}

	var claudeResp messagesResponse
	if err := json.NewDecoder(resp.Body).Decode(&claudeResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return p.fromClaudeResponse(&claudeResp), nil
}

// ChatStream sends a streaming chat request.
func (p *Provider) ChatStream(ctx context.Context, req *domain.ChatRequest) (<-chan domain.StreamDelta, error) {
	claudeReq := p.toClaudeRequest(req)
	claudeReq.Stream = true

	body, err := json.Marshal(claudeReq)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/v1/messages", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	p.setHeaders(httpReq)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		respBody, _ := io.ReadAll(resp.Body)
		p.logger.Error("Anthropic API error",
			logger.Int("status", resp.StatusCode),
			logger.String("body", string(respBody)),
		)
		return nil, fmt.Errorf("%w: status %d", errs.ErrProviderError, resp.StatusCode)
	}

	ch := make(chan domain.StreamDelta, 100)
	go p.readStream(resp.Body, ch)

	return ch, nil
}

func (p *Provider) readStream(body io.ReadCloser, ch chan<- domain.StreamDelta) {
	defer close(ch)
	defer body.Close()

	scanner := bufio.NewScanner(body)
	var currentToolID, currentToolName string
	var toolInputJSON strings.Builder

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")

		var event streamEvent
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			p.logger.Warn("failed to parse SSE event", logger.Error(err))
			continue
		}

		switch event.Type {
		case "content_block_start":
			if event.ContentBlock != nil {
				switch event.ContentBlock.Type {
				case "tool_use":
					currentToolID = event.ContentBlock.ID
					currentToolName = event.ContentBlock.Name
					toolInputJSON.Reset()
				case "thinking":
					// Thinking block started
				}
			}

		case "content_block_delta":
			if event.Delta != nil {
				switch event.Delta.Type {
				case "text_delta":
					if event.Delta.Text != "" {
						ch <- domain.StreamDelta{
							Type: "content",
							Content: &domain.ContentPart{
								Type: domain.ContentTypeText,
								Text: event.Delta.Text,
							},
						}
					}
				case "thinking_delta":
					if event.Delta.Thinking != "" {
						ch <- domain.StreamDelta{
							Type: "thinking",
							Content: &domain.ContentPart{
								Type:     domain.ContentTypeThinking,
								Thinking: event.Delta.Thinking,
							},
						}
					}
				case "input_json_delta":
					toolInputJSON.WriteString(event.Delta.PartialJSON)
				}
			}

		case "content_block_stop":
			// If we were accumulating tool input, emit the tool use
			if currentToolID != "" {
				var input map[string]any
				if toolInputJSON.Len() > 0 {
					json.Unmarshal([]byte(toolInputJSON.String()), &input)
				}
				ch <- domain.StreamDelta{
					Type: "tool_use",
					Content: &domain.ContentPart{
						Type:      domain.ContentTypeToolUse,
						ToolID:    currentToolID,
						ToolName:  currentToolName,
						ToolInput: input,
					},
				}
				currentToolID = ""
				currentToolName = ""
				toolInputJSON.Reset()
			}

		case "message_delta":
			if event.Delta != nil && event.Delta.StopReason != "" {
				delta := domain.StreamDelta{
					Type:         "done",
					FinishReason: mapStopReason(event.Delta.StopReason),
				}
				if event.Usage != nil {
					delta.Usage = &domain.TokenUsage{
						PromptTokens:     event.Usage.InputTokens,
						CompletionTokens: event.Usage.OutputTokens,
						TotalTokens:      event.Usage.InputTokens + event.Usage.OutputTokens,
					}
				}
				ch <- delta
			}

		case "message_stop":
			// Stream ended
		}
	}
}

func mapStopReason(reason string) domain.FinishReason {
	switch reason {
	case "end_turn", "stop_sequence":
		return domain.FinishReasonStop
	case "max_tokens":
		return domain.FinishReasonLength
	case "tool_use":
		return domain.FinishReasonToolCalls
	default:
		return domain.FinishReasonStop
	}
}

func (p *Provider) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
}

func (p *Provider) toClaudeRequest(req *domain.ChatRequest) *messagesRequest {
	claudeReq := &messagesRequest{
		Model:         req.Model,
		System:        req.System,
		MaxTokens:     req.MaxTokens,
		Temperature:   req.Temperature,
		TopP:          req.TopP,
		TopK:          req.TopK,
		StopSequences: req.StopSequences,
	}

	if claudeReq.MaxTokens == 0 {
		claudeReq.MaxTokens = 4096 // Default
	}

	// Convert messages, merging consecutive tool messages
	var pendingToolResults []contentBlock
	for _, m := range req.Messages {
		// Skip system messages (handled separately)
		if m.Role == domain.RoleSystem {
			if claudeReq.System == "" {
				claudeReq.System = m.GetTextContent()
			}
			continue
		}

		// 如果是工具结果消息，收集起来
		if m.Role == domain.RoleTool {
			for _, part := range m.Content {
				if part.Type == domain.ContentTypeToolResult {
					pendingToolResults = append(pendingToolResults, contentBlock{
						Type:      "tool_result",
						ToolUseID: part.ToolUseID,
						Content:   part.Text,
						IsError:   part.IsError,
					})
				}
			}
			continue
		}

		// 如果遇到非工具消息，先把收集的工具结果作为一个 user 消息添加
		if len(pendingToolResults) > 0 {
			claudeReq.Messages = append(claudeReq.Messages, claudeMessage{
				Role:    "user",
				Content: pendingToolResults,
			})
			pendingToolResults = nil
		}

		claudeReq.Messages = append(claudeReq.Messages, p.toClaudeMessage(m))
	}

	// 添加剩余的工具结果
	if len(pendingToolResults) > 0 {
		claudeReq.Messages = append(claudeReq.Messages, claudeMessage{
			Role:    "user",
			Content: pendingToolResults,
		})
	}

	// Convert tools
	for _, t := range req.Tools {
		claudeReq.Tools = append(claudeReq.Tools, claudeTool{
			Name:        t.Name,
			Description: t.Description,
			InputSchema: t.InputSchema,
		})
	}

	// 转换 tool_choice
	if req.ToolChoice != nil {
		claudeReq.ToolChoice = p.toClaudeToolChoice(req.ToolChoice)
	}

	// 转换 thinking
	if req.Thinking != nil {
		claudeReq.Thinking = &claudeThinking{
			Type:         req.Thinking.Type,
			BudgetTokens: req.Thinking.BudgetTokens,
		}
	}

	return claudeReq
}

// toClaudeToolChoice 将统一格式的 ToolChoice 转换为 Claude 格式
func (p *Provider) toClaudeToolChoice(tc *domain.ToolChoice) *claudeToolChoice {
	if tc == nil {
		return nil
	}

	choice := &claudeToolChoice{
		DisableParallelToolUse: tc.DisableParallelToolUse,
	}

	switch tc.Type {
	case domain.ToolChoiceAuto:
		choice.Type = "auto"
	case domain.ToolChoiceNone:
		choice.Type = "none"
	case domain.ToolChoiceAny:
		choice.Type = "any"
	case domain.ToolChoiceTool:
		choice.Type = "tool"
		choice.Name = tc.Name
	default:
		choice.Type = "auto"
	}

	return choice
}

func (p *Provider) toClaudeMessage(m domain.Message) claudeMessage {
	msg := claudeMessage{
		Role: string(m.Role),
	}

	// Convert to tool role if needed
	if m.Role == domain.RoleTool {
		msg.Role = "user" // Claude uses user role with tool_result content
	}

	for _, part := range m.Content {
		switch part.Type {
		case domain.ContentTypeText:
			msg.Content = append(msg.Content, contentBlock{
				Type: "text",
				Text: part.Text,
			})
		case domain.ContentTypeImage:
			var source *imageSource
			if part.URL != "" {
				// URL 格式图像
				source = &imageSource{
					Type: "url",
					URL:  part.URL,
				}
			} else {
				// base64 格式图像
				source = &imageSource{
					Type:      "base64",
					MediaType: part.MediaType,
					Data:      part.Data,
				}
			}
			msg.Content = append(msg.Content, contentBlock{
				Type:   "image",
				Source: source,
			})
		case domain.ContentTypeToolUse:
			input := part.ToolInput
			if input == nil {
				input = make(map[string]any) // Anthropic 要求 input 字段必须存在
			}
			msg.Content = append(msg.Content, contentBlock{
				Type:  "tool_use",
				ID:    part.ToolID,
				Name:  part.ToolName,
				Input: input,
			})
		case domain.ContentTypeToolResult:
			msg.Content = append(msg.Content, contentBlock{
				Type:      "tool_result",
				ToolUseID: part.ToolUseID,
				Content:   part.Text,
				IsError:   part.IsError,
			})
		case domain.ContentTypeThinking:
			msg.Content = append(msg.Content, contentBlock{
				Type:     "thinking",
				Thinking: part.Thinking,
			})
		}
	}

	return msg
}

func (p *Provider) fromClaudeResponse(resp *messagesResponse) *domain.ChatResponse {
	result := &domain.ChatResponse{
		ID:           resp.ID,
		Model:        resp.Model,
		FinishReason: mapStopReason(resp.StopReason),
	}

	for _, block := range resp.Content {
		switch block.Type {
		case "text":
			result.Content = append(result.Content, domain.ContentPart{
				Type: domain.ContentTypeText,
				Text: block.Text,
			})
		case "tool_use":
			result.Content = append(result.Content, domain.ContentPart{
				Type:      domain.ContentTypeToolUse,
				ToolID:    block.ID,
				ToolName:  block.Name,
				ToolInput: block.Input,
			})
		case "thinking":
			result.Content = append(result.Content, domain.ContentPart{
				Type:     domain.ContentTypeThinking,
				Thinking: block.Thinking,
			})
		}
	}

	if resp.Usage != nil {
		result.Usage = &domain.TokenUsage{
			PromptTokens:     resp.Usage.InputTokens,
			CompletionTokens: resp.Usage.OutputTokens,
			TotalTokens:      resp.Usage.InputTokens + resp.Usage.OutputTokens,
		}
	}

	return result
}

// ListModels returns a list of available Claude models.
func (p *Provider) ListModels(ctx context.Context) ([]string, error) {
	// Anthropic doesn't have a public models API, return known models
	return []string{
		"claude-3-5-sonnet-20241022",
		"claude-3-5-haiku-20241022",
		"claude-3-opus-20240229",
		"claude-3-sonnet-20240229",
		"claude-3-haiku-20240307",
	}, nil
}
