package codecv1

import (
	"encoding/binary"
)

type CodecV1 struct {
	key []byte
}

func NewV1(key []byte) *CodecV1 {
	return &CodecV1{
		key: key,
	}
}

func (cdc *CodecV1) EncodeSend(msgId uint16, data []byte) []byte {
	var headers = []byte{0, 0, 0}
	binary.LittleEndian.PutUint16(headers[1:], msgId)
	return append(headers, data...)
}

func (cdc *CodecV1) EncodeCall(callId uint8, msgId uint16, data []byte) []byte {
	var callFlag uint8 = (1 << 7) + (callId & 0x7f)
	var headers = []byte{callFlag, 0, 0}
	binary.LittleEndian.PutUint16(headers[1:], msgId)
	return append(headers, data...)
}

func (cdc *CodecV1) EncodeReturn(callId uint8, data []byte) []byte {
	var callFlag uint8 = (callId & 0x7f)
	return append([]byte{callFlag}, data...)
}

func (cdc *CodecV1) DecodeCallFlag(callFlag byte) (isCall bool, callId uint8) {
	isCall = ((callFlag >> 7) == 1)
	callId = (callFlag & 0x7f)
	return
}

// func (cdc *CodecV1) EncodeCallFlag(callFlag byte) (isCall bool, callId uint8) {
// 	isCall = ((callFlag >> 7) == 1)
// 	callId = (callFlag & 0x7f)
// 	return
// }
