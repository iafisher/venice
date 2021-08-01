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

func (vm *VirtualMachine) Execute(program []*Bytecode) (VeniceValue, bool) {
	index := 0
	for index < len(program) {
		bytecode := program[index]
		jump, ok := vm.executeOne(bytecode)
		if !ok {
			return nil, false
		}

		index += jump
	}

	if len(vm.stack) > 0 {
		return vm.stack[len(vm.stack)-1], true
	} else {
		return nil, true
	}
}

func (vm *VirtualMachine) executeOne(bytecode *Bytecode) (int, bool) {
	switch bytecode.Name {
	case "BINARY_ADD":
		left, right, ok := vm.popTwoInts()
		if !ok {
			return -1, false
		}
		vm.pushStack(&VeniceInteger{left.Value + right.Value})
	case "BINARY_DIV":
		left, right, ok := vm.popTwoInts()
		if !ok {
			return -1, false
		}
		vm.pushStack(&VeniceInteger{left.Value / right.Value})
	case "BINARY_MUL":
		left, right, ok := vm.popTwoInts()
		if !ok {
			return -1, false
		}
		vm.pushStack(&VeniceInteger{left.Value * right.Value})
	case "BINARY_SUB":
		left, right, ok := vm.popTwoInts()
		if !ok {
			return -1, false
		}
		vm.pushStack(&VeniceInteger{left.Value - right.Value})
	case "CALL_BUILTIN":
		if v, ok := bytecode.Args[0].(*VeniceString); ok {
			switch v.Value {
			case "print":
				fmt.Println(vm.popStack().Serialize())
			default:
				return -1, false
			}
		} else {
			return -1, false
		}
	case "PUSH_CONST":
		vm.pushStack(bytecode.Args[0])
	case "PUSH_NAME":
		symbol := bytecode.Args[0].(*VeniceString).Value
		value, ok := vm.env.Get(symbol)
		if !ok {
			return -1, false
		}
		vm.pushStack(value)
	case "STORE_NAME":
		symbol := bytecode.Args[0].(*VeniceString).Value
		vm.env.Put(symbol, vm.popStack())
	default:
		return -1, false
	}
	return 1, true
}

func (vm *VirtualMachine) pushStack(values ...VeniceValue) {
	vm.stack = append(vm.stack, values...)
}

func (vm *VirtualMachine) popStack() VeniceValue {
	ret := vm.stack[len(vm.stack)-1]
	vm.stack = vm.stack[:len(vm.stack)-1]
	return ret
}

func (vm *VirtualMachine) popTwoInts() (*VeniceInteger, *VeniceInteger, bool) {
	right, ok1 := vm.popStack().(*VeniceInteger)
	left, ok2 := vm.popStack().(*VeniceInteger)
	return left, right, ok1 && ok2
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
