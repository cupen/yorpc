package yorpc

import (
	"encoding/binary"
	"fmt"
	"log"
	"sync"

	"golang.org/x/exp/slog"
)

type ServerSession struct {
	opts *Options
	conn Conn
	// conn   *WebsocketConn
	mux    sync.Mutex
	peer   *Client
	server IServer
}

func NewSession(conn Conn, s IServer, opts *Options) *ServerSession {
	if s == nil {
		panic(fmt.Errorf("nil handlershub"))
	}
	return &ServerSession{
		opts:   opts,
		conn:   conn,
		server: s,
		peer:   NewClientByConn(conn),
	}
}

func (s *ServerSession) Start() error {
	return s.Run()
}

func (s *ServerSession) Stop() error {
	defer s.mux.Unlock()
	s.mux.Lock()
	if s.conn != nil {
		return s.conn.Close()
	}
	return nil
}

func (s *ServerSession) Run() error {
	if s.conn == nil {
		panic(fmt.Errorf("nil connection"))
	}
	for {
		msg, err := s.conn.ReadMessage()
		if err != nil {
			return err
		}

		if len(msg) <= 0 {
			if s.enableLog() {
				slog.Info("server received empty message")
			}
			continue
		}

		if err = s.onMessage(msg); err != nil {
			if s.enableLog() {
				log.Printf("server handle message failed. err:%v  data:%v", err, msg)
			}
			if s.opts.OnError != nil {
				s.opts.OnError(0, msg, err)
			}
			continue
		}
	}
}

func (s *ServerSession) writeMessage(data []byte) {
	s.conn.WriteMessage(data)
}

func (s *ServerSession) OnCall(id uint16, args []byte) ([]byte, error) {
	return s.server.OnCall(id, args)
}

func (s *ServerSession) OnSend(id uint16, args []byte) {
	s.server.OnSend(id, args)
}

func (s *ServerSession) enableLog() bool {
	return s.opts != nil && s.opts.DebugLog
}

func (s *ServerSession) onMessage(msg []byte) error {
	isCall, callId := codec.DecodeCallFlag(msg[0])
	_ = msg[2]

	var err error
	// call
	if isCall {
		msgId := binary.LittleEndian.Uint16(msg[1:3])
		if s.enableLog() {
			slog.Info("on-call", "callId", callId)
		}

		msgBody := msg[3:]
		var callRs []byte = nil
		defer func() {
			msg := codec.EncodeReturn(callId, callRs)
			s.writeMessage(msg)
			if s.enableLog() {
				slog.Info("on-return", "callId", callId, "result", callRs)
			}
		}()
		callRs, err = s.OnCall(msgId, msgBody)
		return err
	}
	// send
	if callId <= 0 {
		if s.enableLog() {
			slog.Info("on-send", "callId", callId)
		}
		msgId := binary.LittleEndian.Uint16(msg[1:3])
		msgBody := msg[3:]
		s.OnSend(msgId, msgBody)
		return nil
	}

	if s.enableLog() {
		slog.Info("on-callback", "callId", callId)
	}
	// callback
	msgBody := msg[1:]
	return s.peer.OnCallback(callId, msgBody)
}

func (s *ServerSession) Client() *Client {
	return s.peer
}
