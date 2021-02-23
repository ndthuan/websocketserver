# websocketserver
Websocketserver helps to organize code for handling different formats of websocket message payloads sent by clients.

# Features
* Standardized websocket messages format
```json
{
  "type": "message-type",
  "payload": "any string, usually json formatted"
}
```
* Messages are routed to registered handlers based on message type
```go
wsServer.On("login", loginHandler())
wsServer.On("logout", logoutHandler())
wsServer.On("human-message", humanMessageHandler())
```
* A standalone runner can be started only once in a separate goroutine to proactively send messages to clients without receiving a message
```go
wsServer.SetStandaloneRunner(standaloneRunner())
```
* Can be integrated with any web server framework

# Examples
See [a basic chat server example](./examples/chat-server/).
