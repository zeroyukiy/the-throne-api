package entity

import "time"

type Message struct {
	Id        string    `db:"id" json:"id"`
	UserId    string    `db:"user_id" json:"user_id"`
	Text      string    `db:"text" json:"text"`
	ChatId    string    `db:"chat_id" json:"chat_id"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}
