package internal

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/zeroyukiy/the-throne-api/internal/event"
)

type Status int64

const (
	Close Status = iota
	Open
)

type Room struct {
	id             string
	clients        map[string]*Client
	broadcast      chan []byte
	join           chan *Client
	leave          chan *Client
	roomRepository RoomRepository
	status         Status
}

func NewRoom(id string, roomRepo RoomRepository) *Room {
	return &Room{
		id:             id,
		clients:        make(map[string]*Client),
		broadcast:      make(chan []byte),
		join:           make(chan *Client),
		leave:          make(chan *Client),
		roomRepository: roomRepo,
		status:         Open,
	}
}

func (r *Room) Run() {
	go func() {
		for message := range r.broadcast {
			for _, client := range r.clients {
				client.send <- message
			}
		}
	}()

	for {
		select {
		case client := <-r.join:
			r.clients[client.userId] = client
			client.roomId = r.id
			fmt.Printf("%s joined the room %s\n", client.username, r.id)

		case client := <-r.leave:
			delete(r.clients, client.userId)
			client.roomId = ""
			fmt.Printf("%s left the room %s\n", client.username, r.id)

			fmt.Println(len(r.clients))

			if len(r.clients) > 0 {
				// DataResponse.
				res := event.WebsocketMessage{
					Event: event.Message,
					Payload: event.Payload{
						UserId:   client.userId,
						Username: client.username,
						Avatar:   client.avatar,
						RoomId:   r.id,
						Message:  fmt.Sprintf("%s left the room", client.username),
					},
					Timestamp: time.Now().UTC().Format("15:23:31"),
				}
				b, err := json.Marshal(res)
				if err != nil {
					fmt.Println(err)
				}
				r.broadcast <- b
			} else {
				ok := r.roomRepository.RemoveRoom(r.id)
				if ok {
					fmt.Printf("room %s removed\n", r.id)
				}
			}
		}
	}
}

func (r *Room) GetRoomId() string {
	return r.id
}

func (r *Room) GetClients() map[string]*Client {
	return r.clients
}

func (r *Room) GetStatus() string {
	switch r.status {
	case Close:
		return "close"
	case Open:
		return "open"
	default:
		return "unknown"
	}
}

func (r *Room) GetUniqueUsers() int {
	n := 0
	whitelist := []string{}
	for _, u := range r.clients {
		found := false
		for _, w := range whitelist {
			if u.username == w {
				found = true
			}
		}
		if found == false {
			n++
			whitelist = append(whitelist, u.username)
		}
	}
	return n
}

func (r *Room) UpdateRoom(id string) {
	m := sync.RWMutex{}
	m.Lock()
	defer m.Unlock()
	r.id = id
	for _, c := range r.clients {
		c.roomId = id
	}
}
