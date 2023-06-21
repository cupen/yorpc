package yorpc

type Conn interface {
	// Write a message
	WriteMessage([]byte) error

	// Blocking until a message received
	ReadMessage() ([]byte, error)

	// Close ...
	Close() error
}
