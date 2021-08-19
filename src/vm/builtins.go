package vm

import (
	"fmt"
	"github.com/iafisher/venice/src/vval"
	"strings"
)

func builtinLength(args ...vval.VeniceValue) vval.VeniceValue {
	if len(args) != 1 {
		return nil
	}

	var n int
	switch arg := args[0].(type) {
	case *vval.VeniceList:
		n = len(arg.Values)
	case *vval.VeniceMap:
		n = len(arg.Pairs)
	case *vval.VeniceString:
		n = len(arg.Value)
	default:
		return nil
	}
	return &vval.VeniceInteger{n}
}

func builtinPrint(args ...vval.VeniceValue) vval.VeniceValue {
	if len(args) != 1 {
		return nil
	}

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

func builtinStringFind(args ...vval.VeniceValue) vval.VeniceValue {
	if len(args) != 2 {
		return nil
	}

	stringArg, ok1 := args[0].(*vval.VeniceString)
	charArg, ok2 := args[1].(*vval.VeniceCharacter)
	if !ok1 || !ok2 {
		return nil
	}

	index := strings.IndexByte(stringArg.Value, charArg.Value)
	if index == -1 {
		return vval.VENICE_OPTIONAL_NONE
	} else {
		return vval.VeniceOptionalOf(&vval.VeniceInteger{index})
	}
}

func builtinStringToUpper(args ...vval.VeniceValue) vval.VeniceValue {
	if len(args) != 1 {
		return nil
	}

	stringArg, ok := args[0].(*vval.VeniceString)
	if !ok {
		return nil
	}

	return &vval.VeniceString{strings.ToUpper(stringArg.Value)}
}

func builtinStringToLower(args ...vval.VeniceValue) vval.VeniceValue {
	if len(args) != 1 {
		return nil
	}

	stringArg, ok := args[0].(*vval.VeniceString)
	if !ok {
		return nil
	}

	return &vval.VeniceString{strings.ToLower(stringArg.Value)}
}

// If a method is added here, make sure to also add it to the appropriate place in
// compiler/compiler.go - `NewBuiltinSymbolTable` if it is a global built-in,
// `stringBuiltins` if it is a string built-in, etc.
var builtins = map[string]func(args ...vval.VeniceValue) vval.VeniceValue{
	// Global built-ins
	"length": builtinLength,
	"print":  builtinPrint,
	// List built-ins
	"list__length": builtinLength,
	// String built-ins
	"string__find":     builtinStringFind,
	"string__length":   builtinLength,
	"string__to_lower": builtinStringToLower,
	"string__to_upper": builtinStringToUpper,
}
