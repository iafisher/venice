package main

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
	case "PUSH_CONST":
		vm.pushStack(bytecode.Args[0])
	default:
		return -1, false
	}
	return 1, true
}

func (vm *VirtualMachine) pushStack(values ...VeniceValue) {
	vm.stack = append(vm.stack, values...)
}
