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
