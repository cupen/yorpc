package yorpc

import (
	"encoding/binary"
	"log"
	"net"
	"net/http"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

type WebsocketConn struct {
	conn net.Conn
	ss   ServerSession
}

func NewWebsocketByURL(r *http.Request, w http.ResponseWriter) (*WebsocketConn, error) {
	conn, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		return nil, err
	}
	return &WebsocketConn{
		conn: conn,
	}, nil
}

func NewWebsocketByHTTP(r *http.Request, w http.ResponseWriter) (*WebsocketConn, error) {
	conn, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		return nil, err
	}
	return &WebsocketConn{
		conn: conn,
	}, nil
}

func (this *WebsocketConn) Start() error {
	return this.run()
}

func (this *WebsocketConn) Stop() error {
	return this.conn.Close()
}

func (this *WebsocketConn) run() error {
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
			err = this.OnMessage(msg)
		case ws.OpText:
			err = this.OnMessage(msg)
		// case ws.OpClose:
		default:
			log.Printf("unexpected opcode = %v", op)
		}
		if err != nil {
			return err
		}
	}
}

func (this *WebsocketConn) OnMessage(msg []byte) error {
	msgBody := msg
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
		return this.ss.OnSend(msgId, msgData)
	}

	var err error
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
		return err
	}

	// callback
	msgData = msgBody[1:]
	callback, _ := this.Callbacks[callSeqId]
	if callback == nil {
		log.Printf("Invalid callSeqId: %d. callFlag: %d\n", callSeqId, callFlag)
		return nil
	}
	callback(msgData)
	delete(this.Callbacks, callSeqId)
	return nil
}

// func (this *RpcSession) ReturnMsg(callSeqId uint8, data []byte) {
// 	// log.Printf("return call %d\n", callSeqId)
// 	// isReq + callSeqId
// 	var callFlag uint8 = (0 << 7) + (callSeqId & 0x7f)
// 	data = append([]byte{callFlag}, data...)
// 	if this.ws == nil {
// 		log.Printf("websocket was nil\n")
// 		return
// 	}
// 	err := this.write(websocket.BinaryMessage, data)
// 	if err != nil {
// 		log.Printf("write error %v", err)
// 	}
// }

func (this *WebsocketConn) WriteMessage(msg []byte) error {
	return wsutil.WriteServerMessage(this.conn, ws.OpBinary, msg)
}

func (this *WebsocketConn) SetSession(cs ClientSession, ss ServerSession) error {
	this.ss = ss
}
