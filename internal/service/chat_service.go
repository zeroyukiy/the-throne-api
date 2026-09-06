package service

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/zeroyukiy/the-throne-api/database/repository"
	"github.com/zeroyukiy/the-throne-api/database/viewmodel"
)

type ChatService struct {
	conn        *sqlx.DB
	chatRepo    repository.ChatRepository
	messageRepo repository.MessageRepository
}

func NewChatService(conn *sqlx.DB) *ChatService {
	chatRepo := repository.NewChatRepository(conn)
	messageRepo := repository.NewMessageRepository(conn)

	return &ChatService{
		conn:        conn,
		chatRepo:    chatRepo,
		messageRepo: messageRepo,
	}
}

func (s *ChatService) GetChatAndMessages(ctx context.Context, chat_slug string) (*viewmodel.ChatAndMessages, error) {
	result := &viewmodel.ChatAndMessages{}

	chat, err := s.chatRepo.GetBySlug(ctx, chat_slug)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	messages, err := s.messageRepo.SelectAllByChatId(ctx, chat.Id)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	result.ChatAndUser = *chat
	result.Messages = messages

	return result, nil
}
