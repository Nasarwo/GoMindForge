package ai

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"
)

const (
	gigachatAPIURL       = "https://gigachat.devices.sberbank.ru/api/v1/chat/completions"
	gigachatOAuthURL     = "https://ngw.devices.sberbank.ru:9443/api/v2/oauth"
	gigachatDefaultModel = "GigaChat"
	gigachatScope        = "GIGACHAT_API_PERS"
	// Токен действует 30 минут, обновляем за 5 минут до истечения
	tokenRefreshMargin = 5 * time.Minute
)

type GigaChatProvider struct {
	authKey      string
	clientID     string
	baseURL      string
	client       *http.Client
	tokenMutex   sync.RWMutex
	accessToken  string
	tokenExpires time.Time
}

// NewGigaChatProvider создает новый провайдер GigaChat
func NewGigaChatProvider() *GigaChatProvider {
	// Получаем Authorization key (Base64 encoded client_id:client_secret)
	authKey := os.Getenv("GIGACHAT_AUTH_KEY")
	if authKey == "" {
		// Используем предоставленный ключ по умолчанию
		authKey = "MDE5YjBhOWEtOTdiNS03MmVlLWI5NGMtYjYyN2EwMjhhNWRkOjUzNTUxM2VmLTQ4MWItNGE0Yi1hNDkzLWMyYTRjZDQ2NmVhNA=="
	}

	clientID := os.Getenv("GIGACHAT_CLIENT_ID")
	if clientID == "" {
		// Используем предоставленный client ID по умолчанию
		clientID = "019b0a9a-97b5-72ee-b94c-b627a028a5dd"
	}

	// Если указан прямой access token (для обратной совместимости)
	accessToken := os.Getenv("GIGACHAT_ACCESS_TOKEN")
	if accessToken == "" {
		accessToken = os.Getenv("GIGACHAT_API_KEY")
	}

	// Создаем HTTP клиент с поддержкой пропуска проверки SSL для OAuth endpoint
	// (GigaChat использует самоподписанный сертификат)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	provider := &GigaChatProvider{
		authKey:  authKey,
		clientID: clientID,
		baseURL:  gigachatAPIURL,
		client: &http.Client{
			Timeout:   60 * time.Second,
			Transport: tr,
		},
	}

	// Если указан прямой токен, используем его (но он может быть устаревшим)
	if accessToken != "" {
		provider.accessToken = accessToken
		// Устанавливаем время истечения в прошлом, чтобы при первом запросе получить новый токен
		provider.tokenExpires = time.Now().Add(-1 * time.Minute)
	}

	return provider
}

// GetName возвращает имя провайдера
func (p *GigaChatProvider) GetName() string {
	return "gigachat"
}

// GetDefaultModel возвращает модель по умолчанию
func (p *GigaChatProvider) GetDefaultModel() string {
	return gigachatDefaultModel
}

// getAccessToken получает или обновляет access token через OAuth
func (p *GigaChatProvider) getAccessToken(ctx context.Context) (string, error) {
	p.tokenMutex.RLock()
	// Проверяем, не истек ли токен (с запасом времени)
	if p.accessToken != "" && time.Now().Before(p.tokenExpires.Add(-tokenRefreshMargin)) {
		token := p.accessToken
		p.tokenMutex.RUnlock()
		return token, nil
	}
	p.tokenMutex.RUnlock()

	// Токен истек или отсутствует, получаем новый
	p.tokenMutex.Lock()
	defer p.tokenMutex.Unlock()

	// Двойная проверка (на случай, если другой горутина уже обновила токен)
	if p.accessToken != "" && time.Now().Before(p.tokenExpires.Add(-tokenRefreshMargin)) {
		return p.accessToken, nil
	}

	// Генерируем уникальный RqUID (UUID v4)
	rqUID := generateUUID()

	// Подготавливаем запрос
	data := url.Values{}
	data.Set("scope", gigachatScope)

	req, err := http.NewRequestWithContext(ctx, "POST", gigachatOAuthURL, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create OAuth request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("RqUID", rqUID)
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", p.authKey))

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get access token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("OAuth request failed: status %d, body: %s", resp.StatusCode, string(body))
	}

	var oauthResp struct {
		AccessToken string `json:"access_token"`
		ExpiresAt   int64  `json:"expires_at"`
		TokenType   string `json:"token_type"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&oauthResp); err != nil {
		return "", fmt.Errorf("failed to decode OAuth response: %w", err)
	}

	if oauthResp.AccessToken == "" {
		return "", fmt.Errorf("access token is empty in OAuth response")
	}

	// Сохраняем токен и время истечения
	p.accessToken = oauthResp.AccessToken
	if oauthResp.ExpiresAt > 0 {
		p.tokenExpires = time.Unix(oauthResp.ExpiresAt, 0)
	} else {
		// Если expires_at не указан, используем 30 минут по умолчанию
		p.tokenExpires = time.Now().Add(30 * time.Minute)
	}

	return p.accessToken, nil
}

// Chat отправляет запрос к GigaChat API
func (p *GigaChatProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	// Получаем актуальный access token
	accessToken, err := p.getAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
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
		"model":       model,
		"messages":    gigachatMessages,
		"stream":      false,
		"temperature": 0.7,
		"max_tokens":  2000,
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
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrAPIRequestFailed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		// Если токен истек (401), попробуем получить новый и повторить запрос
		if resp.StatusCode == http.StatusUnauthorized {
			// Сбрасываем токен и получаем новый
			p.tokenMutex.Lock()
			p.accessToken = ""
			p.tokenMutex.Unlock()

			// Повторяем запрос с новым токеном
			return p.Chat(ctx, req)
		}
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

// generateUUID генерирует UUID v4 (для RqUID)
func generateUUID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		// Fallback: используем время, если crypto/rand недоступен
		for i := range b {
			b[i] = byte(time.Now().UnixNano() % 256)
		}
	}
	// Устанавливаем версию (4) и вариант (10)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80

	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
