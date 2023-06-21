package yorpc

import (
	"time"
)

type _id interface {
	string | int | int64 | uint | uint64
}

type _data[Key _id] interface {
	GetID() Key
}

type Session[Key _id, D _data[Key]] interface {
	GetID() Key
	GetData() D
	OnEvent(Event) error
	KeepAlive(time.Duration)
	IsAlive(now time.Time) bool
}
