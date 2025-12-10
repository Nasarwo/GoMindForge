package database

import (
	"database/sql"
	"errors"
	"time"
)

type MessageModel struct {
	DB *sql.DB
}

type Message struct {
	ID        int       `json:"id"`
	ChatID    int       `json:"chat_id"`
	Role      string    `json:"role"` // "user" или "assistant"
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// Create создает новое сообщение
func (m MessageModel) Create(chatID int, role, content string) (*Message, error) {
	// Валидация роли
	if role != "user" && role != "assistant" && role != "system" {
		return nil, errors.New("invalid role: must be 'user', 'assistant', or 'system'")
	}

	// Валидация контента
	if len(content) == 0 {
		return nil, errors.New("content cannot be empty")
	}
	if len(content) > 10000 {
		return nil, errors.New("content too long (max 10000 characters)")
	}

	query := `
		INSERT INTO messages (chat_id, role, content, created_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP)
		RETURNING id`

	var id int
	err := m.DB.QueryRow(query, chatID, role, content).Scan(&id)
	if err != nil {
		return nil, err
	}

	return m.GetByID(id)
}

// GetByID получает сообщение по ID
func (m MessageModel) GetByID(id int) (*Message, error) {
	query := `
		SELECT id, chat_id, role, content, created_at
		FROM messages
		WHERE id = $1`

	var msg Message
	var createdAt sql.NullTime
	err := m.DB.QueryRow(query, id).Scan(
		&msg.ID,
		&msg.ChatID,
		&msg.Role,
		&msg.Content,
		&createdAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("message not found")
		}
		return nil, err
	}

	if createdAt.Valid {
		msg.CreatedAt = createdAt.Time
	}

	return &msg, nil
}

// GetByChatID получает все сообщения чата
func (m MessageModel) GetByChatID(chatID int) ([]*Message, error) {
	query := `
		SELECT id, chat_id, role, content, created_at
		FROM messages
		WHERE chat_id = $1
		ORDER BY created_at ASC`

	rows, err := m.DB.Query(query, chatID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*Message
	for rows.Next() {
		var msg Message
		var createdAt sql.NullTime
		err := rows.Scan(
			&msg.ID,
			&msg.ChatID,
			&msg.Role,
			&msg.Content,
			&createdAt,
		)
		if err != nil {
			return nil, err
		}

		if createdAt.Valid {
			msg.CreatedAt = createdAt.Time
		}

		messages = append(messages, &msg)
	}

	return messages, rows.Err()
}

// DeleteByChatID удаляет все сообщения чата
func (m MessageModel) DeleteByChatID(chatID int) error {
	query := `DELETE FROM messages WHERE chat_id = $1`
	_, err := m.DB.Exec(query, chatID)
	return err
}
