package test

import (
	"testing"
)

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

func TestEnumRedeclaration(t *testing.T) {
	assertTypecheckError(
		t,
		`
		enum MyBoolean { True, False }
		enum MyBoolean { True, False, Maybe }
		`,
		"symbol `MyBoolean` is already defined",
	)

	// Functions and types are in different namespaces, so identical names don't conflict
	// with one another.
	assertEqual(
		t,
		`
		enum MyBoolean { True, False }

		func MyBoolean() -> bool {
			return true
		}

		MyBoolean()
		`,
		B(true),
	)
}

func TestEnumAsFunctionParameter(t *testing.T) {
	/* TODO(2021-09-12)
	assertEqual(
		t,
		`
		func unwrap(o: Optional<int>) -> int {
			match o {
				case Some(x) {
					return x
				}
				case None {
					return -1
				}
			}
		}
		(unwrap(Optional::Some(10)), unwrap(Optional::None))
		`,
		Tup(I(10), I(-1)),
	)
	*/
}
