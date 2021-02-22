package main

import (
	"bufio"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/ndthuan/websocketserver"
	"os"
)

func main() {
	c, _, err := websocket.DefaultDialer.Dial("ws://localhost:8989/ws", nil)
	if err != nil {
		panic(err)
	}
	defer c.Close()

	err = c.WriteJSON(websocketserver.Message{
		Type:    "login",
		Payload: fmt.Sprintf("PID %d", os.Getpid()),
	})

	if err != nil {
		panic(err)
	}

	quit := make(chan int)

	go func() {
		var msg websocketserver.Message

		for {
			err := c.ReadJSON(&msg)

			if err != nil {
				if _, ok := err.(*websocket.CloseError); ok {
					println("Server disconnected")
				}

				quit <- 0
				break
			}

			println(fmt.Sprintf("%v", msg))
		}
	}()

	go func() {
		s := bufio.NewScanner(os.Stdin)

		for {
			if !s.Scan() {
				break
			}

			input := s.Text()

			switch input {
			case "bye":
				c.WriteJSON(websocketserver.Message{
					Type:    "logout",
					Payload: "Gotta go",
				})

				quit <- 0
				break
			default:
				c.WriteJSON(websocketserver.Message{
					Type:    "human-message",
					Payload: input,
				})
			}
		}
	}()

	select {
	case <-quit:
		return
	}
}
