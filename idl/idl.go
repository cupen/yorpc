package idl

import "regexp"

type NodeMethod struct {
	ProtoID uint16
	Name    string
	Args    []NodeArg
}

type NodeArg struct {
	Name string
	Type string
}

// EXAMPLE:
// rpc Player {
//    101 info(id: int) Player
//    102 useItem(itemId: int, count: int) (int, int)
// }
// TODO: yacc
func Parse(text string) map[uint16]NodeMethod {
	sytaxHandler := regexp.MustCompile("rpc +(\\{[.\r\n]*\\})")
	sytaxMethod := regexp.MustCompile("")
	_ = sytaxHandler
	_ = sytaxMethod
	return nil
}

func GenerateCodeGo() string {
	return ""
}

func GenerateCodeTypeScript() string {
	return ""
}
