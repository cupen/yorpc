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
	this.Callbacks[this.callSeqNum] = callback
	// this.write(websocket.BinaryMessage, msgBody)
}

func (this *Client) Send(id uint16, args []byte) error {

}
