package yorpc

import (
	"fmt"
)

type Handler func([]byte) ([]byte, error)
type HandlersHub struct {
	handlers map[uint16]Handler
}

func NewHandlersHub() *HandlersHub {
	return &HandlersHub{
		handlers: map[uint16]Handler{},
	}
}

func (h *HandlersHub) RegisterRPC(id uint16, handler Handler) {
	if old, ok := h.handlers[id]; ok {
		panic(fmt.Errorf("already registerted. old=%v handler=%v", GetFuncName(old), GetFuncName(handler)))
	}
	h.handlers[id] = handler
}

func (h *HandlersHub) RegisterRPCAsync(id uint16, handler func([]byte) error) {
	if old, ok := h.handlers[id]; ok {
		panic(fmt.Errorf("already registerted. old=%v handler=%v", GetFuncName(old), GetFuncName(handler)))
	}
	h.handlers[id] = func(args []byte) ([]byte, error) {
		return nil, handler(args)
	}
}

func (h *HandlersHub) GetHandler(id uint16) (Handler, error) {
	if handler, ok := h.handlers[id]; ok {
		return handler, nil
	}
	return nil, fmt.Errorf("no handler(id=%d)", id)
}

func (h *HandlersHub) OnCall(id uint16, args []byte) ([]byte, error) {
	// log.Printf("server.OnCall id=%d args=%v", id, args)
	handler, err := h.GetHandler(id)
	if err != nil {
		return nil, err
	}
	return handler(args)
}

func (h *HandlersHub) OnSend(id uint16, args []byte) error {
	// log.Printf("server.OnSend id=%d args=%v", id, args)
	handler, err := h.GetHandler(id)
	if err != nil {
		return err
	}
	_, err = handler(args)
	return err
}
