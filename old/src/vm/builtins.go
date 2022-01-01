package vm

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

/**
 * Global built-in functions
 */

func builtinInput(args ...VeniceValue) (VeniceValue, error) {
	countArgsOrPanic(args, 1)
	stringArg := args[0].(*VeniceString)

	reader := bufio.NewReader(os.Stdin)
	fmt.Print(stringArg.Value)
	text, err := reader.ReadString('\n')
	if err != nil {
		return VENICE_OPTIONAL_NONE, nil
	}
	return VeniceOptionalOf(&VeniceString{text}), nil
}

func builtinInt(args ...VeniceValue) (VeniceValue, error) {
	countArgsOrPanic(args, 1)
	realArg := args[0].(*VeniceRealNumber)
	return &VeniceInteger{int(realArg.Value)}, nil
}

func builtinMaximum(args ...VeniceValue) (VeniceValue, error) {
	countArgsOrPanic(args, 2)
	intArg1 := args[0].(*VeniceInteger)
	intArg2 := args[1].(*VeniceInteger)

	if intArg1.Value > intArg2.Value {
		return intArg1, nil
	} else {
		return intArg2, nil
	}
}

func builtinMinimum(args ...VeniceValue) (VeniceValue, error) {
	countArgsOrPanic(args, 2)

	intArg1 := args[0].(*VeniceInteger)
	intArg2 := args[1].(*VeniceInteger)

	if intArg1.Value > intArg2.Value {
		return intArg2, nil
	} else {
		return intArg1, nil
	}
}

func builtinPrint(args ...VeniceValue) (VeniceValue, error) {
	countArgsOrPanic(args, 1)
	switch arg := args[0].(type) {
	case *VeniceString:
		fmt.Println(arg.Value)
	default:
		fmt.Println(args[0].String())
	}
	return nil, nil
}

func builtinRange(args ...VeniceValue) (VeniceValue, error) {
	countArgsOrPanic(args, 2)
	intArg1 := args[0].(*VeniceInteger)
	intArg2 := args[1].(*VeniceInteger)

	length := intArg2.Value - intArg1.Value
	if length <= 0 {
		return &VeniceList{[]VeniceValue{}}, nil
	}

	numbers := make([]VeniceValue, 0, intArg2.Value-intArg1.Value)
	for i := intArg1.Value; i < intArg2.Value; i++ {
		numbers = append(numbers, &VeniceInteger{i})
	}
	return &VeniceList{numbers}, nil
}

func builtinReal(args ...VeniceValue) (VeniceValue, error) {
	countArgsOrPanic(args, 1)
	intArg := args[0].(*VeniceInteger)
	return &VeniceRealNumber{float64(intArg.Value)}, nil
}

func builtinString(args ...VeniceValue) (VeniceValue, error) {
	countArgsOrPanic(args, 1)
	switch arg := args[0].(type) {
	case *VeniceBoolean:
		if arg.Value {
			return &VeniceString{"true"}, nil
		} else {
			return &VeniceString{"false"}, nil
		}
	case *VeniceInteger:
		return &VeniceString{strconv.Itoa(arg.Value)}, nil
	case *VeniceString:
		return arg, nil
	default:
		panic("invalid argument to string()")
	}
}

/**
 * List built-ins
 */

func builtinListAppend(args ...VeniceValue) (VeniceValue, error) {
	countArgsOrPanic(args, 2)
	listArg := args[0].(*VeniceList)
	listArg.Values = append(listArg.Values, args[1])
	return nil, nil
}

func builtinListCopy(args ...VeniceValue) (VeniceValue, error) {
	countArgsOrPanic(args, 1)
	listArg := args[0].(*VeniceList)
	return listArg.Copy(), nil
}

func builtinListExtend(args ...VeniceValue) (VeniceValue, error) {
	countArgsOrPanic(args, 2)
	listArg1 := args[0].(*VeniceList)
	listArg2 := args[1].(*VeniceList)
	listArg1.Values = append(listArg1.Values, listArg2.Values...)
	return nil, nil
}

func builtinListFind(args ...VeniceValue) (VeniceValue, error) {
	countArgsOrPanic(args, 2)
	listArg := args[0].(*VeniceList)

	for i, value := range listArg.Values {
		if args[1].Equals(value) {
			return VeniceOptionalOf(&VeniceInteger{i}), nil
		}
	}

	return VENICE_OPTIONAL_NONE, nil
}

