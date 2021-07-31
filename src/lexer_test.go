package main

import "testing"

func TestNextTokenWithIntegerLiterals(t *testing.T) {
	input := "100"
	lexer := NewLexer(input)

	token := lexer.NextToken()
	checkToken(t, token, TOKEN_INT, "100", 1, 1)

	token = lexer.NextToken()
	checkToken(t, token, TOKEN_EOF, "", 1, 4)
}

func TestNextTokenWithSymbols(t *testing.T) {
	input := "a b c abc a1_0"
	lexer := NewLexer(input)

	token := lexer.NextToken()
	checkToken(t, token, TOKEN_SYMBOL, "a", 1, 1)

	token = lexer.NextToken()
	checkToken(t, token, TOKEN_SYMBOL, "b", 1, 3)

	token = lexer.NextToken()
	checkToken(t, token, TOKEN_SYMBOL, "c", 1, 5)

	token = lexer.NextToken()
	checkToken(t, token, TOKEN_SYMBOL, "abc", 1, 7)

	token = lexer.NextToken()
	checkToken(t, token, TOKEN_SYMBOL, "a1_0", 1, 11)

	token = lexer.NextToken()
	checkToken(t, token, TOKEN_EOF, "", 1, 15)
}

func checkToken(t *testing.T, token *Token, ttype string, value string, line int, column int) {
	if token.Type != ttype {
		t.Fatalf("Wrong token type (line %d, col %d): got %q, expected %q", token.Loc.Line, token.Loc.Column, token.Type, ttype)
	}
	if token.Value != value {
		t.Fatalf("Wrong token value (line %d, col %d): got %q, expected %q", token.Loc.Line, token.Loc.Column, token.Value, value)
	}
	if token.Loc.Line != line {
		t.Fatalf("Wrong line: got %d, expected %d", token.Loc.Line, line)
	}
	if token.Loc.Column != column {
		t.Fatalf("Wrong column: got %d, expected %d", token.Loc.Column, column)
	}
}
