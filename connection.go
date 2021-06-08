package yorpc

type Connection interface {
	OnMessage([]byte) error
	WriteMessage([]byte) error
	SetSession(ClientSession, ServerSession)
}
