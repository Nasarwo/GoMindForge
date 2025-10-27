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
	grokAPIURL       = "https://api.x.ai/v1/chat/completions"
	grokDefaultModel = "grok-beta"
)

type GrokProvider struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// NewGrokProvider создает новый провайдер Grok
func NewGrokProvider() *GrokProvider {
	apiKey := os.Getenv("GROK_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("XAI_API_KEY") // Альтернативное имя переменной
	}

	return &GrokProvider{
		apiKey:  apiKey,
		baseURL: grokAPIURL,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// GetName возвращает имя провайдера
func (p *GrokProvider) GetName() string {
	return "grok"
}

// GetDefaultModel возвращает модель по умолчанию
func (p *GrokProvider) GetDefaultModel() string {
	return grokDefaultModel
}

// Chat отправляет запрос к Grok API
func (p *GrokProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	if p.apiKey == "" {
		return nil, fmt.Errorf("%w: GROK_API_KEY or XAI_API_KEY not set", ErrAPIKeyMissing)
	}

	// Подготавливаем запрос в формате Grok (OpenAI-совместимый)
	grokReq := map[string]interface{}{
		"model":    req.Model,
		"messages": convertMessages(req.Messages),
		"stream":   false,
	}

	// Если модель не указана, используем модель по умолчанию
	if grokReq["model"] == "" {
		grokReq["model"] = grokDefaultModel
	}

	jsonData, err := json.Marshal(grokReq)
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

	var grokResp struct {
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

	if err := json.NewDecoder(resp.Body).Decode(&grokResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(grokResp.Choices) == 0 {
		return nil, fmt.Errorf("%w: no choices in response", ErrAPIRequestFailed)
	}

	response := &ChatResponse{
		Content: grokResp.Choices[0].Message.Content,
		Model:   grokResp.Model,
	}
	response.Usage.PromptTokens = grokResp.Usage.PromptTokens
	response.Usage.CompletionTokens = grokResp.Usage.CompletionTokens
	response.Usage.TotalTokens = grokResp.Usage.TotalTokens

	return response, nil
}
