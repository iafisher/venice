package test

import (
	"testing"
)

func TestForLoopOverString(t *testing.T) {
	assertEqual(
		t,
		`
		let l: [string] = []
		for c in "123" {
			l.append(c)
		}
		l
		`,
		L(S("1"), S("2"), S("3")),
	)

	assertEqual(
		t,
		`
		let l: [string] = []
		for c in "Привет" {
			l.append(c)
		}
		l
		`,
		L(S("П"), S("р"), S("и"), S("в"), S("е"), S("т")),
	)
}

func TestBreakInForLoop(t *testing.T) {
	assertEqual(
		t,
		`
		let l: [int] = []
		for i in range(0, 100) {
			l.append(i)
			if i == 5 {
				break
			}
		}
		l
		`,
		L(I(0), I(1), I(2), I(3), I(4), I(5)),
	)
}

func TestContinueInForLoop(t *testing.T) {
	assertEqual(
		t,
		`
		let l: [int] = []
		for i in range(0, 5) {
			if i == 3 {
				continue
			}
			l.append(i)
		}
		l
		`,
		L(I(0), I(1), I(2), I(4)),
	)
}

func TestNestedLoopsWithContinue(t *testing.T) {
	assertEqual(
		t,
		`
		let l: [(int, int)] = []
		var i = 1
		while i < 4 {
			var j = 1
			while j < 4 {
				if j == 2 {
					j += 1
					continue
				}
				l.append((i, j))
				j += 1
			}
			i += 1
		}
		l
		`,
		L(
			Tup(I(1), I(1)), Tup(I(1), I(3)),
			Tup(I(2), I(1)), Tup(I(2), I(3)),
			Tup(I(3), I(1)), Tup(I(3), I(3)),
		),
	)

	/*
		assertEqual(
			t,
			`
			let l: [(int, int)] = []
			var i = 1
			while i < 4 {
				var j = 1
				for j in range(1, 4) {
					if j == 2 {
						continue
					}
					l.append((i, j))
				}
				i += 1
			}
			l
			`,
			L(
				Tup(I(1), I(1)), Tup(I(1), I(3)),
				Tup(I(2), I(1)), Tup(I(2), I(3)),
				Tup(I(3), I(1)), Tup(I(3), I(3)),
			),
		)
	*/
}
