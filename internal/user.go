package internal

type User struct {
	Username  string `json:"username"`
	Age       int    `json:"age"`
	CreatedAt int64  `json:"created_at"`
}
