package yorpc

import (
	"reflect"
	"runtime"
)

func GetFuncName(f interface{}) string {
	if f == nil {
		return ""
	}
	v := reflect.ValueOf(f)
	if v.IsNil() {
		return "nil"
	}
	return runtime.FuncForPC(v.Pointer()).Name()
}
