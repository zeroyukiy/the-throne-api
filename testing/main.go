package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/zeroyukiy/the-throne-api/internal/event"
)

func main() {
	wg := sync.WaitGroup{}

	for i := range 2 {
		wg.Add(1)
		time.Sleep(time.Duration(rand.Intn(20)) * time.Millisecond)
		go testConnection(i, &wg)
	}
	wg.Wait()
}

func testConnection(n int, wg *sync.WaitGroup) {
	t := time.Now().Add(time.Duration(rand.Intn(20)) * time.Second).UnixMilli()
	conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:8000/ws", nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	// defer func() {`

	conn.SetPingHandler(func(appData string) error {
		// fmt.Println("received a ping from server")
		err = conn.WriteMessage(websocket.PongMessage, nil)
		if err != nil {
			fmt.Println("problem sending pong message -> ", err)
		}
		return nil
	})

	data := event.WebsocketMessage{
		Event: event.Message,
		Payload: event.Payload{
			Message: "hello",
		},
		Timestamp: "",
	}
	b, err := json.Marshal(data)
	if err != nil {
		fmt.Println(err)
	}

	// for range 2 {
	err = conn.WriteMessage(websocket.TextMessage, b)
	if err != nil {
		fmt.Println(err)
		return
	}
	// time.Sleep(900 * time.Millisecond)
	// }

	go func() {
		for {
			now := time.Now().UnixMilli()
			if now > t {
				// fmt.Println("disconnect")
				wg.Done()
				closeConnection(conn)
				break
			}
			time.Sleep(1000 * time.Millisecond)
		}
	}()

	for {
		var data *event.WebsocketMessage
		err = conn.ReadJSON(&data)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				fmt.Printf("error: %v\n", err)
			}
			break
		}

		// 	switch data.Event {
		// 	case event.Message:
		// 		// fmt.Println(data.Message)
		// 		command := ""

		// 		words := strings.Split(data.Payload.Message, "@")
		// 		if len(words) > 1 {
		// 			command = words[1]
		// 		}

		// 		switch command {
		// 		case "welcome":
		// 			data := internal.DataRequest{
		// 				EventType: internal.Message,
		// 				Message:   "hey there, welcome in this chat.",
		// 			}
		// 			b, err := json.Marshal(&data)
		// 			if err != nil {
		// 				fmt.Println(err)
		// 			}
		// 			conn.WriteMessage(websocket.TextMessage, b)
		// 		default:
		// 			fmt.Println(".")
		// 		}

		// 	default:
		// 		fmt.Println("message type not recognized.")
		// 	}
	}
}

func closeConnection(conn *websocket.Conn) {
	err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseGoingAway, ""))
	if err != nil {
		fmt.Println("error sending close going away message ", err)
		return
	}
}
