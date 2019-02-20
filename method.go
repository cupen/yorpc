package yorpc

type MethodID uint16

type MsgHandler func(session Session, msgBody []byte) []byte

type Method struct {
	ID      uint16
	Handler MsgHandler
}
