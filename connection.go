package yorpc

import (
	"log"
	"net"
	"net/http"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

type Connection struct {
	conn net.Conn
}

func NewConnectionByURL(r *http.Request, w http.ResponseWriter) (*Connection, error) {
	conn, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		return nil, err
	}
	return &Connection{
		conn: conn,
	}, nil
}

func NewConnectionByHTTP(r *http.Request, w http.ResponseWriter) (*Connection, error) {
	conn, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		return nil, err
	}
	return &Connection{
		conn: conn,
	}, nil
}

func (this *Connection) Start() error {
	return this.run()
}

func (this *Connection) Stop() error {
	return this.conn.Close()
}

func (this *Connection) run() error {
	for {
		msg, op, err := wsutil.ReadClientData(this.conn)
		if err != nil {
			return err
		}

		switch op {
		case ws.OpPing:
			err = wsutil.WriteServerMessage(this.conn, ws.OpPong, msg)
		case ws.OpPong:
		case ws.OpBinary:
			err = this.onMessage(msg)
		case ws.OpText:
			err = this.onMessage(msg)
		// case ws.OpClose:
		default:
			log.Printf("unexpected opcode = %v", op)
		}
		if err != nil {
			return err
		}
	}
}

func (this *Connection) onMessage(msg []byte) error {
	// err = wsutil.WriteServerMessage(this.conn, ws.OpBinary, msg)
	return nil
}
