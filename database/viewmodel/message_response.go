package viewmodel

import (
	"time"

	"github.com/zeroyukiy/the-throne-api/database/model"
)

type MessageAndUser struct {
	Id        string     `db:"id" json:"id"`
	User      model.User `db:"user" json:"user"`
	Text      string     `db:"text" json:"text"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt time.Time  `db:"updated_at" json:"updated_at"`
}
