package test

import (
	"testing"
)

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
		"incompatible types for operator `==`: Box1 and Box2",
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
