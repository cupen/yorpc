package yorpc

type ClientSession interface {
	// Send a message.
	Send(id uint16, args []byte) error

	// Send a message and receive a result.
	Call(id uint16, args []byte) ([]byte, error)

	// forward a message to call.
	OnCallback(id uint8, args []byte) error
}

type ServerSession interface {
	// Receive a message
	OnSend(id uint16, args []byte) error

	// Request-Response
	OnCall(id uint16, args []byte) ([]byte, error)
}
