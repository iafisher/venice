package vm

import (
	"fmt"
	"github.com/iafisher/venice/src/vval"
	"strconv"
	"strings"
)

/**
 * Global built-in functions
 */

func builtinInt(args ...vval.VeniceValue) vval.VeniceValue {
	if len(args) != 1 {
		return nil
	}

	realArg, ok := args[0].(*vval.VeniceRealNumber)
	if !ok {
		return nil
	}

	return &vval.VeniceInteger{int(realArg.Value)}
}

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

func builtinRange(args ...vval.VeniceValue) vval.VeniceValue {
	if len(args) != 2 {
		return nil
	}

	intArg1, ok1 := args[0].(*vval.VeniceInteger)
	intArg2, ok2 := args[1].(*vval.VeniceInteger)
	if !ok1 || !ok2 {
		return nil
	}

	length := intArg2.Value - intArg1.Value
	if length <= 0 {
		return &vval.VeniceList{[]vval.VeniceValue{}}
	}

	numbers := make([]vval.VeniceValue, 0, intArg2.Value-intArg1.Value)
	for i := intArg1.Value; i < intArg2.Value; i++ {
		numbers = append(numbers, &vval.VeniceInteger{i})
	}
	return &vval.VeniceList{numbers}
}

func builtinReal(args ...vval.VeniceValue) vval.VeniceValue {
	if len(args) != 1 {
		return nil
	}

	intArg, ok := args[0].(*vval.VeniceInteger)
	if !ok {
		return nil
	}

	return &vval.VeniceRealNumber{float64(intArg.Value)}
}

func builtinString(args ...vval.VeniceValue) vval.VeniceValue {
	if len(args) != 1 {
		return nil
	}

	switch arg := args[0].(type) {
	case *vval.VeniceBoolean:
		if arg.Value {
			return &vval.VeniceString{"true"}
		} else {
			return &vval.VeniceString{"false"}
		}
	case *vval.VeniceCharacter:
		return &vval.VeniceString{string(arg.Value)}
	case *vval.VeniceInteger:
		return &vval.VeniceString{strconv.Itoa(arg.Value)}
	case *vval.VeniceString:
		return arg
	default:
		return nil
	}
}

/**
 * List built-ins
 */

func builtinListAppend(args ...vval.VeniceValue) vval.VeniceValue {
	if len(args) != 2 {
		return nil
	}

	listArg, ok := args[0].(*vval.VeniceList)
	if !ok {
		return nil
	}

	listArg.Values = append(listArg.Values, args[1])
	return nil
}

func builtinListExtend(args ...vval.VeniceValue) vval.VeniceValue {
	if len(args) != 2 {
		return nil
	}

	listArg1, ok1 := args[0].(*vval.VeniceList)
	listArg2, ok2 := args[1].(*vval.VeniceList)
	if !ok1 || !ok2 {
		return nil
	}

	listArg1.Values = append(listArg1.Values, listArg2.Values...)
	return nil
}

func builtinListRemove(args ...vval.VeniceValue) vval.VeniceValue {
	if len(args) != 2 {
		return nil
	}

	listArg, ok1 := args[0].(*vval.VeniceList)
	intArg, ok2 := args[1].(*vval.VeniceInteger)
	if !ok1 || !ok2 {
		return nil
	}

	index := intArg.Value
	if index < 0 || index >= len(listArg.Values) {
		// TODO(2021-08-23): Handle this better.
		panic("index out of bounds for list.remove")
	}

	listArg.Values = append(listArg.Values[:index], listArg.Values[index+1:]...)
	return nil
}

func builtinListSlice(args ...vval.VeniceValue) vval.VeniceValue {
	if len(args) != 3 {
		return nil
	}

	listArg, ok1 := args[0].(*vval.VeniceList)
	startIndexArg, ok2 := args[1].(*vval.VeniceInteger)
	endIndexArg, ok3 := args[2].(*vval.VeniceInteger)
	if !ok1 || !ok2 || !ok3 {
		return nil
	}

	// TODO(2021-08-25): Handle out-of-bounds error.
	return &vval.VeniceList{listArg.Values[startIndexArg.Value:endIndexArg.Value]}
}

/**
 * Map built-ins
 */

func builtinMapEntries(args ...vval.VeniceValue) vval.VeniceValue {
	if len(args) != 1 {
		return nil
	}

	mapArg, ok := args[0].(*vval.VeniceMap)
	if !ok {
		return nil
	}

	return mapArg.Entries()
}

