package yorpc

type Conn interface {
	Start() error
	Stop()
}
