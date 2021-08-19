package vm

import (
	"fmt"
	"github.com/iafisher/venice/src/vval"
)

func builtinLength(args ...vval.VeniceValue) vval.VeniceValue {
	var n int
	switch arg := args[0].(type) {
	case *vval.VeniceList:
		n = len(arg.Values)
	case *vval.VeniceMap:
		n = len(arg.Pairs)
	case *vval.VeniceString:
		n = len(arg.Value)
	default:
		n = -1
	}
	return &vval.VeniceInteger{n}
}

func builtinPrint(args ...vval.VeniceValue) vval.VeniceValue {
	switch arg := args[0].(type) {
	case *vval.VeniceCharacter:
		fmt.Println(string(arg.Value))
	case *vval.VeniceString:
		fmt.Println(arg.Value)
	default:
		fmt.Println(args[0].String())
	}
	return nil
}

var builtins = map[string]func(args ...vval.VeniceValue) vval.VeniceValue{
	"length":       builtinLength,
	"list__length": builtinLength,
	"print":        builtinPrint,
}
