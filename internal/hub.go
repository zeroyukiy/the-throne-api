package internal

import (
	"fmt"
)

type Hub struct {
	clients          map[*Client]bool
	registerClient   chan *Client
	unregisterClient chan *Client
	broadcast        chan []byte
}

func NewHub() *Hub {
	return &Hub{
		clients:          make(map[*Client]bool),
		registerClient:   make(chan *Client),
		unregisterClient: make(chan *Client),
		broadcast:        make(chan []byte),
	}
}

func (h *Hub) Run() {
	fmt.Println("Hub is running.")
	for {
		select {
		case client := <-h.registerClient:
			fmt.Println("client connected")
			h.clients[client] = true
			fmt.Println("clients: ", len(h.clients))
		case client := <-h.unregisterClient:
			fmt.Println("client disconnected")
			// h.clients[client] = false
			close(client.send)
			client.conn.Close()
			delete(h.clients, client)
			fmt.Println("clients: ", len(h.clients))
		case msg := <-h.broadcast:
			for client, ok := range h.clients {
				if ok {
					client.send <- msg
				}
			}
		}
	}
}

func (h *Hub) RegisterClient(client *Client) {
	h.registerClient <- client
}

func (h *Hub) UnregisterClient(client *Client) {
	h.unregisterClient <- client
}

func (h *Hub) Broadcast(client *Client, b []byte) {
	for c, ok := range h.clients {
		if ok {
			if c.conn != client.conn {
				c.send <- b
			}
		}
	}
}
