# WebSocketServer
WebSocketServer helps you to organize code by introducing a consistent message format. Messages are routed to handlers based on their types.

This is a wrapper of the lower level library `gorilla/websocket`.

# Installation
```shell
go get github.com/ndthuan/websocketserver
```
Import:
```go
package main

import "github.com/ndthuan/websocketserver"
```

# Features
* Works with JSON standardized messages
```json
{
  "type": "message-type",
  "payload": "any kind of string, usually json-encoded"
}
```
* Super easy to use: messages are routed to registered handlers based on message type (think router)
```go
wsServer.On("login", loginHandler())
wsServer.On("logout", logoutHandler())
wsServer.On("human-message", humanMessageHandler())
```
* A standalone runner can be started in a separate goroutine to proactively broadcast messages to all clients without receiving a message
```go
package main

import (
	"github.com/gorilla/websocket"
	"github.com/ndthuan/websocketserver"
	"time"
)

func broadcastMessageBuilder(conn *websocket.Conn) (*websocketserver.Message, bool) {
	msg := &websocketserver.Message{
		Type:    "system-broadcast",
		Payload: "Hi, how are you?",
	}
	
	return msg, true
}

func standaloneRunner() websocketserver.StandaloneRunner {
	return func(server *websocketserver.Server) {
		for {
			time.Sleep(time.Second)

			server.Broadcast(broadcastMessageBuilder)
		}
	}
}

wsServer.SetStandaloneRunner(standaloneRunner())
```
* Can be integrated with any framework because the underlying library `gorilla/websocket` works with the standard Go HTTP server
* Hooks for connection events
```go
wsServer.OnConnected(connectedCallback())
wsServer.OnDisconnected(disconnectedCallback())
```

# Examples
See [a simple chat server](./examples/chat-server/) for an example of how it is used.
