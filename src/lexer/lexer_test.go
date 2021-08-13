package lexer

import "testing"

func TestIntegerLiterals(t *testing.T) {
	tokens := getTokens("100 -12 0xabc 0o123 0b101011 0")
	checkToken(t, tokens[0], TOKEN_INT, "100", 1, 1)
	checkToken(t, tokens[1], TOKEN_INT, "-12", 1, 5)
	checkToken(t, tokens[2], TOKEN_INT, "0xabc", 1, 9)
	checkToken(t, tokens[3], TOKEN_INT, "0o123", 1, 15)
	checkToken(t, tokens[4], TOKEN_INT, "0b101011", 1, 21)
	checkToken(t, tokens[5], TOKEN_INT, "0", 1, 30)
	checkTokensLength(t, tokens, 6)
}

func TestInvalidIntegerLiterals(t *testing.T) {
	tokens := getTokens("12a 0b12 0o9 0xg 01 -01")
	checkToken(t, tokens[0], TOKEN_ERROR, "invalid character in integer literal: 'a'", 1, 1)
	checkToken(t, tokens[1], TOKEN_ERROR, "invalid character in binary integer literal: '2'", 1, 5)
	checkToken(t, tokens[2], TOKEN_ERROR, "invalid character in octal integer literal: '9'", 1, 10)
	checkToken(t, tokens[3], TOKEN_ERROR, "invalid character in hexadecimal integer literal: 'g'", 1, 14)
	checkToken(t, tokens[4], TOKEN_ERROR, "integer literal cannot start with '0'", 1, 18)
	checkToken(t, tokens[5], TOKEN_ERROR, "integer literal cannot start with '0'", 1, 21)
	checkTokensLength(t, tokens, 6)
}

func TestSymbols(t *testing.T) {
	tokens := getTokens("a b c abc a1_0")
	checkToken(t, tokens[0], TOKEN_SYMBOL, "a", 1, 1)
	checkToken(t, tokens[1], TOKEN_SYMBOL, "b", 1, 3)
	checkToken(t, tokens[2], TOKEN_SYMBOL, "c", 1, 5)
	checkToken(t, tokens[3], TOKEN_SYMBOL, "abc", 1, 7)
	checkToken(t, tokens[4], TOKEN_SYMBOL, "a1_0", 1, 11)
	checkTokensLength(t, tokens, 5)
}

func TestOneCharTokens(t *testing.T) {
	tokens := getTokens("[{()}]=+-*/")
	checkToken(t, tokens[0], TOKEN_LEFT_SQUARE, "[", 1, 1)
	checkToken(t, tokens[1], TOKEN_LEFT_CURLY, "{", 1, 2)
	checkToken(t, tokens[2], TOKEN_LEFT_PAREN, "(", 1, 3)
	checkToken(t, tokens[3], TOKEN_RIGHT_PAREN, ")", 1, 4)
	checkToken(t, tokens[4], TOKEN_RIGHT_CURLY, "}", 1, 5)
	checkToken(t, tokens[5], TOKEN_RIGHT_SQUARE, "]", 1, 6)
	checkToken(t, tokens[6], TOKEN_ASSIGN, "=", 1, 7)
	checkToken(t, tokens[7], TOKEN_PLUS, "+", 1, 8)
	checkToken(t, tokens[8], TOKEN_MINUS, "-", 1, 9)
	checkToken(t, tokens[9], TOKEN_ASTERISK, "*", 1, 10)
	checkToken(t, tokens[10], TOKEN_SLASH, "/", 1, 11)
	checkTokensLength(t, tokens, 11)
}

func TestTwoCharTokens(t *testing.T) {
	tokens := getTokens("-> ++ :: == >= <= !=")
	checkToken(t, tokens[0], TOKEN_ARROW, "->", 1, 1)
	checkToken(t, tokens[1], TOKEN_DOUBLE_PLUS, "++", 1, 4)
	checkToken(t, tokens[2], TOKEN_DOUBLE_COLON, "::", 1, 7)
	checkToken(t, tokens[3], TOKEN_EQUALS, "==", 1, 10)
	checkToken(t, tokens[4], TOKEN_GREATER_THAN_OR_EQUALS, ">=", 1, 13)
	checkToken(t, tokens[5], TOKEN_LESS_THAN_OR_EQUALS, "<=", 1, 16)
	checkToken(t, tokens[6], TOKEN_NOT_EQUALS, "!=", 1, 19)
	checkTokensLength(t, tokens, 7)
}

func TestKeywordTokens(t *testing.T) {
	tokens := getTokens("and break class continue else enum false fn for if in let or private public return true while void")
	checkToken(t, tokens[0], TOKEN_AND, "and", 1, 1)
	checkToken(t, tokens[1], TOKEN_BREAK, "break", 1, 5)
	checkToken(t, tokens[2], TOKEN_CLASS, "class", 1, 11)
	checkToken(t, tokens[3], TOKEN_CONTINUE, "continue", 1, 17)
	checkToken(t, tokens[4], TOKEN_ELSE, "else", 1, 26)
	checkToken(t, tokens[5], TOKEN_ENUM, "enum", 1, 31)
	checkToken(t, tokens[6], TOKEN_FALSE, "false", 1, 36)
	checkToken(t, tokens[7], TOKEN_FN, "fn", 1, 42)
	checkToken(t, tokens[8], TOKEN_FOR, "for", 1, 45)
	checkToken(t, tokens[9], TOKEN_IF, "if", 1, 49)
	checkToken(t, tokens[10], TOKEN_IN, "in", 1, 52)
	checkToken(t, tokens[11], TOKEN_LET, "let", 1, 55)
	checkToken(t, tokens[12], TOKEN_OR, "or", 1, 59)
	checkToken(t, tokens[13], TOKEN_PRIVATE, "private", 1, 62)
	checkToken(t, tokens[14], TOKEN_PUBLIC, "public", 1, 70)
	checkToken(t, tokens[15], TOKEN_RETURN, "return", 1, 77)
	checkToken(t, tokens[16], TOKEN_TRUE, "true", 1, 84)
	checkToken(t, tokens[17], TOKEN_WHILE, "while", 1, 89)
	checkToken(t, tokens[18], TOKEN_VOID, "void", 1, 95)
	checkTokensLength(t, tokens, 19)
}

func getTokens(program string) []*Token {
	tokens := []*Token{}
	lexer := NewLexer(program)
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
		t.Fatalf("expected %d token(s), got %d", expectedLength, len(tokens))
	}
}

func checkToken(t *testing.T, token *Token, ttype string, value string, line int, column int) {
	if token.Value != value {
		if token.Type != ttype {
			t.Fatalf("Wrong token value and type (line %d, col %d): got %q (%q), expected %q (%q)", token.Location.Line, token.Location.Column, token.Value, token.Type, value, ttype)
		} else {
			t.Fatalf("Wrong token value (line %d, col %d): got %q, expected %q", token.Location.Line, token.Location.Column, token.Value, value)
		}
	}
	if token.Type != ttype {
		t.Fatalf("Wrong token type (line %d, col %d): got %q, expected %q", token.Location.Line, token.Location.Column, token.Type, ttype)
	}
	if token.Location.Line != line {
		t.Fatalf("Wrong line: got %d, expected %d", token.Location.Line, line)
	}
	if token.Location.Column != column {
		t.Fatalf("Wrong column: got %d, expected %d", token.Location.Column, column)
	}
}
