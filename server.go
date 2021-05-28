package yorpc

import (
	"log"
	"net"

	"encoding/binary"
	"net/http"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

type MsgHandlerV2 func(ServerSession, []byte)

type Server struct {
	conn       net.Conn
	Handlers   map[uint16]MsgHandlerV2
	Callbacks  map[uint8]func([]byte)
	callSeqNum uint8
}

func NewServer(opts *Options) *Server {
	return &Server{
		Handlers:   nil,
		Callbacks:  nil,
		callSeqNum: 0,
	}
}

func (this *Server) Start(r *http.Request, w http.ResponseWriter) error {
	return this.run(r, w)
}

func (this *Server) Stop() error {
	if this.conn != nil {
		// frame := ws.NewCloseFrame()
		// ws.WriteFrame(this.conn, frame)
		return this.conn.Close()
	}
	return nil
}

func (this *Server) _connect(r *http.Request, w http.ResponseWriter) error {
	conn, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		return err
	}
	this.conn = conn
	return nil
}

func (this *Server) run(r *http.Request, w http.ResponseWriter) error {
	if err := this._connect(r, w); err != nil {
		return err
	}
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

func (this *Server) write(data []byte) error {
	return wsutil.WriteServerMessage(this.conn, ws.OpBinary, data)
}

func (this *Server) onMessage(msgBody []byte) error {
	callFlag := msgBody[0]
	isCall := (callFlag >> 7) == 1
	callSeqId := (callFlag & 0x7f)

	var msgId uint16
	var msgData []byte
	if callSeqId <= 0 {
		// send
		// byte-2~3
		msgId = binary.LittleEndian.Uint16(msgBody[1:3])
		msgData = msgBody[3:]
		return this.OnSend(msgId, msgData)
	}

	var callRs []byte = nil
	if isCall {
		// call
		// byte-2~3
		msgId = binary.LittleEndian.Uint16(msgBody[1:3])
		msgData = msgBody[3:]
		defer func() {
			// log.Printf("return msg callSeqId:%d. callFlag:%d data:%v\n", callSeqId, callFlag, callRs)
			this.ReturnMsg(callSeqId, callRs)
		}()
		callRs, err = this.OnCall(msgId, msgData)
		return
	}

	// callback
	msgData = msgBody[1:]
	callback, _ := this.Callbacks[callSeqId]
	if callback == nil {
		log.Printf("Invalid callSeqId: %d. callFlag: %d\n", callSeqId, callFlag)
		return
	}
	callback(msgData)
	delete(this.Callbacks, callSeqId)
	return nil
}

func (this *Server) Call(msgId uint16, data []byte, callback func([]byte)) {
	// log.Printf("call %d\n", msgId)
	this.callSeqNum++
	callSeqId := this.callSeqNum

	var callFlag uint8 = (1 << 7) + (callSeqId & 0x7f)

	msgIdBytes := []byte{0, 0}
	binary.LittleEndian.PutUint16(msgIdBytes, msgId)
	msgBody := []byte{callFlag}
	msgBody = append(msgBody, msgIdBytes...)
	msgBody = append(msgBody, data...)
	this.Callbacks[this.callSeqNum] = callback
	this.write(msgBody)
}

func (this *Server) SendMsg(msgId uint16, data []byte) {
	// log.Printf("send %d\n", msgId)
	msgIdBytes := []byte{0, 0}
	binary.LittleEndian.PutUint16(msgIdBytes, msgId)

	var callFlag uint8 = 0
	msgBody := append([]byte{callFlag}, msgIdBytes...)
	msgBody = append(msgBody, data...)
	if this.conn == nil {
		log.Printf("websocket was nil\n")
		return
	}
	this.write(msgBody)
}

func (this *Server) ReturnMsg(callSeqId uint8, data []byte) {
	// log.Printf("return call %d\n", callSeqId)
	// isReq + callSeqId
	var callFlag uint8 = (0 << 7) + (callSeqId & 0x7f)
	data = append([]byte{callFlag}, data...)
	err := this.write(data)
	if err != nil {
		log.Printf("write error %v", err)
	}
}

func (this *Server) OnCall(id uint16, data []byte) ([]byte, error) {
	return nil
}

func (this *Server) OnSend(id uint16, data []byte) error {
	return nil
}
