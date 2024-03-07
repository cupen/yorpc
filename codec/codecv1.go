package codec

import (
	"encoding/binary"
	"fmt"

	"github.com/gorilla/websocket"
)

var (
	ErrCodecV1InvalidMsgType = fmt.Errorf("invalid message type")

	_ Codec = &CodecV1{}
)

type CodecV1 struct {
	conn *websocket.Conn
	key  []byte
}

func NewV1(conn *websocket.Conn, key []byte) *CodecV1 {
	return &CodecV1{
		conn: conn,
		key:  key,
	}
}

func (cdc *CodecV1) ReadPacket() (bool, uint8, uint16, []byte, error) {
	msgType, packet, err := cdc.conn.ReadMessage()
	if err != nil {
		return false, 0, 0, nil, err
	}
	for msgType != websocket.BinaryMessage && msgType != websocket.TextMessage {
		msgType, packet, err = cdc.conn.ReadMessage()
		if err != nil {
			return false, 0, 0, nil, err
		}
	}
	var callFlag = packet[0]
	var isCall = (callFlag >> 7) == 1
	var callId = (callFlag & 0x7f)
	if callId > 0 {
		if isCall {
			// on-call
			msgId := binary.LittleEndian.Uint16(packet[1:3])
			msgBody := packet[3:]
			return isCall, callId, msgId, msgBody, nil

		} else {
			// on-callback
			msgBody := packet[1:]
			return false, callId, 0, msgBody, nil
		}
	} else {
		// on-send
		msgId := binary.LittleEndian.Uint16(packet[1:3])
		msgBody := packet[3:]
		return false, 0, msgId, msgBody, nil
	}
}

func (cdc *CodecV1) WritePacket(isCall bool, callId uint8, msgId uint16, body []byte) error {
	var callFlag uint8 = callId & 0x7f
	if isCall {
		callFlag |= (1 << 7)
	}
	var head = []byte{callFlag, 0, 0}
	binary.LittleEndian.PutUint16(head[1:], msgId)
	data := append(head, body...)
	cdc.patchData(data)
	return cdc.conn.WriteMessage(websocket.BinaryMessage, data)
}

func (cdc *CodecV1) ReturnMsg(callId uint8, body []byte) []byte {
	var callFlag uint8 = (0 << 7) + (callId & 0x7f)
	data := append([]byte{callFlag}, body...)
	cdc.patchData(data)
	return data
}

func (cdc *CodecV1) patchData(data []byte) {
	if cdc.key != nil {
		xor(data, cdc.key)
	}
}
