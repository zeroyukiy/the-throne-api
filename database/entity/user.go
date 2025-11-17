package entity

import "time"

type User struct {
	Id        string    `db:"id" json:"id"`
	Username  string    `db:"username" json:"username"`
	Password  string    `db:"password" json:"password"`
	Avatar    string    `db:"avatar" json:"avatar"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}
