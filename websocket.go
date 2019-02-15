package yorpc

import (
	"net/http"

	"github.com/gorilla/websocket"
)

type WebSocketHandler struct {
	Upgrader websocket.Upgrader
	options  Options
}

func NewWebSocketHandler(options Options) *WebSocketHandler {
	return &WebSocketHandler{
		options:  options,
		Upgrader: websocket.Upgrader{},
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

func (this *WebSocketHandler) ServeHttp(w http.ResponseWriter, r *http.Request, session Connection) error {
	ws, err := this.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}
	return session.Start(ws)
}

// func (this *WebSocketServer) Shutdown(timeout int) error {
// 	timeoutDrt := time.Duration(timeout) * time.Second
// 	ctx, _ := context.WithTimeout(context.Background(), timeoutDrt)
// 	return http.Shutdown(ctx)
// }
