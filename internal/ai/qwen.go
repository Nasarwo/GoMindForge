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
	qwenDefaultAPIURL = "https://api.mulerouter.ai/vendors/openai/v1/chat/completions"
	qwenDefaultModel  = "qwen3-max"
)

type QwenProvider struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// NewQwenProvider создает новый провайдер Qwen
func NewQwenProvider() *QwenProvider {
	apiKey := os.Getenv("QWEN_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("DASHSCOPE_API_KEY") // Альтернативное имя переменной
	}

	baseURL := os.Getenv("QWEN_API_BASE_URL")
	if baseURL == "" {
		baseURL = qwenDefaultAPIURL
	}

	return &QwenProvider{
		apiKey:  apiKey,
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// GetName возвращает имя провайдера
func (p *QwenProvider) GetName() string {
	return "qwen"
}

// GetDefaultModel возвращает модель по умолчанию
func (p *QwenProvider) GetDefaultModel() string {
	return qwenDefaultModel
}

// Chat отправляет запрос к Qwen API
func (p *QwenProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	if p.apiKey == "" {
		return nil, fmt.Errorf("%w: QWEN_API_KEY or DASHSCOPE_API_KEY not set", ErrAPIKeyMissing)
	}

	// Определяем модель
	model := req.Model
	if model == "" {
		model = qwenDefaultModel
	}

	// Конвертируем сообщения в формат OpenAI (MuleRouter использует OpenAI-совместимый формат)
	qwenMessages := convertMessages(req.Messages)

	// Подготавливаем запрос в формате OpenAI (MuleRouter совместим с OpenAI API)
	qwenReq := map[string]interface{}{
		"model":    model,
		"messages": qwenMessages,
		"stream":   false,
	}

	jsonData, err := json.Marshal(qwenReq)
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

	var qwenResp struct {
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

	if err := json.NewDecoder(resp.Body).Decode(&qwenResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(qwenResp.Choices) == 0 {
		return nil, fmt.Errorf("%w: no choices in response", ErrAPIRequestFailed)
	}

	response := &ChatResponse{
		Content: qwenResp.Choices[0].Message.Content,
		Model:   qwenResp.Model,
	}
	if response.Model == "" {
		response.Model = model
	}
	response.Usage.PromptTokens = qwenResp.Usage.PromptTokens
	response.Usage.CompletionTokens = qwenResp.Usage.CompletionTokens
	response.Usage.TotalTokens = qwenResp.Usage.TotalTokens

	return response, nil
}
