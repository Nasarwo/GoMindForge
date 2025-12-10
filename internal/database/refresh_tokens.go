package database

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"time"
)

type RefreshTokenModel struct {
	DB *sql.DB
}

type RefreshToken struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// GenerateSafeToken генерирует криптографически безопасный токен
func GenerateSafeToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// Create создает новый refresh token
func (m RefreshTokenModel) Create(userID int, token string, expiresAt time.Time) error {
	query := `
		INSERT INTO refresh_tokens (user_id, token, expires_at, created_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP)`

	_, err := m.DB.Exec(query, userID, token, expiresAt)
	return err
}

// GetByToken получает refresh token по токену
func (m RefreshTokenModel) GetByToken(token string) (*RefreshToken, error) {
	query := `
		SELECT id, user_id, token, expires_at, created_at
		FROM refresh_tokens
		WHERE token = $1`

	var rt RefreshToken
	var expiresAt, createdAt sql.NullTime
	err := m.DB.QueryRow(query, token).Scan(
		&rt.ID,
		&rt.UserID,
		&rt.Token,
		&expiresAt,
		&createdAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("refresh token not found")
		}
		return nil, err
	}

	// PostgreSQL возвращает time.Time напрямую
	if expiresAt.Valid {
		rt.ExpiresAt = expiresAt.Time
	}
	if createdAt.Valid {
		rt.CreatedAt = createdAt.Time
	}

	return &rt, nil
}

// Delete удаляет refresh token
func (m RefreshTokenModel) Delete(token string) error {
	query := `DELETE FROM refresh_tokens WHERE token = $1`
	_, err := m.DB.Exec(query, token)
	return err
}

// DeleteByUserID удаляет все refresh tokens пользователя
func (m RefreshTokenModel) DeleteByUserID(userID int) error {
	query := `DELETE FROM refresh_tokens WHERE user_id = $1`
	_, err := m.DB.Exec(query, userID)
	return err
}

// DeleteExpired удаляет истекшие токены
func (m RefreshTokenModel) DeleteExpired() error {
	query := `DELETE FROM refresh_tokens WHERE expires_at < CURRENT_TIMESTAMP`
	_, err := m.DB.Exec(query)
	return err
}

// IsValid проверяет, не истек ли токен
func (rt *RefreshToken) IsValid() bool {
	return time.Now().Before(rt.ExpiresAt)
}
