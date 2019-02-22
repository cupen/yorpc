package yorpc

import (
	"encoding/binary"
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

type Session interface {
	GetID() string
	GetPlayer() interface{}
	SendMsg(uint16, []byte)
	ReturnMsg(uint8, []byte)
}

type RpcSession struct {
	id        string
	ws        *websocket.Conn
	Handlers  map[uint16]MsgHandler
	CallBakcs map[uint8]func([]byte)

	msgQueue   chan []byte
	msgQueue2  chan []byte
	callSeqNum uint8
}

func NewRpcSession(id string, handlers map[uint16]MsgHandler) *RpcSession {
	return &RpcSession{
		id:        id,
		Handlers:  handlers,
		CallBakcs: map[uint8]func([]byte){},
	}
}

func (this *RpcSession) StartWithUrl(url string) error {
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return err
	}
	return this.Start(conn)
}

func (this *RpcSession) EnableHeartBeat(intervDrt time.Duration) {
	go func() {
		t := time.NewTicker(intervDrt)
		for {
			<-t.C
			if this.ws == nil {
				continue
			}
			err := this.ws.WriteMessage(websocket.PingMessage, nil)
			if err != nil {
				this.ws.Close()
				break
			}
		}
	}()
}

// 开始接收并处理消息
// TODO: 用 epoll 取代 goroutine
func (this *RpcSession) Start(ws *websocket.Conn) error {
	this.ws = ws
	for {
		msgType, msgBody, err := this.ws.ReadMessage()
		if err != nil {
			break
		}
		log.Printf("msgType:%d msgBody:%v\n", msgType, msgBody)
		switch msgType {
		case websocket.BinaryMessage, websocket.TextMessage:
			this.onMessage(msgBody)
		case websocket.PingMessage:
			log.Printf("ping.\n")
		case websocket.CloseMessage:
			log.Printf("close.\n")
			break
		}
	}
	return nil
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
				log.Printf("return msg callSeqId:%d. callFlag:%d data:%v\n", callSeqId, callFlag, callRs)
				this.ReturnMsg(callSeqId, callRs)
			}()

		} else {
			msgData = msgBody[1:]
			callback, _ := this.CallBakcs[callSeqId]
			if callback == nil {
				log.Printf("Invalid callSeqId: %d. callFlag: %d\n", callSeqId, callFlag)
				return
			}
			callback(msgData)
			this.CallBakcs[callSeqId] = nil
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

func (this *RpcSession) Stop() {
}

func (this *RpcSession) Call(msgId uint16, data []byte, callback func([]byte)) {
	log.Printf("call %d\n", msgId)
	this.callSeqNum++
	callSeqId := this.callSeqNum

	var callFlag uint8 = (1 << 7) + (callSeqId & 0x7f)

	msgIdBytes := []byte{0, 0}
	binary.LittleEndian.PutUint16(msgIdBytes, msgId)
	msgBody := []byte{callFlag}
	msgBody = append(msgBody, msgIdBytes...)
	msgBody = append(msgBody, data...)
	this.CallBakcs[this.callSeqNum] = callback
	this.ws.WriteMessage(2, msgBody)
}

func (this *RpcSession) SendMsg(msgId uint16, data []byte) {
	log.Printf("send %d\n", msgId)
	msgIdBytes := []byte{0, 0}
	binary.LittleEndian.PutUint16(msgIdBytes, msgId)

	var callFlag uint8 = 0
	msgBody := append([]byte{callFlag}, msgIdBytes...)
	msgBody = append(msgBody, data...)
	this.ws.WriteMessage(websocket.BinaryMessage, msgBody)
}

func (this *RpcSession) ReturnMsg(callSeqId uint8, data []byte) {
	log.Printf("return call %d\n", callSeqId)
	// isReq + callSeqId
	var callFlag uint8 = (0 << 7) + (callSeqId & 0x7f)
	data = append([]byte{callFlag}, data...)
	this.ws.WriteMessage(websocket.BinaryMessage, data)
}

func (this *RpcSession) GetPlayer() interface{} {
	panic(fmt.Errorf("Not implement"))
}

func (this *RpcSession) GetID() string {
	return this.id
}
