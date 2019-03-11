package yorpc

type EventType uint8

// enum
var Events = struct {
	Ping      EventType
	Pong      EventType
	Connected EventType
	ReConnect EventType
	Close     EventType
}{
	Ping:      1,
	Pong:      2,
	Connected: 3,
	ReConnect: 4,
	Close:     5,
}
