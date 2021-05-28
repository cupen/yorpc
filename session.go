package yorpc

import (
	"time"
)

type Session interface {
	GetType() string
	GetID() string
	GetPlayer() interface{}
	GetToken() string
	OnEvent(Event) error
	SendMsg(uint16, []byte)
	// Call(uint16, []byte, func([]byte))
	ReturnMsg(uint8, []byte)
	KeepAlive(time.Duration)
	IsAlive(now time.Time) bool
	Stop()
}
