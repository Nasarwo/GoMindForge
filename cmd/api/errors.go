package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// APIError представляет структурированную ошибку API
type APIError struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

// Error реализует интерфейс error
func (e *APIError) Error() string {
	return e.Message
}

// Predefined errors
var (
	ErrInvalidCredentials = &APIError{
		Status:  http.StatusUnauthorized,
		Message: "invalid credentials",
		Code:    "INVALID_CREDENTIALS",
	}
	ErrUserAlreadyExists = &APIError{
		Status:  http.StatusConflict,
		Message: "user with this email already exists",
		Code:    "USER_ALREADY_EXISTS",
	}
	ErrInvalidToken = &APIError{
		Status:  http.StatusUnauthorized,
		Message: "invalid token",
		Code:    "INVALID_TOKEN",
	}
	ErrUnauthorized = &APIError{
		Status:  http.StatusUnauthorized,
		Message: "unauthorized",
		Code:    "UNAUTHORIZED",
	}
)

// errorResponse отправляет структурированный ответ об ошибке
func errorResponse(c *gin.Context, err *APIError) {
	c.JSON(err.Status, err)
	c.Abort()
}

// validationErrorResponse отправляет ответ об ошибке валидации
func validationErrorResponse(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, gin.H{
		"status":  http.StatusBadRequest,
		"message": err.Error(),
		"code":    "VALIDATION_ERROR",
	})
	c.Abort()
}

// internalErrorResponse отправляет ответ о внутренней ошибке сервера
func internalErrorResponse(c *gin.Context, err error) {
	// В production не логируем детали ошибки клиенту
	c.JSON(http.StatusInternalServerError, gin.H{
		"status":  http.StatusInternalServerError,
		"message": "internal server error",
		"code":    "INTERNAL_ERROR",
	})
	c.Abort()
}
