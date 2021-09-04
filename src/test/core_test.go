package test

import (
	"github.com/iafisher/venice/src/vm"
	"testing"
)

func TestAndOr(t *testing.T) {
	// Make sure the `and` and `or` short-circuit before evaluating an expression that
	// would cause an error.
	assertEqual(t, `let l = [1]; l.size() >= 2 and l[1] == 42`, B(false))
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
		let f = int
		f(1.0)
		`,
		I(1),
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

	assertTypecheckError(
		t,
		`
		func f(x: int) -> int {
			return x + 1
		}
		x
		`,
		"undefined symbol `x`",
	)
}

func TestIndexing(t *testing.T) {
	assertEqual(t, `let l = [1, 2, 3]; l[1]`, I(2))
	assertEqual(t, `{1: "one", 2: "two", 3: "three"}[3]`, Some(S("three")))
	assertEqual(t, `let s = "123"; s[1]`, S("2"))
	assertEqual(t, `"Привет"[0]`, S("П"))
	assertEqual(t, `"Привет"[5]`, S("т"))
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
	assertTypecheckError(
		t,
		`
		let l = []
		l
		`,
		"empty list has unknown type",
	)
	assertEqual(
		t,
		`
		let l: [int] = []
		l
		`,
		L(),
	)
	assertTypecheckError(
		t,
		`
		let l: string = []
		l
		`,
		"empty list has unknown type",
	)
	assertTypecheckError(
		t,
		`
		let m = {}
		m
		`,
		"empty map has unknown type",
	)
	assertEqual(
		t,
		`
		let m: {int: int} = {}
		m
		`,
		vm.NewVeniceMap(),
	)
	assertTypecheckError(
		t,
		`
		if (true) {
			let x = 10
		}
		x
		`,
		"undefined symbol `x`",
	)
}

func TestListBuiltins(t *testing.T) {
	assertEqual(t, `let l = [1, 2, 3]; l.size()`, I(3))
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
	assertEqual(
		t,
		`
		let l = [1, 2, 3]
		let l2 = l.copy()
		l2[0] = 10
		l
		`,
		L(I(1), I(2), I(3)),
	)
	assertEqual(
		t,
		`
		let l = [4, 3, 2, 1]
		l.sort_in_place()
		l
		`,
		L(I(1), I(2), I(3), I(4)),
	)
	assertEqual(
		t,
		`
		let l = ["Venezuela", "Paraguay", "Bolivia", "Argentina"]
		l.sorted()
		`,
		L(S("Argentina"), S("Bolivia"), S("Paraguay"), S("Venezuela")),
	)
	assertEqual(
		t,
		`
		let l = ["Venezuela", "Paraguay", "Bolivia", "Argentina"]
		let l2 = l.sorted()
		l
		`,
		L(S("Venezuela"), S("Paraguay"), S("Bolivia"), S("Argentina")),
	)
	assertEqual(
		t,
		`
		let l = [4, 3, 2, 1]
		l.reversed()
		`,
		L(I(1), I(2), I(3), I(4)),
	)
	assertEqual(
		t,
		`
		let l = [4, 3, 2, 1]
		l.reverse_in_place()
		l
		`,
		L(I(1), I(2), I(3), I(4)),
	)
	assertEqual(
		t,
		`
		let l = range(0, 100)
		l.find(14)
		`,
		Some(I(14)),
	)
	assertEqual(
		t,
		`
		let l = range(0, 100)
		l.find(-1)
		`,
		None(),
	)
	assertEqual(
		t,
		`
		let l = [4, 0, 4]
		l.find_last(4)
		`,
		Some(I(2)),
	)
	assertEqual(
		t,
		`
		let l = range(0, 100)
		l.find_last(-1)
		`,
		None(),
	)
	assertEqual(
		t,
		`
		let l = ["a", "b", "c"]
		l.join(" ")
		`,
		S("a b c"),
	)
	/* TODO(#115):
	assertTypecheckError(
		t,
		`[1, 2, 3].join(" ")`,
		"`join` is only available for lists of strings",
	)
	*/
	assertTypecheckError(
		t,
		`
		let l = [1, 2, 3]
		l.append("4")
		`,
		"wrong function parameter type",
	)
}

func TestListIndexAssignment(t *testing.T) {
	assertEqual(t, `let l = [1]; l[0] = 42; l[0]`, I(42))
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
