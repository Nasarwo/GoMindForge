package main

import (
	"context"
	"net/http"
	"strings"

	"mindforge/internal/ai"

	"github.com/gin-gonic/gin"
)

type createChatRequest struct {
	AIModel string `json:"ai_model" binding:"required"`
	Title   string `json:"title,omitempty"`
}

type chatResponse struct {
	ID        int    `json:"id"`
	UserID    int    `json:"user_id"`
	AIModel   string `json:"ai_model"`
	Title     string `json:"title"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type createMessageRequest struct {
	Content string `json:"content" binding:"required"`
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

	chat, err := app.models.Chats.Create(userID.(int), req.AIModel, req.Title)
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

	// Создаем сообщение пользователя и сразу сохраняем в БД
	userMessage, err := app.models.Messages.Create(chatID, "user", req.Content)
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
	go app.processAIResponse(chatID, chat.AIModel, userMessage.ID)
}

// processAIResponse обрабатывает ответ AI в фоне
func (app *application) processAIResponse(chatID int, aiModel string, lastUserMessageID int) {
	// Получаем историю сообщений для контекста
	history, err := app.models.Messages.GetByChatID(chatID)
	if err != nil {
		app.logger.Error("Error getting message history", "error", err, "chat_id", chatID)
		return
	}

	// Получаем AI провайдера на основе модели чата
	providerName := strings.ToLower(aiModel)
	if providerName == "" {
		providerName = "deepseek" // По умолчанию DeepSeek
	}

	// Маппинг названий моделей на провайдеров
	if strings.Contains(providerName, "deepseek") || strings.Contains(providerName, "deep-seek") {
		providerName = "deepseek"
	} else if strings.Contains(providerName, "grok") {
		providerName = "grok"
	} else {
		providerName = "deepseek" // По умолчанию
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

	// Отправляем запрос к AI
	aiReq := ai.ChatRequest{
		Model:    provider.GetDefaultModel(),
		Messages: aiMessages,
	}

	ctx := context.Background()
	aiResp, err := provider.Chat(ctx, aiReq)
	if err != nil {
		app.logger.Error("Error calling AI provider", "error", err, "chat_id", chatID, "provider", providerName)
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

	app.logger.Info("AI response saved",
		"chat_id", chatID,
		"message_id", assistantMessage.ID,
		"model", aiResp.Model,
		"tokens", aiResp.Usage.TotalTokens,
		"prompt_tokens", aiResp.Usage.PromptTokens,
		"completion_tokens", aiResp.Usage.CompletionTokens,
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
