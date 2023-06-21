package websocket

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

type Conn struct {
	conn  net.Conn
	state ws.State
}

func NewClient(url string, timeout time.Duration) (*Conn, error) {
	ctx := context.Background()
	ctx, _ = context.WithTimeout(ctx, timeout)
	conn, _, _, err := ws.Dial(ctx, url)
	if err != nil {
		return nil, err
	}
	return &Conn{
		conn:  conn,
		state: ws.StateClientSide,
	}, nil
}

func NewServer(r *http.Request, w http.ResponseWriter) (*Conn, error) {
	conn, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		return nil, err
	}
	return &Conn{
		conn:  conn,
		state: ws.StateServerSide,
	}, nil
}

func (wc *Conn) WriteMessage(msg []byte) error {
	return wsutil.WriteMessage(wc.conn, wc.state, ws.OpBinary, msg)
}

func (wc *Conn) Ping() error {
	frame := ws.NewPingFrame(nil)
	return ws.WriteFrame(wc.conn, frame)
}

func (wc *Conn) Pong() error {
	frame := ws.NewPongFrame(nil)
	return ws.WriteFrame(wc.conn, frame)
}

func (wc *Conn) ReadMessage() ([]byte, error) {
	data, opCode, err := wsutil.ReadData(wc.conn, wc.state)
	if opCode.IsData() {

	}
	return data, err
}

func (wc *Conn) Close() error {
	if wc.conn == nil {
		return nil
	}

	frame := ws.NewCloseFrame(nil)
	if err := ws.WriteFrame(wc.conn, frame); err != nil {
		return err
	}

	wc.conn.Close()
	wc.conn = nil
	return nil
}
