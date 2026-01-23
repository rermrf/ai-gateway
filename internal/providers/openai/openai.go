// Package openai 实现 OpenAI 提供商适配器。
package openai

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

// Provider 为 OpenAI API 实现 Provider 接口。
type Provider struct {
	apiKey  string
	baseURL string
	client  *http.Client
	logger  logger.Logger
}

// NewProvider 创建一个新的 OpenAI 提供商。
func NewProvider(apiKey, baseURL string, client *http.Client, l logger.Logger) *Provider {
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	if client == nil {
		client = http.DefaultClient
	}
	return &Provider{
		apiKey:  apiKey,
		baseURL: baseURL,
		client:  client,
		logger:  l.With(logger.String("provider", "openai")),
	}
}

func (p *Provider) Name() string { return "openai" }

func (p *Provider) SupportsStreaming() bool { return true }
func (p *Provider) SupportsTools() bool     { return true }
func (p *Provider) SupportsVision() bool    { return true }

// OpenAI API 类型
type chatRequest struct {
	Model             string          `json:"model"`
	Messages          []message       `json:"messages"`
	Stream            bool            `json:"stream,omitempty"`
	MaxTokens         int             `json:"max_tokens,omitempty"`
	Temperature       *float64        `json:"temperature,omitempty"`
	TopP              *float64        `json:"top_p,omitempty"`
	Stop              []string        `json:"stop,omitempty"`
	Tools             []tool          `json:"tools,omitempty"`
	ToolChoice        interface{}     `json:"tool_choice,omitempty"` // string or object
	ResponseFormat    *responseFormat `json:"response_format,omitempty"`
	PresencePenalty   *float64        `json:"presence_penalty,omitempty"`
	FrequencyPenalty  *float64        `json:"frequency_penalty,omitempty"`
	StreamOptions     *streamOptions  `json:"stream_options,omitempty"`
	ParallelToolCalls *bool           `json:"parallel_tool_calls,omitempty"`
}

type responseFormat struct {
	Type       string            `json:"type"` // text, json_object, json_schema
	JSONSchema *jsonSchemaConfig `json:"json_schema,omitempty"`
}

type jsonSchemaConfig struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Schema      map[string]any `json:"schema"`
	Strict      bool           `json:"strict,omitempty"`
}

type streamOptions struct {
	IncludeUsage bool `json:"include_usage"`
}

type message struct {
	Role             string      `json:"role"`
	Content          interface{} `json:"content"`                     // 字符串或 []contentPart
	ReasoningContent string      `json:"reasoning_content,omitempty"` // DeepSeek R1 思考内容
	Name             string      `json:"name,omitempty"`
	ToolCalls        []toolCall  `json:"tool_calls,omitempty"`
	ToolCallID       string      `json:"tool_call_id,omitempty"`
}

type contentPart struct {
	Type     string    `json:"type"`
	Text     string    `json:"text,omitempty"`
	ImageURL *imageURL `json:"image_url,omitempty"`
}

type imageURL struct {
	URL    string `json:"url"`
	Detail string `json:"detail,omitempty"`
}

type tool struct {
	Type     string   `json:"type"`
	Function function `json:"function"`
}

type function struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Parameters  map[string]any `json:"parameters,omitempty"`
}

type toolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function functionCall `json:"function"`
}

type functionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type chatResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Model   string   `json:"model"`
	Choices []choice `json:"choices"`
	Usage   *usage   `json:"usage,omitempty"`
}

type choice struct {
	Index        int      `json:"index"`
	Message      message  `json:"message"`
	Delta        *message `json:"delta,omitempty"`
	FinishReason string   `json:"finish_reason,omitempty"`
}

type usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Chat 发送非流式聊天请求。
func (p *Provider) Chat(ctx context.Context, req *domain.ChatRequest) (*domain.ChatResponse, error) {
	oaiReq := p.toOpenAIRequest(req)
	oaiReq.Stream = false

	body, err := json.Marshal(oaiReq)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewReader(body))
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
		p.logger.Error("OpenAI API error",
			logger.Int("status", resp.StatusCode),
			logger.String("body", string(respBody)),
		)
		return nil, fmt.Errorf("%w: status %d", errs.ErrProviderError, resp.StatusCode)
	}

	var oaiResp chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&oaiResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return p.fromOpenAIResponse(&oaiResp), nil
}

