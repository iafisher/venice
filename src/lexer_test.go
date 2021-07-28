package main

import "testing"

func TestNextToken(t *testing.T) {
	input := "100"
	lexer := NewLexer(input)

	token := lexer.NextToken()
	checkToken(t, token, TOKEN_INT, "100", 1, 1)

	token = lexer.NextToken()
	checkToken(t, token, TOKEN_EOF, "", 1, 4)
}

func checkToken(t *testing.T, token *Token, ttype string, value string, line int, column int) {
	if token.Type != ttype {
		t.Fatalf("Wrong token type: got %q, expected %q", token.Type, ttype)
	}
	if token.Value != value {
		t.Fatalf("Wrong token value: got %q, expected %q", token.Value, value)
	}
	if token.Loc.Line != line {
		t.Fatalf("Wrong line: got %d, expected %d", token.Loc.Line, line)
	}
	if token.Loc.Column != column {
		t.Fatalf("Wrong column: got %d, expected %d", token.Loc.Column, column)
	}
}