func builtinListFindLast(args ...VeniceValue) (VeniceValue, error) {
	countArgsOrPanic(args, 2)
	listArg := args[0].(*VeniceList)

	for i := len(listArg.Values) - 1; i >= 0; i-- {
		if args[1].Equals(listArg.Values[i]) {
			return VeniceOptionalOf(&VeniceInteger{i}), nil
		}
	}

	return VENICE_OPTIONAL_NONE, nil
}

func builtinListPop(args ...VeniceValue) (VeniceValue, error) {
	countArgsOrPanic(args, 1)
	listArg := args[0].(*VeniceList)

	if len(listArg.Values) == 0 {
		return nil, &PanicError{"index out of bounds"}
	}

	r := listArg.Values[len(listArg.Values)-1]
	listArg.Values = listArg.Values[:len(listArg.Values)-1]
	return r, nil
}

func builtinListRemove(args ...VeniceValue) (VeniceValue, error) {
	countArgsOrPanic(args, 2)
	listArg := args[0].(*VeniceList)
	intArg := args[1].(*VeniceInteger)

	index := intArg.Value
	if index < 0 || index >= len(listArg.Values) {
		return nil, &PanicError{"index out of bounds"}
	}

	listArg.Values = append(listArg.Values[:index], listArg.Values[index+1:]...)
	return nil, nil
}

func builtinListReversed(args ...VeniceValue) (VeniceValue, error) {
	countArgsOrPanic(args, 1)
	listArg := args[0].(*VeniceList)
	copiedList := listArg.Copy()
	builtinListReverseInPlace(copiedList)
	return copiedList, nil
}

func builtinListReverseInPlace(args ...VeniceValue) (VeniceValue, error) {
	countArgsOrPanic(args, 1)
	listArg := args[0].(*VeniceList)

	for i, j := 0, len(listArg.Values)-1; i < j; i, j = i+1, j-1 {
		listArg.Values[i], listArg.Values[j] = listArg.Values[j], listArg.Values[i]
	}

	return nil, nil
}

func builtinListSize(args ...VeniceValue) (VeniceValue, error) {
	countArgsOrPanic(args, 1)
	listArg := args[0].(*VeniceList)
	return &VeniceInteger{len(listArg.Values)}, nil
}

func builtinListSlice(args ...VeniceValue) (VeniceValue, error) {
	countArgsOrPanic(args, 3)
	listArg := args[0].(*VeniceList)
	startIndexArg := args[1].(*VeniceInteger)
	endIndexArg := args[2].(*VeniceInteger)

	if startIndexArg.Value < 0 || startIndexArg.Value >= len(listArg.Values) {
		return nil, &PanicError{"index out of bounds"}
	}

	if endIndexArg.Value < 0 || endIndexArg.Value > len(listArg.Values) {
		return nil, &PanicError{"index out of bounds"}
	}

	return &VeniceList{listArg.Values[startIndexArg.Value:endIndexArg.Value]}, nil
}

func builtinListSorted(args ...VeniceValue) (VeniceValue, error) {
	countArgsOrPanic(args, 1)
	listArg := args[0].(*VeniceList)
	copiedList := listArg.Copy()
	builtinListSortInPlace(copiedList)
	return copiedList, nil
}

func builtinListSortInPlace(args ...VeniceValue) (VeniceValue, error) {
	countArgsOrPanic(args, 1)
	listArg := args[0].(*VeniceList)
	sort.Slice(listArg.Values, func(i, j int) bool {
		return listArg.Values[i].Compare(listArg.Values[j])
	})
	return nil, nil
}

/**
 * Map built-ins
 */

func builtinMapClear(args ...VeniceValue) (VeniceValue, error) {
	countArgsOrPanic(args, 1)
	mapArg := args[0].(*VeniceMap)
	mapArg.Clear()
	return nil, nil
}

func builtinMapCopy(args ...VeniceValue) (VeniceValue, error) {
	countArgsOrPanic(args, 1)
	mapArg := args[0].(*VeniceMap)
	return mapArg.Copy(), nil
}

func builtinMapEntries(args ...VeniceValue) (VeniceValue, error) {
	countArgsOrPanic(args, 1)
	mapArg := args[0].(*VeniceMap)
	return mapArg.Entries(), nil
}

func builtinMapKeys(args ...VeniceValue) (VeniceValue, error) {
	countArgsOrPanic(args, 1)
	mapArg := args[0].(*VeniceMap)
	return mapArg.Keys(), nil
}

