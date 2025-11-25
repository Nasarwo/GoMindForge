package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (app *application) initializeRoutes() http.Handler {
	g := gin.Default()

	// Health check endpoint (не требует аутентификации)
	g.GET("/health", app.handleHealth)

	v1 := g.Group("/api/v1")
	{
		// Аутентификация
		v1.POST("/register", app.handleRegister)
		v1.POST("/login", app.handleLogin)
		v1.POST("/refresh", app.handleRefresh)
		v1.POST("/logout", app.handleLogout)
		v1.GET("/profile", app.jwtAuthMiddleware(), app.handleGetProfile)

		// Чаты (требуют аутентификации)
		chats := v1.Group("/chats", app.jwtAuthMiddleware())
		{
			chats.POST("", app.handleCreateChat)
			chats.GET("", app.handleGetChats)
			chats.GET("/:id", app.handleGetChat)
			chats.PUT("/:id/title", app.handleUpdateChatTitle)
			chats.DELETE("/:id", app.handleDeleteChat)
			chats.POST("/:id/messages", app.handleCreateMessage)
			chats.GET("/:id/messages", app.handleGetMessages)
		}
	}

	return g
}
