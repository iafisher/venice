package vm

import (
	"fmt"
	"github.com/iafisher/venice/src/bytecode"
	"github.com/iafisher/venice/src/vval"
	"strings"
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

func (vm *VirtualMachine) Execute(compiledProgram *bytecode.CompiledProgram, debug bool) (vval.VeniceValue, error) {
	if compiledProgram.Version != 1 {
		return nil, &ExecutionError{fmt.Sprintf("unsupported bytecode version: %d", compiledProgram.Version)}
	}

	return vm.executeFunction(compiledProgram, "main", debug)
}

func (vm *VirtualMachine) executeFunction(compiledProgram *bytecode.CompiledProgram, functionName string, debug bool) (vval.VeniceValue, error) {
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

func (vm *VirtualMachine) executeOne(bcodeAny bytecode.Bytecode, compiledProgram *bytecode.CompiledProgram, debug bool) (int, error) {
	switch bcode := bcodeAny.(type) {
	case *bytecode.BinaryAdd:
		left, right, err := vm.popTwoInts()
		if err != nil {
			return -1, err
		}
		vm.pushStack(&vval.VeniceInteger{left.Value + right.Value})
	case *bytecode.BinaryAnd:
		left, right, err := vm.popTwoBools()
		if err != nil {
			return -1, err
		}
		vm.pushStack(&vval.VeniceBoolean{left.Value && right.Value})
	case *bytecode.BinaryConcat:
		rightAny := vm.popStack()
		leftAny := vm.popStack()

		switch right := rightAny.(type) {
		case *vval.VeniceList:
			left := leftAny.(*vval.VeniceList)
			vm.pushStack(&vval.VeniceList{append(left.Values, right.Values...)})
		case *vval.VeniceString:
			left := leftAny.(*vval.VeniceString)
			vm.pushStack(&vval.VeniceString{left.Value + right.Value})
		default:
			return -1, &ExecutionError{fmt.Sprintf("BINARY_CONCAT requires list or string on top of stack, got %s (%T)", rightAny.String(), rightAny)}
		}
	case *bytecode.BinaryDiv:
		left, right, err := vm.popTwoInts()
		if err != nil {
			return -1, err
		}
		vm.pushStack(&vval.VeniceInteger{left.Value / right.Value})
	case *bytecode.BinaryEq:
		right := vm.popStack()
		left := vm.popStack()
		result := left.Equals(right)
		vm.pushStack(&vval.VeniceBoolean{result})
	case *bytecode.BinaryGt:
		left, right, err := vm.popTwoInts()
		if err != nil {
			return -1, err
		}
		vm.pushStack(&vval.VeniceBoolean{left.Value > right.Value})
	case *bytecode.BinaryGtEq:
		left, right, err := vm.popTwoInts()
		if err != nil {
			return -1, err
		}
		vm.pushStack(&vval.VeniceBoolean{left.Value >= right.Value})
	case *bytecode.BinaryIn:
		rightAny := vm.popStack()
		leftAny := vm.popStack()

		var result bool
		switch right := rightAny.(type) {
		case *vval.VeniceString:
			switch left := leftAny.(type) {
			case *vval.VeniceCharacter:
				result = strings.IndexByte(right.Value, left.Value) != -1
			case *vval.VeniceString:
				result = strings.Contains(right.Value, left.Value)
			default:
				return -1, &ExecutionError{"BINARY_IN expected character or string"}
			}
		case *vval.VeniceList:
			result = false
			for _, value := range right.Values {
				if value.Equals(leftAny) {
					result = true
					break
				}
			}
		case *vval.VeniceMap:
			result = false
			for _, pair := range right.Pairs {
				if pair.Key.Equals(leftAny) {
					result = true
					break
				}
			}
		default:
			return -1, &ExecutionError{"BINARY_IN requires list, map, or string"}
		}
		vm.pushStack(&vval.VeniceBoolean{result})
	case *bytecode.BinaryListIndex:
		indexAny := vm.popStack()
		index, ok := indexAny.(*vval.VeniceInteger)
		if !ok {
			return -1, &ExecutionError{fmt.Sprintf("BINARY_LIST_INDEX requires integer on top of stack, got %s (%T)", indexAny.String(), indexAny)}
		}

		listAny := vm.popStack()
		list, ok := listAny.(*vval.VeniceList)
		if !ok {
			return -1, &ExecutionError{fmt.Sprintf("BINARY_LIST_INDEX requires list on top of stack, got %s (%T)", listAny.String(), listAny)}
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
		vm.pushStack(&vval.VeniceBoolean{left.Value < right.Value})
	case *bytecode.BinaryLtEq:
		left, right, err := vm.popTwoInts()
		if err != nil {
			return -1, err
		}
		vm.pushStack(&vval.VeniceBoolean{left.Value <= right.Value})
	case *bytecode.BinaryMapIndex:
		index := vm.popStack()
		vMapAny := vm.popStack()
		vMap, ok := vMapAny.(*vval.VeniceMap)
		if !ok {
			return -1, &ExecutionError{fmt.Sprintf("BINARY_MAP_INDEX requires map on top of stack, got %s (%T)", vMapAny.String(), vMapAny)}
		}

		for _, pair := range vMap.Pairs {
			if pair.Key.Equals(index) {
				vm.pushStack(pair.Value)
				return 1, nil
			}
		}

		// TODO(2021-08-03): Return a proper error value.
		vm.pushStack(&vval.VeniceInteger{-1})
	case *bytecode.BinaryMul:
		left, right, err := vm.popTwoInts()
		if err != nil {
			return -1, err
		}
		vm.pushStack(&vval.VeniceInteger{left.Value * right.Value})
	case *bytecode.BinaryNotEq:
		right := vm.popStack()
		left := vm.popStack()
		result := left.Equals(right)
		vm.pushStack(&vval.VeniceBoolean{!result})
	case *bytecode.BinaryOr:
		left, right, err := vm.popTwoBools()
		if err != nil {
			return -1, err
		}
		vm.pushStack(&vval.VeniceBoolean{left.Value || right.Value})
	case *bytecode.BinaryStringIndex:
		indexAny := vm.popStack()
		index, ok := indexAny.(*vval.VeniceInteger)
		if !ok {
			return -1, &ExecutionError{fmt.Sprintf("BINARY_STRING_INDEX requires integer on top of stack, got %s (%T)", indexAny.String(), indexAny)}
		}

		stringAny := vm.popStack()
		str, ok := stringAny.(*vval.VeniceString)
		if !ok {
			return -1, &ExecutionError{fmt.Sprintf("BINARY_STRING_INDEX requires string on top of stack, got %s (%T)", stringAny.String(), stringAny)}
		}

		if index.Value < 0 || index.Value >= len(str.Value) {
			return -1, &ExecutionError{"index out of bounds"}
		}

		vm.pushStack(&vval.VeniceCharacter{str.Value[index.Value]})
	case *bytecode.BinarySub:
		left, right, err := vm.popTwoInts()
		if err != nil {
			return -1, err
		}
		vm.pushStack(&vval.VeniceInteger{left.Value - right.Value})
	case *bytecode.BuildClass:
		values := []vval.VeniceValue{}
		for i := 0; i < bcode.N; i++ {
			topOfStack := vm.popStack()
			values = append(values, topOfStack)
		}

		vm.pushStack(&vval.VeniceClassObject{ClassName: bcode.Name, Values: values})
	case *bytecode.BuildList:
		values := []vval.VeniceValue{}
		for i := 0; i < bcode.N; i++ {
			topOfStack := vm.popStack()
			values = append(values, topOfStack)
		}
		vm.pushStack(&vval.VeniceList{values})
	case *bytecode.BuildMap:
		pairs := []*vval.VeniceMapPair{}
		for i := 0; i < bcode.N; i++ {
			value := vm.popStack()
			key := vm.popStack()
			pairs = append(pairs, &vval.VeniceMapPair{key, value})
		}
		vm.pushStack(&vval.VeniceMap{pairs})
	case *bytecode.BuildTuple:
		values := []vval.VeniceValue{}
		for i := 0; i < bcode.N; i++ {
			topOfStack := vm.popStack()
			values = append(values, topOfStack)
		}
		vm.pushStack(&vval.VeniceTuple{values})
	case *bytecode.CallBuiltin:
		functionName, ok := vm.popStack().(*vval.VeniceString)
		if !ok {
			return -1, &ExecutionError{"expected string at top of virtual machine stack for CALL_BUILTIN"}
		}

		builtin, ok := builtins[functionName.Value]
		if !ok {
			return -1, &ExecutionError{fmt.Sprintf("unknown builtin `%s`", functionName.Value)}
		}

		args := []vval.VeniceValue{}
		for i := 0; i < bcode.N; i++ {
			topOfStack := vm.popStack()
			args = append(args, topOfStack)
		}

		result := builtin(args...)
		if result != nil {
			vm.pushStack(result)
		}
	case *bytecode.CallFunction:
		functionName, ok := vm.popStack().(*vval.VeniceString)
		if !ok {
			return -1, &ExecutionError{"expected string at top of virtual machine stack for CALL_FUNCTION"}
		}

		functionEnv := &Environment{vm.Env, map[string]vval.VeniceValue{}}
		functionStack := vm.Stack[len(vm.Stack)-bcode.N:]
		vm.Stack = vm.Stack[:len(vm.Stack)-bcode.N]

		if debug {
			fmt.Println("DEBUG: Calling function in child virtual machine")
		}

		functionVm := &VirtualMachine{functionStack, functionEnv}
		value, err := functionVm.executeFunction(compiledProgram, functionName.Value, debug)
		if err != nil {
			return -1, err
		}

		vm.pushStack(value)
	case *bytecode.ForIter:
		iter := vm.peekStack().(vval.VeniceIterator)

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
		case *vval.VeniceList:
			vm.pushStack(&vval.VeniceListIterator{List: topOfStack, Index: 0})
		case *vval.VeniceMap:
			vm.pushStack(&vval.VeniceMapIterator{Map: topOfStack, Index: 0})
		}
	case *bytecode.LookupMethod:
		switch topOfStack := vm.peekStack().(type) {
		case *vval.VeniceClassObject:
			vm.pushStack(&vval.VeniceString{fmt.Sprintf("%s__%s", topOfStack.ClassName, bcode.Name)})
		case *vval.VeniceList:
			vm.pushStack(&vval.VeniceString{fmt.Sprintf("list__%s", bcode.Name)})
		case *vval.VeniceString:
			vm.pushStack(&vval.VeniceString{fmt.Sprintf("string__%s", bcode.Name)})
		default:
			return -1, &ExecutionError{"expected class, list, map, or string object at top of virtual machine stack for LOOKUP_METHOD"}
		}
	case *bytecode.PushConstBool:
		vm.pushStack(&vval.VeniceBoolean{bcode.Value})
	case *bytecode.PushConstChar:
		vm.pushStack(&vval.VeniceCharacter{bcode.Value})
	case *bytecode.PushConstInt:
		vm.pushStack(&vval.VeniceInteger{bcode.Value})
	case *bytecode.PushConstStr:
		vm.pushStack(&vval.VeniceString{bcode.Value})
	case *bytecode.PushEnum:
		values := []vval.VeniceValue{}
		for i := 0; i < bcode.N; i++ {
			topOfStack := vm.popStack()
			values = append(values, topOfStack)
		}

		// Reverse the array since the values are popped off the stack in reverse order.
		for i, j := 0, len(values)-1; i < j; i, j = i+1, j-1 {
			values[i], values[j] = values[j], values[i]
		}

		vm.pushStack(&vval.VeniceEnumObject{bcode.Name, values})
	case *bytecode.PushField:
		topOfStack, ok := vm.popStack().(*vval.VeniceClassObject)
		if !ok {
			return -1, &ExecutionError{"expected class object at top of virtual machine stack for PUSH_FIELD"}
		}

		vm.pushStack(topOfStack.Values[bcode.Index])
	case *bytecode.PushName:
		value, ok := vm.Env.Get(bcode.Name)
		if !ok {
			return -1, &ExecutionError{fmt.Sprintf("undefined symbol: %s", bcode.Name)}
		}
		vm.pushStack(value)
	case *bytecode.PushTupleField:
		topOfStack, ok := vm.popStack().(*vval.VeniceTuple)
		if !ok {
			return -1, &ExecutionError{"expected tuple at top of virtual machine stack for PUSH_TUPLE_FIELD"}
		}

		vm.pushStack(topOfStack.Values[bcode.Index])
	case *bytecode.RelJump:
		return bcode.N, nil
	case *bytecode.RelJumpIfFalse:
		topOfStack, ok := vm.popStack().(*vval.VeniceBoolean)
		if !ok {
			return -1, &ExecutionError{"expected boolean at top of virtual machine stack for REL_JUMP_IF_FALSE"}
		}

		if !topOfStack.Value {
			return bcode.N, nil
		}
	case *bytecode.RelJumpIfFalseOrPop:
		topOfStack, ok := vm.peekStack().(*vval.VeniceBoolean)
		if !ok {
			return -1, &ExecutionError{"expected boolean at top of virtual machine stack for REL_JUMP_IF_FALSE_OR_POP"}
		}

		if !topOfStack.Value {
			return bcode.N, nil
		} else {
			vm.popStack()
		}
	case *bytecode.RelJumpIfTrueOrPop:
		topOfStack, ok := vm.peekStack().(*vval.VeniceBoolean)
		if !ok {
			return -1, &ExecutionError{"expected boolean at top of virtual machine stack for REL_JUMP_IF_TRUE_OR_POP"}
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

		destination, ok := destinationAny.(*vval.VeniceClassObject)
		if !ok {
			return -1, &ExecutionError{"expected class object at top of virtual machine stack for STORE_FIELD"}
		}

		destination.Values[bcode.Index] = value
	case *bytecode.StoreName:
		topOfStack := vm.popStack()
		vm.Env.Put(bcode.Name, topOfStack)
	case *bytecode.UnaryMinus:
		topOfStack, ok := vm.popStack().(*vval.VeniceInteger)
		if !ok {
			return -1, &ExecutionError{"expected integer at top of virtual machine stack for UNARY_MINUS"}
		}

		topOfStack.Value = -topOfStack.Value
		vm.pushStack(topOfStack)
	case *bytecode.UnaryNot:
		topOfStack, ok := vm.popStack().(*vval.VeniceBoolean)
		if !ok {
			return -1, &ExecutionError{"expected boolean at top of virtual machine stack for UNARY_NOT"}
		}

		topOfStack.Value = !topOfStack.Value
		vm.pushStack(topOfStack)
	default:
		return -1, &ExecutionError{fmt.Sprintf("unknown bytecode instruction: %T", bcode)}
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

func (vm *VirtualMachine) peekStack() vval.VeniceValue {
	return vm.Stack[len(vm.Stack)-1]
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