// ChatStream 发送流式聊天请求。
func (p *Provider) ChatStream(ctx context.Context, req *domain.ChatRequest) (<-chan domain.StreamDelta, error) {
	oaiReq := p.toOpenAIRequest(req)
	oaiReq.Stream = true
	oaiReq.StreamOptions = &streamOptions{IncludeUsage: true}

	body, err := json.Marshal(oaiReq)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewReader(body))
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
		p.logger.Error("OpenAI API error",
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
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		var chunk chatResponse
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			p.logger.Warn("failed to parse SSE chunk", logger.Error(err))
			continue
		}

		if len(chunk.Choices) > 0 {
			delta := p.deltaFromChoice(&chunk.Choices[0], chunk.Usage)
			ch <- delta
		}
	}
}

func (p *Provider) deltaFromChoice(c *choice, u *usage) domain.StreamDelta {
	delta := domain.StreamDelta{Type: "content"}

	if c.Delta != nil {
		// 处理思考内容 (DeepSeek R1 等模型)
		if c.Delta.ReasoningContent != "" {
			delta.Type = "thinking"
			delta.Content = &domain.ContentPart{
				Type:     domain.ContentTypeThinking,
				Thinking: c.Delta.ReasoningContent,
			}
			return delta
		}

		// 处理常规文本内容
		switch content := c.Delta.Content.(type) {
		case string:
			if content != "" {
				delta.Content = &domain.ContentPart{
					Type: domain.ContentTypeText,
					Text: content,
				}
			}
		}

		// 处理工具调用
		if len(c.Delta.ToolCalls) > 0 {
			tc := c.Delta.ToolCalls[0]
			delta.Type = "tool_use"
			delta.Content = &domain.ContentPart{
				Type:     domain.ContentTypeToolUse,
				ToolID:   tc.ID,
				ToolName: tc.Function.Name,
			}
			if tc.Function.Arguments != "" {
				var args map[string]any
				if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err == nil {
					delta.Content.ToolInput = args
				}
			}
		}
	}

	if c.FinishReason != "" {
		delta.Type = "done"
		delta.FinishReason = domain.FinishReason(c.FinishReason)
	}

	if u != nil {
		delta.Usage = &domain.TokenUsage{
			PromptTokens:     u.PromptTokens,
			CompletionTokens: u.CompletionTokens,
			TotalTokens:      u.TotalTokens,
		}
	}

	return delta
}

func (p *Provider) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
}

func (p *Provider) toOpenAIRequest(req *domain.ChatRequest) *chatRequest {
	oaiReq := &chatRequest{
		Model:            req.Model,
		MaxTokens:        req.MaxTokens,
		Temperature:      req.Temperature,
		TopP:             req.TopP,
		Stop:             req.StopSequences,
		PresencePenalty:  req.PresencePenalty,
		FrequencyPenalty: req.FrequencyPenalty,
	}

	// 转换消息
	for _, m := range req.Messages {
		oaiReq.Messages = append(oaiReq.Messages, p.toOpenAIMessage(m))
	}

	// 如果存在系统消息，则先添加系统消息
	if req.System != "" {
		sysMsg := message{Role: "system", Content: req.System}
		oaiReq.Messages = append([]message{sysMsg}, oaiReq.Messages...)
	}

	// 转换工具
	for _, t := range req.Tools {
		oaiReq.Tools = append(oaiReq.Tools, tool{
			Type: "function",
			Function: function{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  t.InputSchema,
			},
		})
	}

	// 转换 tool_choice
	if req.ToolChoice != nil {
		oaiReq.ToolChoice = p.toOpenAIToolChoice(req.ToolChoice)
		// 处理并行工具调用
		if req.ToolChoice.DisableParallelToolUse {
			falseVal := false
			oaiReq.ParallelToolCalls = &falseVal
		}
	}

	// 转换 response_format
	if req.ResponseFormat != nil {
		oaiReq.ResponseFormat = p.toOpenAIResponseFormat(req.ResponseFormat)
	}

	return oaiReq
}

