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
