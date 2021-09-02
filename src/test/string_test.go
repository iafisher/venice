package test

import (
	"testing"
)

func TestStringBuiltins(t *testing.T) {
	// string.length
	assertEqual(t, `let s = "123"; s.length()`, I(3))
	assertEqual(t, `"123".length()`, I(3))

	// string.to_upper
	assertEqual(t, `"abc".to_upper()`, S("ABC"))

	// string.to_lower
	assertEqual(t, `"ABC".to_lower()`, S("abc"))

	// string.slice
	assertEqual(t, `"ABCDE".slice(1, 3)`, S("BC"))

	// string.split_space
	assertEqual(t, `"colorless green ideas".split_space()`, L(S("colorless"), S("green"), S("ideas")))

	// string.split
	assertEqual(t, `"colorless green ideas".split(" ")`, L(S("colorless"), S("green"), S("ideas")))

	// string.ends_with
	assertEqual(t, `"Lorem ipsum".ends_with("ipsum")`, B(true))

	// string.starts_with
	assertEqual(t, `"Lorem ipsum".starts_with("ipsum")`, B(false))

	// string.trim
	assertEqual(t, "\"     word\\t\\t\\n\\r\".trim()", S("word"))

	// string.trim_left
	assertEqual(t, "\"     word\\t\\t\\n\\r\".trim_left()", S("word\t\t\n\r"))

	// string.trim_right
	assertEqual(t, "\"     word\\t\\t\\n\\r\".trim_right()", S("     word"))

	// string.remove_prefix
	assertEqual(t, `"www.example.com".remove_prefix("www.")`, S("example.com"))

	// string.remove_suffix
	assertEqual(t, `"www.example.com".remove_suffix(".com")`, S("www.example"))

	// string.replace_all
	assertEqual(t, `"Mt. Hood and Mt. St. Helens".replace_all("Mt.", "Mount")`, S("Mount Hood and Mount St. Helens"))

	// string.replace_first
	assertEqual(t, `"Mt. Hood and Mt. St. Helens".replace_first("Mt.", "Mount")`, S("Mount Hood and Mt. St. Helens"))

	// string.replace_last
	assertEqual(t, `"Mt. Hood and Mt. St. Helens".replace_last("Mt.", "Mount")`, S("Mt. Hood and Mount St. Helens"))
}
