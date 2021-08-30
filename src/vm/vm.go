package vm

import (
	"fmt"
	"github.com/iafisher/venice/src/common/bytecode"
	"strings"
)

type Environment struct {
	Parent  *Environment
	Symbols map[string]VeniceValue
}

type VirtualMachine struct {
	Stack []VeniceValue
	Env   *Environment
}

func NewVirtualMachine() *VirtualMachine {
	return &VirtualMachine{
		Stack: []VeniceValue{},
		Env: &Environment{
			Parent:  nil,
			Symbols: make(map[string]VeniceValue),
		},
	}
}

func (vm *VirtualMachine) Execute(
	compiledProgram *bytecode.CompiledProgram, debug bool,
) (VeniceValue, error) {
	if compiledProgram.Version != 1 {
		return nil, &ExecutionError{
			fmt.Sprintf("unsupported bytecode version: %d", compiledProgram.Version),
		}
	}

	return vm.executeFunction(compiledProgram, "main", debug)
}

func (vm *VirtualMachine) executeFunction(
	compiledProgram *bytecode.CompiledProgram, functionName string, debug bool,
) (VeniceValue, error) {
	code, ok := compiledProgram.Code[functionName]
	if !ok {
		return nil, &ExecutionError{fmt.Sprintf("function %q not found", functionName)}
	}

	index := 0
	for index < len(code) {
		bytecode := code[index]

		if debug {
			fmt.Printf("DEBUG: Executing %s\n", bytecode)
			fmt.Println("DEBUG: Stack (bottom to top)")
			if len(vm.Stack) > 0 {
				for _, value := range vm.Stack {
					fmt.Printf("DEBUG:   %s\n", value.String())
				}
			} else {
				fmt.Println("DEBUG:   <empty>")
			}
		}

		jump, err := vm.executeOne(bytecode, compiledProgram, debug)
		if err != nil {
			return nil, err
		}

		if jump == 0 {
			break
		}

		index += jump
	}

	if len(vm.Stack) > 0 {
		ret := vm.Stack[len(vm.Stack)-1]
		vm.Stack = nil
		return ret, nil
	} else {
		return nil, nil
	}
}

