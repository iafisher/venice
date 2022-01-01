package lex

import (
	"testing"
)

func TestBlockComments(t *testing.T) {
	tokens := getTokens(
		`
		###
		A comment
		# Not the end of the comment
		## Neither is this
		But this is:
		###
		1 + 1
		`,
	)

	checkToken(t, tokens[0], TOKEN_NEWLINE, "\n")
	checkToken(t, tokens[1], TOKEN_NEWLINE, "\n")
	checkToken(t, tokens[2], TOKEN_INT, "1")
	checkToken(t, tokens[3], TOKEN_PLUS, "+")
	checkToken(t, tokens[4], TOKEN_INT, "1")
	checkToken(t, tokens[5], TOKEN_NEWLINE, "\n")
	checkTokensLength(t, tokens, 6)
}

func TestKeywordTokens(t *testing.T) {
	tokens := getTokens(
		"and break class continue else enum false for func if in let or private public return true while void",
	)
	checkToken(t, tokens[0], TOKEN_AND, "and")
	checkToken(t, tokens[1], TOKEN_BREAK, "break")
	checkToken(t, tokens[2], TOKEN_CLASS, "class")
	checkToken(t, tokens[3], TOKEN_CONTINUE, "continue")
	checkToken(t, tokens[4], TOKEN_ELSE, "else")
	checkToken(t, tokens[5], TOKEN_ENUM, "enum")
	checkToken(t, tokens[6], TOKEN_FALSE, "false")
	checkToken(t, tokens[7], TOKEN_FOR, "for")
	checkToken(t, tokens[8], TOKEN_FUNC, "func")
	checkToken(t, tokens[9], TOKEN_IF, "if")
	checkToken(t, tokens[10], TOKEN_IN, "in")
	checkToken(t, tokens[11], TOKEN_LET, "let")
	checkToken(t, tokens[12], TOKEN_OR, "or")
	checkToken(t, tokens[13], TOKEN_PRIVATE, "private")
	checkToken(t, tokens[14], TOKEN_PUBLIC, "public")
	checkToken(t, tokens[15], TOKEN_RETURN, "return")
	checkToken(t, tokens[16], TOKEN_TRUE, "true")
	checkToken(t, tokens[17], TOKEN_WHILE, "while")
	checkToken(t, tokens[18], TOKEN_VOID, "void")
	checkTokensLength(t, tokens, 19)
}

func TestNumericLiterals(t *testing.T) {
	// Integers
	tokens := getTokens("100 -12 0xabc 0o123 0b101011 0")
	checkToken(t, tokens[0], TOKEN_INT, "100")
	checkToken(t, tokens[1], TOKEN_INT, "-12")
	checkToken(t, tokens[2], TOKEN_INT, "0xabc")
	checkToken(t, tokens[3], TOKEN_INT, "0o123")
	checkToken(t, tokens[4], TOKEN_INT, "0b101011")
	checkToken(t, tokens[5], TOKEN_INT, "0")
	checkTokensLength(t, tokens, 6)

	// Real numbers
	tokens = getTokens("0.0 3.14")
	checkToken(t, tokens[0], TOKEN_REAL_NUMBER, "0.0")
	checkToken(t, tokens[1], TOKEN_REAL_NUMBER, "3.14")
	checkTokensLength(t, tokens, 2)
}

func TestInvalidNumericLiterals(t *testing.T) {
	tokens := getTokens("1a2 0b12 0o9 0xg 01 -01")

	checkToken(
		t, tokens[0], TOKEN_ERROR, "invalid character in integer literal: 'a'",
	)
	checkToken(
		t, tokens[1], TOKEN_ERROR, "invalid character in binary integer literal: '2'",
	)
	checkToken(
		t, tokens[2], TOKEN_ERROR, "invalid character in octal integer literal: '9'",
	)
	checkToken(
		t,
		tokens[3],
		TOKEN_ERROR,
		"invalid character in hexadecimal integer literal: 'g'",
	)
	checkToken(t, tokens[4], TOKEN_ERROR, "numeric literal cannot start with '0'")
	checkToken(t, tokens[5], TOKEN_ERROR, "numeric literal cannot start with '0'")
	checkTokensLength(t, tokens, 6)

	tokens = getTokens("0x0.0 1.1.1 00.1")
	checkToken(t, tokens[0], TOKEN_ERROR, "real numbers must be in base 10")
	checkToken(t, tokens[1], TOKEN_ERROR, "numeric literal may not contain multiple periods")
	checkToken(t, tokens[2], TOKEN_ERROR, "numeric literal cannot start with '0'")
	checkTokensLength(t, tokens, 3)
}

