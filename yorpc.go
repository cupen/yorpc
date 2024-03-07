package yorpc

type Client interface {
	Call(uint16, []byte) (uint16, []byte)
	Send(uint16, []byte)
}

type Server interface {
	OnCall(uint16, []byte) (uint16, []byte)
	OnSend(uint16, []byte)
}
