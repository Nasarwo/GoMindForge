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
	query := `
		INSERT INTO messages (chat_id, role, content, created_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP)`

	result, err := m.DB.Exec(query, chatID, role, content)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return m.GetByID(int(id))
}

// GetByID получает сообщение по ID
func (m MessageModel) GetByID(id int) (*Message, error) {
	query := `
		SELECT id, chat_id, role, content, created_at
		FROM messages
		WHERE id = ?`

	var msg Message
	var createdAtStr sql.NullString
	err := m.DB.QueryRow(query, id).Scan(
		&msg.ID,
		&msg.ChatID,
		&msg.Role,
		&msg.Content,
		&createdAtStr,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("message not found")
		}
		return nil, err
	}

	if createdAtStr.Valid {
		msg.CreatedAt, err = scanTime(createdAtStr.String)
		if err != nil {
			return nil, err
		}
	}

	return &msg, nil
}

// GetByChatID получает все сообщения чата
func (m MessageModel) GetByChatID(chatID int) ([]*Message, error) {
	query := `
		SELECT id, chat_id, role, content, created_at
		FROM messages
		WHERE chat_id = ?
		ORDER BY created_at ASC`

	rows, err := m.DB.Query(query, chatID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*Message
	for rows.Next() {
		var msg Message
		var createdAtStr sql.NullString
		err := rows.Scan(
			&msg.ID,
			&msg.ChatID,
			&msg.Role,
			&msg.Content,
			&createdAtStr,
		)
		if err != nil {
			return nil, err
		}

		if createdAtStr.Valid {
			msg.CreatedAt, err = scanTime(createdAtStr.String)
			if err != nil {
				return nil, err
			}
		}

		messages = append(messages, &msg)
	}

	return messages, rows.Err()
}

// DeleteByChatID удаляет все сообщения чата
func (m MessageModel) DeleteByChatID(chatID int) error {
	query := `DELETE FROM messages WHERE chat_id = ?`
	_, err := m.DB.Exec(query, chatID)
	return err
}