func (vm *VirtualMachine) executeOne(
	bcodeAny bytecode.Bytecode, compiledProgram *bytecode.CompiledProgram, debug bool,
) (int, error) {
	switch bcode := bcodeAny.(type) {
	case *bytecode.BinaryAdd:
		left, right, err := vm.popTwoInts()
		if err != nil {
			return -1, err
		}
		vm.pushStack(&VeniceInteger{left + right})
	case *bytecode.BinaryAnd:
		left, right, err := vm.popTwoBools()
		if err != nil {
			return -1, err
		}
		vm.pushStack(&VeniceBoolean{left.Value && right.Value})
	case *bytecode.BinaryConcat:
		rightAny := vm.popStack()
		leftAny := vm.popStack()

		switch right := rightAny.(type) {
		case *VeniceList:
			left := leftAny.(*VeniceList)
			vm.pushStack(&VeniceList{append(left.Values, right.Values...)})
		case *VeniceString:
			left := leftAny.(*VeniceString)
			vm.pushStack(&VeniceString{left.Value + right.Value})
		default:
			return -1, &ExecutionError{
				fmt.Sprintf(
					"BINARY_CONCAT requires list or string on top of stack, got %s (%T)",
					rightAny.String(),
					rightAny,
				),
			}
		}
	case *bytecode.BinaryEq:
		right := vm.popStack()
		left := vm.popStack()
		result := left.Equals(right)
		vm.pushStack(&VeniceBoolean{result})
	case *bytecode.BinaryGt:
		left, right, err := vm.popTwoInts()
		if err != nil {
			return -1, err
		}
		vm.pushStack(&VeniceBoolean{left > right})
	case *bytecode.BinaryGtEq:
		left, right, err := vm.popTwoInts()
		if err != nil {
			return -1, err
		}
		vm.pushStack(&VeniceBoolean{left >= right})
	case *bytecode.BinaryIn:
		rightAny := vm.popStack()
		leftAny := vm.popStack()

		var result bool
		switch right := rightAny.(type) {
		case *VeniceString:
			switch left := leftAny.(type) {
			case *VeniceCharacter:
				result = strings.IndexByte(right.Value, left.Value) != -1
			case *VeniceString:
				result = strings.Contains(right.Value, left.Value)
			default:
				return -1, &ExecutionError{"BINARY_IN expected character or string"}
			}
		case *VeniceList:
			result = false
			for _, value := range right.Values {
				if value.Equals(leftAny) {
					result = true
					break
				}
			}
		case *VeniceMap:
			result = right.Get(leftAny) != nil
		default:
			return -1, &ExecutionError{"BINARY_IN requires list, map, or string"}
		}
		vm.pushStack(&VeniceBoolean{result})
	case *bytecode.BinaryListIndex:
		indexAny := vm.popStack()
		index, ok := indexAny.(*VeniceInteger)
		if !ok {
			return -1, &ExecutionError{
				fmt.Sprintf(
					"BINARY_LIST_INDEX requires integer on top of stack, got %s (%T)",
					indexAny.String(),
					indexAny,
				),
			}
		}

		listAny := vm.popStack()
		list, ok := listAny.(*VeniceList)
		if !ok {
			return -1, &ExecutionError{
				fmt.Sprintf(
					"BINARY_LIST_INDEX requires list on top of stack, got %s (%T)",
					listAny.String(),
					listAny,
				),
			}
		}

		if index.Value < 0 || index.Value >= len(list.Values) {
			return -1, &ExecutionError{"index out of bounds"}
		}

		vm.pushStack(list.Values[index.Value])
	case *bytecode.BinaryLt:
		left, right, err := vm.popTwoInts()
		if err != nil {
			return -1, err
		}
		vm.pushStack(&VeniceBoolean{left < right})
	case *bytecode.BinaryLtEq:
		left, right, err := vm.popTwoInts()
		if err != nil {
			return -1, err
		}
		vm.pushStack(&VeniceBoolean{left <= right})
	case *bytecode.BinaryMapIndex:
		index := vm.popStack()
		vMapAny := vm.popStack()
		vMap, ok := vMapAny.(*VeniceMap)
		if !ok {
			return -1, &ExecutionError{
				fmt.Sprintf(
					"BINARY_MAP_INDEX requires map on top of stack, got %s (%T)",
					vMapAny.String(),
					vMapAny,
				),
			}
		}

		result := vMap.Get(index)
		var wrappedResult VeniceValue
		if result == nil {
			wrappedResult = &VeniceEnumObject{Label: "None"}
		} else {
			wrappedResult = &VeniceEnumObject{
				Label:  "Some",
				Values: []VeniceValue{result},
			}
		}

		vm.pushStack(wrappedResult)
	case *bytecode.BinaryMul:
		left, right, err := vm.popTwoInts()
		if err != nil {
			return -1, err
		}
		vm.pushStack(&VeniceInteger{left * right})
	case *bytecode.BinaryNotEq:
		right := vm.popStack()
		left := vm.popStack()
		result := left.Equals(right)
		vm.pushStack(&VeniceBoolean{!result})
	case *bytecode.BinaryOr:
		left, right, err := vm.popTwoBools()
		if err != nil {
			return -1, err
		}
		vm.pushStack(&VeniceBoolean{left.Value || right.Value})
	case *bytecode.BinaryRealAdd:
		left, right, err := vm.popTwoReals()
		if err != nil {
			return -1, err
		}
		vm.pushStack(&VeniceRealNumber{left + right})
	case *bytecode.BinaryRealDiv:
		left, right, err := vm.popTwoReals()
		if err != nil {
			return -1, err
		}
		vm.pushStack(&VeniceRealNumber{left / right})
	case *bytecode.BinaryRealMul:
		left, right, err := vm.popTwoReals()
		if err != nil {
			return -1, err
		}
		vm.pushStack(&VeniceRealNumber{left * right})
	case *bytecode.BinaryRealSub:
		left, right, err := vm.popTwoReals()
		if err != nil {
			return -1, err
		}
		vm.pushStack(&VeniceRealNumber{left - right})
	case *bytecode.BinaryStringIndex:
		indexAny := vm.popStack()
		index, ok := indexAny.(*VeniceInteger)
		if !ok {
			return -1, &ExecutionError{
				fmt.Sprintf(
					"BINARY_STRING_INDEX requires integer on top of stack, got %s (%T)",
					indexAny.String(),
					indexAny,
				),
			}
		}

		stringAny := vm.popStack()
		str, ok := stringAny.(*VeniceString)
		if !ok {
			return -1, &ExecutionError{
				fmt.Sprintf(
					"BINARY_STRING_INDEX requires string on top of stack, got %s (%T)",
					stringAny.String(),
					stringAny,
				),
			}
		}

		if index.Value < 0 || index.Value >= len(str.Value) {
			return -1, &ExecutionError{"index out of bounds"}
		}

		vm.pushStack(&VeniceCharacter{str.Value[index.Value]})
	case *bytecode.BinarySub:
		left, right, err := vm.popTwoInts()
		if err != nil {
			return -1, err
		}
		vm.pushStack(&VeniceInteger{left - right})
	case *bytecode.BuildClass:
		values := make([]VeniceValue, 0, bcode.N)
		for i := 0; i < bcode.N; i++ {
			topOfStack := vm.popStack()
			values = append(values, topOfStack)
		}

		vm.pushStack(&VeniceClassObject{ClassName: bcode.Name, Values: values})
	case *bytecode.BuildList:
		values := make([]VeniceValue, 0, bcode.N)
		for i := 0; i < bcode.N; i++ {
			topOfStack := vm.popStack()
			values = append(values, topOfStack)
		}
		vm.pushStack(&VeniceList{values})
	case *bytecode.BuildMap:
		pairs := make([]*VeniceMapPair, 0, bcode.N)
		for i := 0; i < bcode.N; i++ {
			value := vm.popStack()
			key := vm.popStack()
			pairs = append(pairs, &VeniceMapPair{key, value})
		}
		vm.pushStack(&VeniceMap{pairs})
	case *bytecode.BuildTuple:
		values := make([]VeniceValue, 0, bcode.N)
		for i := 0; i < bcode.N; i++ {
			topOfStack := vm.popStack()
			values = append(values, topOfStack)
		}
		vm.pushStack(&VeniceTuple{values})
	case *bytecode.CallFunction:
		functionObject, ok := vm.popStack().(*VeniceFunctionObject)
		if !ok {
			return -1, &ExecutionError{
				"expected function at top of virtual machine stack for CALL_FUNCTION",
			}
		}

		if functionObject.IsBuiltin {
			builtin, ok := builtins[functionObject.Name]
			if !ok {
				return -1, &ExecutionError{
					fmt.Sprintf("unknown builtin `%s`", functionObject.Name),
				}
			}

			args := make([]VeniceValue, 0, bcode.N)
			for i := 0; i < bcode.N; i++ {
				topOfStack := vm.popStack()
				args = append(args, topOfStack)
			}

			result := builtin(args...)
			if result != nil {
				vm.pushStack(result)
			}
		} else {
			functionEnv := &Environment{vm.Env, map[string]VeniceValue{}}
			functionStack := vm.Stack[len(vm.Stack)-bcode.N:]
			vm.Stack = vm.Stack[:len(vm.Stack)-bcode.N]

			if debug {
				fmt.Println("DEBUG: Calling function in child virtual machine")
			}

			functionVm := &VirtualMachine{functionStack, functionEnv}
			value, err := functionVm.executeFunction(compiledProgram, functionObject.Name, debug)
			if err != nil {
				return -1, err
			}

			vm.pushStack(value)
		}
	case *bytecode.CheckLabel:
		enum := vm.peekStack().(*VeniceEnumObject)
		vm.pushStack(&VeniceBoolean{enum.Label == bcode.Name})
	case *bytecode.ForIter:
		iter := vm.peekStack().(VeniceIterator)

		next := iter.Next()
		if next == nil {
			vm.popStack()
			return bcode.N, nil
		} else {
			for _, v := range next {
				vm.pushStack(v)
			}
		}
	case *bytecode.GetIter:
		topOfStackAny := vm.popStack()
		switch topOfStack := topOfStackAny.(type) {
		case *VeniceList:
			vm.pushStack(&VeniceListIterator{List: topOfStack, Index: 0})
		case *VeniceMap:
			vm.pushStack(&VeniceMapIterator{Map: topOfStack, Index: 0})
		}
	case *bytecode.LookupMethod:
		switch topOfStack := vm.peekStack().(type) {
		case *VeniceClassObject:
			vm.pushStack(
				&VeniceFunctionObject{
					fmt.Sprintf("%s__%s", topOfStack.ClassName, bcode.Name),
					false,
				},
			)
		case *VeniceList:
			vm.pushStack(&VeniceFunctionObject{fmt.Sprintf("list__%s", bcode.Name), true})
		case *VeniceMap:
			vm.pushStack(&VeniceFunctionObject{fmt.Sprintf("map__%s", bcode.Name), true})
		case *VeniceString:
			vm.pushStack(&VeniceFunctionObject{fmt.Sprintf("string__%s", bcode.Name), true})
		default:
			return -1, &ExecutionError{
				"expected class, list, map, or string object at top of virtual machine stack for LOOKUP_METHOD",
			}
		}
	case *bytecode.PushConstBool:
		vm.pushStack(&VeniceBoolean{bcode.Value})
	case *bytecode.PushConstChar:
		vm.pushStack(&VeniceCharacter{bcode.Value})
	case *bytecode.PushConstFunction:
		vm.pushStack(&VeniceFunctionObject{bcode.Name, bcode.IsBuiltin})
	case *bytecode.PushConstInt:
		vm.pushStack(&VeniceInteger{bcode.Value})
	case *bytecode.PushConstRealNumber:
		vm.pushStack(&VeniceRealNumber{bcode.Value})
	case *bytecode.PushConstStr:
		vm.pushStack(&VeniceString{bcode.Value})
	case *bytecode.PushEnum:
		values := make([]VeniceValue, 0, bcode.N)
		for i := 0; i < bcode.N; i++ {
			topOfStack := vm.popStack()
			values = append(values, topOfStack)
		}

		// Reverse the array since the values are popped off the stack in reverse order.
		for i, j := 0, len(values)-1; i < j; i, j = i+1, j-1 {
			values[i], values[j] = values[j], values[i]
		}

		vm.pushStack(&VeniceEnumObject{bcode.Name, values})
	case *bytecode.PushEnumIndex:
		topOfStack, ok := vm.peekStack().(*VeniceEnumObject)
		if !ok {
			return -1, &ExecutionError{
				"expected enum object at top of virtual machine stack for PUSH_ENUM_INDEX",
			}
		}
		vm.pushStack(topOfStack.Values[bcode.Index])
	case *bytecode.PushField:
		topOfStack, ok := vm.popStack().(*VeniceClassObject)
		if !ok {
			return -1, &ExecutionError{
				"expected class object at top of virtual machine stack for PUSH_FIELD",
			}
		}

		vm.pushStack(topOfStack.Values[bcode.Index])
	case *bytecode.PushName:
		value, ok := vm.Env.Get(bcode.Name)
		if !ok {
			return -1, &ExecutionError{fmt.Sprintf("undefined symbol: %s", bcode.Name)}
		}
		vm.pushStack(value)
	case *bytecode.PushTupleField:
		topOfStack, ok := vm.popStack().(*VeniceTuple)
		if !ok {
			return -1, &ExecutionError{
				"expected tuple at top of virtual machine stack for PUSH_TUPLE_FIELD",
			}
		}

		vm.pushStack(topOfStack.Values[bcode.Index])
	case *bytecode.RelJump:
		return bcode.N, nil
	case *bytecode.RelJumpIfFalse:
		topOfStack, ok := vm.popStack().(*VeniceBoolean)
		if !ok {
			return -1, &ExecutionError{
				"expected boolean at top of virtual machine stack for REL_JUMP_IF_FALSE",
			}
		}

		if !topOfStack.Value {
			return bcode.N, nil
		}
	case *bytecode.RelJumpIfFalseOrPop:
		topOfStack, ok := vm.peekStack().(*VeniceBoolean)
		if !ok {
			return -1, &ExecutionError{
				"expected boolean at top of virtual machine stack for REL_JUMP_IF_FALSE_OR_POP",
			}
		}

		if !topOfStack.Value {
			return bcode.N, nil
		} else {
			vm.popStack()
		}
	case *bytecode.RelJumpIfTrueOrPop:
		topOfStack, ok := vm.peekStack().(*VeniceBoolean)
		if !ok {
			return -1, &ExecutionError{
				"expected boolean at top of virtual machine stack for REL_JUMP_IF_TRUE_OR_POP",
			}
		}

		if topOfStack.Value {
			return bcode.N, nil
		} else {
			vm.popStack()
		}
	case *bytecode.Return:
		return 0, nil
	case *bytecode.StoreField:
		destinationAny := vm.popStack()
		value := vm.popStack()

		destination, ok := destinationAny.(*VeniceClassObject)
		if !ok {
			return -1, &ExecutionError{
				"expected class object at top of virtual machine stack for STORE_FIELD",
			}
		}

		destination.Values[bcode.Index] = value
	case *bytecode.StoreIndex:
		destinationAny := vm.popStack()
		indexAny := vm.popStack()
		value := vm.popStack()

		destination, ok := destinationAny.(*VeniceList)
		if !ok {
			return -1, &ExecutionError{
				"expected list object at top of virtual machine stack for STORE_INDEX",
			}
		}

		index, ok := indexAny.(*VeniceInteger)
		if !ok {
			return -1, &ExecutionError{
				"expected integer at top of virtual machine stack for STORE_INDEX",
			}
		}

		destination.Values[index.Value] = value
	case *bytecode.StoreMapIndex:
		destinationAny := vm.popStack()
		key := vm.popStack()
		value := vm.popStack()

		destination, ok := destinationAny.(*VeniceMap)
		if !ok {
			return -1, &ExecutionError{
				"expected map object at top of virtual machine stack for STORE_MAP_INDEX",
			}
		}

		destination.Put(key, value)
	case *bytecode.StoreName:
		topOfStack := vm.popStack()
		vm.Env.Put(bcode.Name, topOfStack)
	case *bytecode.UnaryMinus:
		topOfStack, ok := vm.popStack().(*VeniceInteger)
		if !ok {
			return -1, &ExecutionError{
				"expected integer at top of virtual machine stack for UNARY_MINUS",
			}
		}

		topOfStack.Value = -topOfStack.Value
		vm.pushStack(topOfStack)
	case *bytecode.UnaryNot:
		topOfStack, ok := vm.popStack().(*VeniceBoolean)
		if !ok {
			return -1, &ExecutionError{
				"expected boolean at top of virtual machine stack for UNARY_NOT",
			}
		}

		topOfStack.Value = !topOfStack.Value
		vm.pushStack(topOfStack)
	default:
		return -1, &ExecutionError{fmt.Sprintf("unknown bytecode instruction: %T", bcode)}
	}
	return 1, nil
}

