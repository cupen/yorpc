package yorpc

import "fmt"

var (
	ErrNoWsConn       = fmt.Errorf("no websocket connection")
	ErrMissingConn    = newError(1, "missing connection")
	ErrNotImplemented = newError(2, "not implemented")
	defaultErrors     = initErrors()
)

func init() {
	initErrors()
}

type Error struct {
	Code uint16
	Msg  string
}

func newError(code uint16, msg string) *Error {
	return &Error{
		Code: code,
		Msg:  msg,
	}
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	return e.Msg
}

func initErrors() (errs [128]*Error) {
	for i := 0; i < 128; i++ {
		errs[i] = &Error{
			Code: uint16(i),
		}
	}
	for _, err := range []*Error{
		ErrMissingConn,
		ErrNotImplemented,
	} {
		errs[err.Code] = err
	}
	return
}