func builtinMapRemove(args ...VeniceValue) (VeniceValue, error) {
	countArgsOrPanic(args, 2)
	mapArg := args[0].(*VeniceMap)
	mapArg.Remove(args[1])
	return nil, nil
}

func builtinMapSize(args ...VeniceValue) (VeniceValue, error) {
	countArgsOrPanic(args, 1)
	mapArg := args[0].(*VeniceMap)
	return &VeniceInteger{mapArg.Size}, nil
}

func builtinMapValues(args ...VeniceValue) (VeniceValue, error) {
	countArgsOrPanic(args, 1)
	mapArg := args[0].(*VeniceMap)
	return mapArg.Values(), nil
}

/**
 * String built-ins
 */

func builtinStringEndsWith(args ...VeniceValue) (VeniceValue, error) {
	countArgsOrPanic(args, 2)
	stringArg := args[0].(*VeniceString)
	suffixArg := args[1].(*VeniceString)
	return &VeniceBoolean{strings.HasSuffix(stringArg.Value, suffixArg.Value)}, nil
}

func builtinStringFind(args ...VeniceValue) (VeniceValue, error) {
	countArgsOrPanic(args, 2)
	stringArg := args[0].(*VeniceString)
	searchArg := args[1].(*VeniceString)
	index := strings.Index(stringArg.Value, searchArg.Value)
	if index == -1 {
		return VENICE_OPTIONAL_NONE, nil
	} else {
		return VeniceOptionalOf(&VeniceInteger{index}), nil
	}
}

func builtinStringFindLast(args ...VeniceValue) (VeniceValue, error) {
	countArgsOrPanic(args, 2)
	stringArg := args[0].(*VeniceString)
	searchArg := args[1].(*VeniceString)
	index := strings.LastIndex(stringArg.Value, searchArg.Value)
	if index == -1 {
		return VENICE_OPTIONAL_NONE, nil
	} else {
		return VeniceOptionalOf(&VeniceInteger{index}), nil
	}
}

func builtinStringIsControl(args ...VeniceValue) (VeniceValue, error) {
	countArgsOrPanic(args, 1)
	stringArg := args[0].(*VeniceString)

	if len(stringArg.Value) == 0 {
		return &VeniceBoolean{false}, nil
	}

	result := true
	for _, r := range stringArg.Value {
		if !unicode.IsControl(r) {
			result = false
			break
		}
	}

	return &VeniceBoolean{result}, nil
}

func builtinStringIsDigit(args ...VeniceValue) (VeniceValue, error) {
	countArgsOrPanic(args, 1)
	stringArg := args[0].(*VeniceString)

	if len(stringArg.Value) == 0 {
		return &VeniceBoolean{false}, nil
	}

	result := true
	for _, r := range stringArg.Value {
		if !unicode.IsDigit(r) {
			result = false
			break
		}
	}

	return &VeniceBoolean{result}, nil
}

func builtinStringIsGraphic(args ...VeniceValue) (VeniceValue, error) {
	countArgsOrPanic(args, 1)
	stringArg := args[0].(*VeniceString)

	if len(stringArg.Value) == 0 {
		return &VeniceBoolean{false}, nil
	}

	result := true
	for _, r := range stringArg.Value {
		if !unicode.IsGraphic(r) {
			result = false
			break
		}
	}

	return &VeniceBoolean{result}, nil
}

func builtinStringIsLetter(args ...VeniceValue) (VeniceValue, error) {
	countArgsOrPanic(args, 1)
	stringArg := args[0].(*VeniceString)

	if len(stringArg.Value) == 0 {
		return &VeniceBoolean{false}, nil
	}

	result := true
	for _, r := range stringArg.Value {
		if !unicode.IsLetter(r) {
			result = false
			break
		}
	}

	return &VeniceBoolean{result}, nil
}

func builtinStringIsLowercase(args ...VeniceValue) (VeniceValue, error) {
	countArgsOrPanic(args, 1)
	stringArg := args[0].(*VeniceString)

	if len(stringArg.Value) == 0 {
		return &VeniceBoolean{false}, nil
	}

	result := true
	for _, r := range stringArg.Value {
		if !unicode.IsLower(r) {
			result = false
			break
		}
	}

	return &VeniceBoolean{result}, nil
}

func builtinStringIsMark(args ...VeniceValue) (VeniceValue, error) {
	countArgsOrPanic(args, 1)
	stringArg := args[0].(*VeniceString)

	if len(stringArg.Value) == 0 {
		return &VeniceBoolean{false}, nil
	}

	result := true
	for _, r := range stringArg.Value {
		if !unicode.IsMark(r) {
			result = false
			break
		}
	}

	return &VeniceBoolean{result}, nil
}

