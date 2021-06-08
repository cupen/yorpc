package yorpc

type ClientSession interface {
	Call(id uint16, args []byte) ([]byte, error)
	Send(id uint16, args []byte) error
}

type ServerSession interface {
	OnCall(id uint16, args []byte) ([]byte, error)
	OnSend(id uint16, args []byte) error
}
