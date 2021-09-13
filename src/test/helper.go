package test

import (
	compilerPkg "github.com/iafisher/venice/src/compiler"
	"github.com/iafisher/venice/src/vm"
	"strings"
	"testing"
)

func B(b bool) *vm.VeniceBoolean {
	return &vm.VeniceBoolean{b}
}

func F(n float64) *vm.VeniceRealNumber {
	return &vm.VeniceRealNumber{n}
}

func L(values ...vm.VeniceValue) *vm.VeniceList {
	return &vm.VeniceList{values}
}

func I(n int) *vm.VeniceInteger {
	return &vm.VeniceInteger{n}
}

func S(s string) *vm.VeniceString {
	return &vm.VeniceString{s}
}

func Some(v vm.VeniceValue) *vm.VeniceEnumObject {
	return &vm.VeniceEnumObject{
		Label:  "Some",
		Values: []vm.VeniceValue{v},
	}
}

func None() *vm.VeniceEnumObject {
	return &vm.VeniceEnumObject{
		Label:  "None",
		Values: []vm.VeniceValue{},
	}
}

func Tup(values ...vm.VeniceValue) *vm.VeniceTuple {
	return &vm.VeniceTuple{values}
}

func assertEqual(t *testing.T, program string, result vm.VeniceValue) {
	parsedFile, err := compilerPkg.ParseString(program)
	if err != nil {
		t.Helper()
		t.Fatalf("Parse error: %s\n\nInput:\n\n%s", err, program)
	}

	compiler := compilerPkg.NewCompiler()
	compiledProgram, err := compiler.Compile(parsedFile)
	if err != nil {
		t.Helper()
		t.Fatalf("Compile error: %s\n\nInput:\n\n%s", err, program)
	}

	virtualMachine := vm.NewVirtualMachine()
	value, err := virtualMachine.Execute(compiledProgram, false)
	if err != nil {
		t.Helper()
		t.Fatalf("Execution error: %s\n\nInput:\n\n%s", err, program)
	}

	if value == nil {
		t.Helper()
		t.Fatalf("Code snippet did not return a value\n\nInput:\n\n%s", program)
	}

	if !value.Equals(result) {
		t.Helper()
		t.Fatalf(
			"Expected %s, got %s\n\nInput:\n\n%s",
			result.String(),
			value.String(),
			program,
		)
	}
}

func assertTypecheckError(t *testing.T, program string, errorMessage string) {
	parsedFile, err := compilerPkg.ParseString(program)
	if err != nil {
		t.Helper()
		t.Fatalf("Parse error: %s\n\nInput:\n\n%s", err, program)
	}

	compiler := compilerPkg.NewCompiler()
	_, err = compiler.Compile(parsedFile)
	if err == nil {
		t.Helper()
		t.Fatalf(
			"Expected compile error, but program compiled without error\n\nInput:\n\n%s",
			program,
		)
	}

	if !strings.Contains(err.Error(), errorMessage) {
		t.Helper()
		t.Fatalf(
			"Expected compile error to contain substring %q, but it did not\n\nError: %s\n\nInput:\n\n%s",
			errorMessage,
			err.Error(),
			program,
		)
	}
}

func assertPanic(t *testing.T, program string) {
	parsedFile, err := compilerPkg.ParseString(program)
	if err != nil {
		t.Helper()
		t.Fatalf("Parse error: %s\n\nInput:\n\n%s", err, program)
	}

	compiler := compilerPkg.NewCompiler()
	compiledProgram, err := compiler.Compile(parsedFile)
	if err != nil {
		t.Helper()
		t.Fatalf("Compile error: %s\n\nInput:\n\n%s", err, program)
	}

	virtualMachine := vm.NewVirtualMachine()
	_, err = virtualMachine.Execute(compiledProgram, false)
	if err == nil {
		t.Helper()
		t.Fatalf("Expected panic, but program ran without panicking: %s\n\nInput:\n\n%s", err, program)
	}

	if _, ok := err.(*vm.PanicError); !ok {
		t.Helper()
		t.Fatalf("Expected panic, but program suffered an internal error: %s\n\nInput:\n\n%s", err, program)
	}
}
