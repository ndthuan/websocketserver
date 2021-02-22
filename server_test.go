package websocketserver_test

import (
	"github.com/gorilla/websocket"
	"github.com/ndthuan/websocketserver"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestServer_On(t *testing.T) {
	handlerMaker := func() websocketserver.MessageHandler {
		return func(conn *websocket.Conn, server *websocketserver.Server, message websocketserver.Message) error {
			return nil
		}
	}

	type args struct {
		messageType    string
		messageHandler websocketserver.MessageHandler
	}

	tests := []struct {
		name                 string
		args                 []args
		expectedHandlerCount int
	}{
		{
			name: "No handlers",
			args: []args{},
			expectedHandlerCount: 1,
		},
		{
			name: "Single handler",
			args: []args{{
				messageType:    "one",
				messageHandler: handlerMaker(),
			}},
			expectedHandlerCount: 1,
		},
		{
			name: "Multiple handlers",
			args: []args{
				{
					messageType:    "two",
					messageHandler: handlerMaker(),
				},
				{
					messageType:    "three",
					messageHandler: handlerMaker(),
				},
			},
			expectedHandlerCount: 2,
		},
	}

	for _, tt := range tests {
		s := websocketserver.Server{}

		for _, arg := range tt.args {
			s.On(arg.messageType, arg.messageHandler)
		}

		assert.Equal(t, tt.expectedHandlerCount, len(s.Handlers()), tt.name)
	}
}
