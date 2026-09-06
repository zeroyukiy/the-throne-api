package repository

import (
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/zeroyukiy/the-throne-api/database/model"
)

type UserRepository interface {
	Get(username string, password string) (*model.User, error)
	GetCards(user_id int) ([]*model.Card, error)
	FindOne(username string) (*model.User, error)
}

type userRepository struct {
	conn *sqlx.DB
}

func NewUserRepository(conn *sqlx.DB) UserRepository {
	return &userRepository{
		conn: conn,
	}
}

func (r *userRepository) Get(username string, password string) (*model.User, error) {
	user := &model.User{}

	query := `SELECT id, username, avatar, created_at FROM users WHERE username = $1 AND password = $2`
	err := r.conn.Get(user, query, username, password)
	if err != nil {
		log.Println("select error: ", err)
		return nil, err
	}

	return user, nil
}

func (r *userRepository) FindOne(id string) (*model.User, error) {
	user := &model.User{}

	query := `SELECT id, username, avatar, created_at FROM users WHERE id = $1`
	err := r.conn.Get(user, query, id)
	if err != nil {
		log.Println("select error: ", err)
		return nil, err
	}

	return user, nil
}

func (r *userRepository) GetCards(user_id int) ([]*model.Card, error) {
	cards := []*model.Card{}

	query := `SELECT c.id, c.name FROM card_user as cu
		JOIN cards as c on c.id = cu.card_id
		JOIN users as u on u.id = cu.user_id
		WHERE u.id = $1`
	rows, err := r.conn.Queryx(query, user_id)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		card := &model.Card{}
		err := rows.StructScan(card)
		if err != nil {
			return nil, err
		}
		cards = append(cards, card)
	}

	return cards, nil
}
