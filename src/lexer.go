package main

import "fmt"

type Lexer struct {
	program string
	index   int
	loc     Location
}

type Token struct {
	Type  string
	Value string
	Loc   *Location
}

type Location struct {
	Line   int
	Column int
}

func NewLexer(program string) *Lexer {
	return &Lexer{program: program, index: 0, loc: Location{Line: 1, Column: 1}}
}

func (l *Lexer) NextToken() *Token {
	l.skipWhitespaceExceptNewlines()
	return l.nextToken()
}

func (l *Lexer) NextTokenSkipNewlines() *Token {
	l.skipAllWhitespace()
	return l.nextToken()
}

func (l *Lexer) nextToken() *Token {
	loc := l.copyLocation()
	if l.index >= len(l.program) {
		return &Token{Type: TOKEN_EOF, Value: "", Loc: loc}
	}

	if l.index+1 < len(l.program) {
		two_prefix := l.program[l.index : l.index+2]
		token_type, ok := two_char_tokens[two_prefix]
		if ok {
			l.advance()
			l.advance()
			return &Token{Type: token_type, Value: two_prefix, Loc: loc}
		}
	}

	ch := l.program[l.index]
	token_type, ok := one_char_tokens[ch]
	if ok {
		l.advance()
		return &Token{Type: token_type, Value: string(ch), Loc: loc}
	}

	switch {
	case isDigit(ch):
		value := l.readInteger()
		return &Token{Type: TOKEN_INT, Value: value, Loc: loc}
	case isSymbolFirstCharacter(ch):
		value := l.readSymbol()
		if keywordType, ok := keywords[value]; ok {
			return &Token{Type: keywordType, Value: value, Loc: loc}
		} else {
			return &Token{Type: TOKEN_SYMBOL, Value: value, Loc: loc}
		}
	case ch == '"':
		value := l.readString()
		return &Token{Type: TOKEN_STRING, Value: value, Loc: loc}
	default:
		l.advance()
		return &Token{Type: TOKEN_UNKNOWN, Value: string(ch), Loc: loc}
	}
}

func (l *Lexer) skipAllWhitespace() {
	for l.index < len(l.program) && isWhitespace(l.program[l.index]) {
		l.advance()
	}
}

func (l *Lexer) skipWhitespaceExceptNewlines() {
	for l.index < len(l.program) && isWhitespace(l.program[l.index]) && l.program[l.index] != '\n' {
		l.advance()
	}
}

func (l *Lexer) readInteger() string {
	start := l.index
	for l.index < len(l.program) && isDigit(l.program[l.index]) {
		l.advance()
	}
	return l.program[start:l.index]
}

func (l *Lexer) readSymbol() string {
	start := l.index
	l.advance()
	for l.index < len(l.program) && isSymbolCharacter(l.program[l.index]) {
		l.advance()
	}
	return l.program[start:l.index]
}

func (l *Lexer) readString() string {
	start := l.index
	l.advance()
	for l.index < len(l.program) && l.program[l.index] != '"' {
		l.advance()
	}
	l.advance()
	return l.program[start+1 : l.index-1]
}

func (l *Lexer) advance() {
	if l.index < len(l.program) {
		if l.program[l.index] == '\n' {
			l.loc.Line += 1
			l.loc.Column = 1
		} else {
			l.loc.Column += 1
		}
		l.index += 1
	}
}

func (l *Lexer) copyLocation() *Location {
	return &Location{Line: l.loc.Line, Column: l.loc.Column}
}

