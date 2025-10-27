package main

import (
	"strconv"

	"mindforge/internal/database"

	"github.com/gin-gonic/gin"
)

// getChatIDFromParam извлекает и валидирует chatID из параметров URL
func getChatIDFromParam(c *gin.Context) (int, *APIError) {
	idStr := c.Param("id")
	if idStr == "" {
		return 0, &APIError{
			Status:  400,
			Message: "chat id is required",
			Code:    "INVALID_CHAT_ID",
		}
	}

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		return 0, &APIError{
			Status:  400,
			Message: "invalid chat id",
			Code:    "INVALID_CHAT_ID",
		}
	}

	return id, nil
}

// getUserIDFromContext извлекает userID из контекста
func getUserIDFromContext(c *gin.Context) (int, *APIError) {
	userID, exists := c.Get("userID")
	if !exists {
		return 0, ErrUnauthorized
	}

	id, ok := userID.(int)
	if !ok || id <= 0 {
		return 0, &APIError{
			Status:  401,
			Message: "invalid user id",
			Code:    "INVALID_USER_ID",
		}
	}

	return id, nil
}

// validateChatOwnership проверяет принадлежность чата пользователю
func (app *application) validateChatOwnership(c *gin.Context, chatID, userID int) (*database.Chat, *APIError) {
	chat, err := app.models.Chats.GetByID(chatID)
	if err != nil {
		if err.Error() == "chat not found" {
			return nil, &APIError{
				Status:  404,
				Message: "chat not found",
				Code:    "CHAT_NOT_FOUND",
			}
		}
		return nil, &APIError{
			Status:  500,
			Message: "internal server error",
			Code:    "INTERNAL_ERROR",
		}
	}

	if chat.UserID != userID {
		return nil, &APIError{
			Status:  403,
			Message: "forbidden",
			Code:    "FORBIDDEN",
		}
	}

	return chat, nil
}
