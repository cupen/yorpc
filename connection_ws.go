package yorpc

import (
	"net"
	"net/http"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

type WebsocketConn struct {
	conn  net.Conn
	state ws.State
	ss    ServerSession
	cs    ClientSession
}

func NewWebsocketByHTTP(r *http.Request, w http.ResponseWriter) (*WebsocketConn, error) {
	conn, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		return nil, err
	}
	return &WebsocketConn{
		conn:  conn,
		state: ws.StateServerSide,
	}, nil
}

func (wc *WebsocketConn) WriteMessage(msg []byte) error {
	return wsutil.WriteMessage(wc.conn, wc.state, ws.OpBinary, msg)
}

func (wc *WebsocketConn) ReceiveMessage() ([]byte, error) {
	data, _, err := wsutil.ReadData(wc.conn, wc.state)
	return data, err
}

func (wc *WebsocketConn) SetSession(cs ClientSession, ss ServerSession) {
	wc.cs = cs
	wc.ss = ss
}

func (wc *WebsocketConn) Close() error {
	if wc.conn == nil {
		return nil
	}
	wc.conn.Close()
	return nil
}
