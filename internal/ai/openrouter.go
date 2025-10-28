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
	openrouterAPIURL       = "https://openrouter.ai/api/v1/chat/completions"
	openrouterDefaultModel = "deepseek/deepseek-chat"
)

type OpenRouterProvider struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// NewOpenRouterProvider создает новый провайдер OpenRouter
func NewOpenRouterProvider() *OpenRouterProvider {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("OPENROUTER_API_KEY") // Можно использовать значение по умолчанию для тестов
	}

	return &OpenRouterProvider{
		apiKey:  apiKey,
		baseURL: openrouterAPIURL,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// GetName возвращает имя провайдера
func (p *OpenRouterProvider) GetName() string {
	return "openrouter"
}

// GetDefaultModel возвращает модель по умолчанию
func (p *OpenRouterProvider) GetDefaultModel() string {
	return openrouterDefaultModel
}

// Chat отправляет запрос к OpenRouter API
func (p *OpenRouterProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	if p.apiKey == "" {
		return nil, fmt.Errorf("%w: OPENROUTER_API_KEY not set", ErrAPIKeyMissing)
	}

	// Подготавливаем запрос в формате OpenRouter
	openrouterReq := map[string]interface{}{
		"model":    req.Model,
		"messages": convertMessages(req.Messages),
		"stream":   false,
	}

	// Если модель не указана, используем модель по умолчанию
	if openrouterReq["model"] == "" {
		openrouterReq["model"] = openrouterDefaultModel
	}

	jsonData, err := json.Marshal(openrouterReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.apiKey))
	httpReq.Header.Set("HTTP-Referer", "https://mindforge.app") // Опционально, для идентификации приложения
	httpReq.Header.Set("X-Title", "MindForge")                  // Опционально, для идентификации приложения

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrAPIRequestFailed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%w: status %d, body: %s", ErrAPIRequestFailed, resp.StatusCode, string(body))
	}

	var openrouterResp struct {
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

	if err := json.NewDecoder(resp.Body).Decode(&openrouterResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(openrouterResp.Choices) == 0 {
		return nil, fmt.Errorf("%w: no choices in response", ErrAPIRequestFailed)
	}

	response := &ChatResponse{
		Content: openrouterResp.Choices[0].Message.Content,
		Model:   openrouterResp.Model,
	}
	response.Usage.PromptTokens = openrouterResp.Usage.PromptTokens
	response.Usage.CompletionTokens = openrouterResp.Usage.CompletionTokens
	response.Usage.TotalTokens = openrouterResp.Usage.TotalTokens

	return response, nil
}
