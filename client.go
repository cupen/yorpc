package yorpc

import (
	"context"
	"encoding/binary"
	"net"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

type Client struct {
	conn       net.Conn
	callSeqNum uint8
	Callbacks  map[uint8]func([]byte)
}

func NewClient(conn Connection) (*Client, error) {
	return &Client{
		Callbacks: map[uint8]func([]byte){},
	}, nil
}

func (this *Client) Call(ctx context.Context, id uint16, args []byte) ([]byte, error) {
	this.callSeqNum++
	if this.callSeqNum == 0 {
		this.callSeqNum++
	}
	var callSeqId = this.callSeqNum
	var callFlag uint8 = (1 << 7) + (callSeqId & 0x7f)

	msgIdBytes := []byte{0, 0}
	binary.LittleEndian.PutUint16(msgIdBytes, id)
	msgBody := []byte{callFlag}
	msgBody = append(msgBody, msgIdBytes...)
	msgBody = append(msgBody, args...)
	var callRs []byte = nil
	ctx, cancel := context.WithCancel(ctx)
	this.Callbacks[callSeqId] = func(rs []byte) {
		callRs = rs
		cancel()
	}
	err := wsutil.WriteClientMessage(this.conn, ws.OpBinary, msgBody)
	if err != nil {
		cancel()
		return nil, err
	}
	<-ctx.Done()
	return callRs, err
}

func (this *Client) Send(id uint16, args []byte) error {
	return nil

}
