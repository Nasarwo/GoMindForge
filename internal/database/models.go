package database

import "database/sql"

type Models struct {
	Users         UserModel
	RefreshTokens RefreshTokenModel
	Chats         ChatModel
	Messages      MessageModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Users:         UserModel{DB: db},
		RefreshTokens: RefreshTokenModel{DB: db},
		Chats:         ChatModel{DB: db},
		Messages:      MessageModel{DB: db},
	}
}
