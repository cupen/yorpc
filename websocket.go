package yorpc

import (
	"net/http"

	"github.com/gorilla/websocket"
)

type WebSocket struct {
	Upgrader websocket.Upgrader
	options  Options
}

func NewWebSocket(options Options) *WebSocket {
	return &WebSocket{
		options: options,
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			}},
	}
}

func (this *WebSocket) ServeHttp(w http.ResponseWriter, r *http.Request, session Connection) error {
	ws, err := this.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}
	ws.Close()
	return session.Start(ws)
}
