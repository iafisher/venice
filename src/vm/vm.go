package vm

import (
	"fmt"
	"github.com/iafisher/venice/src/vval"
)

type Environment struct {
	Parent  *Environment
	Symbols map[string]vval.VeniceValue
}

type VirtualMachine struct {
	Stack []vval.VeniceValue
	Env   *Environment
}

func NewVirtualMachine() *VirtualMachine {
	return &VirtualMachine{[]vval.VeniceValue{}, &Environment{nil, make(map[string]vval.VeniceValue)}}
}

func (vm *VirtualMachine) Execute(compiledProgram vval.CompiledProgram, debug bool) (vval.VeniceValue, error) {
	return vm.executeFunction(compiledProgram, "main", debug)
}

func (vm *VirtualMachine) executeFunction(compiledProgram vval.CompiledProgram, functionName string, debug bool) (vval.VeniceValue, error) {
	code, ok := compiledProgram[functionName]
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

func (vm *VirtualMachine) executeOne(bytecode *vval.Bytecode, compiledProgram vval.CompiledProgram, debug bool) (int, error) {
	switch bytecode.Name {
	case "BINARY_ADD":
		left, right, err := vm.popTwoInts()
		if err != nil {
			return -1, err
		}
		vm.pushStack(&vval.VeniceInteger{left.Value + right.Value})
	case "BINARY_AND":
		left, right, err := vm.popTwoBools()
		if err != nil {
			return -1, err
		}
		vm.pushStack(&vval.VeniceBoolean{left.Value && right.Value})
	case "BINARY_CONCAT":
		rightInterface := vm.popStack()
		leftInterface := vm.popStack()

		switch right := rightInterface.(type) {
		case *vval.VeniceList:
			left := leftInterface.(*vval.VeniceList)
			vm.pushStack(&vval.VeniceList{append(left.Values, right.Values...)})
		case *vval.VeniceString:
			left := leftInterface.(*vval.VeniceString)
			vm.pushStack(&vval.VeniceString{left.Value + right.Value})
		default:
			return -1, &ExecutionError{fmt.Sprintf("BINARY_CONCAT requires list or string on top of stack, got %s (%T)", rightInterface.String(), rightInterface)}
		}
	case "BINARY_DIV":
		left, right, err := vm.popTwoInts()
		if err != nil {
			return -1, err
		}
		vm.pushStack(&vval.VeniceInteger{left.Value / right.Value})
	case "BINARY_EQ":
		right := vm.popStack()
		left := vm.popStack()
		result := left.Equals(right)
		vm.pushStack(&vval.VeniceBoolean{result})
	case "BINARY_GT":
		left, right, err := vm.popTwoInts()
		if err != nil {
			return -1, err
		}
		vm.pushStack(&vval.VeniceBoolean{left.Value > right.Value})
	case "BINARY_GT_EQ":
		left, right, err := vm.popTwoInts()
		if err != nil {
			return -1, err
		}
		vm.pushStack(&vval.VeniceBoolean{left.Value >= right.Value})
	case "BINARY_LIST_INDEX":
		indexInterface := vm.popStack()
		index, ok := indexInterface.(*vval.VeniceInteger)
		if !ok {
			return -1, &ExecutionError{fmt.Sprintf("BINARY_LIST_INDEX requires integer on top of stack, got %s (%T)", indexInterface.String(), indexInterface)}
		}

		listInterface := vm.popStack()
		list, ok := listInterface.(*vval.VeniceList)
		if !ok {
			return -1, &ExecutionError{fmt.Sprintf("BINARY_LIST_INDEX requires list on top of stack, got %s (%T)", listInterface.String(), listInterface)}
		}

		if index.Value < 0 || index.Value >= len(list.Values) {
			return -1, &ExecutionError{"index out of bounds"}
		}

		vm.pushStack(list.Values[index.Value])
	case "BINARY_LT":
		left, right, err := vm.popTwoInts()
		if err != nil {
			return -1, err
		}
		vm.pushStack(&vval.VeniceBoolean{left.Value < right.Value})
	case "BINARY_LT_EQ":
		left, right, err := vm.popTwoInts()
		if err != nil {
			return -1, err
		}
		vm.pushStack(&vval.VeniceBoolean{left.Value <= right.Value})
	case "BINARY_MAP_INDEX":
		index := vm.popStack()
		vMapInterface := vm.popStack()
		vMap, ok := vMapInterface.(*vval.VeniceMap)
		if !ok {
			return -1, &ExecutionError{fmt.Sprintf("BINARY_MAP_INDEX requires map on top of stack, got %s (%T)", vMapInterface.String(), vMapInterface)}
		}

		for _, pair := range vMap.Pairs {
			if pair.Key.Equals(index) {
				vm.pushStack(pair.Value)
				return 1, nil
			}
		}

		// TODO(2021-08-03): Return a proper error value.
		vm.pushStack(&vval.VeniceInteger{-1})
	case "BINARY_MUL":
		left, right, err := vm.popTwoInts()
		if err != nil {
			return -1, err
		}
		vm.pushStack(&vval.VeniceInteger{left.Value * right.Value})
	case "BINARY_NOT_EQ":
		right := vm.popStack()
		left := vm.popStack()
		result := left.Equals(right)
		vm.pushStack(&vval.VeniceBoolean{!result})
	case "BINARY_OR":
		left, right, err := vm.popTwoBools()
		if err != nil {
			return -1, err
		}
		vm.pushStack(&vval.VeniceBoolean{left.Value || right.Value})
	case "BINARY_STRING_INDEX":
		indexInterface := vm.popStack()
		index, ok := indexInterface.(*vval.VeniceInteger)
		if !ok {
			return -1, &ExecutionError{fmt.Sprintf("BINARY_STRING_INDEX requires integer on top of stack, got %s (%T)", indexInterface.String(), indexInterface)}
		}

		stringInterface := vm.popStack()
		str, ok := stringInterface.(*vval.VeniceString)
		if !ok {
			return -1, &ExecutionError{fmt.Sprintf("BINARY_STRING_INDEX requires string on top of stack, got %s (%T)", stringInterface.String(), stringInterface)}
		}

		if index.Value < 0 || index.Value >= len(str.Value) {
			return -1, &ExecutionError{"index out of bounds"}
		}

		vm.pushStack(&vval.VeniceCharacter{str.Value[index.Value]})
	case "BINARY_SUB":
		left, right, err := vm.popTwoInts()
		if err != nil {
			return -1, err
		}
		vm.pushStack(&vval.VeniceInteger{left.Value - right.Value})
	case "BUILD_CLASS":
		values := []vval.VeniceValue{}
		n := bytecode.Args[0].(*vval.VeniceInteger).Value
		for i := 0; i < n; i++ {
			topOfStack := vm.popStack()
			values = append(values, topOfStack)
		}

		// Reverse the array since the values are popped off the stack in reverse order.
		for i, j := 0, len(values)-1; i < j; i, j = i+1, j-1 {
			values[i], values[j] = values[j], values[i]
		}

		vm.pushStack(&vval.VeniceClassObject{values})
	case "BUILD_LIST":
		values := []vval.VeniceValue{}
		n := bytecode.Args[0].(*vval.VeniceInteger).Value
		for i := 0; i < n; i++ {
			topOfStack := vm.popStack()
			values = append(values, topOfStack)
		}
		vm.pushStack(&vval.VeniceList{values})
	case "BUILD_MAP":
		pairs := []*vval.VeniceMapPair{}
		n := bytecode.Args[0].(*vval.VeniceInteger).Value
		for i := 0; i < n; i++ {
			value := vm.popStack()
			key := vm.popStack()
			pairs = append(pairs, &vval.VeniceMapPair{key, value})
		}
		vm.pushStack(&vval.VeniceMap{pairs})
	case "BUILD_TUPLE":
		values := []vval.VeniceValue{}
		n := bytecode.Args[0].(*vval.VeniceInteger).Value
		for i := 0; i < n; i++ {
			topOfStack := vm.popStack()
			values = append(values, topOfStack)
		}
		vm.pushStack(&vval.VeniceTuple{values})
	case "CALL_BUILTIN":
		if v, ok := bytecode.Args[0].(*vval.VeniceString); ok {
			switch v.Value {
			case "print":
				topOfStackAny := vm.popStack()
				switch topOfStack := topOfStackAny.(type) {
				case *vval.VeniceCharacter:
					fmt.Println(string(topOfStack.Value))
				case *vval.VeniceString:
					fmt.Println(topOfStack.Value)
				default:
					fmt.Println(topOfStackAny.String())
				}
			default:
				return -1, &ExecutionError{fmt.Sprintf("unknown builtin: %s", v.Value)}
			}
		} else {
			return -1, &ExecutionError{"argument to CALL_BUILTIN must be a string"}
		}
	case "CALL_FUNCTION":
		argOne := bytecode.Args[0].(*vval.VeniceString)
		argTwo := bytecode.Args[1].(*vval.VeniceInteger)

		functionName := argOne.Value
		n := argTwo.Value

		functionEnv := &Environment{vm.Env, map[string]vval.VeniceValue{}}
		functionStack := vm.Stack[len(vm.Stack)-n:]
		vm.Stack = vm.Stack[:len(vm.Stack)-n]

		if debug {
			fmt.Println("DEBUG: Calling function in child virtual machine")
		}

		functionVm := &VirtualMachine{functionStack, functionEnv}
		value, err := functionVm.executeFunction(compiledProgram, functionName, debug)
		if err != nil {
			return -1, err
		}

		vm.pushStack(value)
	case "FOR_ITER":
		iter := vm.Stack[len(vm.Stack)-1].(vval.VeniceIterator)

		next := iter.Next()
		if next == nil {
			vm.popStack()
			n := bytecode.Args[0].(*vval.VeniceInteger).Value
			return n, nil
		} else {
			for _, v := range next {
				vm.pushStack(v)
			}
		}
	case "GET_ITER":
		topOfStackAny := vm.popStack()
		switch topOfStack := topOfStackAny.(type) {
		case *vval.VeniceList:
			vm.pushStack(&vval.VeniceListIterator{List: topOfStack, Index: 0})
		case *vval.VeniceMap:
			vm.pushStack(&vval.VeniceMapIterator{Map: topOfStack, Index: 0})
		}
	case "PUSH_CONST":
		vm.pushStack(bytecode.Args[0])
	case "PUSH_ENUM":
		values := []vval.VeniceValue{}
		n := bytecode.Args[1].(*vval.VeniceInteger).Value
		for i := 0; i < n; i++ {
			topOfStack := vm.popStack()
			values = append(values, topOfStack)
		}

		// Reverse the array since the values are popped off the stack in reverse order.
		for i, j := 0, len(values)-1; i < j; i, j = i+1, j-1 {
			values[i], values[j] = values[j], values[i]
		}

		vm.pushStack(&vval.VeniceEnumObject{bytecode.Args[0].(*vval.VeniceString).Value, values})
	case "PUSH_FIELD":
		fieldIndex := bytecode.Args[0].(*vval.VeniceInteger).Value
		topOfStack, ok := vm.popStack().(*vval.VeniceClassObject)
		if !ok {
			return -1, &ExecutionError{"expected class object at top of virtual machine stack"}
		}

		vm.pushStack(topOfStack.Values[fieldIndex])
	case "PUSH_NAME":
		symbol := bytecode.Args[0].(*vval.VeniceString).Value
		value, ok := vm.Env.Get(symbol)
		if !ok {
			return -1, &ExecutionError{fmt.Sprintf("undefined symbol: %s", symbol)}
		}
		vm.pushStack(value)
	case "PUSH_TUPLE_FIELD":
		fieldIndex := bytecode.Args[0].(*vval.VeniceInteger).Value
		topOfStack, ok := vm.popStack().(*vval.VeniceTuple)
		if !ok {
			return -1, &ExecutionError{"expected tuple at top of virtual machine stack"}
		}

		vm.pushStack(topOfStack.Values[fieldIndex])
	case "REL_JUMP":
		value := bytecode.Args[0].(*vval.VeniceInteger).Value
		return value, nil
	case "REL_JUMP_IF_FALSE":
		topOfStack, ok := vm.popStack().(*vval.VeniceBoolean)
		if !ok {
			return -1, &ExecutionError{"expected boolean at top of virtual machine stack"}
		}

		if !topOfStack.Value {
			value := bytecode.Args[0].(*vval.VeniceInteger).Value
			return value, nil
		}
	case "RETURN":
		return 0, nil
	case "STORE_NAME":
		symbol := bytecode.Args[0].(*vval.VeniceString).Value
		topOfStack := vm.popStack()
		vm.Env.Put(symbol, topOfStack)
	case "UNARY_MINUS":
		topOfStack, ok := vm.popStack().(*vval.VeniceInteger)
		if !ok {
			return -1, &ExecutionError{"expected integer at top of virtual machine stack"}
		}

		topOfStack.Value = -topOfStack.Value
		vm.pushStack(topOfStack)
	case "UNARY_NOT":
		topOfStack, ok := vm.popStack().(*vval.VeniceBoolean)
		if !ok {
			return -1, &ExecutionError{"expected boolean at top of virtual machine stack"}
		}

		topOfStack.Value = !topOfStack.Value
		vm.pushStack(topOfStack)
	default:
		return -1, &ExecutionError{fmt.Sprintf("unknown bytecode instruction: %s", bytecode.Name)}
	}
	return 1, nil
}

func (vm *VirtualMachine) pushStack(values ...vval.VeniceValue) {
	vm.Stack = append(vm.Stack, values...)
}

func (vm *VirtualMachine) popStack() vval.VeniceValue {
	ret := vm.Stack[len(vm.Stack)-1]
	vm.Stack = vm.Stack[:len(vm.Stack)-1]
	return ret
}

func (vm *VirtualMachine) popTwoInts() (*vval.VeniceInteger, *vval.VeniceInteger, error) {
	right, ok := vm.popStack().(*vval.VeniceInteger)
	if !ok {
		return nil, nil, &ExecutionError{"expected integer at top of virtual machine stack"}
	}

	left, ok := vm.popStack().(*vval.VeniceInteger)
	if !ok {
		return nil, nil, &ExecutionError{"expected integer at top of virtual machine stack"}
	}

	return left, right, nil
}

func (vm *VirtualMachine) popTwoBools() (*vval.VeniceBoolean, *vval.VeniceBoolean, error) {
	right, ok := vm.popStack().(*vval.VeniceBoolean)
	if !ok {
		return nil, nil, &ExecutionError{"expected boolean at top of virtual machine stack"}
	}

	left, ok := vm.popStack().(*vval.VeniceBoolean)
	if !ok {
		return nil, nil, &ExecutionError{"expected boolean at top of virtual machine stack"}
	}

	return left, right, nil
}

func (env *Environment) Get(symbol string) (vval.VeniceValue, bool) {
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

func (env *Environment) Put(symbol string, value vval.VeniceValue) {
	env.Symbols[symbol] = value
}

type ExecutionError struct {
	Message string
}

func (e *ExecutionError) Error() string {
	return e.Message
}
