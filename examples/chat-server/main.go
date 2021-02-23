package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/ndthuan/websocketserver"
	"log"
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
		log.Println("Login handler triggered")
		username := message.Payload

		authenticatedConnections[connection] = username

		broadcast(fmt.Sprintf("Everyone, please welcome %s to our chat room!", username), connection)

		return nil
	}
}

func logoutHandler() websocketserver.MessageHandler {
	return func(connection *websocket.Conn, listener *websocketserver.Server, message websocketserver.Message) error {
		log.Println("Logout handler triggered")
		username, loggedIn := authenticatedConnections[connection]

		if loggedIn {
			connection.WriteJSON(websocketserver.Message{
				Type:    "system-message",
				Payload: "Bye-bye, " + username,
			})
			delete(authenticatedConnections, connection)
			broadcast(fmt.Sprintf("%s left the room", username), connection)
			_ = connection.WriteMessage(websocket.CloseMessage, []byte{})
		}

		return nil
	}
}

func disconnectedCallback() websocketserver.ConnectionCallback {
	return func(connection *websocket.Conn, listener *websocketserver.Server) error {
		log.Println("Client disconnected")
		loggedInUser, already := authenticatedConnections[connection]
		if already {
			broadcast(fmt.Sprintf("%s was disconnected", loggedInUser), connection)
			delete(authenticatedConnections, connection)
		}
		return nil
	}
}

func connectedCallback() websocketserver.ConnectionCallback {
	return func(connection *websocket.Conn, listener *websocketserver.Server) error {
		log.Println("Client connected")
		return connection.WriteJSON(websocketserver.Message{
			Type:    "system-message",
			Payload: "You are connected",
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

func broadcastMessageBuilder(conn *websocket.Conn) (*websocketserver.Message, bool) {
	if username, loggedIn := authenticatedConnections[conn]; loggedIn {
		now, _ := time.Now().MarshalText()

		msg := &websocketserver.Message{
			Type:    "system-broadcast",
			Payload: fmt.Sprintf("Hi %s, now is %s, how are you?", username, now),
		}

		return msg, true
	}

	return nil, false
}

func standaloneRunner() websocketserver.StandaloneRunner {
	return func(server *websocketserver.Server) {
		log.Println("Standalone runner was started")
		for {
			time.Sleep(time.Second)

			server.Broadcast(broadcastMessageBuilder)
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
