package main

import (
	"net/http"
	"strings"
	"time"

	"mindforge/internal/database"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type registerRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type userResponse struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type authResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token,omitempty"`
	User         userResponse `json:"user"`
}

type refreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type refreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (app *application) handleRegister(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validationErrorResponse(c, err)
		return
	}

	// Нормализуем email (приводим к нижнему регистру)
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.Username = strings.TrimSpace(req.Username)

	// Проверяем, существует ли пользователь с таким email
	_, err := app.models.Users.GetByEmail(req.Email)
	if err == nil {
		errorResponse(c, ErrUserAlreadyExists)
		return
	}

	// Проверяем, существует ли пользователь с таким username
	_, err = app.models.Users.GetByUsername(req.Username)
	if err == nil {
		errorResponse(c, &APIError{
			Status:  http.StatusConflict,
			Message: "user with this username already exists",
			Code:    "USERNAME_ALREADY_EXISTS",
		})
		return
	}

	// Хешируем пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		app.logger.Error("Error hashing password", "error", err)
		internalErrorResponse(c, err)
		return
	}

	// Создаем пользователя
	user, err := app.models.Users.Create(req.Username, req.Email, string(hashedPassword))
	if err != nil {
		app.logger.Error("Error creating user", "error", err, "email", req.Email)
		internalErrorResponse(c, err)
		return
	}

	// Генерируем access и refresh токены
	accessToken, refreshToken, err := app.generateTokenPair(user.ID)
	if err != nil {
		app.logger.Error("Error generating tokens", "error", err, "user_id", user.ID)
		internalErrorResponse(c, err)
		return
	}

	// Устанавливаем refresh token в cookie
	app.setRefreshTokenCookie(c, refreshToken)

	c.JSON(http.StatusCreated, authResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken, // Также возвращаем в JSON для мобильных приложений
		User: userResponse{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
	})
}

func (app *application) handleLogin(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validationErrorResponse(c, err)
		return
	}

	// Нормализуем email
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	// Находим пользователя по email
	user, err := app.models.Users.GetByEmail(req.Email)
	if err != nil {
		// Для безопасности всегда возвращаем одинаковое сообщение
		errorResponse(c, ErrInvalidCredentials)
		return
	}

	// Проверяем пароль
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		errorResponse(c, ErrInvalidCredentials)
		return
	}

	// Генерируем access и refresh токены
	accessToken, refreshToken, err := app.generateTokenPair(user.ID)
	if err != nil {
		app.logger.Error("Error generating tokens", "error", err, "user_id", user.ID)
		internalErrorResponse(c, err)
		return
	}

	// Устанавливаем refresh token в cookie
	app.setRefreshTokenCookie(c, refreshToken)

	c.JSON(http.StatusOK, authResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken, // Также возвращаем в JSON для мобильных приложений
		User: userResponse{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
	})
}

func (app *application) handleGetProfile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		errorResponse(c, ErrUnauthorized)
		return
	}

	user, err := app.models.Users.GetByID(userID.(int))
	if err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  http.StatusNotFound,
				"message": "user not found",
				"code":    "USER_NOT_FOUND",
			})
			return
		}
		app.logger.Error("Error getting user", "error", err, "user_id", userID)
		internalErrorResponse(c, err)
		return
	}

	c.JSON(http.StatusOK, userResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	})
}

// generateTokenPair создает пару access и refresh токенов
func (app *application) generateTokenPair(userID int) (accessToken string, refreshToken string, err error) {
	// Генерируем access token (короткоживущий, 15 минут)
	accessToken, err = app.generateAccessToken(userID)
	if err != nil {
		return "", "", err
	}

	// Генерируем refresh token (долгоживущий, 7 дней)
	refreshToken, err = database.GenerateSafeToken()
	if err != nil {
		return "", "", err
	}

	// Сохраняем refresh token в БД
	expiresAt := time.Now().Add(7 * 24 * time.Hour) // 7 дней
	err = app.models.RefreshTokens.Create(userID, refreshToken, expiresAt)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// generateAccessToken создает JWT access token
func (app *application) generateAccessToken(userID int) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"user_id": userID,
		"iat":     now.Unix(),                       // issued at
		"exp":     now.Add(15 * time.Minute).Unix(), // expiration time (15 минут)
		"nbf":     now.Unix(),                       // not before
		"type":    "access",                         // тип токена
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(app.jwtSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// setRefreshTokenCookie устанавливает refresh token в HTTP-only cookie
func (app *application) setRefreshTokenCookie(c *gin.Context, token string) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "refresh_token",
		Value:    token,
		Path:     "/api/v1/refresh",
		MaxAge:   7 * 24 * 60 * 60, // 7 дней
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})
}

// handleRefresh обновляет access token используя refresh token
func (app *application) handleRefresh(c *gin.Context) {
	// Пытаемся получить refresh token из cookie или из тела запроса
	var refreshToken string

	// Сначала проверяем cookie
	cookieToken, err := c.Cookie("refresh_token")
	if err == nil && cookieToken != "" {
		refreshToken = cookieToken
	} else {
		// Если нет в cookie, пытаемся получить из JSON body
		var req refreshTokenRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			errorResponse(c, &APIError{
				Status:  401,
				Message: "refresh token required",
				Code:    "REFRESH_TOKEN_REQUIRED",
			})
			return
		}
		refreshToken = req.RefreshToken
	}

	if refreshToken == "" {
		errorResponse(c, &APIError{
			Status:  401,
			Message: "refresh token required",
			Code:    "REFRESH_TOKEN_REQUIRED",
		})
		return
	}

	// Проверяем refresh token в БД
	rt, err := app.models.RefreshTokens.GetByToken(refreshToken)
	if err != nil {
		errorResponse(c, ErrInvalidToken)
		return
	}

	// Проверяем, не истек ли токен
	if !rt.IsValid() {
		// Удаляем истекший токен
		app.models.RefreshTokens.Delete(refreshToken)
		errorResponse(c, ErrInvalidToken)
		return
	}

	// Генерируем новую пару токенов
	accessToken, newRefreshToken, err := app.generateTokenPair(rt.UserID)
	if err != nil {
		app.logger.Error("Error generating tokens", "error", err, "user_id", rt.UserID)
		internalErrorResponse(c, err)
		return
	}

	// Удаляем старый refresh token (rotation)
	app.models.RefreshTokens.Delete(refreshToken)

	// Устанавливаем новый refresh token в cookie
	app.setRefreshTokenCookie(c, newRefreshToken)

	c.JSON(http.StatusOK, refreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	})
}

// handleLogout выходит из системы и удаляет refresh token
func (app *application) handleLogout(c *gin.Context) {
	// Получаем refresh token из cookie
	refreshToken, err := c.Cookie("refresh_token")
	if err == nil && refreshToken != "" {
		// Удаляем refresh token из БД
		app.models.RefreshTokens.Delete(refreshToken)
	}

	// Удаляем cookie
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/api/v1/refresh",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "logged out successfully",
	})
}

// handleHealth обрабатывает health check запросы
func (app *application) handleHealth(c *gin.Context) {
	// Проверяем подключение к базе данных
	if err := app.db.Ping(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "unhealthy",
			"message": "database connection failed",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"message": "service is running",
	})
}