func builtinMapKeys(args ...vval.VeniceValue) vval.VeniceValue {
	if len(args) != 1 {
		return nil
	}

	mapArg, ok := args[0].(*vval.VeniceMap)
	if !ok {
		return nil
	}

	return mapArg.Keys()
}

func builtinMapRemove(args ...vval.VeniceValue) vval.VeniceValue {
	if len(args) != 2 {
		return nil
	}

	mapArg, ok := args[0].(*vval.VeniceMap)
	if !ok {
		return nil
	}

	mapArg.Remove(args[1])
	return nil
}

func builtinMapValues(args ...vval.VeniceValue) vval.VeniceValue {
	if len(args) != 1 {
		return nil
	}

	mapArg, ok := args[0].(*vval.VeniceMap)
	if !ok {
		return nil
	}

	return mapArg.Values()
}

/**
 * String built-ins
 */

func builtinStringFind(args ...vval.VeniceValue) vval.VeniceValue {
	if len(args) != 2 {
		return nil
	}

	stringArg, ok1 := args[0].(*vval.VeniceString)
	searchArg, ok2 := args[1].(*vval.VeniceString)
	if !ok1 || !ok2 {
		return nil
	}

	index := strings.Index(stringArg.Value, searchArg.Value)
	if index == -1 {
		return vval.VENICE_OPTIONAL_NONE
	} else {
		return vval.VeniceOptionalOf(&vval.VeniceInteger{index})
	}
}

func builtinStringFindLast(args ...vval.VeniceValue) vval.VeniceValue {
	if len(args) != 2 {
		return nil
	}

	stringArg, ok1 := args[0].(*vval.VeniceString)
	searchArg, ok2 := args[1].(*vval.VeniceString)
	if !ok1 || !ok2 {
		return nil
	}

	index := strings.LastIndex(stringArg.Value, searchArg.Value)
	if index == -1 {
		return vval.VENICE_OPTIONAL_NONE
	} else {
		return vval.VeniceOptionalOf(&vval.VeniceInteger{index})
	}
}

func builtinStringSlice(args ...vval.VeniceValue) vval.VeniceValue {
	if len(args) != 3 {
		return nil
	}

	stringArg, ok1 := args[0].(*vval.VeniceString)
	startIndexArg, ok2 := args[1].(*vval.VeniceInteger)
	endIndexArg, ok3 := args[2].(*vval.VeniceInteger)
	if !ok1 || !ok2 || !ok3 {
		return nil
	}

	// TODO(2021-08-25): Handle out-of-bounds error.
	return &vval.VeniceString{stringArg.Value[startIndexArg.Value:endIndexArg.Value]}
}

func builtinStringSplit(args ...vval.VeniceValue) vval.VeniceValue {
	if len(args) != 2 {
		return nil
	}

	stringArg, ok1 := args[0].(*vval.VeniceString)
	splitterArg, ok2 := args[1].(*vval.VeniceString)
	if !ok1 || !ok2 {
		return nil
	}

	list := &vval.VeniceList{}
	for _, word := range strings.Split(stringArg.Value, splitterArg.Value) {
		list.Values = append(list.Values, &vval.VeniceString{word})
	}
	return list
}

func builtinStringSplitSpace(args ...vval.VeniceValue) vval.VeniceValue {
	if len(args) != 1 {
		return nil
	}

	stringArg, ok := args[0].(*vval.VeniceString)
	if !ok {
		return nil
	}

	list := &vval.VeniceList{}
	for _, word := range strings.Fields(stringArg.Value) {
		list.Values = append(list.Values, &vval.VeniceString{word})
	}
	return list
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
	"int":    builtinInt,
	"length": builtinLength,
	"print":  builtinPrint,
	"range":  builtinRange,
	"real":   builtinReal,
	"string": builtinString,
	// List built-ins
	"list__append": builtinListAppend,
	"list__extend": builtinListExtend,
	"list__length": builtinLength,
	"list__remove": builtinListRemove,
	"list__slice":  builtinListSlice,
	// Map built-ins
	"map__entries": builtinMapEntries,
	"map__keys":    builtinMapKeys,
	"map__remove":  builtinMapRemove,
	"map__values":  builtinMapValues,
	// String built-ins
	"string__find":        builtinStringFind,
	"string__find_last":   builtinStringFindLast,
	"string__length":      builtinLength,
	"string__slice":       builtinStringSlice,
	"string__split_space": builtinStringSplitSpace,
	"string__split":       builtinStringSplit,
	"string__to_lower":    builtinStringToLower,
	"string__to_upper":    builtinStringToUpper,
}