func builtinStringIsNumber(args ...VeniceValue) (VeniceValue, error) {
	countArgsOrPanic(args, 1)
	stringArg := args[0].(*VeniceString)

	if len(stringArg.Value) == 0 {
		return &VeniceBoolean{false}, nil
	}

	result := true
	for _, r := range stringArg.Value {
		if !unicode.IsNumber(r) {
			result = false
			break
		}
	}

	return &VeniceBoolean{result}, nil
}

func builtinStringIsPrintable(args ...VeniceValue) (VeniceValue, error) {
	countArgsOrPanic(args, 1)
	stringArg := args[0].(*VeniceString)

	if len(stringArg.Value) == 0 {
		return &VeniceBoolean{false}, nil
	}

	result := true
	for _, r := range stringArg.Value {
		if !unicode.IsPrint(r) {
			result = false
			break
		}
	}

	return &VeniceBoolean{result}, nil
}

func builtinStringIsPunctuation(args ...VeniceValue) (VeniceValue, error) {
	countArgsOrPanic(args, 1)
	stringArg := args[0].(*VeniceString)

	if len(stringArg.Value) == 0 {
		return &VeniceBoolean{false}, nil
	}

	result := true
	for _, r := range stringArg.Value {
		if !unicode.IsPunct(r) {
			result = false
			break
		}
	}

	return &VeniceBoolean{result}, nil
}

func builtinStringIsSymbol(args ...VeniceValue) (VeniceValue, error) {
	countArgsOrPanic(args, 1)
	stringArg := args[0].(*VeniceString)

	if len(stringArg.Value) == 0 {
		return &VeniceBoolean{false}, nil
	}

	result := true
	for _, r := range stringArg.Value {
		if !unicode.IsSymbol(r) {
			result = false
			break
		}
	}

	return &VeniceBoolean{result}, nil
}

func builtinStringIsTitleCase(args ...VeniceValue) (VeniceValue, error) {
	countArgsOrPanic(args, 1)
	stringArg := args[0].(*VeniceString)

	if len(stringArg.Value) == 0 {
		return &VeniceBoolean{false}, nil
	}

	result := true
	for _, r := range stringArg.Value {
		if !unicode.IsTitle(r) {
			result = false
			break
		}
	}

	return &VeniceBoolean{result}, nil
}

func builtinStringIsUppercase(args ...VeniceValue) (VeniceValue, error) {
	countArgsOrPanic(args, 1)
	stringArg := args[0].(*VeniceString)

	if len(stringArg.Value) == 0 {
		return &VeniceBoolean{false}, nil
	}

	result := true
	for _, r := range stringArg.Value {
		if !unicode.IsUpper(r) {
			result = false
			break
		}
	}

	return &VeniceBoolean{result}, nil
}

func builtinStringIsWhitespace(args ...VeniceValue) (VeniceValue, error) {
	countArgsOrPanic(args, 1)
	stringArg := args[0].(*VeniceString)

	if len(stringArg.Value) == 0 {
		return &VeniceBoolean{false}, nil
	}

	result := true
	for _, r := range stringArg.Value {
		if !unicode.IsSpace(r) {
			result = false
			break
		}
	}

	return &VeniceBoolean{result}, nil
}

func builtinStringQuoted(args ...VeniceValue) (VeniceValue, error) {
	countArgsOrPanic(args, 1)
	stringArg := args[0].(*VeniceString)
	return &VeniceString{strconv.Quote(stringArg.Value)}, nil
}

func builtinStringRemovePrefix(args ...VeniceValue) (VeniceValue, error) {
	countArgsOrPanic(args, 2)
	stringArg := args[0].(*VeniceString)
	prefixArg := args[1].(*VeniceString)

	if strings.HasPrefix(stringArg.Value, prefixArg.Value) {
		return &VeniceString{stringArg.Value[len(prefixArg.Value):]}, nil
	} else {
		return stringArg, nil
	}
}

func builtinStringRemoveSuffix(args ...VeniceValue) (VeniceValue, error) {
	countArgsOrPanic(args, 2)
	stringArg := args[0].(*VeniceString)
	suffixArg := args[1].(*VeniceString)

	if strings.HasSuffix(stringArg.Value, suffixArg.Value) {
		return &VeniceString{stringArg.Value[:len(stringArg.Value)-len(suffixArg.Value)]}, nil
	} else {
		return stringArg, nil
	}
}

