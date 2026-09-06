package repository

import (
	"context"
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/zeroyukiy/the-throne-api/database/viewmodel"
)

type MessageRepository interface {
	SelectAllByChatId(ctx context.Context, chat_id string) ([]*viewmodel.MessageAndUser, error)
}

type messageRepository struct {
	conn *sqlx.DB
}

func NewMessageRepository(conn *sqlx.DB) MessageRepository {
	return &messageRepository{
		conn: conn,
	}
}

func (r *messageRepository) SelectAllByChatId(ctx context.Context, chat_id string) ([]*viewmodel.MessageAndUser, error) {
	messages := []*viewmodel.MessageAndUser{}
	query2 := `SELECT
	m.id,
	m.text,
	m.created_at,
	m.updated_at,
	u.id as "user.id",
	u.username as "user.username",
	u.avatar as "user.avatar",
	u.created_at as "user.created_at"
	FROM messages as m
	LEFT JOIN users as u ON m.user_id = u.id
	WHERE chat_id = $1`
	rows, err := r.conn.QueryxContext(ctx, query2, chat_id)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	for rows.Next() {
		message := &viewmodel.MessageAndUser{}
		err = rows.StructScan(message)
		if err != nil {
			log.Fatal(err)
		}
		messages = append(messages, message)
	}

	return messages, nil
}
