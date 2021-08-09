package main

import "testing"

func TestIntegerLiterals(t *testing.T) {
	input := "100 -12"
	lexer := NewLexer(input)

	token := lexer.NextToken()
	checkToken(t, token, TOKEN_INT, "100", 1, 1)

	token = lexer.NextToken()
	checkToken(t, token, TOKEN_INT, "-12", 1, 5)

	token = lexer.NextToken()
	checkToken(t, token, TOKEN_EOF, "", 1, 8)
}

func TestSymbols(t *testing.T) {
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

func TestOneCharTokens(t *testing.T) {
	input := "[{()}]=+-*/"
	lexer := NewLexer(input)

	token := lexer.NextToken()
	checkToken(t, token, TOKEN_LEFT_SQUARE, "[", 1, 1)

	token = lexer.NextToken()
	checkToken(t, token, TOKEN_LEFT_CURLY, "{", 1, 2)

	token = lexer.NextToken()
	checkToken(t, token, TOKEN_LEFT_PAREN, "(", 1, 3)

	token = lexer.NextToken()
	checkToken(t, token, TOKEN_RIGHT_PAREN, ")", 1, 4)

	token = lexer.NextToken()
	checkToken(t, token, TOKEN_RIGHT_CURLY, "}", 1, 5)

	token = lexer.NextToken()
	checkToken(t, token, TOKEN_RIGHT_SQUARE, "]", 1, 6)

	token = lexer.NextToken()
	checkToken(t, token, TOKEN_ASSIGN, "=", 1, 7)

	token = lexer.NextToken()
	checkToken(t, token, TOKEN_PLUS, "+", 1, 8)

	token = lexer.NextToken()
	checkToken(t, token, TOKEN_MINUS, "-", 1, 9)

	token = lexer.NextToken()
	checkToken(t, token, TOKEN_ASTERISK, "*", 1, 10)

	token = lexer.NextToken()
	checkToken(t, token, TOKEN_SLASH, "/", 1, 11)

	token = lexer.NextToken()
	checkToken(t, token, TOKEN_EOF, "", 1, 12)
}

func TestTwoCharTokens(t *testing.T) {
	input := "->"
	lexer := NewLexer(input)

	token := lexer.NextToken()
	checkToken(t, token, TOKEN_ARROW, "->", 1, 1)

	token = lexer.NextToken()
	checkToken(t, token, TOKEN_EOF, "", 1, 3)
}

func TestKeywordTokens(t *testing.T) {
	input := "let"
	lexer := NewLexer(input)

	token := lexer.NextToken()
	checkToken(t, token, TOKEN_LET, "let", 1, 1)

	token = lexer.NextToken()
	checkToken(t, token, TOKEN_EOF, "", 1, 4)
}

func checkToken(t *testing.T, token *Token, ttype string, value string, line int, column int) {
	if token.Type != ttype {
		t.Fatalf("Wrong token type (line %d, col %d): got %q, expected %q", token.Location.Line, token.Location.Column, token.Type, ttype)
	}
	if token.Value != value {
		t.Fatalf("Wrong token value (line %d, col %d): got %q, expected %q", token.Location.Line, token.Location.Column, token.Value, value)
	}
	if token.Location.Line != line {
		t.Fatalf("Wrong line: got %d, expected %d", token.Location.Line, line)
	}
	if token.Location.Column != column {
		t.Fatalf("Wrong column: got %d, expected %d", token.Location.Column, column)
	}
}
