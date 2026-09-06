package internal

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/zeroyukiy/the-throne-api/database/model"
	"github.com/zeroyukiy/the-throne-api/internal/event"
)

var (
	// pongTime = 60 * time.Second
	pongTime = 20 * time.Second
	pingTime = 9 * pongTime / 10

	writeTime = 5 * time.Second
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// if r.Header.Get("Origin") == "http://localhost:5173" {
		return true
		// }
		// return false
	},
}

type Client struct {
	conn     *websocket.Conn
	send     chan []byte
	hub      *Hub
	latency  time.Time
	userId   string
	username string
	avatar   string
	roomId   string
}

func NewClient(conn *websocket.Conn, hub *Hub, user_id string, username string, avatar string) *Client {
	return &Client{
		conn:     conn,
		hub:      hub,
		send:     make(chan []byte, 256),
		userId:   user_id,
		username: username,
		avatar:   avatar,
		roomId:   "",
	}
}

func (c *Client) Run() {
	go c.fromWebsocketToHub()
	go c.fromHubToWebsocket()
}

func (c *Client) GetUserId() string {
	return c.userId
}

func (c *Client) fromWebsocketToHub() {
	defer func() {
		fmt.Println("unregistering client..")
		c.hub.unregisterClient <- c
	}()

	c.conn.SetReadDeadline(time.Now().Add(pongTime))
	c.conn.SetPongHandler(func(appData string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongTime)) // receiving a pong message from the client is resetting the read dead line
		return nil
	})

	for {
		var data *event.WebsocketMessage
		// _, msg, err := c.conn.ReadMessage()
		err := c.conn.ReadJSON(&data)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v\n", err)
			}
			break
		}

		switch data.Event {
		case event.Message:
			if c.roomId != "" {
				response := &event.WebsocketMessage{
					Event: event.Message,
					Payload: event.Payload{
						UserId:   c.userId,
						Username: c.username,
						Avatar:   c.avatar,
						RoomId:   c.roomId,
						Message:  data.Payload.Message,
					},
					Timestamp: time.Now().UTC().Format("15:04:22"),
				}
				b, err := json.Marshal(response)
				if err != nil {
					fmt.Println(err)
					return
				}
				fmt.Println(response)
				room, ok := c.hub.roomRepository.GetRoom(c.roomId)
				if ok {
					room.broadcast <- b
				}
			}
		case event.JoinRoom:
			room, ok := c.hub.roomRepository.GetRoom(data.Payload.RoomId)
			if ok {
				room.join <- c
			} else {
				if room = c.hub.roomRepository.CreateRoom(data.Payload.RoomId); room != nil {
					fmt.Printf("room %s created\n", room.id)
					room.join <- c
				}
			}

		case event.LeaveRoom:
			room, ok := c.hub.roomRepository.GetRoom(c.roomId)
			if ok {
				room.leave <- c
			}

		default:
			log.Println("event not recognized")
		}
	}
}

func (c *Client) fromHubToWebsocket() {
	ticker := time.NewTicker(pingTime)
	defer func() {
		ticker.Stop()
	}()
	for {
		select {
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeTime))
			if !ok {
				// c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				fmt.Println(err)
				return
			}
			w.Write(msg)

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeTime)) // without setting a write deadline the server will not wait for the pong message from the client
			err := c.conn.WriteMessage(websocket.PingMessage, nil)
			if err != nil {
				return
			}
		}
	}
}

// func LatencyStatus(t time.Time) LatencyQuality {
// 	l := time.Since(t).Milliseconds()
// 	if l <= 30 {
// 		return Optimal
// 	} else if l > 30 && l <= 100 {
// 		return Good
// 	} else if l > 100 && l <= 150 {
// 		return Normal
// 	} else if l > 150 && l <= 250 {
// 		return Bad
// 	} else {
// 		return Awful
// 	}
// }

func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request, user *model.User) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	// uid := uuid.New()
	// client := NewClient(conn, hub, uid.String(), user.Username, user.Avatar)
	client := NewClient(conn, hub, user.Id, user.Username, user.Avatar)
	client.hub.registerClient <- client

	client.Run()
}