func (vm *VirtualMachine) pushStack(values ...VeniceValue) {
	vm.Stack = append(vm.Stack, values...)
}

func (vm *VirtualMachine) popStack() VeniceValue {
	ret := vm.Stack[len(vm.Stack)-1]
	vm.Stack = vm.Stack[:len(vm.Stack)-1]
	return ret
}

func (vm *VirtualMachine) peekStack() VeniceValue {
	return vm.Stack[len(vm.Stack)-1]
}

func (vm *VirtualMachine) popTwoInts() (int, int, error) {
	right, err := vm.popInt()
	if err != nil {
		return 0, 0, err
	}

	left, err := vm.popInt()
	if err != nil {
		return 0, 0, err
	}

	return left, right, nil
}

func (vm *VirtualMachine) popInt() (int, error) {
	topOfStack, ok := vm.popStack().(*VeniceInteger)
	if !ok {
		return 0, &ExecutionError{"expected integer at top of virtual machine stack"}
	}
	return topOfStack.Value, nil
}

func (vm *VirtualMachine) popTwoReals() (float64, float64, error) {
	right, err := vm.popReal()
	if err != nil {
		return 0, 0, err
	}

	left, err := vm.popReal()
	if err != nil {
		return 0, 0, err
	}

	return left, right, nil
}

