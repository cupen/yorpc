package yorpc

import (
	"encoding/binary"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Event int

var Events = struct {
	Starting      Event
	Stopping      Event
	KeepAliveTick Event
}{
	Starting:      1,
	Stopping:      2,
	KeepAliveTick: 3,
}

type Session interface {
	GetType() string
	GetID() string
	GetPlayer() interface{}
	GetToken() string
	OnEvent(Event) error
	SendMsg(uint16, []byte)
	// Call(uint16, []byte, func([]byte))
	ReturnMsg(uint8, []byte)
	KeepAlive(time.Duration)
	IsAlive(now time.Time) bool
	Stop()
}

type RpcSession struct {
	id         string
	token      string
	ws         *websocket.Conn
	Handlers   map[uint16]MsgHandler
	Callbacks  map[uint8]func([]byte)
	callSeqNum uint8
	closedAt   time.Time
	mux        sync.Mutex
}

func NewRpcSession(id, token string, handlers map[uint16]MsgHandler) *RpcSession {
	return &RpcSession{
		id:        id,
		token:     token,
		Handlers:  handlers,
		Callbacks: map[uint8]func([]byte){},
	}
}

func (this *RpcSession) Connect(ws *WebSocket) error {
	return this.Connect2(ws.Conn)
}

func (this *RpcSession) Connect2(ws *websocket.Conn) error {
	if ws == nil {
		return fmt.Errorf("Invalid websocket")
	}
	this.ws = ws
	return nil
}

func (this *RpcSession) KeepAlive(ttl time.Duration) {
	this.closedAt = time.Now().Add(ttl)
}

func (this *RpcSession) IsAlive(now time.Time) bool {
	if this.closedAt.IsZero() {
		return true
	}
	return now.Before(this.closedAt)
}

func (this *RpcSession) Start(opts Options) error {
	if this.ws == nil {
		return fmt.Errorf("Invalid ws")
	}
	this.OnEvent(Events.Starting)
	for {
		msgType, msgBody, err := this.ws.ReadMessage()
		if err != nil {
			this.OnEvent(Events.Stopping)
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
	return nil
}

func (this *RpcSession) write(msgType int, data []byte) error {
	this.mux.Lock()
	defer this.mux.Unlock()
	return this.ws.WriteMessage(msgType, data)
}

func (this *RpcSession) onMessage(msgBody []byte) {
	callFlag := msgBody[0]
	isReq := (callFlag >> 7) == 1
	callSeqId := (callFlag & 0x7f)

	// callback
	var msgId uint16
	var msgData []byte
	var callRs []byte = nil
	if callSeqId > 0 {
		if isReq {
			// byte-2~3
			msgId = binary.LittleEndian.Uint16(msgBody[1:3])
			msgData = msgBody[3:]
			defer func() {
				// log.Printf("return msg callSeqId:%d. callFlag:%d data:%v\n", callSeqId, callFlag, callRs)
				this.ReturnMsg(callSeqId, callRs)
			}()

		} else {
			msgData = msgBody[1:]
			callback, _ := this.Callbacks[callSeqId]
			if callback == nil {
				log.Printf("Invalid callSeqId: %d. callFlag: %d\n", callSeqId, callFlag)
				return
			}
			callback(msgData)
			this.Callbacks[callSeqId] = nil
			return
		}
	} else {
		// byte-2~3
		msgId = binary.LittleEndian.Uint16(msgBody[1:3])
		msgData = msgBody[3:]
	}
	handler, _ := this.Handlers[msgId]
	if handler == nil {
		log.Printf("Invalid msgId: %d. callFlag: %d\n", msgId, callFlag)
		return
	}
	callRs = handler(this, msgData)
}

func (this *RpcSession) Call(msgId uint16, data []byte, callback func([]byte)) {
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
	if this.ws == nil {
		log.Printf("websocket was nil\n")
		return
	}

	this.write(websocket.BinaryMessage, msgBody)
}

func (this *RpcSession) SendMsg(msgId uint16, data []byte) {
	// log.Printf("send %d\n", msgId)
	msgIdBytes := []byte{0, 0}
	binary.LittleEndian.PutUint16(msgIdBytes, msgId)

	var callFlag uint8 = 0
	msgBody := append([]byte{callFlag}, msgIdBytes...)
	msgBody = append(msgBody, data...)
	if this.ws == nil {
		log.Printf("websocket was nil\n")
		return
	}
	this.write(websocket.BinaryMessage, msgBody)
}

func (this *RpcSession) ReturnMsg(callSeqId uint8, data []byte) {
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

func (this *RpcSession) GetType() string {
	return "player"
}

func (this *RpcSession) GetID() string {
	return this.id
}

func (this *RpcSession) GetToken() string {
	return this.token
}

func (this *RpcSession) GetPlayer() interface{} {
	// panic(fmt.Errorf("Not implement"))
	return nil
}

func (this *RpcSession) OnEvent(e Event) error {
	switch e {
	case Events.Starting:
	case Events.Stopping:
	default:
	}
	return nil
}

func (this *RpcSession) Stop() {
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
