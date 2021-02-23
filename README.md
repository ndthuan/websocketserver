# websocketserver
Helps to organize code for handling different formats of websocket message payloads sent by clients.

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
wsServer.SetStandaloneRunner(standaloneRunner())
```
* Can be integrated with any framework because the underlying library `gorilla/websocket` works with the standard Go HTTP server
* Hooks for connection events
```go
wsServer.OnConnected(connectedCallback())
wsServer.OnDisconnected(disconnectedCallback())
```

# Examples
See [a basic chat server example](./examples/chat-server/).
