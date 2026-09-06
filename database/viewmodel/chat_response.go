package viewmodel

import (
	"time"

	"github.com/zeroyukiy/the-throne-api/database/model"
)

type ChatAndUser struct {
	Id          string     `db:"id" json:"id"`
	Name        string     `db:"name" json:"name"`
	Slug        string     `db:"slug" json:"slug"`
	Location    string     `db:"location" json:"location"`
	Description string     `db:"description" json:"description"`
	Open        string     `db:"is_open" json:"is_open"`
	User        model.User `db:"user" json:"user"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at" json:"updated_at"`
}

type ChatAndMessages struct {
	ChatAndUser
	Messages []*MessageAndUser `json:"messages"`
}
