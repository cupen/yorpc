package yorpc

type Event int

var Events = struct {
	Starting      Event
	Stopping      Event
	KeepAliveTick Event
}{
	Starting:      1,
	Stopping:      2,
	KeepAliveTick: 3,
}
