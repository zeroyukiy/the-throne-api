package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/zeroyukiy/the-throne-api/internal"
)

type ChatHandler struct {
	hub *internal.Hub
}

func NewChatHandler(hub *internal.Hub) *ChatHandler {
	return &ChatHandler{
		hub: hub,
	}
}

func (h *ChatHandler) GetRoom(c echo.Context) error {
	// OK
	v := struct {
		Status int
	}{
		Status: http.StatusOK,
	}
	return c.JSON(http.StatusOK, v)
}
