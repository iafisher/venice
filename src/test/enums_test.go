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

func TestEnumAsFunctionParameter(t *testing.T) {
	/* TODO(2021-09-11)
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
		unwrap(Optional::Some(10)
		`,
		I(10),
	)
	*/
}
