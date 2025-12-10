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
	gigachatAPIURL       = "https://gigachat.devices.sberbank.ru/api/v1/chat/completions"
	gigachatDefaultModel = "GigaChat"
)

type GigaChatProvider struct {
	accessToken string
	baseURL     string
	client      *http.Client
}

// NewGigaChatProvider создает новый провайдер GigaChat
func NewGigaChatProvider() *GigaChatProvider {
	accessToken := os.Getenv("GIGACHAT_ACCESS_TOKEN")
	if accessToken == "" {
		accessToken = os.Getenv("GIGACHAT_API_KEY") // Альтернативное имя для совместимости
	}

	return &GigaChatProvider{
		accessToken: accessToken,
		baseURL:     gigachatAPIURL,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// GetName возвращает имя провайдера
func (p *GigaChatProvider) GetName() string {
	return "gigachat"
}

// GetDefaultModel возвращает модель по умолчанию
func (p *GigaChatProvider) GetDefaultModel() string {
	return gigachatDefaultModel
}

// Chat отправляет запрос к GigaChat API
func (p *GigaChatProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	if p.accessToken == "" {
		return nil, fmt.Errorf("%w: GIGACHAT_ACCESS_TOKEN or GIGACHAT_API_KEY not set", ErrAPIKeyMissing)
	}

	// Определяем модель
	model := req.Model
	if model == "" {
		model = gigachatDefaultModel
	}

	// Конвертируем сообщения в формат GigaChat (OpenAI-совместимый формат)
	gigachatMessages := convertMessages(req.Messages)

	// Подготавливаем запрос в формате GigaChat
	gigachatReq := map[string]interface{}{
		"model":    model,
		"messages": gigachatMessages,
		"stream":   false,
		"temperature": 0.7,
		"max_tokens": 2000,
	}

	jsonData, err := json.Marshal(gigachatReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.accessToken))

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrAPIRequestFailed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%w: status %d, body: %s", ErrAPIRequestFailed, resp.StatusCode, string(body))
	}

	var gigachatResp struct {
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

	if err := json.NewDecoder(resp.Body).Decode(&gigachatResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(gigachatResp.Choices) == 0 {
		return nil, fmt.Errorf("%w: no choices in response", ErrAPIRequestFailed)
	}

	response := &ChatResponse{
		Content: gigachatResp.Choices[0].Message.Content,
		Model:   gigachatResp.Model,
	}
	response.Usage.PromptTokens = gigachatResp.Usage.PromptTokens
	response.Usage.CompletionTokens = gigachatResp.Usage.CompletionTokens
	response.Usage.TotalTokens = gigachatResp.Usage.TotalTokens

	return response, nil
}
