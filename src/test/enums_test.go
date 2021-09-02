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
