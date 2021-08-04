package main

import "fmt"

type Environment struct {
	parent  *Environment
	symbols map[string]VeniceValue
}

type VirtualMachine struct {
	stack []VeniceValue
	env   *Environment
}

func NewVirtualMachine() *VirtualMachine {
	return &VirtualMachine{[]VeniceValue{}, &Environment{nil, make(map[string]VeniceValue)}}
}

type ExecutionError struct {
	Message string
}

func (e *ExecutionError) Error() string {
	return e.Message
}

func NewEmptyStackError() *ExecutionError {
	return &ExecutionError{"virtual machine stack is empty"}
}

func (vm *VirtualMachine) Execute(program []*Bytecode) (VeniceValue, error) {
	index := 0
	for index < len(program) {
		bytecode := program[index]
		jump, err := vm.executeOne(bytecode)
		if err != nil {
			return nil, err
		}

		if jump == 0 {
			break
		}

		index += jump
	}

	if len(vm.stack) > 0 {
		ret := vm.stack[len(vm.stack)-1]
		vm.stack = nil
		return ret, nil
	} else {
		return nil, nil
	}
}

func (vm *VirtualMachine) executeOne(bytecode *Bytecode) (int, error) {
	switch bytecode.Name {
	case "BINARY_ADD":
		left, right, err := vm.popTwoInts()
		if err != nil {
			return -1, err
		}
		vm.pushStack(&VeniceInteger{left.Value + right.Value})
	case "BINARY_DIV":
		left, right, err := vm.popTwoInts()
		if err != nil {
			return -1, err
		}
		vm.pushStack(&VeniceInteger{left.Value / right.Value})
	case "BINARY_EQ":
		right, ok := vm.popStack()
		if !ok {
			return -1, NewEmptyStackError()
		}

		left, ok := vm.popStack()
		if !ok {
			return -1, NewEmptyStackError()
		}

		result := left.Equals(right)
		vm.pushStack(&VeniceBoolean{result})
	case "BINARY_LIST_INDEX":
		indexUntyped, ok := vm.popStack()
		if !ok {
			return -1, NewEmptyStackError()
		}

		index, ok := indexUntyped.(*VeniceInteger)
		if !ok {
			return -1, &ExecutionError{fmt.Sprintf("BINARY_LIST_INDEX requires integer on top of stack, got %s (%T)", indexUntyped.Serialize(), indexUntyped)}
		}

		listUntyped, ok := vm.popStack()
		if !ok {
			return -1, NewEmptyStackError()
		}

		list, ok := listUntyped.(*VeniceList)
		if !ok {
			return -1, &ExecutionError{fmt.Sprintf("BINARY_LIST_INDEX requires list on top of stack, got %s (%T)", listUntyped.Serialize(), listUntyped)}
		}

		// TODO(2021-08-03): Handle out-of-bounds index.
		vm.pushStack(list.Values[index.Value])
	case "BINARY_MAP_INDEX":
		index, ok := vm.popStack()
		if !ok {
			return -1, NewEmptyStackError()
		}

		vMapUntyped, ok := vm.popStack()
		if !ok {
			return -1, NewEmptyStackError()
		}

		vMap, ok := vMapUntyped.(*VeniceMap)
		if !ok {
			return -1, &ExecutionError{fmt.Sprintf("BINARY_MAP_INDEX requires map on top of stack, got %s (%T)", vMapUntyped.Serialize(), vMapUntyped)}
		}

		for _, pair := range vMap.Pairs {
			if pair.Key.Equals(index) {
				vm.pushStack(pair.Value)
				return 1, nil
			}
		}

		// TODO(2021-08-03): Return a proper error value.
		vm.pushStack(&VeniceInteger{-1})
	case "BINARY_MUL":
		left, right, err := vm.popTwoInts()
		if err != nil {
			return -1, err
		}
		vm.pushStack(&VeniceInteger{left.Value * right.Value})
	case "BINARY_SUB":
		left, right, err := vm.popTwoInts()
		if err != nil {
			return -1, err
		}
		vm.pushStack(&VeniceInteger{left.Value - right.Value})
	case "BUILD_LIST":
		values := []VeniceValue{}
		n := bytecode.Args[0].(*VeniceInteger).Value
		for i := 0; i < n; i++ {
			topOfStack, ok := vm.popStack()
			if !ok {
				return -1, NewEmptyStackError()
			}
			values = append(values, topOfStack)
		}
		vm.pushStack(&VeniceList{values})
	case "BUILD_MAP":
		pairs := []*VeniceMapPair{}
		n := bytecode.Args[0].(*VeniceInteger).Value
		for i := 0; i < n; i++ {
			value, ok := vm.popStack()
			if !ok {
				return -1, NewEmptyStackError()
			}
			key, ok := vm.popStack()
			if !ok {
				return -1, NewEmptyStackError()
			}
			pairs = append(pairs, &VeniceMapPair{key, value})
		}
		vm.pushStack(&VeniceMap{pairs})
	case "CALL_BUILTIN":
		if v, ok := bytecode.Args[0].(*VeniceString); ok {
			switch v.Value {
			case "print":
				topOfStack, ok := vm.popStack()
				if !ok {
					return -1, NewEmptyStackError()
				}
				fmt.Println(topOfStack.SerializePrintable())
			default:
				return -1, &ExecutionError{fmt.Sprintf("unknown builtin: %s", v.Value)}
			}
		} else {
			return -1, &ExecutionError{"argument to CALL_BUILTIN must be a string"}
		}
	case "CALL_FUNCTION":
		if v, ok := bytecode.Args[0].(*VeniceInteger); ok {
			n := v.Value

			topOfStack, ok := vm.popStack()
			if !ok {
				return -1, NewEmptyStackError()
			}

			if function, ok := topOfStack.(*VeniceFunction); ok {
				fEnv := &Environment{vm.env, map[string]VeniceValue{}}
				for i := 0; i < n; i++ {
					topOfStack, ok = vm.popStack()
					if !ok {
						return -1, NewEmptyStackError()
					}

					fEnv.symbols[function.Params[len(function.Params)-(i+1)]] = topOfStack
				}

				subVm := NewVirtualMachine()
				subVm.env = fEnv
				val, err := subVm.Execute(function.Body)
				if err != nil {
					return -1, nil
				}
				vm.pushStack(val)
			} else {
				return -1, &ExecutionError{"CALL_FUNCTION requires function on top of stack"}
			}
		} else {
			return -1, &ExecutionError{"argument to CALL_FUNCTION must be an integer"}
		}
	case "PUSH_CONST":
		vm.pushStack(bytecode.Args[0])
	case "PUSH_NAME":
		symbol := bytecode.Args[0].(*VeniceString).Value
		value, ok := vm.env.Get(symbol)
		if !ok {
			return -1, &ExecutionError{fmt.Sprintf("undefined symbol: %s", symbol)}
		}
		vm.pushStack(value)
	case "REL_JUMP":
		value := bytecode.Args[0].(*VeniceInteger).Value
		return value, nil
	case "REL_JUMP_IF_FALSE":
		topOfStack, ok := vm.popStack()
		if !ok {
			return -1, NewEmptyStackError()
		}

		topOfStackAsBool, ok := topOfStack.(*VeniceBoolean)
		if !ok {
			return -1, &ExecutionError{"expected boolean at top of virtual machine stack"}
		}

		if !topOfStackAsBool.Value {
			value := bytecode.Args[0].(*VeniceInteger).Value
			return value, nil
		}
	case "RETURN":
		return 0, nil
	case "STORE_NAME":
		symbol := bytecode.Args[0].(*VeniceString).Value
		topOfStack, ok := vm.popStack()
		if !ok {
			return -1, NewEmptyStackError()
		}
		vm.env.Put(symbol, topOfStack)
	default:
		return -1, &ExecutionError{fmt.Sprintf("unknown bytecode instruction: %s", bytecode.Name)}
	}
	return 1, nil
}

