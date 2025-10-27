package database

import (
	"database/sql"
	"errors"
	"time"
)

// scanTime парсит SQLite timestamp из строки в time.Time
func scanTime(value interface{}) (time.Time, error) {
	switch v := value.(type) {
	case time.Time:
		return v, nil
	case string:
		// SQLite может возвращать timestamp в разных форматах
		formats := []string{
			time.RFC3339,
			time.RFC3339Nano,
			"2006-01-02 15:04:05-07:00",
			"2006-01-02 15:04:05",
			"2006-01-02T15:04:05Z07:00",
			"2006-01-02T15:04:05",
		}
		for _, format := range formats {
			if t, err := time.Parse(format, v); err == nil {
				return t, nil
			}
		}
		return time.Time{}, errors.New("unable to parse time: " + v)
	case []byte:
		return scanTime(string(v))
	case nil:
		return time.Time{}, nil
	default:
		return time.Time{}, errors.New("unsupported time type")
	}
}

type UserModel struct {
	DB *sql.DB
}

type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (m UserModel) Create(username, email, passwordHash string) (*User, error) {
	query := `
		INSERT INTO users (username, email, password_hash, created_at, updated_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`

	result, err := m.DB.Exec(query, username, email, passwordHash)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	// Получаем созданного пользователя
	return m.GetByID(int(id))
}

func (m UserModel) GetByEmail(email string) (*User, error) {
	query := `
		SELECT id, username, email, password_hash, created_at, updated_at
		FROM users
		WHERE email = ?`

	var user User
	var createdAtStr, updatedAtStr sql.NullString
	err := m.DB.QueryRow(query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&createdAtStr,
		&updatedAtStr,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	// Парсим timestamps
	if createdAtStr.Valid {
		user.CreatedAt, err = scanTime(createdAtStr.String)
		if err != nil {
			return nil, err
		}
	}
	if updatedAtStr.Valid {
		user.UpdatedAt, err = scanTime(updatedAtStr.String)
		if err != nil {
			return nil, err
		}
	}

	return &user, nil
}

func (m UserModel) GetByID(id int) (*User, error) {
	query := `
		SELECT id, username, email, password_hash, created_at, updated_at
		FROM users
		WHERE id = ?`

	var user User
	var createdAtStr, updatedAtStr sql.NullString
	err := m.DB.QueryRow(query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&createdAtStr,
		&updatedAtStr,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	// Парсим timestamps
	if createdAtStr.Valid {
		user.CreatedAt, err = scanTime(createdAtStr.String)
		if err != nil {
			return nil, err
		}
	}
	if updatedAtStr.Valid {
		user.UpdatedAt, err = scanTime(updatedAtStr.String)
		if err != nil {
			return nil, err
		}
	}

	return &user, nil
}
