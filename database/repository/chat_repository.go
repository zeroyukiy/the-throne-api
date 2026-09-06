package repository

import (
	"context"
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/zeroyukiy/the-throne-api/database/model"
	"github.com/zeroyukiy/the-throne-api/database/viewmodel"
)

type ChatRepository interface {
	SelectAll(context.Context) ([]*model.Chat, error)
	GetBySlug(context.Context, string) (*viewmodel.ChatAndUser, error)
}

type chatRepository struct {
	conn *sqlx.DB
}

func NewChatRepository(conn *sqlx.DB) ChatRepository {
	return &chatRepository{
		conn: conn,
	}
}

func (r *chatRepository) SelectAll(ctx context.Context) ([]*model.Chat, error) {
	chats := []*model.Chat{}

	query := `SELECT * FROM chats LIMIT 10`
	if err := r.conn.SelectContext(ctx, &chats, query); err != nil {
		log.Println(err)
		return nil, err
	}
	return chats, nil
}

func (r *chatRepository) GetBySlug(ctx context.Context, slug string) (*viewmodel.ChatAndUser, error) {
	chat := &viewmodel.ChatAndUser{}
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
	row := r.conn.QueryRowContext(ctx, query, slug)
	// err := r.conn.GetContext(ctx, chat, query, slug)
	if row.Err() != nil {
		log.Println(row.Err())
		return nil, row.Err()
	}

	err := row.Scan(&chat.Id, &chat.Name, &chat.Slug, &chat.Location, &chat.Description, &chat.Open, &chat.User.Id, &chat.User.Username, &chat.User.Avatar, &chat.User.CreatedAt, &chat.CreatedAt, &chat.UpdatedAt)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return chat, nil
}
