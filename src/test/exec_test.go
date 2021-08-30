package test

import (
	compilerPkg "github.com/iafisher/venice/src/compiler"
	"github.com/iafisher/venice/src/vm"
	"strings"
	"testing"
)

func TestAndOr(t *testing.T) {
	// Make sure the `and` and `or` short-circuit before evaluating an expression that
	// would cause an error.
	assertEqual(t, `let l = [1]; length(l) >= 2 and l[1] == 42`, B(false))
	assertEqual(t, `true or [0][1] == 10`, B(true))
}

func TestAssignStatements(t *testing.T) {
	assertEqual(
		t,
		`
		var i = 0
		i = 42
		i
		`,
		I(42),
	)

	assertEqual(
		t,
		`
		var i = 41
		i += 1
		i
		`,
		I(42),
	)

	assertEqual(
		t,
		`
		var i = 21
		i *= 2
		i
		`,
		I(42),
	)

	assertEqual(
		t,
		`
		var i = 44
		i -= 2
		i
		`,
		I(42),
	)

	assertEqual(
		t,
		`
		var i = 126.0
		i /= 3
		i
		`,
		F(42.0),
	)
}

func TestBuiltinFunctions(t *testing.T) {
	assertEqual(t, `range(0, 5)`, L(I(0), I(1), I(2), I(3), I(4)))
	assertEqual(t, `range(10, 5)`, L())
	assertEqual(t, `int(1.0)`, I(1))
	assertEqual(t, `real(1)`, F(1.0))
}

func TestBreakStatement(t *testing.T) {
	assertEqual(
		t,
		`
		var i = 41
		while true {
			i += 1
			break
		}
		i
		`,
		I(42),
	)
	assertTypecheckError(t, "break", "break statement outside of loop")
}

func TestClassConstructor(t *testing.T) {
	assertEqual(
		t,
		`
		class Point {
			public x: int
			public y: int
		}

		let p = new Point { x: 1, y: 2 }
		(p.x, p.y)
		`,
		Tup(I(1), I(2)),
	)
	/*
		assertEqual(
			t,
			`
			class SurpriseBox no constructor {
				private value: int

				constructor(self) {
					self.value = 42
					return self
				}

				public func get_value(self) -> int {
					return self.value
				}
			}

			let box = SurpriseBox()
			box.get_value()
			`,
			I(42),
		)

		assertTypecheckError(
			t,
			`
			class SurpriseBox no constructor {
				private value: int

				constructor(self) {}
			}
			`,
			"field `value` not set in constructor",
		)
	*/
}

func TestClassEquality(t *testing.T) {
	assertEqual(
		t,
		`
		class Box {
			public value: int
		}

		let b1 = new Box { value: 42 }
		let b2 = new Box { value: 42 }
		b1 == b2
		`,
		B(true),
	)

	assertTypecheckError(
		t,
		`
		class Box1 {
			public value: int
		}

		class Box2 {
			public value: int
		}

		let b1 = new Box1 { value: 42 }
		let b2 = new Box2 { value: 42 }
		b1 == b2
		`,
		"invalid type for right operand of ==",
	)
}

func TestClassDeclaration(t *testing.T) {
	assertTypecheckError(
		t,
		`
		class SecretBox {
			private secret: int
		}
		let box = new SecretBox { secret: 42 }
		box.secret
		`,
		"use of private field",
	)
}

func TestClassFieldAssignment(t *testing.T) {
	assertEqual(
		t,
		`
		class Box {
		  public value: int
		}

		let b = new Box { value: 0 }
		b.value = 42
		b.value
		`,
		I(42),
	)
	assertTypecheckError(
		t,
		`
		class Box {
		  public value: int
		}

		let b = new Box { value: 0 }
		b.value = "42"
		`,
		"expected type int, got string",
	)
	/*
		assertTypecheckError(
			t,
			`
			class Box<T> {
			  public object: T
			}

			let x = Box("41")
			x + 1
			`,
			"invalid type for left operand of +",
		)
	*/
}

