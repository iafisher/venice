package test

import (
	"github.com/iafisher/venice/src/vm"
	"testing"
)

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

func TestFunctionCallTypeInferenceForEmptyContainers(t *testing.T) {
	assertEqual(
		t,
		`
		func get_first_or_zero(items: [int]) -> int {
			return items[0] if items.size() > 0 else 0
		}
		get_first_or_zero([])
		`,
		I(0),
	)

	assertEqual(
		t,
		`
		func get_value_or_zero(items: {int: int}) -> int {
			match items[0] {
				case Some(x) {
					return x
				}
				case None {
					return 0
				}
			}
		}
		get_value_or_zero({})
		`,
		I(0),
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

	assertEqual(
		t,
		`
		func empty() -> [int] {
			return []
		}
		empty()
		`,
		L(),
	)

	assertEqual(
		t,
		`
		func empty_map() -> {int: int} {
			return {}
		}
		empty_map()
		`,
		vm.NewVeniceMap(),
	)

	assertEqual(
		t,
		`
		func f() -> int {
			if true {
				return 42
			} else {
				return -1
			}
		}
		f()
		`,
		I(42),
	)

	assertEqual(
		t,
		`
		func f(o: Optional<int>) -> int {
			match o {
				case Some(x) {
					return x
				}
				case None {
					return -1
				}
			}
		}
		f(Optional::Some(10))
		`,
		I(10),
	)

	assertTypecheckError(
		t,
		`
		func f() -> int {
			let x = 42
		}
		`,
		"non-void function does not end with return statement",
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

func TestFunctionWithWrongReturnType(t *testing.T) {
	assertTypecheckError(
		t,
		`
		func f() -> int {
			return "abc"
		}
		`,
		"conflicting function return types: got string, expected int",
	)

	assertTypecheckError(
		t,
		`
		func f() -> int {
			if true {
				return 42
			} else {
				return "bleh"
			}
		}
		`,
		"conflicting function return types: got string, expected int",
	)
}

func TestFunctionWithoutReturn(t *testing.T) {
	assertTypecheckError(
		t,
		`
		func f() -> int {
			let x = 42
		}
		`,
		"non-void function does not end with return statement",
	)

	assertTypecheckError(
		t,
		`
		func f() -> int {
			if false {
				return 42
			}
		}
		`,
		"non-void function does not end with return statement",
	)

	assertTypecheckError(
		t,
		`
		func f() -> int {
			while true {
				return 42
			}
		}
		`,
		"non-void function does not end with return statement",
	)
}
