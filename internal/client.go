package internal

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var (
	pongTime = 60 * time.Second
	pingTime = 9 * pongTime / 10

	writeTime = 10 * time.Second
)

type LatencyQuality string

const (
	Optimal LatencyQuality = "#1B5E20"
	Good    LatencyQuality = "#4CAF50"
	Normal  LatencyQuality = "#FAFAFA"
	Bad     LatencyQuality = "#FF9800"
	Awful   LatencyQuality = "#B71C1C"
)

type Data struct {
	EventType EventType `json:"event_type"`
	Username  string    `json:"username"`
	Message   string    `json:"message"`
}

type Client struct {
	conn     *websocket.Conn
	send     chan []byte
	hub      *Hub
	latency  time.Time
	mu       sync.RWMutex
	username string
}

func NewClient(conn *websocket.Conn, hub *Hub, username string) *Client {
	return &Client{
		conn:     conn,
		hub:      hub,
		send:     make(chan []byte),
		username: username,
	}
}

func (c *Client) Run() {
	go c.fromWebsocketToHub()
	go c.fromHubToWebsocket()
}

func (c *Client) fromWebsocketToHub() {
	defer func() {
		c.hub.unregisterClient <- c
		c.conn.Close()
	}()

	// done := make(chan struct{})
	c.conn.SetReadDeadline(time.Now().Add(pongTime))
	c.conn.SetPongHandler(func(appData string) error {
		// fmt.Println("received a pong from the client")
		latency_status := LatencyStatus(c.latency)
		latency := &Data{
			EventType: Latency,
			// Message:   fmt.Sprintf("%.1fms", float32(time.Since(c.latency).Microseconds())/1000),
			Message: string(latency_status),
		}
		c.conn.SetReadDeadline(time.Now().Add(pongTime)) // receiving a pong message from the client is resetting the read dead line
		c.conn.WriteJSON(latency)
		return nil
	})

	var data *Data
	for {
		// _, msg, err := c.conn.ReadMessage()
		err := c.conn.ReadJSON(&data)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v\n", err)
			}
			fmt.Println(err)
			break
		}
		switch data.EventType {
		case Message:
			data.Username = c.username
			fmt.Println("data: ", data)
			b, err := json.Marshal(data)
			if err != nil {
				fmt.Println(err)
			}
			// c.hub.broadcast <- b
			c.hub.Broadcast(c, b)
		case Join:
			fmt.Println("user", c.username, "joined room: ", data.Message)
		case Leave:
			fmt.Println("user", c.username, "left room")
		}
	}
}

func (c *Client) fromHubToWebsocket() {
	ticker := time.NewTicker(pingTime)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeTime))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				fmt.Println(err)
				return
			}
			w.Write(msg)

			// n := len(c.send)
			// for range n {
			// 	w.Write([]byte("\n"))
			// 	w.Write(<-c.send)
			// }

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeTime)) // without setting a write deadline the server will not wait for the pong message from the client
			err := c.conn.WriteMessage(websocket.PingMessage, nil)
			if err != nil {
				c.hub.unregisterClient <- c
				return
			}
			c.mu.Lock()
			c.latency = time.Now()
			c.mu.Unlock()
		}
	}
}

func LatencyStatus(t time.Time) LatencyQuality {
	l := time.Since(t).Milliseconds()
	if l <= 30 {
		return Optimal
	} else if l > 30 && l <= 100 {
		return Good
	} else if l > 100 && l <= 150 {
		return Normal
	} else if l > 150 && l <= 250 {
		return Bad
	} else {
		return Awful
	}
}