func (token *Token) asString() string {
	return fmt.Sprintf("%s (%q) at line %d, column %d", token.Type, token.Value, token.Loc.Line, token.Loc.Column)
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func isSymbolFirstCharacter(ch byte) bool {
	return ('a' <= ch && ch <= 'z') || ('A' <= ch && ch < +'Z') || (ch == '_')
}

func isSymbolCharacter(ch byte) bool {
	return isSymbolFirstCharacter(ch) || isDigit(ch)
}

func isWhitespace(ch byte) bool {
	return ch == ' ' || ch == '\n' || ch == '\t' || ch == '\v' || ch == '\f' || ch == '\r'
}

const (
	TOKEN_ARROW        = "TOKEN_ARROW"
	TOKEN_ASSIGN       = "TOKEN_ASSIGN"
	TOKEN_ASTERISK     = "TOKEN_ASTERISK"
	TOKEN_BREAK        = "TOKEN_BREAK"
	TOKEN_COLON        = "TOKEN_COLON"
	TOKEN_COMMA        = "TOKEN_COMMA"
	TOKEN_CONTINUE     = "TOKEN_CONTINUE"
	TOKEN_ELSE         = "TOKEN_ELSE"
	TOKEN_EQ           = "TOKEN_EQ"
	TOKEN_FALSE        = "TOKEN_FALSE"
	TOKEN_FN           = "TOKEN_FN"
	TOKEN_FOR          = "TOKEN_FOR"
	TOKEN_IF           = "TOKEN_IF"
	TOKEN_INT          = "TOKEN_INT"
	TOKEN_EOF          = "TOKEN_EOF"
	TOKEN_LEFT_CURLY   = "TOKEN_LEFT_CURLY"
	TOKEN_LEFT_PAREN   = "TOKEN_LEFT_PAREN"
	TOKEN_LEFT_SQUARE  = "TOKEN_LEFT_SQUARE"
	TOKEN_LET          = "TOKEN_LET"
	TOKEN_MINUS        = "TOKEN_MINUS"
	TOKEN_NEWLINE      = "TOKEN_NEWLINE"
	TOKEN_PLUS         = "TOKEN_PLUS"
	TOKEN_RETURN       = "TOKEN_RETURN"
	TOKEN_RIGHT_CURLY  = "TOKEN_RIGHT_CURLY"
	TOKEN_RIGHT_PAREN  = "TOKEN_RIGHT_PAREN"
	TOKEN_RIGHT_SQUARE = "TOKEN_RIGHT_SQUARE"
	TOKEN_SEMICOLON    = "TOKEN_SEMICOLON"
	TOKEN_SLASH        = "TOKEN_SLASH"
	TOKEN_STRING       = "TOKEN_STRING"
	TOKEN_SYMBOL       = "TOKEN_SYMBOL"
	TOKEN_TRUE         = "TOKEN_TRUE"
	TOKEN_UNKNOWN      = "TOKEN_UNKNOWN"
	TOKEN_VOID         = "TOKEN_VOID"
	TOKEN_WHILE        = "TOKEN_WHILE"
)

var keywords = map[string]string{
	"break":    TOKEN_BREAK,
	"continue": TOKEN_CONTINUE,
	"else":     TOKEN_ELSE,
	"false":    TOKEN_FALSE,
	"fn":       TOKEN_FN,
	"for":      TOKEN_FOR,
	"if":       TOKEN_IF,
	"let":      TOKEN_LET,
	"return":   TOKEN_RETURN,
	"true":     TOKEN_TRUE,
	"while":    TOKEN_WHILE,
	"void":     TOKEN_VOID,
}

var one_char_tokens = map[byte]string{
	'=':  TOKEN_ASSIGN,
	'*':  TOKEN_ASTERISK,
	':':  TOKEN_COLON,
	',':  TOKEN_COMMA,
	'{':  TOKEN_LEFT_CURLY,
	'(':  TOKEN_LEFT_PAREN,
	'[':  TOKEN_LEFT_SQUARE,
	'-':  TOKEN_MINUS,
	'\n': TOKEN_NEWLINE,
	'+':  TOKEN_PLUS,
	'}':  TOKEN_RIGHT_CURLY,
	')':  TOKEN_RIGHT_PAREN,
	']':  TOKEN_RIGHT_SQUARE,
	';':  TOKEN_SEMICOLON,
	'/':  TOKEN_SLASH,
}

var two_char_tokens = map[string]string{
	"->": TOKEN_ARROW,
	"==": TOKEN_EQ,
}
