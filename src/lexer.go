package main

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
	l.skipWhitespace()

	loc := l.copyLocation()
	if l.index >= len(l.program) {
		return &Token{Type: TOKEN_EOF, Value: "", Loc: loc}
	}

	ch := l.program[l.index]
	switch {
	case isDigit(ch):
		value := l.readInteger()
		return &Token{Type: TOKEN_INT, Value: value, Loc: loc}
	case isSymbolFirstCharacter(ch):
		value := l.readSymbol()
		return &Token{Type: TOKEN_SYMBOL, Value: value, Loc: loc}
	default:
		l.advance()
		return &Token{Type: TOKEN_UNKNOWN, Value: string(ch), Loc: loc}
	}
}

func (l *Lexer) skipWhitespace() {
	for l.index < len(l.program) && isWhitespace(l.program[l.index]) {
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
	TOKEN_INT     = "TOKEN_INT"
	TOKEN_EOF     = "TOKEN_EOF"
	TOKEN_SYMBOL  = "TOKEN_SYMBOL"
	TOKEN_UNKNOWN = "TOKEN_UNKNOWN"
)
