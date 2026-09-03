package repository

import (
	"context"
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/zeroyukiy/the-throne-api/database/entity"
)

type ChatRepository interface {
	SelectAll(context.Context) ([]*entity.Chat, error)
	GetBySlug(context.Context, string) (*entity.Chat, error)
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

func (r *chatRepository) GetBySlug(ctx context.Context, slug string) (*entity.Chat, error) {
	chat := &entity.Chat{}

	// query := `SELECT * FROM chats LEFT JOIN messages ON chats.id = messages.chat_id WHERE slug = $1`
	query := `SELECT * FROM chats WHERE slug = $1`
	err := r.conn.GetContext(ctx, chat, query, slug)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	query2 := `SELECT * FROM messages WHERE chat_id = $1`
	rows, err := r.conn.QueryxContext(ctx, query2, chat.Id)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	for rows.Next() {
		message := &entity.Message{}
		err = rows.StructScan(message)
		if err != nil {
			fmt.Println(err)
		}
		chat.Messages = append(chat.Messages, message)
	}

	return chat, nil
}