func builtinStringReplaceAll(args ...VeniceValue) (VeniceValue, error) {
	countArgsOrPanic(args, 3)
	stringArg := args[0].(*VeniceString)
	beforeArg := args[1].(*VeniceString)
	afterArg := args[2].(*VeniceString)
	return &VeniceString{strings.ReplaceAll(stringArg.Value, beforeArg.Value, afterArg.Value)}, nil
}

func builtinStringReplaceFirst(args ...VeniceValue) (VeniceValue, error) {
	countArgsOrPanic(args, 3)
	stringArg := args[0].(*VeniceString)
	beforeArg := args[1].(*VeniceString)
	afterArg := args[2].(*VeniceString)
	return &VeniceString{strings.Replace(stringArg.Value, beforeArg.Value, afterArg.Value, 1)}, nil
}

func builtinStringReplaceLast(args ...VeniceValue) (VeniceValue, error) {
	countArgsOrPanic(args, 3)
	stringArg := args[0].(*VeniceString)
	beforeArg := args[1].(*VeniceString)
	afterArg := args[2].(*VeniceString)

	index := strings.LastIndex(stringArg.Value, beforeArg.Value)
	if index == -1 {
		return stringArg, nil
	}
	return &VeniceString{
		stringArg.Value[:index] +
			afterArg.Value +
			stringArg.Value[index+len(beforeArg.Value):],
	}, nil
}

func builtinStringSize(args ...VeniceValue) (VeniceValue, error) {
	countArgsOrPanic(args, 1)
	stringArg := args[0].(*VeniceString)
	return &VeniceInteger{utf8.RuneCountInString(stringArg.Value)}, nil
}

func builtinStringSlice(args ...VeniceValue) (VeniceValue, error) {
	countArgsOrPanic(args, 3)
	stringArg := args[0].(*VeniceString)
	startIndexArg := args[1].(*VeniceInteger)
	endIndexArg := args[2].(*VeniceInteger)

	if startIndexArg.Value < 0 || startIndexArg.Value >= len(stringArg.Value) {
		return nil, &PanicError{"index out of bounds"}
	}

	if endIndexArg.Value < 0 || endIndexArg.Value > len(stringArg.Value) {
		return nil, &PanicError{"index out of bounds"}
	}

	return &VeniceString{getUtf8Slice(stringArg.Value, startIndexArg.Value, endIndexArg.Value)}, nil
}

func builtinStringSplit(args ...VeniceValue) (VeniceValue, error) {
	countArgsOrPanic(args, 2)
	stringArg := args[0].(*VeniceString)
	splitterArg := args[1].(*VeniceString)

	list := &VeniceList{}
	for _, word := range strings.Split(stringArg.Value, splitterArg.Value) {
		list.Values = append(list.Values, &VeniceString{word})
	}
	return list, nil
}

func builtinStringSplitSpace(args ...VeniceValue) (VeniceValue, error) {
	countArgsOrPanic(args, 1)
	stringArg := args[0].(*VeniceString)

	list := &VeniceList{}
	for _, word := range strings.Fields(stringArg.Value) {
		list.Values = append(list.Values, &VeniceString{word})
	}
	return list, nil
}

func builtinStringStartsWith(args ...VeniceValue) (VeniceValue, error) {
	countArgsOrPanic(args, 2)
	stringArg := args[0].(*VeniceString)
	prefixArg := args[1].(*VeniceString)
	return &VeniceBoolean{strings.HasPrefix(stringArg.Value, prefixArg.Value)}, nil
}

func builtinStringToUppercase(args ...VeniceValue) (VeniceValue, error) {
	countArgsOrPanic(args, 1)
	stringArg := args[0].(*VeniceString)
	return &VeniceString{strings.ToUpper(stringArg.Value)}, nil
}

func builtinStringToLowercase(args ...VeniceValue) (VeniceValue, error) {
	countArgsOrPanic(args, 1)
	stringArg := args[0].(*VeniceString)
	return &VeniceString{strings.ToLower(stringArg.Value)}, nil
}

func builtinStringTrim(args ...VeniceValue) (VeniceValue, error) {
	countArgsOrPanic(args, 1)
	stringArg := args[0].(*VeniceString)
	return &VeniceString{strings.TrimSpace(stringArg.Value)}, nil
}

