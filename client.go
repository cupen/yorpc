package yorpc

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"
)

type Client struct {
	conn       Conn
	callbacks  [256]func([]byte)
	callIdSeed int32
	timeout    time.Duration
}

func NewClient(conn Conn, timeout time.Duration) *Client {
	c := &Client{
		conn:      conn,
		callbacks: [256]func([]byte){},
		timeout:   timeout,
	}
	c.Start()
	return c
}

func NewClientByConn(conn Conn) *Client {
	return &Client{
		conn:      conn,
		callbacks: [256]func([]byte){},
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
		msg, err := c.conn.ReadMessage()
		if err != nil {
			return err
		}
		if len(msg) <= 0 {
			continue
		}
		if err = c.onMessage(msg); err != nil {
			return err
		}
	}
}

func (c *Client) Call(ctx context.Context, id uint16, args []byte) ([]byte, error) {
	var callId = c.nextCallId()
	var callRsCh = make(chan []byte)
	c.callbacks[callId] = func(rs []byte) {
		callRsCh <- rs
	}
	var cancel context.CancelFunc
	if ctx == nil {
		ctx = context.TODO()
		if c.timeout > 0 {
			ctx, cancel = context.WithTimeout(ctx, c.timeout)
			defer cancel()
		}
	}

	msg := codec.EncodeCall(callId, id, args)
	err := c.conn.WriteMessage(msg)
	if err != nil {
		return nil, err
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case rs := <-callRsCh:
		return rs, nil
	}

}

func (c *Client) Send(id uint16, args []byte) error {
	// slog.Info("[client] send", "id", id, "args", args)
	msg := codec.EncodeSend(id, args)
	return c.conn.WriteMessage(msg)
}

func (c *Client) OnCallback(callId uint8, result []byte) error {
	// slog.Info("[client] callback", "callId", callId, "result", result)
	cb := c.callbacks[callId]
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
