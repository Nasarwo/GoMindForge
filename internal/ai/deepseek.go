package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	deepseekAPIURL       = "https://api.deepseek.com/v1/chat/completions"
	deepseekDefaultModel = "deepseek-chat"
)

type DeepSeekProvider struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// NewDeepSeekProvider создает новый провайдер DeepSeek
func NewDeepSeekProvider() *DeepSeekProvider {
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("DEEPSEEK_API_KEY") // Можно использовать значение по умолчанию для тестов
	}

	return &DeepSeekProvider{
		apiKey:  apiKey,
		baseURL: deepseekAPIURL,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// GetName возвращает имя провайдера
func (p *DeepSeekProvider) GetName() string {
	return "deepseek"
}

// GetDefaultModel возвращает модель по умолчанию
func (p *DeepSeekProvider) GetDefaultModel() string {
	return deepseekDefaultModel
}

// Chat отправляет запрос к DeepSeek API
func (p *DeepSeekProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	if p.apiKey == "" {
		return nil, fmt.Errorf("%w: DEEPSEEK_API_KEY not set", ErrAPIKeyMissing)
	}

	// Подготавливаем запрос в формате DeepSeek
	deepseekReq := map[string]interface{}{
		"model":    req.Model,
		"messages": convertMessages(req.Messages),
		"stream":   false,
	}

	// Если модель не указана, используем модель по умолчанию
	if deepseekReq["model"] == "" {
		deepseekReq["model"] = deepseekDefaultModel
	}

	jsonData, err := json.Marshal(deepseekReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.apiKey))

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrAPIRequestFailed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%w: status %d, body: %s", ErrAPIRequestFailed, resp.StatusCode, string(body))
	}

	var deepseekResp struct {
		ID      string `json:"id"`
		Choices []struct {
			Message struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Model string `json:"model"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&deepseekResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(deepseekResp.Choices) == 0 {
		return nil, fmt.Errorf("%w: no choices in response", ErrAPIRequestFailed)
	}

	response := &ChatResponse{
		Content: deepseekResp.Choices[0].Message.Content,
		Model:   deepseekResp.Model,
	}
	response.Usage.PromptTokens = deepseekResp.Usage.PromptTokens
	response.Usage.CompletionTokens = deepseekResp.Usage.CompletionTokens
	response.Usage.TotalTokens = deepseekResp.Usage.TotalTokens

	return response, nil
}

// convertMessages конвертирует сообщения в формат API
func convertMessages(messages []Message) []map[string]string {
	result := make([]map[string]string, len(messages))
	for i, msg := range messages {
		result[i] = map[string]string{
			"role":    msg.Role,
			"content": msg.Content,
		}
	}
	return result
}
