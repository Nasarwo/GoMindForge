package main

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func (app *application) jwtAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			errorResponse(c, &APIError{
				Status:  401,
				Message: "authorization header required",
				Code:    "AUTHORIZATION_HEADER_REQUIRED",
			})
			return
		}

		// Проверяем формат "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			errorResponse(c, &APIError{
				Status:  401,
				Message: "invalid authorization header format",
				Code:    "INVALID_AUTH_HEADER",
			})
			return
		}

		tokenString := strings.TrimSpace(parts[1])
		if tokenString == "" {
			errorResponse(c, ErrInvalidToken)
			return
		}

		// Парсим и проверяем токен
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Проверяем метод подписи
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(app.jwtSecret), nil
		})

		if err != nil {
			errorResponse(c, ErrInvalidToken)
			return
		}

		if !token.Valid {
			errorResponse(c, ErrInvalidToken)
			return
		}

		// Извлекаем userID из claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			errorResponse(c, &APIError{
				Status:  401,
				Message: "invalid token claims",
				Code:    "INVALID_TOKEN_CLAIMS",
			})
			return
		}

		userID, ok := claims["user_id"].(float64)
		if !ok {
			errorResponse(c, &APIError{
				Status:  401,
				Message: "invalid user id in token",
				Code:    "INVALID_USER_ID",
			})
			return
		}

		// Сохраняем userID в контексте
		c.Set("userID", int(userID))
		c.Next()
	}
}
