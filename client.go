package yorpc

import (
	"context"
	"fmt"
	"log"
	"sync/atomic"

	"github.com/gobwas/ws"
)

type Client struct {
	conn       Connection
	callbacks  map[uint8]func([]byte)
	callIdSeed int32
}

func NewClient(url string) (*Client, error) {
	ctx := context.Background()
	conn, _, _, err := ws.Dial(ctx, url)
	if err != nil {
		return nil, err
	}

	return &Client{
		conn: &WebsocketConn{
			conn:  conn,
			state: ws.StateClientSide,
		},
		callbacks: map[uint8]func([]byte){},
	}, nil
}

func NewClientByConn(conn Connection) *Client {
	return &Client{
		conn:      conn,
		callbacks: map[uint8]func([]byte){},
	}
}

func (c *Client) nextCallId() uint8 {
	maxValue := int32(128)
	callId := atomic.AddInt32(&c.callIdSeed, 1) % maxValue
	if callId <= 0 {
		callId = atomic.AddInt32(&c.callIdSeed, 1) % maxValue
		if callId <= 0 {
			panic(fmt.Errorf("invalid call id %d", callId))
		}
	}
	return uint8(callId)
}

func (c *Client) Start() {
	go c.Run()
}

func (c *Client) Run() error {
	if c.conn == nil {
		panic(fmt.Errorf("nil connection"))
	}
	for {
		msg, err := c.conn.ReceiveMessage()
		if err != nil {
			return err
		}

		log.Printf("server received message: %v", msg)
		if len(msg) <= 0 {
			continue
		}

		err = c.onMessage(msg)
		if err != nil {
			return err
		}
	}
}

func (c *Client) Call(ctx context.Context, id uint16, args []byte) ([]byte, error) {
	var callId = c.nextCallId()
	// var callRs []byte = nil
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	var callRsCh = make(chan []byte)
	c.callbacks[callId] = func(rs []byte) {
		// callRs = rs
		callRsCh <- rs
		// cancel()
	}

	msg := codec.EncodeCall(callId, id, args)
	err := c.conn.WriteMessage(msg)
	if err != nil {
		return nil, err
	}
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case rs := <-callRsCh:
			return rs, nil
		}
	}
}

func (c *Client) Send(id uint16, args []byte) error {
	log.Printf("client.Send id=%d args=%v", id, args)
	msg := codec.EncodeSend(id, args)
	return c.conn.WriteMessage(msg)
}

func (c *Client) OnCallback(callId uint8, result []byte) error {
	log.Printf("client.OnCallBack callId=%d result=%v", callId, result)
	cb, ok := c.callbacks[callId]
	_ = ok
	if cb == nil {
		return fmt.Errorf("no callback for callId(%d)", callId)
	}
	cb(result)
	return nil
}

func (c *Client) onMessage(msg []byte) error {
	isCall, callId := codec.DecodeCallFlag(msg[0])
	// server(send)
	if callId <= 0 || isCall {
		return fmt.Errorf("client is not server")
	}
	msgBody := msg[1:]
	return c.OnCallback(callId, msgBody)
}