func builtinStringTrimLeft(args ...VeniceValue) (VeniceValue, error) {
	countArgsOrPanic(args, 1)
	stringArg := args[0].(*VeniceString)
	return &VeniceString{strings.TrimLeftFunc(stringArg.Value, unicode.IsSpace)}, nil
}

func builtinStringTrimRight(args ...VeniceValue) (VeniceValue, error) {
	countArgsOrPanic(args, 1)
	stringArg := args[0].(*VeniceString)
	return &VeniceString{strings.TrimRightFunc(stringArg.Value, unicode.IsSpace)}, nil
}

/**
 * Miscellaneous functions.
 */

func getUtf8Slice(s string, start int, end int) string {
	byteIndex := 0
	charIndex := 0

	for charIndex < start {
		_, size := utf8.DecodeRuneInString(s[byteIndex:])
		byteIndex += size
		charIndex++
	}

	var sb strings.Builder
	for charIndex < end {
		r, size := utf8.DecodeRuneInString(s[byteIndex:])
		sb.WriteRune(r)
		byteIndex += size
		charIndex++
	}

	return sb.String()
}

func countArgsOrPanic(args []VeniceValue, count int) {
	if len(args) != count {
		panic("wrong number of arguments to built-in function")
	}
}

/**
 * Function table
 */

// If a method is added here, make sure to also add it to the appropriate place in
// compiler/compiler.go - `NewBuiltinSymbolTable` if it is a global built-in,
// `stringBuiltins` if it is a string built-in, etc.
var builtins = map[string]func(args ...VeniceValue) (VeniceValue, error){
	// Global built-ins
	"input":   builtinInput,
	"int":     builtinInt,
	"maximum": builtinMaximum,
	"minimum": builtinMinimum,
	"print":   builtinPrint,
	"range":   builtinRange,
	"real":    builtinReal,
	"string":  builtinString,
	// List built-ins
	"list__append":           builtinListAppend,
	"list__copy":             builtinListCopy,
	"list__extend":           builtinListExtend,
	"list__find":             builtinListFind,
	"list__find_last":        builtinListFindLast,
	"list__pop":              builtinListPop,
	"list__remove":           builtinListRemove,
	"list__reversed":         builtinListReversed,
	"list__reverse_in_place": builtinListReverseInPlace,
	"list__slice":            builtinListSlice,
	"list__sorted":           builtinListSorted,
	"list__sort_in_place":    builtinListSortInPlace,
	"list__size":             builtinListSize,
	// Map built-ins
	"map__clear":   builtinMapClear,
	"map__copy":    builtinMapCopy,
	"map__entries": builtinMapEntries,
	"map__keys":    builtinMapKeys,
	"map__remove":  builtinMapRemove,
	"map__values":  builtinMapValues,
	"map__size":    builtinMapSize,
	// String built-ins
	"string__ends_with":      builtinStringEndsWith,
	"string__find":           builtinStringFind,
	"string__find_last":      builtinStringFindLast,
	"string__is_control":     builtinStringIsControl,
	"string__is_digit":       builtinStringIsDigit,
	"string__is_graphic":     builtinStringIsGraphic,
	"string__is_letter":      builtinStringIsLetter,
	"string__is_lowercase":   builtinStringIsLowercase,
	"string__is_mark":        builtinStringIsMark,
	"string__is_number":      builtinStringIsNumber,
	"string__is_printable":   builtinStringIsPrintable,
	"string__is_punctuation": builtinStringIsPunctuation,
	"string__is_symbol":      builtinStringIsSymbol,
	"string__is_title_case":  builtinStringIsTitleCase,
	"string__is_uppercase":   builtinStringIsUppercase,
	"string__is_whitespace":  builtinStringIsWhitespace,
	"string__quoted":         builtinStringQuoted,
	"string__remove_prefix":  builtinStringRemovePrefix,
	"string__remove_suffix":  builtinStringRemoveSuffix,
	"string__replace_all":    builtinStringReplaceAll,
	"string__replace_first":  builtinStringReplaceFirst,
	"string__replace_last":   builtinStringReplaceLast,
	"string__size":           builtinStringSize,
	"string__slice":          builtinStringSlice,
	"string__split_space":    builtinStringSplitSpace,
	"string__split":          builtinStringSplit,
	"string__starts_with":    builtinStringStartsWith,
	"string__to_lowercase":   builtinStringToLowercase,
	"string__to_uppercase":   builtinStringToUppercase,
	"string__trim":           builtinStringTrim,
	"string__trim_left":      builtinStringTrimLeft,
	"string__trim_right":     builtinStringTrimRight,
}
