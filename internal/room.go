package internal

import "fmt"

type Room struct {
	id      string
	clients []*Client
}

func NewRoom(id string) *Room {
	return &Room{
		id:      id,
		clients: make([]*Client, 0),
	}
}

func (r *Room) Join(client *Client) {
	r.clients = append(r.clients, client)
	fmt.Println("client joined room: ", r.id)
}

func (r *Room) Leave(client *Client) {
	for k, c := range r.clients {
		if client.conn == c.conn {
			if len(r.clients) > 1 {
				r.clients = append(r.clients[:k], r.clients[k+1:]...)
			} else {
				r.clients = nil
			}
		}
	}
	fmt.Println("client left room: ", r.id)
}
