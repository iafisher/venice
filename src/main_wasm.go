package main

import "syscall/js"

func main() {
	c := make(chan struct{}, 0)
	js.Global().Set("veniceExecute", js.FuncOf(func(this js.Value, p []js.Value) interface{} {
		vm := NewVirtualMachine()
		compiler := NewCompiler()
		line := p[0].String()

		tree, err := NewParser(NewLexer(line)).Parse()
		if err != nil {
			return js.ValueOf(nil)
		}

		bytecodes, err := compiler.Compile(tree)
		if err != nil {
			return js.ValueOf(nil)
		}

		value, err := vm.Execute(bytecodes)
		if err != nil {
			return js.ValueOf(nil)
		}

		if value != nil {
			return js.ValueOf(value.Serialize())
		} else {
			return js.ValueOf(nil)
		}
	}))
	<-c
}
