package internal

import (
	"fmt"
)

type MessageToBroadcast struct {
	client  *Client
	message []byte
}

type Hub struct {
	clients          map[*Client]bool
	roomRepository   RoomRepository
	registerClient   chan *Client
	unregisterClient chan *Client
	broadcast        chan MessageToBroadcast
}

func NewHub() *Hub {
	return &Hub{
		clients:          make(map[*Client]bool),
		roomRepository:   &RoomManager{},
		registerClient:   make(chan *Client),
		unregisterClient: make(chan *Client),
		broadcast:        make(chan MessageToBroadcast),
	}
}

func (h *Hub) InitRoomManager() {
	h.roomRepository = NewRoomManager()
}

func (h *Hub) Run() {
	h.InitRoomManager()
	fmt.Println("Hub is running.")
	for {
		select {
		case client := <-h.registerClient:
			h.clients[client] = true
			fmt.Println(client.username, "connected")

		case client := <-h.unregisterClient:
			room, ok := h.roomRepository.GetRoom(client.roomId)
			if ok {
				room.leave <- client
			}
			close(client.send)
			client.conn.Close()
			delete(h.clients, client)
			fmt.Println("client disconnected", client.username)

		case b := <-h.broadcast:
			c := b.client
			message := b.message
			for client, ok := range h.clients {
				if ok {
					if client.conn != c.conn {
						client.send <- message
					}
				}
			}
		}
	}
}

func (h *Hub) GetRoomRepository() RoomRepository {
	return h.roomRepository
}
