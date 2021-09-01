package vm

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

/**
 * Global built-in functions
 */

func builtinInput(args ...VeniceValue) VeniceValue {
	if len(args) != 1 {
		return nil
	}

	stringArg, ok := args[0].(*VeniceString)
	if !ok {
		return nil
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Print(stringArg.Value)
	text, err := reader.ReadString('\n')
	if err != nil {
		return VENICE_OPTIONAL_NONE
	}
	return VeniceOptionalOf(&VeniceString{text})
}

func builtinInt(args ...VeniceValue) VeniceValue {
	if len(args) != 1 {
		return nil
	}

	realArg, ok := args[0].(*VeniceRealNumber)
	if !ok {
		return nil
	}

	return &VeniceInteger{int(realArg.Value)}
}

func builtinLength(args ...VeniceValue) VeniceValue {
	if len(args) != 1 {
		return nil
	}

	var n int
	switch arg := args[0].(type) {
	case *VeniceList:
		n = len(arg.Values)
	case *VeniceMap:
		n = arg.Size
	case *VeniceString:
		n = len(arg.Value)
	default:
		return nil
	}
	return &VeniceInteger{n}
}

func builtinPrint(args ...VeniceValue) VeniceValue {
	if len(args) != 1 {
		return nil
	}

	switch arg := args[0].(type) {
	case *VeniceCharacter:
		fmt.Println(string(arg.Value))
	case *VeniceString:
		fmt.Println(arg.Value)
	default:
		fmt.Println(args[0].String())
	}
	return nil
}

func builtinRange(args ...VeniceValue) VeniceValue {
	if len(args) != 2 {
		return nil
	}

	intArg1, ok1 := args[0].(*VeniceInteger)
	intArg2, ok2 := args[1].(*VeniceInteger)
	if !ok1 || !ok2 {
		return nil
	}

	length := intArg2.Value - intArg1.Value
	if length <= 0 {
		return &VeniceList{[]VeniceValue{}}
	}

	numbers := make([]VeniceValue, 0, intArg2.Value-intArg1.Value)
	for i := intArg1.Value; i < intArg2.Value; i++ {
		numbers = append(numbers, &VeniceInteger{i})
	}
	return &VeniceList{numbers}
}

func builtinReal(args ...VeniceValue) VeniceValue {
	if len(args) != 1 {
		return nil
	}

	intArg, ok := args[0].(*VeniceInteger)
	if !ok {
		return nil
	}

	return &VeniceRealNumber{float64(intArg.Value)}
}

func builtinString(args ...VeniceValue) VeniceValue {
	if len(args) != 1 {
		return nil
	}

	switch arg := args[0].(type) {
	case *VeniceBoolean:
		if arg.Value {
			return &VeniceString{"true"}
		} else {
			return &VeniceString{"false"}
		}
	case *VeniceCharacter:
		return &VeniceString{string(arg.Value)}
	case *VeniceInteger:
		return &VeniceString{strconv.Itoa(arg.Value)}
	case *VeniceString:
		return arg
	default:
		return nil
	}
}

/**
 * List built-ins
 */

func builtinListAppend(args ...VeniceValue) VeniceValue {
	if len(args) != 2 {
		return nil
	}

	listArg, ok := args[0].(*VeniceList)
	if !ok {
		return nil
	}

	listArg.Values = append(listArg.Values, args[1])
	return nil
}

func builtinListCopy(args ...VeniceValue) VeniceValue {
	if len(args) != 1 {
		return nil
	}

	listArg, ok := args[0].(*VeniceList)
	if !ok {
		return nil
	}

	return listArg.Copy()
}

func builtinListExtend(args ...VeniceValue) VeniceValue {
	if len(args) != 2 {
		return nil
	}

	listArg1, ok1 := args[0].(*VeniceList)
	listArg2, ok2 := args[1].(*VeniceList)
	if !ok1 || !ok2 {
		return nil
	}

	listArg1.Values = append(listArg1.Values, listArg2.Values...)
	return nil
}

func builtinListFind(args ...VeniceValue) VeniceValue {
	if len(args) != 2 {
		return nil
	}

	listArg, ok := args[0].(*VeniceList)
	if !ok {
		return nil
	}

	for i, value := range listArg.Values {
		if args[1].Equals(value) {
			return VeniceOptionalOf(&VeniceInteger{i})
		}
	}

	return VENICE_OPTIONAL_NONE
}

func builtinListFindLast(args ...VeniceValue) VeniceValue {
	if len(args) != 2 {
		return nil
	}

	listArg, ok := args[0].(*VeniceList)
	if !ok {
		return nil
	}

	for i := len(listArg.Values) - 1; i >= 0; i-- {
		if args[1].Equals(listArg.Values[i]) {
			return VeniceOptionalOf(&VeniceInteger{i})
		}
	}

	return VENICE_OPTIONAL_NONE
}

func builtinListJoin(args ...VeniceValue) VeniceValue {
	if len(args) != 2 {
		return nil
	}

	listArg, ok1 := args[0].(*VeniceList)
	stringArg, ok2 := args[1].(*VeniceString)
	if !ok1 || !ok2 {
		return nil
	}

	var sb strings.Builder
	joiner := stringArg.Value
	for i, value := range listArg.Values {
		valueAsString, ok := value.(*VeniceString)
		if !ok {
			return nil
		}

		sb.WriteString(valueAsString.Value)
		if i != len(listArg.Values)-1 {
			sb.WriteString(joiner)
		}
	}
	return &VeniceString{sb.String()}
}

func builtinListRemove(args ...VeniceValue) VeniceValue {
	if len(args) != 2 {
		return nil
	}

	listArg, ok1 := args[0].(*VeniceList)
	intArg, ok2 := args[1].(*VeniceInteger)
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

func builtinListReversed(args ...VeniceValue) VeniceValue {
	if len(args) != 1 {
		return nil
	}

	listArg, ok := args[0].(*VeniceList)
	if !ok {
		return nil
	}

	copiedList := listArg.Copy()
	builtinListReverseInPlace(copiedList)
	return copiedList
}

func builtinListReverseInPlace(args ...VeniceValue) VeniceValue {
	if len(args) != 1 {
		return nil
	}

	listArg, ok := args[0].(*VeniceList)
	if !ok {
		return nil
	}

	for i, j := 0, len(listArg.Values)-1; i < j; i, j = i+1, j-1 {
		listArg.Values[i], listArg.Values[j] = listArg.Values[j], listArg.Values[i]
	}

	return nil
}

func builtinListSlice(args ...VeniceValue) VeniceValue {
	if len(args) != 3 {
		return nil
	}

	listArg, ok1 := args[0].(*VeniceList)
	startIndexArg, ok2 := args[1].(*VeniceInteger)
	endIndexArg, ok3 := args[2].(*VeniceInteger)
	if !ok1 || !ok2 || !ok3 {
		return nil
	}

	// TODO(2021-08-25): Handle out-of-bounds error.
	return &VeniceList{listArg.Values[startIndexArg.Value:endIndexArg.Value]}
}

func builtinListSorted(args ...VeniceValue) VeniceValue {
	if len(args) != 1 {
		return nil
	}

	listArg, ok := args[0].(*VeniceList)
	if !ok {
		return nil
	}

	copiedList := listArg.Copy()
	builtinListSortInPlace(copiedList)
	return copiedList
}

func builtinListSortInPlace(args ...VeniceValue) VeniceValue {
	if len(args) != 1 {
		return nil
	}

	listArg, ok := args[0].(*VeniceList)
	if !ok {
		return nil
	}

	sort.Slice(listArg.Values, func(i, j int) bool {
		return listArg.Values[i].Compare(listArg.Values[j])
	})
	return nil
}

/**
 * Map built-ins
 */

func builtinMapClear(args ...VeniceValue) VeniceValue {
	if len(args) != 1 {
		return nil
	}

	mapArg, ok := args[0].(*VeniceMap)
	if !ok {
		return nil
	}
	mapArg.Clear()
	return nil
}

func builtinMapCopy(args ...VeniceValue) VeniceValue {
	if len(args) != 1 {
		return nil
	}

	mapArg, ok := args[0].(*VeniceMap)
	if !ok {
		return nil
	}
	return mapArg.Copy()
}

func builtinMapEntries(args ...VeniceValue) VeniceValue {
	if len(args) != 1 {
		return nil
	}

	mapArg, ok := args[0].(*VeniceMap)
	if !ok {
		return nil
	}

	return mapArg.Entries()
}

func builtinMapKeys(args ...VeniceValue) VeniceValue {
	if len(args) != 1 {
		return nil
	}

	mapArg, ok := args[0].(*VeniceMap)
	if !ok {
		return nil
	}

	return mapArg.Keys()
}

func builtinMapRemove(args ...VeniceValue) VeniceValue {
	if len(args) != 2 {
		return nil
	}

	mapArg, ok := args[0].(*VeniceMap)
	if !ok {
		return nil
	}

	mapArg.Remove(args[1])
	return nil
}

func builtinMapValues(args ...VeniceValue) VeniceValue {
	if len(args) != 1 {
		return nil
	}

	mapArg, ok := args[0].(*VeniceMap)
	if !ok {
		return nil
	}

	return mapArg.Values()
}

/**
 * String built-ins
 */

func builtinStringFind(args ...VeniceValue) VeniceValue {
	if len(args) != 2 {
		return nil
	}

	stringArg, ok1 := args[0].(*VeniceString)
	searchArg, ok2 := args[1].(*VeniceString)
	if !ok1 || !ok2 {
		return nil
	}

	index := strings.Index(stringArg.Value, searchArg.Value)
	if index == -1 {
		return VENICE_OPTIONAL_NONE
	} else {
		return VeniceOptionalOf(&VeniceInteger{index})
	}
}

func builtinStringFindLast(args ...VeniceValue) VeniceValue {
	if len(args) != 2 {
		return nil
	}

	stringArg, ok1 := args[0].(*VeniceString)
	searchArg, ok2 := args[1].(*VeniceString)
	if !ok1 || !ok2 {
		return nil
	}

	index := strings.LastIndex(stringArg.Value, searchArg.Value)
	if index == -1 {
		return VENICE_OPTIONAL_NONE
	} else {
		return VeniceOptionalOf(&VeniceInteger{index})
	}
}

func builtinStringSlice(args ...VeniceValue) VeniceValue {
	if len(args) != 3 {
		return nil
	}

	stringArg, ok1 := args[0].(*VeniceString)
	startIndexArg, ok2 := args[1].(*VeniceInteger)
	endIndexArg, ok3 := args[2].(*VeniceInteger)
	if !ok1 || !ok2 || !ok3 {
		return nil
	}

	// TODO(2021-08-25): Handle out-of-bounds error.
	return &VeniceString{stringArg.Value[startIndexArg.Value:endIndexArg.Value]}
}

func builtinStringSplit(args ...VeniceValue) VeniceValue {
	if len(args) != 2 {
		return nil
	}

	stringArg, ok1 := args[0].(*VeniceString)
	splitterArg, ok2 := args[1].(*VeniceString)
	if !ok1 || !ok2 {
		return nil
	}

	list := &VeniceList{}
	for _, word := range strings.Split(stringArg.Value, splitterArg.Value) {
		list.Values = append(list.Values, &VeniceString{word})
	}
	return list
}

func builtinStringSplitSpace(args ...VeniceValue) VeniceValue {
	if len(args) != 1 {
		return nil
	}

	stringArg, ok := args[0].(*VeniceString)
	if !ok {
		return nil
	}

	list := &VeniceList{}
	for _, word := range strings.Fields(stringArg.Value) {
		list.Values = append(list.Values, &VeniceString{word})
	}
	return list
}

func builtinStringToUpper(args ...VeniceValue) VeniceValue {
	if len(args) != 1 {
		return nil
	}

	stringArg, ok := args[0].(*VeniceString)
	if !ok {
		return nil
	}

	return &VeniceString{strings.ToUpper(stringArg.Value)}
}

func builtinStringToLower(args ...VeniceValue) VeniceValue {
	if len(args) != 1 {
		return nil
	}

	stringArg, ok := args[0].(*VeniceString)
	if !ok {
		return nil
	}

	return &VeniceString{strings.ToLower(stringArg.Value)}
}

// If a method is added here, make sure to also add it to the appropriate place in
// compiler/compiler.go - `NewBuiltinSymbolTable` if it is a global built-in,
// `stringBuiltins` if it is a string built-in, etc.
var builtins = map[string]func(args ...VeniceValue) VeniceValue{
	// Global built-ins
	"input":  builtinInput,
	"int":    builtinInt,
	"length": builtinLength,
	"print":  builtinPrint,
	"range":  builtinRange,
	"real":   builtinReal,
	"string": builtinString,
	// List built-ins
	"list__append":           builtinListAppend,
	"list__copy":             builtinListCopy,
	"list__extend":           builtinListExtend,
	"list__find":             builtinListFind,
	"list__find_last":        builtinListFindLast,
	"list__join":             builtinListJoin,
	"list__length":           builtinLength,
	"list__remove":           builtinListRemove,
	"list__reversed":         builtinListReversed,
	"list__reverse_in_place": builtinListReverseInPlace,
	"list__slice":            builtinListSlice,
	"list__sorted":           builtinListSorted,
	"list__sort_in_place":    builtinListSortInPlace,
	// Map built-ins
	"map__clear":   builtinMapClear,
	"map__copy":    builtinMapCopy,
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
