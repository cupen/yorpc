package yorpc

type CodecV1 struct {
}

func (this *CodecV1) EncodeSend(callSeqId uint8, data []byte) []byte {
	var callFlag uint8 = (0 << 7) + (callSeqId & 0x7f)
	return append([]byte{callFlag}, data...)
}

func (this *CodecV1) EncodeCall(callSeqId uint8, data []byte) []byte {
	var callFlag uint8 = (0 << 7) + (callSeqId & 0x7f)
	return append([]byte{callFlag}, data...)
}

func (this *CodecV1) EncodeReturn(callSeqId uint8, data []byte) []byte {
	var callFlag uint8 = (0 << 7) + (callSeqId & 0x7f)
	return append([]byte{callFlag}, data...)
}

func (this *CodecV1) Decode(data []byte) {
	isReq := (callFlag >> 7) == 1
	callSeqId := (callFlag & 0x7f)

}
