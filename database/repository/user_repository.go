package repository

import (
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/zeroyukiy/the-throne-api/database/entity"
)

type UserRepository interface {
	Get(username string, password string) (*entity.User, error)
	GetCards(user_id int) ([]*entity.Card, error)
}

type userRepository struct {
	conn *sqlx.DB
}

func NewUserRepository(conn *sqlx.DB) UserRepository {
	return &userRepository{
		conn: conn,
	}
}

func (r *userRepository) Get(username string, password string) (*entity.User, error) {
	user := &entity.User{}

	query := `SELECT id, username, avatar, created_at FROM users WHERE username = ? AND password = ?`
	err := r.conn.Get(user, query, username, password)
	if err != nil {
		log.Println("select error: ", err)
		return nil, err
	}

	return user, nil
}

func (r *userRepository) GetCards(user_id int) ([]*entity.Card, error) {
	cards := []*entity.Card{}

	query := `SELECT c.id, c.name FROM card_user as cu
		JOIN cards as c on c.id = cu.card_id
		JOIN users as u on u.id = cu.user_id
		WHERE u.id = ?`
	rows, err := r.conn.Queryx(query, user_id)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		card := &entity.Card{}
		err := rows.StructScan(card)
		if err != nil {
			return nil, err
		}
		cards = append(cards, card)
	}

	return cards, nil
}
