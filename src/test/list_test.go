package test

import (
	"testing"
)

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

func TestListIndexOutOfBounds(t *testing.T) {
	assertPanic(t, "let l = [1, 2, 3]; l[500]")
	assertPanic(t, "let l = [1, 2, 3]; l[500] = 4")
	assertPanic(t, "let l = [1, 2, 3]; l.remove(500)")
	assertPanic(t, "let l = [1, 2, 3]; l.slice(500, 600)")
}
