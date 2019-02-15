package yorpc

import "github.com/gorilla/websocket"

type Connection interface {
	Start(*websocket.Conn) error
	Stop()
}