func TestOneCharTokens(t *testing.T) {
	tokens := getTokens("[{()}]=+-*/")
	checkToken(t, tokens[0], TOKEN_LEFT_SQUARE, "[")
	checkToken(t, tokens[1], TOKEN_LEFT_CURLY, "{")
	checkToken(t, tokens[2], TOKEN_LEFT_PAREN, "(")
	checkToken(t, tokens[3], TOKEN_RIGHT_PAREN, ")")
	checkToken(t, tokens[4], TOKEN_RIGHT_CURLY, "}")
	checkToken(t, tokens[5], TOKEN_RIGHT_SQUARE, "]")
	checkToken(t, tokens[6], TOKEN_ASSIGN, "=")
	checkToken(t, tokens[7], TOKEN_PLUS, "+")
	checkToken(t, tokens[8], TOKEN_MINUS, "-")
	checkToken(t, tokens[9], TOKEN_ASTERISK, "*")
	checkToken(t, tokens[10], TOKEN_SLASH, "/")
	checkTokensLength(t, tokens, 11)
}

func TestStringLiterals(t *testing.T) {
	tokens := getTokens(`"hello" "\"" "\\\\\u263a"`)
	checkToken(t, tokens[0], TOKEN_STRING, "hello")
	checkToken(t, tokens[1], TOKEN_STRING, "\"")
	checkToken(t, tokens[2], TOKEN_STRING, "\\\\☺")

	tokens = getTokens(`"`)
	checkToken(t, tokens[0], TOKEN_ERROR, "invalid string literal")

	tokens = getTokens(`"\xGG"`)
	checkToken(t, tokens[0], TOKEN_ERROR, "invalid string literal")
}

func TestSymbols(t *testing.T) {
	tokens := getTokens("a b c abc a1_0")
	checkToken(t, tokens[0], TOKEN_SYMBOL, "a")
	checkToken(t, tokens[1], TOKEN_SYMBOL, "b")
	checkToken(t, tokens[2], TOKEN_SYMBOL, "c")
	checkToken(t, tokens[3], TOKEN_SYMBOL, "abc")
	checkToken(t, tokens[4], TOKEN_SYMBOL, "a1_0")
	checkTokensLength(t, tokens, 5)
}

func TestTwoCharTokens(t *testing.T) {
	tokens := getTokens("-> ++ :: == >= <= !=")
	checkToken(t, tokens[0], TOKEN_ARROW, "->")
	checkToken(t, tokens[1], TOKEN_DOUBLE_PLUS, "++")
	checkToken(t, tokens[2], TOKEN_DOUBLE_COLON, "::")
	checkToken(t, tokens[3], TOKEN_EQUALS, "==")
	checkToken(t, tokens[4], TOKEN_GREATER_THAN_OR_EQUALS, ">=")
	checkToken(t, tokens[5], TOKEN_LESS_THAN_OR_EQUALS, "<=")
	checkToken(t, tokens[6], TOKEN_NOT_EQUALS, "!=")
	checkTokensLength(t, tokens, 7)
}

func getTokens(program string) []*Token {
	tokens := []*Token{}
	lexer := NewLexer("", program)
	for {
		token := lexer.NextToken()
		if token.Type == TOKEN_EOF {
			break
		}
		tokens = append(tokens, token)
	}
	return tokens
}

func checkTokensLength(t *testing.T, tokens []*Token, expectedLength int) {
	if len(tokens) != expectedLength {
		t.Helper()
		t.Fatalf("expected %d token(s), got %d", expectedLength, len(tokens))
	}
}

func checkToken(t *testing.T, token *Token, ttype string, value string) {
	if token.Value != value {
		t.Helper()
		if token.Type != ttype {
			t.Fatalf(
				"Wrong token value and type (line %d, col %d): got %q (%q), expected %q (%q)",
				token.Location.Line,
				token.Location.Column,
				token.Value,
				token.Type,
				value,
				ttype,
			)
		} else {
			t.Fatalf(
				"Wrong token value (line %d, col %d): got %q, expected %q",
				token.Location.Line,
				token.Location.Column,
				token.Value,
				value,
			)
		}
	}
	if token.Type != ttype {
		t.Helper()
		t.Fatalf(
			"Wrong token type (line %d, col %d): got %q, expected %q",
			token.Location.Line,
			token.Location.Column,
			token.Type,
			ttype,
		)
	}
}

func checkTokenAndLocation(t *testing.T, token *Token, ttype string, value string, line int, column int) {
	checkToken(t, token, ttype, value)

	if token.Location.Line != line {
		t.Helper()
		t.Fatalf("Wrong line: got %d, expected %d", token.Location.Line, line)
	}
	if token.Location.Column != column {
		t.Helper()
		t.Fatalf("Wrong column: got %d, expected %d", token.Location.Column, column)
	}
}
