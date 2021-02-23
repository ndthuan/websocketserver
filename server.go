package websocketserver

import (
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
)

//Message defines the standardized message structure.
type Message struct {
	Type    string
	Payload string
}

//MessageHandler handles a specific command.
type MessageHandler func(conn *websocket.Conn, server *Server, message Message) error

//ConnectionCallback handles a connection event such as connected, disconnected
type ConnectionCallback func(conn *websocket.Conn, server *Server) error

//StandaloneRunner is a func that runs in a separate goroutine. It's started only when the first client is connected.
//It can be used to periodically send messages to connected clients without receiving a message.
type StandaloneRunner func(server *Server)

//BroadcastMessageBuilder is the callback on each connection when the server broadcasts. It should return true or false
//to signal if the message should be sent to that client.
type BroadcastMessageBuilder func(conn *websocket.Conn) (*Message, bool)

//Server acts like a router that routes messages to handlers based on type specified in each message.
type Server struct {
	Upgrader         websocket.Upgrader
	allHandler       MessageHandler
	handlers         map[string]MessageHandler
	onDisconnected   ConnectionCallback
	onConnected      ConnectionCallback
	standaloneRunner StandaloneRunner
	connections      map[*websocket.Conn]bool
	locker           sync.RWMutex
	once             sync.Once
}

func (s *Server) addConnection(conn *websocket.Conn) {
	s.locker.Lock()
	defer s.locker.Unlock()

	if s.connections == nil {
		s.connections = make(map[*websocket.Conn]bool)
	}

	s.connections[conn] = true
}

//On registers a handler on a message type. The existing handler of the same type will be overridden.
func (s *Server) On(t string, messageHandler MessageHandler) {
	if s.handlers == nil {
		s.handlers = make(map[string]MessageHandler)
	}

	s.handlers[t] = messageHandler
}

//Handlers returns all registered handlers
func (s *Server) Handlers() map[string]MessageHandler {
	return s.handlers
}

//OnAll registers a catch-all handler that will be executed on all message types.
//If there is a specific handler registered on a type, this will be executed after.
func (s *Server) OnAll(handler MessageHandler) {
	s.allHandler = handler
}

//OnDisconnected registers a callback when a connection is closed.
func (s *Server) OnDisconnected(callback ConnectionCallback) {
	s.onDisconnected = callback
}

//OnConnected registers a callback when a connection is opened.
func (s *Server) OnConnected(callback ConnectionCallback) {
	s.onConnected = callback
}

//SetStandaloneRunner registers a StandaloneRunner.
func (s *Server) SetStandaloneRunner(runner StandaloneRunner) {
	s.standaloneRunner = runner
}

//Broadcast sends messages to every connected client.
//The messageBuilder func can build a different message for a different client,
//it can also signal if the message should be sent to that client.
func (s *Server) Broadcast(messageBuilder BroadcastMessageBuilder) {
	s.locker.RLock()
	defer s.locker.RUnlock()

	for c := range s.connections {
		if message, goAhead := messageBuilder(c); goAhead {
			_ = c.WriteJSON(*message)
		}
	}
}

//Start upgrades an HTTP connection and starts listening for the messages.
func (s *Server) Start(w http.ResponseWriter, r *http.Request, responseHeader http.Header) error {
	c, err := s.Upgrader.Upgrade(w, r, responseHeader)
	if err != nil {
		return err
	}

	if s.onConnected != nil {
		if err := s.onConnected(c, s); err != nil {
			return err
		}
	}

	s.addConnection(c)

	if s.standaloneRunner != nil {
		s.once.Do(func() {
			go s.standaloneRunner(s)
		})
	}

	defer func() { _ = c.Close() }()

	var msg Message

	for {
		err := c.ReadJSON(&msg)
		if _, isCloseError := err.(*websocket.CloseError); isCloseError {
			if s.onDisconnected != nil {
				_ = s.onDisconnected(c, s)
			}
			return nil
		} else if err != nil {
			return err
		}

		if handler, exists := s.handlers[msg.Type]; exists {
			if err := handler(c, s, msg); err != nil {
				return err
			}
		}

		if s.allHandler != nil {
			if err := s.allHandler(c, s, msg); err != nil {
				return err
			}
		}
	}
}
