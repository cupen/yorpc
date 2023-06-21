package handlerhub

import (
	"fmt"
)

type Handler func([]byte) ([]byte, error)
type Hub struct {
	handlers map[uint16]Handler
}

func New() *Hub {
	return &Hub{
		handlers: map[uint16]Handler{},
	}
}

func (h *Hub) Register(id uint16, handler Handler) {
	if old, ok := h.handlers[id]; ok {
		panic(fmt.Errorf("already registerted. old=%v handler=%v", GetFuncName(old), GetFuncName(handler)))
	}
	h.handlers[id] = handler
}

func (h *Hub) Register2(id uint16, handler func([]byte)) {
	if old, ok := h.handlers[id]; ok {
		panic(fmt.Errorf("already registerted. old=%v handler=%v", GetFuncName(old), GetFuncName(handler)))
	}
	h.handlers[id] = func(args []byte) ([]byte, error) {
		handler(args)
		return nil, nil
	}
}

func (h *Hub) GetHandler(id uint16) (Handler, error) {
	if handler, ok := h.handlers[id]; ok {
		return handler, nil
	}
	return nil, fmt.Errorf("no handler(id=%d)", id)
}

func (h *Hub) OnCall(id uint16, args []byte) ([]byte, error) {
	handler, err := h.GetHandler(id)
	if err != nil {
		return nil, err
	}
	return handler(args)
}

func (h *Hub) OnSend(id uint16, args []byte) {
	handler, err := h.GetHandler(id)
	if err != nil {
		h.onError(id, args, err)
		return
	}
	_, err = handler(args)
	h.onError(id, args, err)
}

func (h *Hub) onError(id uint16, args []byte, err error) {
}