func (vm *VirtualMachine) popReal() (float64, error) {
	topOfStackAny := vm.popStack()
	switch topOfStack := topOfStackAny.(type) {
	case *VeniceInteger:
		return float64(topOfStack.Value), nil
	case *VeniceRealNumber:
		return topOfStack.Value, nil
	default:
		return 0, &ExecutionError{"expected real number at top of virtual machine stack"}
	}
}

func (vm *VirtualMachine) popTwoBools() (*VeniceBoolean, *VeniceBoolean, error) {
	right, ok := vm.popStack().(*VeniceBoolean)
	if !ok {
		return nil, nil, &ExecutionError{"expected boolean at top of virtual machine stack"}
	}

	left, ok := vm.popStack().(*VeniceBoolean)
	if !ok {
		return nil, nil, &ExecutionError{"expected boolean at top of virtual machine stack"}
	}

	return left, right, nil
}

func (env *Environment) Get(symbol string) (VeniceValue, bool) {
	value, ok := env.Symbols[symbol]
	if !ok {
		if env.Parent != nil {
			return env.Parent.Get(symbol)
		} else {
			return nil, false
		}
	}
	return value, true
}

func (env *Environment) Put(symbol string, value VeniceValue) {
	env.Symbols[symbol] = value
}

type ExecutionError struct {
	Message string
}

func (e *ExecutionError) Error() string {
	return e.Message
}
