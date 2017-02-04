package ws

import (
	"logger"

	"golang.org/x/net/websocket"
)

type WsMessage struct {
	Message map[string]interface{}
	Error   error
}

// WsConnection represents web socket connection, implements Connection interface.
type WsConnection struct {
	connnection *websocket.Conn
	ToClient    chan WsMessage
	interrupt   chan bool
}

func NewConnection(connnection *websocket.Conn) *WsConnection {
	logger.Info("WsConnection", "NewConnection", "New connection")
	conn := WsConnection{
		connnection: connnection,
		ToClient:    make(chan WsMessage, 100),
		interrupt:   make(chan bool),
	}

	go func() {
		for {
			logger.Info("WsConnection", "NewConnection", "Waiting for message")
			select {
			case <-conn.interrupt:
				return
			default:

				msg := make(map[string]interface{})
				err := websocket.JSON.Receive(conn.connnection, &msg)
				conn.ToClient <- WsMessage{
					Message: msg,
					Error:   err,
				}
			}
		}
	}()
	return &conn
}

// Send sends message through the connection.
func (c WsConnection) Send(msg interface{}) error {
	return websocket.JSON.Send(c.connnection, msg)
}

// Close closes the connection
func (c WsConnection) Close() error {
	c.interrupt <- true
	return c.connnection.Close()
}
