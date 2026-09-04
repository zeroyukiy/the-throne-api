package repository

import (
	"context"
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/zeroyukiy/the-throne-api/database/entity"
)

type ChatRepository interface {
	SelectAll(context.Context) ([]*entity.Chat, error)
	GetBySlug(context.Context, string) (*entity.ChatPlusUserPlusMessages, error)
}

type chatRepository struct {
	conn *sqlx.DB
}

func NewChatRepository(conn *sqlx.DB) ChatRepository {
	return &chatRepository{
		conn: conn,
	}
}

func (r *chatRepository) SelectAll(ctx context.Context) ([]*entity.Chat, error) {
	chats := []*entity.Chat{}

	query := `SELECT * FROM chats LIMIT 10`
	if err := r.conn.SelectContext(ctx, &chats, query); err != nil {
		log.Println(err)
		return nil, err
	}
	return chats, nil
}

func (r *chatRepository) GetBySlug(ctx context.Context, slug string) (*entity.ChatPlusUserPlusMessages, error) {
	chat := &entity.ChatPlusUserPlusMessages{
		Messages: []*entity.CustomMessage{},
	}
	query := `SELECT
	c.id,
	c.name,
	c.slug,
	c.location,
	c.description,
	c.is_open,
	u.id as "user.id",
	u.username as "user.username",
	u.avatar as "user.avatar",
	u.created_at as "user.created_at",
	c.created_at,
	c.updated_at
	FROM chats as c
	LEFT JOIN users as u ON c.user_id = u.id
	WHERE slug = $1`
	err := r.conn.GetContext(ctx, chat, query, slug)
	if err != nil {
		log.Println(err)
		return nil, err
	}

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
	rows, err := r.conn.QueryxContext(ctx, query2, chat.Id)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	for rows.Next() {
		message := &entity.CustomMessage{}
		err = rows.StructScan(message)
		if err != nil {
			log.Fatal(err)
		}
		chat.Messages = append(chat.Messages, message)
	}

	return chat, nil
}
