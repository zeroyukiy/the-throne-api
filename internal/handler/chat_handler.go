package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/zeroyukiy/the-throne-api/database/repository"
)

type ChatHandler struct {
	chatRepo repository.ChatRepository
}

func NewChatHandler(conn *sqlx.DB) *ChatHandler {
	repo := repository.NewChatRepository(conn)

	return &ChatHandler{
		chatRepo: repo,
	}
}

func (h *ChatHandler) Index(c *gin.Context) {
	chatList, err := h.chatRepo.SelectAll(c)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": fmt.Sprint("not found"),
		})
		return
	}
	c.JSON(http.StatusOK, chatList)
}

func (h *ChatHandler) Show(c *gin.Context) {
	slug := c.Param("slug")
	chat, err := h.chatRepo.GetBySlug(c, slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": fmt.Sprint("not found"),
		})
		return
	}
	c.JSON(http.StatusOK, chat)
}
