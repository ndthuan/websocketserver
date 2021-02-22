package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/ndthuan/websocketserver"
	"math/rand"
	"time"
)

// This map is not thread safe, it's for demonstration purpose only. You should implement your own lock mechanism.
var authenticatedConnections map[*websocket.Conn]string

func init() {
	authenticatedConnections = make(map[*websocket.Conn]string)
}

func broadcast(msg string, excluded *websocket.Conn) {
	for conn := range authenticatedConnections {
		if conn == excluded {
			continue
		}
		_ = conn.WriteJSON(websocketserver.Message{
			Type:    "system-broadcast",
			Payload: msg,
		})
	}
}

func loginHandler() websocketserver.MessageHandler {
	return func(connection *websocket.Conn, listener *websocketserver.Server, message websocketserver.Message) error {
		username := message.Payload

		authenticatedConnections[connection] = username

		broadcast(fmt.Sprintf("Everyone, please welcome %s to our chat room!", username), connection)

		return nil
	}
}

func logoutHandler() websocketserver.MessageHandler {
	return func(connection *websocket.Conn, listener *websocketserver.Server, message websocketserver.Message) error {
		username, loggedIn := authenticatedConnections[connection]

		if loggedIn {
			delete(authenticatedConnections, connection)
			broadcast(fmt.Sprintf("%s left the room", username), connection)
			_ = connection.WriteMessage(websocket.CloseMessage, []byte{})
		}

		return nil
	}
}

func disconnectedCallback() websocketserver.ConnectionCallback {
	return func(connection *websocket.Conn, listener *websocketserver.Server) error {
		loggedInUser, already := authenticatedConnections[connection]
		if already {
			broadcast(fmt.Sprintf("%s disconnected", loggedInUser), connection)
			delete(authenticatedConnections, connection)
		}
		return nil
	}
}

func connectedCallback() websocketserver.ConnectionCallback {
	return func(connection *websocket.Conn, listener *websocketserver.Server) error {
		return connection.WriteJSON(websocketserver.Message{
			Type:    "system-message",
			Payload: "Hi there, this is private message to you. Welcome!",
		})
	}
}

func humanMessageHandler() websocketserver.MessageHandler {
	return func(connection *websocket.Conn, listener *websocketserver.Server, message websocketserver.Message) error {
		username, ok := authenticatedConnections[connection]

		if ok {
			broadcast(fmt.Sprintf("%s: %s", username, message.Payload), connection)
		}

		return nil
	}
}

func standaloneRunner() websocketserver.StandaloneRunner {
	return func(server *websocketserver.Server) {
		for {
			time.Sleep(time.Duration(rand.Intn(10)) * time.Second)

			server.Broadcast(func(conn *websocket.Conn) (*websocketserver.Message, bool) {
				if username, loggedIn := authenticatedConnections[conn]; loggedIn {
					now, _ := time.Now().MarshalText()

					msg := &websocketserver.Message{
						Type:    "system-broadcast",
						Payload: fmt.Sprintf("Hi %s, now is %s, how are you?", username, now),
					}

					return msg, true
				}

				return nil, false
			})
		}
	}
}

func main() {
	wsServer := websocketserver.Server{
		Upgrader: websocket.Upgrader{},
	}

	wsServer.OnConnected(connectedCallback())
	wsServer.OnDisconnected(disconnectedCallback())
	wsServer.SetStandaloneRunner(standaloneRunner())
	wsServer.On("login", loginHandler())
	wsServer.On("logout", logoutHandler())
	wsServer.On("human-message", humanMessageHandler())

	e := echo.New()
	e.Use(middleware.Logger())
	e.GET("/ws", func(c echo.Context) error {
		return wsServer.Start(c.Response(), c.Request(), nil)
	})

	e.Logger.Fatal(e.Start(":8989"))
}
