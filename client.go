package yorpc

import (
	"context"
	"encoding/binary"
	"net"
)

type Client struct {
	conn       net.Conn
	callSeqNum uint8
	Callbacks  map[uint8]func([]byte)
}

func NewClient(url string) (*Client, error) {

}

func (this *Client) run() error {
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
	this.Callbacks[this.callSeqNum] = func(rs []byte) {
		callRs = rs

		d := ctx.Done()
		d <- struct{}{}
	}
	// this.write(websocket.BinaryMessage, msgBody)
	return callRs, nil
}

func (this *Client) Send(id uint16, args []byte) error {

}
