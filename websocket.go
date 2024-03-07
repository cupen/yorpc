package yorpc

import (
	"net/http"

	"github.com/gorilla/websocket"
)

type WebSocket struct {
	Upgrader websocket.Upgrader
	Conn     *websocket.Conn
	options  Options
}

func OpenWebSocket(opts Options, w http.ResponseWriter, r *http.Request) (*WebSocket, error) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}
	ws := WebSocket{
		options:  opts,
		Upgrader: upgrader,
		Conn:     conn,
	}
	return &ws, nil
}