// toOpenAIToolChoice 将统一格式的 ToolChoice 转换为 OpenAI 格式
func (p *Provider) toOpenAIToolChoice(tc *domain.ToolChoice) interface{} {
	if tc == nil {
		return nil
	}

	switch tc.Type {
	case domain.ToolChoiceNone:
		return "none"
	case domain.ToolChoiceAuto:
		return "auto"
	case domain.ToolChoiceAny:
		return "required"
	case domain.ToolChoiceTool:
		return map[string]interface{}{
			"type": "function",
			"function": map[string]string{
				"name": tc.Name,
			},
		}
	default:
		return "auto"
	}
}

// toOpenAIResponseFormat 将统一格式的 ResponseFormat 转换为 OpenAI 格式
func (p *Provider) toOpenAIResponseFormat(rf *domain.ResponseFormat) *responseFormat {
	if rf == nil {
		return nil
	}

	format := &responseFormat{
		Type: string(rf.Type),
	}

	if rf.Type == domain.ResponseFormatJSONSchema && rf.JSONSchema != nil {
		format.JSONSchema = &jsonSchemaConfig{
			Name:        rf.JSONSchema.Name,
			Description: rf.JSONSchema.Description,
			Schema:      rf.JSONSchema.Schema,
			Strict:      rf.JSONSchema.Strict,
		}
	}

	return format
}

func (p *Provider) toOpenAIMessage(m domain.Message) message {
	msg := message{
		Role: string(m.Role),
		Name: m.Name,
	}

	// 检查是否所有内容都是纯文本
	allText := true
	for _, part := range m.Content {
		if part.Type != domain.ContentTypeText {
			allText = false
			break
		}
	}

	if allText && len(m.Content) == 1 {
		// 简单文本消息
		msg.Content = m.Content[0].Text
	} else {
		// 多部分内容
		var parts []contentPart
		for _, part := range m.Content {
			switch part.Type {
			case domain.ContentTypeText:
				parts = append(parts, contentPart{Type: "text", Text: part.Text})
			case domain.ContentTypeImage:
				url := part.URL
				if url == "" && part.Data != "" {
					url = fmt.Sprintf("data:%s;base64,%s", part.MediaType, part.Data)
				}
				parts = append(parts, contentPart{
					Type:     "image_url",
					ImageURL: &imageURL{URL: url},
				})
			case domain.ContentTypeToolUse:
				// OpenAI 在 assistant 消息中使用 tool_calls
				args, _ := json.Marshal(part.ToolInput)
				msg.ToolCalls = append(msg.ToolCalls, toolCall{
					ID:   part.ToolID,
					Type: "function",
					Function: functionCall{
						Name:      part.ToolName,
						Arguments: string(args),
					},
				})
			case domain.ContentTypeToolResult:
				msg.Role = "tool"
				msg.ToolCallID = part.ToolUseID
				msg.Content = part.Text
			}
		}
		if len(parts) > 0 {
			msg.Content = parts
		}
	}

	return msg
}

func (p *Provider) fromOpenAIResponse(resp *chatResponse) *domain.ChatResponse {
	result := &domain.ChatResponse{
		ID:    resp.ID,
		Model: resp.Model,
	}

	if len(resp.Choices) > 0 {
		c := resp.Choices[0]
		result.FinishReason = domain.FinishReason(c.FinishReason)

		// 解析内容
		switch content := c.Message.Content.(type) {
		case string:
			result.Content = append(result.Content, domain.ContentPart{
				Type: domain.ContentTypeText,
				Text: content,
			})
		}

		// 解析工具调用
		for _, tc := range c.Message.ToolCalls {
			var args map[string]any
			json.Unmarshal([]byte(tc.Function.Arguments), &args)
			result.Content = append(result.Content, domain.ContentPart{
				Type:      domain.ContentTypeToolUse,
				ToolID:    tc.ID,
				ToolName:  tc.Function.Name,
				ToolInput: args,
			})
		}
	}

	if resp.Usage != nil {
		result.Usage = &domain.TokenUsage{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		}
	}

	return result
}

// ListModels 返回可用模型列表。
func (p *Provider) ListModels(ctx context.Context) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", p.baseURL+"/models", nil)
	if err != nil {
		return nil, err
	}
	p.setHeaders(req)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	models := make([]string, len(result.Data))
	for i, m := range result.Data {
		models[i] = m.ID
	}
	return models, nil
}