func TestConcat(t *testing.T) {
	assertEqual(t, `"abc" ++ "def"`, S("abcdef"))
	assertEqual(t, `[1, 2, 3] ++ [4, 5, 6]`, L(I(1), I(2), I(3), I(4), I(5), I(6)))
}

func TestContinueStatement(t *testing.T) {
	assertEqual(
		t,
		`
		var i = 41
		while i != 42 {
			i += 1
			continue
			i = 666
		}
		i
		`,
		I(42),
	)
	assertTypecheckError(t, "continue", "continue statement outside of loop")
}

func TestEnumDeclaration(t *testing.T) {
	assertTypecheckError(
		t,
		`
		enum MyBoolean { True, False }
		MyBoolean::Maybe
		`,
		"enum MyBoolean does not have case Maybe",
	)
	assertTypecheckError(
		t,
		`
		enum IntResult {
		  Some(int),
		  None,
		}

		IntResult::Some("hello")
		`,
		"wrong function parameter type",
	)
}

func TestFunctionCall(t *testing.T) {
	assertEqual(
		t,
		`
		func add_one(x: int) -> int {
			return x + 1
		}

		add_one(41)
		`,
		I(42),
	)
	assertTypecheckError(
		t,
		`
		func f(x: int) -> int {
			return x + 1
		}

		f("not an integer")
		`,
		"wrong function parameter type",
	)
	assertTypecheckError(
		t,
		`
		func add_one(x: int) -> int {
			return x + 1
		}

		x
		`,
		"undefined symbol",
	)
	assertTypecheckError(t, "int::y", "cannot use double colon after non-enum type")

	assertEqual(
		t,
		`
		let f = length
		f([1, 2, 3])
		`,
		I(3),
	)
}

func TestFunctionDeclaration(t *testing.T) {
	assertEqual(
		t,
		`
		func f() -> [int] {
			return [1, 2, 3]
		}
		f()
		`,
		L(I(1), I(2), I(3)),
	)

	assertTypecheckError(
		t,
		`
		func f() -> int {
			let x = 42
		}
		`,
		"non-void function has no return statement",
	)
}

func TestIndexing(t *testing.T) {
	assertEqual(t, `let l = [1, 2, 3]; l[1]`, I(2))
	assertEqual(t, `{1: "one", 2: "two", 3: "three"}[3]`, Some(S("three")))
	assertEqual(t, `let s = "123"; s[1]`, C('2'))
}

func TestLength(t *testing.T) {
	assertEqual(t, `length([1, 2, 3])`, I(3))
	assertEqual(t, `length("abcdef")`, I(6))
	// assertEqual(t, `length([])`, I(0))
	assertEqual(t, `length("")`, I(0))
}

func TestLetStatement(t *testing.T) {
	assertTypecheckError(
		t,
		`
		let x = 10
		let x = 11
		`,
		"re-declaration of symbol",
	)
	assertTypecheckError(
		t,
		`
		func f(x: int) {
		  let x = 10
		}
		`,
		"re-declaration of symbol",
	)

	assertEqual(
		t,
		`
		let x: int = 42
		x
		`,
		I(42),
	)
	assertTypecheckError(
		t,
		`
		let x: string = 42
		`,
		"expected string, got int",
	)
}

func TestListBuiltins(t *testing.T) {
	assertEqual(t, `let l = [1, 2, 3]; l.length()`, I(3))
	assertEqual(t, `let l = [1, 2]; l.append(3); l[2]`, I(3))
	assertEqual(t, `let l = [1, 2]; l.extend([3, 4]); l`, L(I(1), I(2), I(3), I(4)))
	assertEqual(
		t,
		`
		let l = [1, 2, 3];
		l.remove(1);
		l
		`,
		L(I(1), I(3)),
	)
	assertEqual(t, `[1, 2, 3].slice(1, 2)`, L(I(2)))
	assertEqual(
		t,
		`
		let l = [1, 2, 3];
		let l2 = l.slice(0, 3);
		l2.remove(2);
		(l, l2)
		`,
		Tup(L(I(1), I(2), I(3)), L(I(1), I(2))),
	)
}

