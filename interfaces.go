package yorpc

import "context"

type IServer interface {
	OnCall(uint16, []byte) ([]byte, error)
	OnSend(uint16, []byte)
}

type IClient interface {
	ICaller
	ISender
}

type ICaller interface {
	Call(context.Context, uint16, []byte) ([]byte, error)
}
type ISender interface {
	Send(uint16, []byte)
}
