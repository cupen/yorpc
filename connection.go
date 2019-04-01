package yorpc

type Connection interface {
	Start(*Options) error
	Stop()
}
