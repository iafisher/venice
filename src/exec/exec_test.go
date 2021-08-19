package exec

import (
	"github.com/iafisher/venice/src/compiler"
	"github.com/iafisher/venice/src/parser"
	vm_mod "github.com/iafisher/venice/src/vm"
	"github.com/iafisher/venice/src/vval"
	"testing"
)

func TestAndOr(t *testing.T) {
	// Make sure the `and` and `or` short-circuit before evaluating an expression that
	// would cause an error.
	assertEqual(t, `let l = [1]; length(l) >= 2 and l[1] == 42`, B(false))
	assertEqual(t, `true or [0][1] == 10`, B(true))
}

func TestConcat(t *testing.T) {
	assertEqual(t, `"abc" ++ "def"`, S("abcdef"))
	assertEqual(t, `[1, 2, 3] ++ [4, 5, 6]`, L(I(1), I(2), I(3), I(4), I(5), I(6)))
}

func TestIndexing(t *testing.T) {
	assertEqual(t, `let l = [1, 2, 3]; l[1]`, I(2))
	assertEqual(t, `{1: "one", 2: "two", 3: "three"}[3]`, S("three"))
	assertEqual(t, `let s = "123"; s[1]`, C('2'))
}

func TestLength(t *testing.T) {
	assertEqual(t, `length([1, 2, 3])`, I(3))
	assertEqual(t, `length("abcdef")`, I(6))
	// assertEqual(t, `length([])`, I(0))
	assertEqual(t, `length("")`, I(0))
}

func TestTernaryIf(t *testing.T) {
	assertEqual(t, `(41 if true else 665) + 1`, I(42))
}

func TestTuples(t *testing.T) {
	assertEqual(t, `let t = (1, "two", [3]); t`, Tup(I(1), S("two"), L(I(3))))
	assertEqual(t, `let t = (1, "two", [3]); t.1`, S("two"))
}

func B(b bool) *vval.VeniceBoolean {
	return &vval.VeniceBoolean{b}
}

func C(ch byte) *vval.VeniceCharacter {
	return &vval.VeniceCharacter{ch}
}

func L(values ...vval.VeniceValue) *vval.VeniceList {
	return &vval.VeniceList{values}
}

func I(n int) *vval.VeniceInteger {
	return &vval.VeniceInteger{n}
}

func S(s string) *vval.VeniceString {
	return &vval.VeniceString{s}
}

func Tup(values ...vval.VeniceValue) *vval.VeniceTuple {
	return &vval.VeniceTuple{values}
}

func assertEqual(t *testing.T, program string, result vval.VeniceValue) {
	parsedFile, err := parser.NewParser().ParseString(program)
	if err != nil {
		t.Fatalf("Parse error: %s\n\nInput: %q", err, program)
	}

	compiler := compiler.NewCompiler()
	compiledProgram, err := compiler.Compile(parsedFile)
	if err != nil {
		t.Fatalf("Compile error: %s\n\nInput: %q", err, program)
	}

	vm := vm_mod.NewVirtualMachine()
	value, err := vm.Execute(compiledProgram, false)
	if err != nil {
		t.Fatalf("Execution error: %s\n\nInput: %q", err, program)
	}

	if !value.Equals(result) {
		t.Fatalf("Expected %s, got %s\n\nInput: %q", result.String(), value.String(), program)
	}
}
