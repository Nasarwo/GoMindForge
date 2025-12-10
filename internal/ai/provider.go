package ai

import (
	"context"
)

// Message представляет сообщение в диалоге
type Message struct {
	Role    string `json:"role"` // "user", "assistant", "system"
	Content string `json:"content"`
}

// ChatRequest представляет запрос к AI провайдеру
type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream,omitempty"`
}

// ChatResponse представляет ответ от AI провайдера
type ChatResponse struct {
	Content string `json:"content"`
	Model   string `json:"model"`
	Usage   struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// Provider определяет интерфейс для работы с AI провайдерами
type Provider interface {
	// Chat отправляет запрос к AI и возвращает ответ
	Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error)

	// GetDefaultModel возвращает модель по умолчанию для провайдера
	GetDefaultModel() string

	// GetName возвращает имя провайдера
	GetName() string
}

// ProviderFactory создает провайдера по имени
type ProviderFactory struct {
	providers map[string]Provider
}

// NewProviderFactory создает новую фабрику провайдеров
func NewProviderFactory() *ProviderFactory {
	factory := &ProviderFactory{
		providers: make(map[string]Provider),
	}

	// Регистрируем провайдеры
	factory.Register("deepseek", NewDeepSeekProvider())
	factory.Register("openrouter", NewOpenRouterProvider())
	factory.Register("grok", NewGrokProvider())
	factory.Register("gigachat", NewGigaChatProvider())

	return factory
}

// Register регистрирует провайдера
func (f *ProviderFactory) Register(name string, provider Provider) {
	f.providers[name] = provider
}

// Get возвращает провайдера по имени
func (f *ProviderFactory) Get(name string) (Provider, error) {
	provider, exists := f.providers[name]
	if !exists {
		return nil, ErrProviderNotFound
	}
	return provider, nil
}

// List возвращает список доступных провайдеров
func (f *ProviderFactory) List() []string {
	var names []string
	for name := range f.providers {
		names = append(names, name)
	}
	return names
}