func TestListIndexAssignment(t *testing.T) {
	assertEqual(t, `let l = [1]; l[0] = 42; l[0]`, I(42))
}

func TestMapBuiltins(t *testing.T) {
	assertEqual(
		t,
		`
		let m = {1: "one", 2: "two"}
		m.remove(2)
		2 in m
		`,
		B(false),
	)

	assertEqual(
		t,
		`
		let m = {1: "one", 2: "two"}
		m.keys()
		`,
		L(I(1), I(2)),
	)

	assertEqual(
		t,
		`
		let m = {1: "one", 2: "two"}
		m.values()
		`,
		L(S("one"), S("two")),
	)

	assertEqual(
		t,
		`
		let m = {1: "one", 2: "two"}
		m.entries()
		`,
		L(Tup(I(1), S("one")), Tup(I(2), S("two"))),
	)
}

func TestMapIndexAssignment(t *testing.T) {
	assertEqual(
		t,
		`
		let m = {"one": 1}
		m["forty two"] = 42
		m["forty two"]
		`,
		Some(I(42)),
	)
}

func TestMatchStatements(t *testing.T) {
	assertEqual(
		t,
		`
		var answer = 0
		match Optional::Some(42) {
			case Some(x) {
				answer = x
			}
			case None {
				answer = -1
			}
		}
		answer
		`,
		I(42),
	)

	assertEqual(
		t,
		`
		var answer = 0
		match Optional::None {
			case Some(x) {
			}
			case None {
				answer = 42
			}
		}
		answer
		`,
		I(42),
	)

	assertTypecheckError(
		t,
		`
		match "abc" {
			case Some(x) {}
			case None {}
		}
		`,
		"cannot match a non-enum type",
	)
}

func TestRealNumbers(t *testing.T) {
	assertEqual(t, `1 / 2`, F(0.5))
	assertEqual(t, `1 + 2.0`, F(3.0))
	assertEqual(t, `3.0 * 2.0`, F(6.0))
}

func TestStringBuiltins(t *testing.T) {
	assertEqual(t, `let s = "123"; s.length()`, I(3))
	assertEqual(t, `"123".length()`, I(3))
	assertEqual(t, `"abc".to_upper()`, S("ABC"))
	assertEqual(t, `"ABC".to_lower()`, S("abc"))
	assertEqual(t, `"ABCDE".slice(1, 3)`, S("BC"))
	assertEqual(t, `"colorless green ideas".split_space()`, L(S("colorless"), S("green"), S("ideas")))
	assertEqual(t, `"colorless green ideas".split(" ")`, L(S("colorless"), S("green"), S("ideas")))
}

func TestTernaryIf(t *testing.T) {
	assertEqual(t, `(41 if true else 665) + 1`, I(42))
}

func TestTuples(t *testing.T) {
	assertEqual(t, `let t = (1, "two", [3]); t`, Tup(I(1), S("two"), L(I(3))))
	assertEqual(t, `let t = (1, "two", [3]); t.1`, S("two"))
}

func TestVarStatements(t *testing.T) {
	assertEqual(
		t,
		`
		var i = 0
		i += 1
		i
		`,
		I(1),
	)
}

func B(b bool) *vm.VeniceBoolean {
	return &vm.VeniceBoolean{b}
}

func C(ch byte) *vm.VeniceCharacter {
	return &vm.VeniceCharacter{ch}
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

func Tup(values ...vm.VeniceValue) *vm.VeniceTuple {
	return &vm.VeniceTuple{values}
}

func assertEqual(t *testing.T, program string, result vm.VeniceValue) {
	parsedFile, err := compilerPkg.NewParser().ParseString(program)
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
	parsedFile, err := compilerPkg.NewParser().ParseString(program)
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
