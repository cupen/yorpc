package yorpc

import (
	"encoding/binary"
	"fmt"
	"log"
	"sync"
)

type MsgHandlerV2 func(ServerSession, []byte)

type Server struct {
	conn     Connection
	handlers *HandlersHub
	mux      sync.Mutex
	peer     *Client
}

func NewServer(conn Connection, handlers *HandlersHub, opts *Options) *Server {
	return &Server{
		conn:     conn,
		handlers: handlers,
		peer:     NewClientByConn(conn),
	}
}

func (s *Server) Start() error {
	return s.Run()
}

func (s *Server) Stop() error {
	defer s.mux.Unlock()
	s.mux.Lock()
	if s.conn != nil {
		return s.conn.Close()
	}
	return nil
}

func (s *Server) Run() error {
	if s.conn == nil {
		panic(fmt.Errorf("nil connection"))
	}
	for {
		msg, err := s.conn.ReceiveMessage()
		if err != nil {
			return err
		}

		log.Printf("server received message: %v", msg)
		if len(msg) <= 0 {
			continue
		}

		err = s.onMessage(msg)
		if err != nil {
			return err
		}
	}
}

func (s *Server) writeMessage(msg []byte) {
	log.Printf("server write %v", msg)
	s.conn.WriteMessage(msg)
}

func (s *Server) OnCall(id uint16, args []byte) ([]byte, error) {
	log.Printf("server.OnCall id=%d args=%v", id, args)
	handler, err := s.handlers.GetHandler(id)
	if err != nil {
		return nil, err
	}
	return handler(args)
}

func (s *Server) OnSend(id uint16, args []byte) error {
	log.Printf("server.OnSend id=%d args=%v", id, args)
	handler, err := s.handlers.GetHandler(id)
	if err != nil {
		return err
	}
	_, err = handler(args)
	return err
}

func (s *Server) onMessage(msg []byte) error {
	isCall, callId := codec.DecodeCallFlag(msg[0])
	// server(send)
	if callId <= 0 {
		_ = msg[2]
		// TODO: move to codec
		// send(byte:2~3)
		msgId := binary.LittleEndian.Uint16(msg[1:3])
		msgBody := msg[3:]
		return s.OnSend(msgId, msgBody)
	}

	// server(call)
	var err error
	if isCall {
		_ = msg[2]
		// TODO: move to codec
		// call(byte:2~3)
		msgId := binary.LittleEndian.Uint16(msg[1:3])
		msgBody := msg[3:]
		var callRs []byte = nil
		defer func() {
			msg := codec.EncodeReturn(callId, callRs)
			s.writeMessage(msg)
			// log.Printf("return msg callSeqId:%d. callFlag:%d data:%v\n", callSeqId, callFlag, callRs)
		}()
		callRs, err = s.OnCall(msgId, msgBody)
		return err
	}
	log.Printf("server.peer on call")
	msgBody := msg[1:]
	return s.peer.OnCallback(callId, msgBody)
}

func (s *Server) Client() *Client {
	return s.peer
}
