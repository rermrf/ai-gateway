// 多模态/图片分析示例
// 演示如何通过 AI Gateway 发送图片给 LLM 进行视觉分析
package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

const (
	GatewayURL = "http://localhost:8081/v1/chat/completions"
	Model      = "Qwen/Qwen3-VL-32B-Thinking" // 使用支持视觉的模型
)

// 请求结构
type Message struct {
	Role    string        `json:"role"`
	Content []ContentPart `json:"content,omitempty"`
}

type ContentPart struct {
	Type     string    `json:"type"`
	Text     string    `json:"text,omitempty"`
	ImageURL *ImageURL `json:"image_url,omitempty"`
}

type ImageURL struct {
	URL    string `json:"url"`
	Detail string `json:"detail,omitempty"` // auto, low, high
}

type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type ChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func main() {
	fmt.Println("=== AI Gateway 多模态/图片分析示例 ===\n")

	// 检查命令行参数
	if len(os.Args) < 2 {
		fmt.Println("用法: go run main.go <图片路径或URL>")
		fmt.Println()
		fmt.Println("示例:")
		fmt.Println("  go run main.go ./photo.jpg           # 本地图片")
		fmt.Println("  go run main.go https://example.com/image.png  # 网络图片")
		return
	}

	imagePath := os.Args[1]
	var imageURL string

	// 判断是 URL 还是本地文件
	if strings.HasPrefix(imagePath, "http://") || strings.HasPrefix(imagePath, "https://") {
		imageURL = imagePath
		fmt.Printf("📷 使用网络图片: %s\n", imageURL)
	} else {
		// 读取本地图片并转为 base64
		data, err := os.ReadFile(imagePath)
		if err != nil {
			fmt.Printf("错误: 无法读取图片文件: %v\n", err)
			return
		}

		// 检测图片类型
		mediaType := detectMediaType(imagePath)
		base64Data := base64.StdEncoding.EncodeToString(data)
		imageURL = fmt.Sprintf("data:%s;base64,%s", mediaType, base64Data)

		fmt.Printf("📷 本地图片: %s (%s, %d bytes)\n", imagePath, mediaType, len(data))
	}

	// 构建多模态消息
	message := Message{
		Role: "user",
		Content: []ContentPart{
			{
				Type: "text",
				Text: "请详细描述这张图片的内容，包括你看到的所有细节。用中文回答。",
			},
			{
				Type: "image_url",
				ImageURL: &ImageURL{
					URL:    imageURL,
					Detail: "auto",
				},
			},
		},
	}

	fmt.Println("\n📤 发送图片给 LLM 进行分析...")
	fmt.Println()

	resp, err := chat([]Message{message})
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}

	if len(resp.Choices) > 0 {
		fmt.Printf("🤖 LLM 分析结果:\n\n%s\n", resp.Choices[0].Message.Content)
	}
}

func detectMediaType(path string) string {
	lower := strings.ToLower(path)
	switch {
	case strings.HasSuffix(lower, ".png"):
		return "image/png"
	case strings.HasSuffix(lower, ".gif"):
		return "image/gif"
	case strings.HasSuffix(lower, ".webp"):
		return "image/webp"
	default:
		return "image/jpeg"
	}
}

func chat(messages []Message) (*ChatResponse, error) {
	reqBody := ChatRequest{
		Model:    Model,
		Messages: messages,
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
	req.Header.Set("Authorization", "Bearer test-key")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return nil, err
	}

	return &chatResp, nil
}
