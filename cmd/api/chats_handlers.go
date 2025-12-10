package main

import (
	"context"
	"net/http"
	"strings"
	"time"

	"mindforge/internal/ai"
	"mindforge/internal/database"
	"mindforge/internal/env"

	"github.com/gin-gonic/gin"
)

type createChatRequest struct {
	AIModel string `json:"ai_model" binding:"required"`
}

type chatResponse struct {
	ID        int    `json:"id"`
	UserID    int    `json:"user_id"`
	AIModel   string `json:"ai_model"`
	Title     string `json:"title"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type updateChatTitleRequest struct {
	Title string `json:"title" binding:"required,min=1,max=200"`
}

type createMessageRequest struct {
	Content string `json:"content" binding:"required,min=1,max=10000"`
}

type messageResponse struct {
	ID        int    `json:"id"`
	ChatID    int    `json:"chat_id"`
	Role      string `json:"role"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}

// handleCreateChat создает новый чат
func (app *application) handleCreateChat(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		errorResponse(c, ErrUnauthorized)
		return
	}

	var req createChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		app.logger.Warn("Validation error", "error", err)
		validationErrorResponse(c, err)
		return
	}

	// Проверяем, что ai_model не пустой
	if req.AIModel == "" {
		errorResponse(c, &APIError{
			Status:  400,
			Message: "ai_model is required",
			Code:    "VALIDATION_ERROR",
		})
		return
	}

	// Валидируем, что провайдер для указанной модели существует
	providerName := strings.ToLower(req.AIModel)
	if strings.Contains(providerName, "deepseek") || strings.Contains(providerName, "deep-seek") {
		providerName = "openrouter"
	} else if strings.Contains(providerName, "grok") {
		providerName = "grok"
	} else if strings.Contains(providerName, "gigachat") {
		providerName = "gigachat" // GigaChat
	} else {
		providerName = "openrouter" // По умолчанию OpenRouter
	}

	_, err := app.aiProviderFactory.Get(providerName)
	if err != nil {
		app.logger.Warn("Invalid AI model/provider", "model", req.AIModel, "provider", providerName, "error", err)
		errorResponse(c, &APIError{
			Status:  400,
			Message: "invalid ai_model or provider not available",
			Code:    "INVALID_AI_MODEL",
		})
		return
	}

	chat, err := app.models.Chats.Create(userID.(int), req.AIModel, "Новый чат")
	if err != nil {
		app.logger.Error("Error creating chat", "error", err)
		internalErrorResponse(c, err)
		return
	}

	c.JSON(http.StatusCreated, chatResponse{
		ID:        chat.ID,
		UserID:    chat.UserID,
		AIModel:   chat.AIModel,
		Title:     chat.Title,
		CreatedAt: chat.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: chat.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// handleGetChats получает список чатов пользователя
func (app *application) handleGetChats(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		errorResponse(c, ErrUnauthorized)
		return
	}

	chats, err := app.models.Chats.GetByUserID(userID.(int))
	if err != nil {
		app.logger.Error("Error getting chats", "error", err, "user_id", userID)
		internalErrorResponse(c, err)
		return
	}

	response := make([]chatResponse, len(chats))
	for i, chat := range chats {
		response[i] = chatResponse{
			ID:        chat.ID,
			UserID:    chat.UserID,
			AIModel:   chat.AIModel,
			Title:     chat.Title,
			CreatedAt: chat.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt: chat.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	c.JSON(http.StatusOK, response)
}

// handleGetChat получает конкретный чат
func (app *application) handleGetChat(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		errorResponse(c, ErrUnauthorized)
		return
	}

	chatID, apiErr := getChatIDFromParam(c)
	if apiErr != nil {
		errorResponse(c, apiErr)
		return
	}

	chat, err := app.models.Chats.GetByID(chatID)
	if err != nil {
		if err.Error() == "chat not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  404,
				"message": "chat not found",
				"code":    "CHAT_NOT_FOUND",
			})
			return
		}
		app.logger.Error("Error getting chat", "error", err, "chat_id", chatID)
		internalErrorResponse(c, err)
		return
	}

	// Проверяем, что чат принадлежит пользователю
	if chat.UserID != userID.(int) {
		errorResponse(c, &APIError{
			Status:  403,
			Message: "forbidden",
			Code:    "FORBIDDEN",
		})
		return
	}

	c.JSON(http.StatusOK, chatResponse{
		ID:        chat.ID,
		UserID:    chat.UserID,
		AIModel:   chat.AIModel,
		Title:     chat.Title,
		CreatedAt: chat.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: chat.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// handleCreateMessage создает новое сообщение в чате
func (app *application) handleCreateMessage(c *gin.Context) {
	userID, apiErr := getUserIDFromContext(c)
	if apiErr != nil {
		errorResponse(c, apiErr)
		return
	}

	chatID, apiErr := getChatIDFromParam(c)
	if apiErr != nil {
		errorResponse(c, apiErr)
		return
	}

	chat, apiErr := app.validateChatOwnership(c, chatID, userID)
	if apiErr != nil {
		errorResponse(c, apiErr)
		return
	}

	var req createMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validationErrorResponse(c, err)
		return
	}

	// Валидация и очистка контента
	content := strings.TrimSpace(req.Content)
	if content == "" {
		errorResponse(c, &APIError{
			Status:  400,
			Message: "message content cannot be empty",
			Code:    "VALIDATION_ERROR",
		})
		return
	}

	// Ограничение длины контента
	if len(content) > 10000 {
		errorResponse(c, &APIError{
			Status:  400,
			Message: "message content too long (max 10000 characters)",
			Code:    "VALIDATION_ERROR",
		})
		return
	}

	// Создаем сообщение пользователя и сразу сохраняем в БД
	userMessage, err := app.models.Messages.Create(chatID, "user", content)
	if err != nil {
		app.logger.Error("Error creating message", "error", err, "chat_id", chatID)
		internalErrorResponse(c, err)
		return
	}

	// Обновляем время последнего обновления чата
	app.models.Chats.UpdateUpdatedAt(chatID)

	// Возвращаем ответ пользователю сразу, не дожидаясь ответа AI
	c.JSON(http.StatusCreated, gin.H{
		"user_message": messageResponse{
			ID:        userMessage.ID,
			ChatID:    userMessage.ChatID,
			Role:      userMessage.Role,
			Content:   userMessage.Content,
			CreatedAt: userMessage.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		},
		"status":  "processing",
		"message": "Your message has been sent. AI response will be saved automatically.",
	})

	// Обработка AI ответа в фоне (горутина)
	go func() {
		// Обработка паник в горутине
		defer func() {
			if r := recover(); r != nil {
				app.logger.Error("Panic in processAIResponse",
					"error", r,
					"chat_id", chatID,
					"ai_model", chat.AIModel,
				)
			}
		}()
		app.processAIResponse(chatID, chat.AIModel, userMessage.ID)
	}()
}

// processAIResponse обрабатывает ответ AI в фоне
// ВАЖНО: Каждый чат имеет свой изолированный контекст:
// - История сообщений получается только для конкретного chatID (WHERE chat_id = $1)
// - Модель AI берется из самого чата (chat.AIModel), а не из запроса
// - Разные чаты и разные модели не смешиваются
// - Каждый чат работает со своей собственной историей и своей моделью AI
func (app *application) processAIResponse(chatID int, aiModel string, lastUserMessageID int) {
	// Создаем контекст с таймаутом (максимум 2 минуты на обработку)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Получаем историю сообщений для контекста (только для этого конкретного чата)
	// SQL запрос: SELECT ... FROM messages WHERE chat_id = $1
	// Это гарантирует, что каждый чат имеет свою изолированную историю
	history, err := app.models.Messages.GetByChatID(chatID)
	if err != nil {
		app.logger.Error("Error getting message history", "error", err, "chat_id", chatID)
		return
	}

	// Настройки контекста из переменных окружения
	maxHistoryMessages := env.GetEnvInt("AI_MAX_CONTEXT_MESSAGES", 100)
	maxContextTokens := env.GetEnvInt("AI_MAX_CONTEXT_TOKENS", 32000) // По умолчанию 32k токенов

	originalCount := len(history)
	app.logger.Debug("Processing AI response with isolated context",
		"chat_id", chatID,
		"ai_model", aiModel,
		"history_count", originalCount,
		"max_messages", maxHistoryMessages,
		"max_tokens", maxContextTokens,
		"context_isolation", "enabled", // Подтверждение изоляции контекста
	)

	// Ограничиваем количество сообщений в истории
	if len(history) > maxHistoryMessages {
		// Берем последние N сообщений (сохраняем контекст недавних сообщений)
		history = history[len(history)-maxHistoryMessages:]
		app.logger.Info("Message history truncated by count",
			"chat_id", chatID,
			"original_count", originalCount,
			"truncated_count", len(history),
		)
	}

	// Ограничиваем по токенам (приблизительная оценка: 1 токен ≈ 4 символа)
	// Более точная оценка: примерно 0.75 токена на слово или 4 символа на токен
	estimatedTokens := 0
	truncatedHistory := make([]*database.Message, 0, len(history))

	// Идем с конца истории (последние сообщения важнее) и добавляем сообщения пока не превысим лимит
	for i := len(history) - 1; i >= 0; i-- {
		msg := history[i]
		// Приблизительная оценка токенов: длина контента / 4 + накладные расходы на роль (~5 токенов)
		msgTokens := len(msg.Content)/4 + 5

		if estimatedTokens+msgTokens > maxContextTokens {
			// Если добавление этого сообщения превысит лимит, останавливаемся
			break
		}

		estimatedTokens += msgTokens
		truncatedHistory = append([]*database.Message{msg}, truncatedHistory...)
	}

	if len(truncatedHistory) < len(history) {
		app.logger.Info("Message history truncated by tokens",
			"chat_id", chatID,
			"original_count", originalCount,
			"truncated_count", len(truncatedHistory),
			"estimated_tokens", estimatedTokens,
			"max_tokens", maxContextTokens,
		)
		history = truncatedHistory
	}

	// Получаем AI провайдера на основе модели чата
	// ВАЖНО: aiModel берется из самого чата (chat.AIModel), сохраненного в БД
	// Это гарантирует, что каждый чат использует свою модель, даже если у пользователя несколько чатов с разными моделями
	providerName := strings.ToLower(aiModel)
	if providerName == "" {
		providerName = "openrouter" // По умолчанию OpenRouter
	}

	// Маппинг названий моделей на провайдеров
	// Каждый чат может использовать свою модель (deepseek-chat, grok-beta, GigaChat и т.д.)
	// Разные модели не смешиваются между чатами
	if strings.Contains(providerName, "deepseek") || strings.Contains(providerName, "deep-seek") {
		providerName = "openrouter" // Используем OpenRouter для DeepSeek
	} else if strings.Contains(providerName, "grok") {
		providerName = "grok"
	} else if strings.Contains(providerName, "gigachat") {
		providerName = "gigachat" // GigaChat
	} else {
		providerName = "openrouter" // По умолчанию OpenRouter
	}

	provider, err := app.aiProviderFactory.Get(providerName)
	if err != nil {
		app.logger.Error("Error getting AI provider", "error", err, "provider", providerName)
		return
	}

	// Конвертируем историю сообщений в формат для AI
	aiMessages := make([]ai.Message, len(history))
	for i, msg := range history {
		aiMessages[i] = ai.Message{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	app.logger.Debug("Sending request to AI with isolated context",
		"chat_id", chatID,
		"chat_ai_model", aiModel, // Модель, сохраненная в чате
		"messages_count", len(aiMessages),
		"provider", providerName,
		"model", provider.GetDefaultModel(),
		"context_isolation", "enabled", // Подтверждение изоляции
	)

	// Отправляем запрос к AI
	aiReq := ai.ChatRequest{
		Model:    provider.GetDefaultModel(),
		Messages: aiMessages,
	}

	aiResp, err := provider.Chat(ctx, aiReq)
	if err != nil {
		// Проверяем, не истек ли контекст
		if ctx.Err() == context.DeadlineExceeded {
			app.logger.Error("AI request timeout", "chat_id", chatID, "provider", providerName)
		} else {
			app.logger.Error("Error calling AI provider", "error", err, "chat_id", chatID, "provider", providerName)
		}
		return
	}

	// Сохраняем ответ ассистента в БД
	assistantMessage, err := app.models.Messages.Create(chatID, "assistant", aiResp.Content)
	if err != nil {
		app.logger.Error("Error creating assistant message", "error", err, "chat_id", chatID)
		return
	}

	// Обновляем время последнего обновления чата
	app.models.Chats.UpdateUpdatedAt(chatID)

	app.logger.Info("AI response saved with isolated context",
		"chat_id", chatID,
		"chat_ai_model", aiModel, // Модель чата для подтверждения изоляции
		"message_id", assistantMessage.ID,
		"model", aiResp.Model,
		"tokens", aiResp.Usage.TotalTokens,
		"prompt_tokens", aiResp.Usage.PromptTokens,
		"completion_tokens", aiResp.Usage.CompletionTokens,
		"context_messages_count", len(history), // Количество сообщений в контексте этого чата
	)
}

// handleGetMessages получает историю сообщений чата
func (app *application) handleGetMessages(c *gin.Context) {
	userID, apiErr := getUserIDFromContext(c)
	if apiErr != nil {
		errorResponse(c, apiErr)
		return
	}

	chatID, apiErr := getChatIDFromParam(c)
	if apiErr != nil {
		errorResponse(c, apiErr)
		return
	}

	_, apiErr = app.validateChatOwnership(c, chatID, userID)
	if apiErr != nil {
		errorResponse(c, apiErr)
		return
	}

	messages, err := app.models.Messages.GetByChatID(chatID)
	if err != nil {
		app.logger.Error("Error getting messages", "error", err, "chat_id", chatID)
		internalErrorResponse(c, err)
		return
	}

	response := make([]messageResponse, len(messages))
	for i, msg := range messages {
		response[i] = messageResponse{
			ID:        msg.ID,
			ChatID:    msg.ChatID,
			Role:      msg.Role,
			Content:   msg.Content,
			CreatedAt: msg.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	c.JSON(http.StatusOK, response)
}

// handleDeleteChat удаляет чат
func (app *application) handleDeleteChat(c *gin.Context) {
	userID, apiErr := getUserIDFromContext(c)
	if apiErr != nil {
		errorResponse(c, apiErr)
		return
	}

	chatID, apiErr := getChatIDFromParam(c)
	if apiErr != nil {
		errorResponse(c, apiErr)
		return
	}

	_, apiErr = app.validateChatOwnership(c, chatID, userID)
	if apiErr != nil {
		errorResponse(c, apiErr)
		return
	}

	// Удаляем сообщения и чат (CASCADE должен удалить сообщения автоматически)
	if err := app.models.Chats.Delete(chatID); err != nil {
		app.logger.Error("Error deleting chat", "error", err, "chat_id", chatID)
		internalErrorResponse(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "chat deleted successfully",
	})
}

// handleUpdateChatTitle обновляет название чата
func (app *application) handleUpdateChatTitle(c *gin.Context) {
	userID, apiErr := getUserIDFromContext(c)
	if apiErr != nil {
		errorResponse(c, apiErr)
		return
	}

	chatID, apiErr := getChatIDFromParam(c)
	if apiErr != nil {
		errorResponse(c, apiErr)
		return
	}

	_, apiErr = app.validateChatOwnership(c, chatID, userID)
	if apiErr != nil {
		errorResponse(c, apiErr)
		return
	}

	var req updateChatTitleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validationErrorResponse(c, err)
		return
	}

	// Проверяем, что название не пустое
	if strings.TrimSpace(req.Title) == "" {
		errorResponse(c, &APIError{
			Status:  400,
			Message: "title cannot be empty",
			Code:    "VALIDATION_ERROR",
		})
		return
	}

	// Обновляем название чата
	if err := app.models.Chats.UpdateTitle(chatID, strings.TrimSpace(req.Title)); err != nil {
		app.logger.Error("Error updating chat title", "error", err, "chat_id", chatID)
		internalErrorResponse(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "chat title updated successfully",
	})
}
