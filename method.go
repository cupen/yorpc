package yorpc

type MethodID uint16

type MsgHandler func(session Session, callSeqNum uint8, msgBody []byte)

type Method struct {
	ID      uint16
	Handler MsgHandler
}
