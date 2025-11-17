package repository

import (
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/zeroyukiy/the-throne-api/database/entity"
)

type UserRepository interface {
	Get(username string, password string) *entity.User
}

type userRepository struct {
	conn *sqlx.DB
}

func NewUserRepository(conn *sqlx.DB) UserRepository {
	return &userRepository{
		conn: conn,
	}
}

func (r *userRepository) Get(username string, password string) *entity.User {
	user := &entity.User{}

	query := `SELECT id, username, avatar, created_at FROM users WHERE username = ? && password = ?`
	err := sqlx.Get(r.conn, user, query, username, password)
	if err != nil {
		log.Println("select error: ", err)
	}

	return user
}
