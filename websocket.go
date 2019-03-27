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

// func (this *WebSocketServer) AddPath(path string, spawnHandler SpawnHandler) {
// 	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
// 		ws, err := this.Upgrader.Upgrade(w, r, nil)
// 		if err != nil {
// 			log.Printf("connect failed!!! err:%v\n", err)
// 			panic(err)
// 		}

// 		defer ws.Close()
// 		conn := spawnHandler(ws)
// 		conn.Start()

// 		// intervDrt := time.Duration(this.options.HeartBeat) * time.Second
// 		// ws.StartHeartBeat(intervDrt)
// 	})
// }

func (this *WebSocket) ServeHttp(w http.ResponseWriter, r *http.Request, session Connection) error {
	ws, err := this.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}
	return session.Start(ws, &this.options)
}

// func (this *WebSocketServer) Shutdown(timeout int) error {
// 	timeoutDrt := time.Duration(timeout) * time.Second
// 	ctx, _ := context.WithTimeout(context.Background(), timeoutDrt)
// 	return http.Shutdown(ctx)
// }
