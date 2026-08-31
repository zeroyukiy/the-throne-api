package event

type WebsocketEvent int64

const (
	Message WebsocketEvent = iota + 1
	Typing
	JoinRoom
	LeaveRoom
)

type WebsocketMessage struct {
	Event     WebsocketEvent `json:"event"`
	Payload   Payload        `json:"payload"`
	Timestamp string         `json:"timestamp"`
}

type Payload struct {
	UserId   string `json:"user_id,omitempty"`
	Username string `json:"username,omitempty"`
	Avatar   string `json:"avatar,omitempty"`
	RoomId   string `json:"room_id,omitempty"`
	Message  string `json:"message,omitempty"`
}