func (vm *VirtualMachine) pushStack(values ...VeniceValue) {
	vm.stack = append(vm.stack, values...)
}

func (vm *VirtualMachine) popStack() (VeniceValue, bool) {
	if len(vm.stack) == 0 {
		return nil, false
	}

	ret := vm.stack[len(vm.stack)-1]
	vm.stack = vm.stack[:len(vm.stack)-1]
	return ret, true
}

func (vm *VirtualMachine) popTwoInts() (*VeniceInteger, *VeniceInteger, error) {
	right, ok := vm.popStack()
	if !ok {
		return nil, nil, NewEmptyStackError()
	}

	rightAsInteger, ok := right.(*VeniceInteger)
	if !ok {
		return nil, nil, &ExecutionError{"expected integer at top of virtual machine stack"}
	}

	left, ok := vm.popStack()
	if !ok {
		return nil, nil, NewEmptyStackError()
	}

	leftAsInteger, ok := left.(*VeniceInteger)
	if !ok {
		return nil, nil, &ExecutionError{"expected integer at top of virtual machine stack"}
	}

	return leftAsInteger, rightAsInteger, nil
}

func (env *Environment) Get(symbol string) (VeniceValue, bool) {
	value, ok := env.symbols[symbol]
	if !ok {
		if env.parent != nil {
			return env.parent.Get(symbol)
		} else {
			return nil, false
		}
	}
	return value, true
}

func (env *Environment) Put(symbol string, value VeniceValue) {
	env.symbols[symbol] = value
}
