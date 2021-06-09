package yorpc

type Connection interface {
	// OnMessage([]byte) error
	// WriteMessage([]byte) error
	// SetSession(ClientSession, ServerSession)

	// Write a message
	WriteMessage([]byte) error

	// Blocking until a message received
	ReceiveMessage() ([]byte, error)

	// Close ...
	Close() error
}
