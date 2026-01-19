// 工具调用 (Function Calling) 示例
// 演示如何通过 AI Gateway 使用 LLM 的工具调用功能
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	GatewayURL = "http://localhost:8081/v1/chat/completions"
	Model      = "claude-sonnet-4-5" // 或其他支持工具调用的模型
)

// 请求/响应结构
type Message struct {
	Role       string     `json:"role"`
	Content    string     `json:"content,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
}

type ToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

type Tool struct {
	Type     string   `json:"type"`
	Function Function `json:"function"`
}

type Function struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Tools    []Tool    `json:"tools,omitempty"`
}

type ChatResponse struct {
	Choices []struct {
		Message      Message `json:"message"`
		FinishReason string  `json:"finish_reason"`
	} `json:"choices"`
}

// 模拟的工具函数
func getWeather(city string) string {
	// 模拟天气数据
	weathers := map[string]string{
		"北京": "晴天，25°C",
		"上海": "多云，28°C",
		"深圳": "阴天，30°C",
	}
	if w, ok := weathers[city]; ok {
		return w
	}
	return fmt.Sprintf("%s: 晴，22°C", city)
}

func getCurrentTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

// 执行工具调用
func executeToolCall(name string, args map[string]any) string {
	switch name {
	case "get_weather":
		city, _ := args["city"].(string)
		return getWeather(city)
	case "get_current_time":
		return getCurrentTime()
	default:
		return "未知工具"
	}
}

func main() {
	fmt.Println("=== AI Gateway 工具调用示例 ===\n")

	// 定义可用的工具
	tools := []Tool{
		{
			Type: "function",
			Function: Function{
				Name:        "get_weather",
				Description: "获取指定城市的天气信息",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"city": map[string]any{
							"type":        "string",
							"description": "城市名称，如北京、上海",
						},
					},
					"required": []string{"city"},
				},
			},
		},
		{
			Type: "function",
			Function: Function{
				Name:        "get_current_time",
				Description: "获取当前时间",
				Parameters: map[string]any{
					"type":       "object",
					"properties": map[string]any{},
				},
			},
		},
	}

	// 用户问题
	userQuestion := "现在几点了？北京和上海的天气怎么样？"
	fmt.Printf("用户: %s\n\n", userQuestion)

	messages := []Message{
		{Role: "user", Content: userQuestion},
	}

	// 第一次调用：让 LLM 决定使用哪些工具
	fmt.Println("📤 发送请求给 LLM...")
	resp, err := chat(messages, tools)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}

	if len(resp.Choices) == 0 {
		fmt.Println("无响应")
		return
	}

	assistantMsg := resp.Choices[0].Message
	finishReason := resp.Choices[0].FinishReason

	// 检查是否需要调用工具
	if finishReason == "tool_calls" && len(assistantMsg.ToolCalls) > 0 {
		fmt.Printf("🔧 LLM 请求调用 %d 个工具:\n", len(assistantMsg.ToolCalls))

		// 添加助手消息到历史
		messages = append(messages, assistantMsg)

		// 执行每个工具调用
		for _, tc := range assistantMsg.ToolCalls {
			var args map[string]any
			json.Unmarshal([]byte(tc.Function.Arguments), &args)

			fmt.Printf("   - %s(%v)\n", tc.Function.Name, args)

			// 执行工具
			result := executeToolCall(tc.Function.Name, args)
			fmt.Printf("     结果: %s\n", result)

			// 添加工具结果到消息
			messages = append(messages, Message{
				Role:       "tool",
				ToolCallID: tc.ID,
				Content:    result,
			})
		}

		// 第二次调用：让 LLM 基于工具结果生成最终回答
		fmt.Println("\n📤 发送工具结果给 LLM...")
		resp, err = chat(messages, nil) // 第二次不需要传 tools
		if err != nil {
			fmt.Printf("错误: %v\n", err)
			return
		}

		if len(resp.Choices) > 0 {
			fmt.Printf("\n助手: %s\n", resp.Choices[0].Message.Content)
		}
	} else {
		// LLM 直接回答，没有调用工具
		fmt.Printf("助手: %s\n", assistantMsg.Content)
	}
}

func chat(messages []Message, tools []Tool) (*ChatResponse, error) {
	reqBody := ChatRequest{
		Model:    Model,
		Messages: messages,
		Tools:    tools,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", GatewayURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer sk-gateway-test-key")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", string(respBody))
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return nil, err
	}

	return &chatResp, nil
}
