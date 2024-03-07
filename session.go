package yorpc

import (
	"encoding/binary"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type ID interface {
	int64 | uint64 | string
}

type Event int

var Events = struct {
	Start     Event
	Stop      Event
	KeepAlive Event
}{
	Start:     1,
	Stop:      2,
	KeepAlive: 3,
}

type Session[id ID] interface {
	GetID() id
	GetType() string
	SendMsg(uint16, []byte)
	IsAlive(now time.Time) bool
	KeepAlive(time.Duration)
	OnEvent(Event) error
	Close()
	GetToken() string
	// GetPlayer() interface{}
	// Call(uint16, []byte, func([]byte))
	// ReturnMsg(uint8, []byte)
}

type RpcSession[id ID] struct {
	id         id
	token      string
	ws         *websocket.Conn
	handler    MsgHandler[id]
	callbacks  map[uint8]func([]byte)
	callSeqNum uint8
	closedAt   time.Time
	mux        sync.Mutex
}

func NewRpcSession[id ID](_id id, token string, handler MsgHandler[id]) *RpcSession[id] {
	return &RpcSession[id]{
		id:        _id,
		token:     token,
		handler:   handler,
		callbacks: map[uint8]func([]byte){},
	}
}

func (this *RpcSession[id]) Connect(ws *WebSocket) error {
	return this.Connect2(ws.Conn)
}

func (this *RpcSession[id]) Connect2(ws *websocket.Conn) error {
	if ws == nil {
		return fmt.Errorf("invalid websocket")
	}
	this.ws = ws
	return nil
}

func (this *RpcSession[id]) KeepAlive(ttl time.Duration) {
	this.closedAt = time.Now().Add(ttl)
}

func (this *RpcSession[id]) IsAlive(now time.Time) bool {
	if this.closedAt.IsZero() {
		return true
	}
	return now.Before(this.closedAt)
}

func (this *RpcSession[id]) Start(opts Options) error {
	if this.ws == nil {
		return fmt.Errorf("no ws connection")
	}
	this.OnEvent(Events.Start)
	for {
		msgType, msgBody, err := this.ws.ReadMessage()
		if err != nil {
			this.OnEvent(Events.Stop)
			return err
		}
		// log.Printf("msgType:%d msgBody:%v\n", msgType, msgBody)
		switch msgType {
		case websocket.BinaryMessage, websocket.TextMessage:
			this.onMessage(msgBody)
		case websocket.PingMessage:
			this.write(websocket.PongMessage, nil)
		case websocket.PongMessage:
			this.KeepAlive(opts.GetHeartBeatDrt())
		case websocket.CloseMessage:
			this.ws.Close()
			return nil
		}
	}
	// return nil
}

func (this *RpcSession[id]) write(msgType int, data []byte) error {
	this.mux.Lock()
	defer this.mux.Unlock()
	return this.ws.WriteMessage(msgType, data)
}

func (this *RpcSession[id]) onMessage(msgBody []byte) {
	var callFlag = msgBody[0]
	var isCall = (callFlag >> 7) == 1
	var callId = (callFlag & 0x7f)

	var msgId uint16
	var msgData []byte
	var callRs []byte = nil
	if callId > 0 {
		if isCall {
			// on-call
			msgId = binary.LittleEndian.Uint16(msgBody[1:3])
			msgData = msgBody[3:]
			defer func() {
				// log.Printf("return msg callSeqId:%d. callFlag:%d data:%v\n", callSeqId, callFlag, callRs)
				this.ReturnMsg(callId, callRs)
			}()

		} else {
			// on-callback
			msgData = msgBody[1:]
			callback := this.callbacks[callId]
			if callback == nil {
				log.Printf("Invalid callSeqId: %d. callFlag: %d\n", callId, callFlag)
				return
			}
			callback(msgData)
			delete(this.callbacks, callId)
			return
		}
	} else {
		// on-send
		msgId = binary.LittleEndian.Uint16(msgBody[1:3])
		msgData = msgBody[3:]
	}
	code, respData := this.handler(this, msgId, msgData)
	callRs = append([]byte{0, 0}, respData...)
	binary.LittleEndian.PutUint16(callRs, code)
}

func (this *RpcSession[id]) Call(msgId uint16, data []byte, callback func([]byte)) {
	// log.Printf("call %d\n", msgId)
	this.callSeqNum++
	callSeqId := this.callSeqNum

	var callFlag uint8 = (1 << 7) + (callSeqId & 0x7f)
	var head = []byte{callFlag, 0, 0}
	binary.LittleEndian.PutUint16(head[1:], msgId)
	var msgBody = append(head, data...)
	this.callbacks[this.callSeqNum] = callback
	if this.ws == nil {
		log.Printf("websocket was nil\n")
		return
	}

	this.write(websocket.BinaryMessage, msgBody)
}

func (this *RpcSession[id]) SendMsg(msgId uint16, data []byte) {
	// log.Printf("send %d\n", msgId)
	var head = []byte{0, 0, 0}
	binary.LittleEndian.PutUint16(head[1:], msgId)

	var msgBody = append(head, data...)
	if this.ws == nil {
		log.Printf("websocket was nil\n")
		return
	}
	this.write(websocket.BinaryMessage, msgBody)
}

func (this *RpcSession[id]) ReturnMsg(callSeqId uint8, data []byte) {
	// log.Printf("return call %d\n", callSeqId)
	// isReq + callSeqId
	var callFlag uint8 = (0 << 7) + (callSeqId & 0x7f)
	data = append([]byte{callFlag}, data...)
	if this.ws == nil {
		log.Printf("websocket was nil\n")
		return
	}
	err := this.write(websocket.BinaryMessage, data)
	if err != nil {
		log.Printf("write error %v", err)
	}
}

func (this *RpcSession[id]) GetType() string {
	return "player"
}

func (this *RpcSession[id]) GetID() id {
	return this.id
}

func (this *RpcSession[id]) GetToken() string {
	return this.token
}

func (this *RpcSession[id]) GetPlayer() interface{} {
	panic(fmt.Errorf("not implement"))
}

func (this *RpcSession[id]) OnEvent(e Event) error {
	switch e {
	case Events.Start:
	case Events.Stop:
	default:
	}
	return nil
}

func (this *RpcSession[id]) Close() {
	this.mux.Lock()
	defer this.mux.Unlock()
	if this.ws != nil {
		err := this.ws.WriteControl(websocket.CloseMessage, nil, time.Now().Add(10*time.Second))
		if err != nil {
			log.Printf("websocket close failed err=%v", err)
		}
		this.ws.Close()
	}
}
