package test

import (
	"testing"
)

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

	assertEqual(
		t,
		`
		let m = {"Canada": "Ottawa"}
		let mCopy = m.copy()
		m["Canada"] = "Toronto"
		mCopy["Canada"]
		`,
		Some(S("Ottawa")),
	)

	assertEqual(
		t,
		`
		let m = {"Guatemala": 14000000}
		m.clear()
		length(m)
		`,
		I(0),
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
