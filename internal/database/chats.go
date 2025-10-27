package database

import (
	"database/sql"
	"errors"
	"time"
)

type ChatModel struct {
	DB *sql.DB
}

type Chat struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	AIModel   string    `json:"ai_model"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Create создает новый чат
func (m ChatModel) Create(userID int, aiModel, title string) (*Chat, error) {
	query := `
		INSERT INTO chats (user_id, ai_model, title, created_at, updated_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`

	result, err := m.DB.Exec(query, userID, aiModel, title)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return m.GetByID(int(id))
}

// GetByID получает чат по ID
func (m ChatModel) GetByID(id int) (*Chat, error) {
	query := `
		SELECT id, user_id, ai_model, title, created_at, updated_at
		FROM chats
		WHERE id = ?`

	var chat Chat
	var createdAtStr, updatedAtStr sql.NullString
	err := m.DB.QueryRow(query, id).Scan(
		&chat.ID,
		&chat.UserID,
		&chat.AIModel,
		&chat.Title,
		&createdAtStr,
		&updatedAtStr,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("chat not found")
		}
		return nil, err
	}

	if createdAtStr.Valid {
		chat.CreatedAt, err = scanTime(createdAtStr.String)
		if err != nil {
			return nil, err
		}
	}
	if updatedAtStr.Valid {
		chat.UpdatedAt, err = scanTime(updatedAtStr.String)
		if err != nil {
			return nil, err
		}
	}

	return &chat, nil
}

// GetByUserID получает все чаты пользователя
func (m ChatModel) GetByUserID(userID int) ([]*Chat, error) {
	query := `
		SELECT id, user_id, ai_model, title, created_at, updated_at
		FROM chats
		WHERE user_id = ?
		ORDER BY updated_at DESC`

	rows, err := m.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chats []*Chat
	for rows.Next() {
		var chat Chat
		var createdAtStr, updatedAtStr sql.NullString
		err := rows.Scan(
			&chat.ID,
			&chat.UserID,
			&chat.AIModel,
			&chat.Title,
			&createdAtStr,
			&updatedAtStr,
		)
		if err != nil {
			return nil, err
		}

		if createdAtStr.Valid {
			chat.CreatedAt, err = scanTime(createdAtStr.String)
			if err != nil {
				return nil, err
			}
		}
		if updatedAtStr.Valid {
			chat.UpdatedAt, err = scanTime(updatedAtStr.String)
			if err != nil {
				return nil, err
			}
		}

		chats = append(chats, &chat)
	}

	return chats, rows.Err()
}

// UpdateTitle обновляет заголовок чата
func (m ChatModel) UpdateTitle(chatID int, title string) error {
	query := `
		UPDATE chats
		SET title = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`

	_, err := m.DB.Exec(query, title, chatID)
	return err
}

// Delete удаляет чат
func (m ChatModel) Delete(chatID int) error {
	query := `DELETE FROM chats WHERE id = ?`
	_, err := m.DB.Exec(query, chatID)
	return err
}

// UpdateUpdatedAt обновляет время последнего обновления чата
func (m ChatModel) UpdateUpdatedAt(chatID int) error {
	query := `UPDATE chats SET updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := m.DB.Exec(query, chatID)
	return err
}
