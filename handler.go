package yorpc

type MsgHandler[id ID] func(s Session[id], msgId uint16, msgBody []byte) (uint16, []byte)
