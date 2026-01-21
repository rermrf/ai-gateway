// 通过 AI Gateway 与 LLM 对话的示例（流式版本 + 思考支持）
// 使用 OpenAI 协议访问网关服务
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

const (
	// 网关地址
	GatewayURL = "http://localhost:8081/v1/chat/completions"
	// 使用的模型（支持思考的模型如 o1, deepseek-r1 等）
	Model = "deepseek-ai/DeepSeek-R1"
)

// Message 表示一条对话消息
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest 请求结构
type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

// StreamChunk SSE 流式响应的一个块
type StreamChunk struct {
	Choices []struct {
		Delta struct {
			Content          string `json:"content"`
			ReasoningContent string `json:"reasoning_content"` // 思考内容（部分模型支持）
		} `json:"delta"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
}

func main() {
	fmt.Println("=== AI Gateway 流式聊天示例（支持思考） ===")
	fmt.Println("输入消息与 LLM 对话，输入 'exit' 退出")
	fmt.Println()

	// 保存对话历史
	var history []Message

	// 可选：添加系统提示
	history = append(history, Message{
		Role:    "system",
		Content: "你是一个有帮助的助手。请用中文回答。",
	})

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("你: ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}
		if input == "exit" {
			fmt.Println("再见！")
			break
		}

		// 添加用户消息到历史
		history = append(history, Message{
			Role:    "user",
			Content: input,
		})

		// 调用流式 API
		response, err := chatStream(history)
		if err != nil {
			fmt.Printf("\n错误: %v\n", err)
			// 移除失败的用户消息
			history = history[:len(history)-1]
			continue
		}

		// 添加助手回复到历史
		history = append(history, Message{
			Role:    "assistant",
			Content: response,
		})
		fmt.Println()
	}
}

func chatStream(messages []Message) (string, error) {
	reqBody := ChatRequest{
		Model:    Model,
		Messages: messages,
		Stream:   true,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", GatewayURL, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer sk-ad0b8f5b588778a1ce89769ef28ccdc65e8b976b778bad2ada88d1f9bec053c7")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error: status %d", resp.StatusCode)
	}

	// 读取 SSE 流
	scanner := bufio.NewScanner(resp.Body)
	var fullContent strings.Builder
	inThinking := false

	for scanner.Scan() {
		line := scanner.Text()

		if line == "" {
			continue
		}

		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")

		if data == "[DONE]" {
			break
		}

		var chunk StreamChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}

		if len(chunk.Choices) > 0 {
			delta := chunk.Choices[0].Delta

			// 处理思考内容
			if delta.ReasoningContent != "" {
				if !inThinking {
					fmt.Print("\n💭 思考中: ")
					inThinking = true
				}
				fmt.Print(delta.ReasoningContent)
			}

			// 处理正常回复内容
			if delta.Content != "" {
				if inThinking {
					fmt.Print("\n\n助手: ")
					inThinking = false
				} else if fullContent.Len() == 0 {
					fmt.Print("助手: ")
				}
				fmt.Print(delta.Content)
				fullContent.WriteString(delta.Content)
			}
		}
	}

	fmt.Println()
	return fullContent.String(), nil
}
