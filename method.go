package yorpc

type MsgHandler func(session Session, msgId uint16, msgBody []byte) (uint16, []byte)
