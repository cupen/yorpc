package yorpc

type Handler interface {
	OnCall(protoId uint16, msg []byte) []byte
	OnEvent(EventType)
}

type Caller interface {
	Call(protoId uint16, msg []byte) []byte
	OnEvent(EventType)
}
