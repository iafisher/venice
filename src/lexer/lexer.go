package lexer

import (
	"fmt"
	"strconv"
	"strings"
)

type Lexer struct {
	program  string
	index    int
	location Location
}

type Token struct {
	Type     string
	Value    string
	Location *Location
}

type Location struct {
	Line     int
	Column   int
	FilePath string
}

func NewLexer(filePath string, program string) *Lexer {
	return &Lexer{
		program: program,
		index:   0,
		location: Location{
			Line:     1,
			Column:   1,
			FilePath: filePath,
		},
	}
}

func (l *Lexer) NextToken() *Token {
	l.skipCommentsAndWhitespaceExceptNewlines()
	return l.nextToken()
}

func (l *Lexer) NextTokenSkipNewlines() *Token {
	l.skipCommentsAndAllWhitespace()
	return l.nextToken()
}

func (l *Lexer) nextToken() *Token {
	location := l.copyLocation()
	if l.index >= len(l.program) {
		return &Token{Type: TOKEN_EOF, Value: "", Location: location}
	}

	if l.index+2 < len(l.program) {
		three_prefix := l.program[l.index : l.index+3]
		token_type, ok := three_char_tokens[three_prefix]
		if ok {
			l.advance()
			l.advance()
			l.advance()
			return &Token{Type: token_type, Value: three_prefix, Location: location}
		}
	}

	if l.index+1 < len(l.program) {
		two_prefix := l.program[l.index : l.index+2]
		token_type, ok := two_char_tokens[two_prefix]
		if ok {
			l.advance()
			l.advance()
			return &Token{Type: token_type, Value: two_prefix, Location: location}
		}
	}

	ch := l.program[l.index]

	// To simplify parsing, integer literals beginning with a minus sign are lexed as a
	// single TOKEN_INT instead of a TOKEN_MINUS followed by a TOKEN_INT, so we have to
	// check this case before we look at the one-character tokens in the next block.
	if ch == '-' && l.index+1 < len(l.program) && isDigit(l.program[l.index+1]) {
		value, err := l.readInteger()
		if err != nil {
			return &Token{Type: TOKEN_ERROR, Value: err.Error(), Location: location}
		}
		return &Token{Type: TOKEN_INT, Value: value, Location: location}
	}

	token_type, ok := one_char_tokens[ch]
	if ok {
		l.advance()
		return &Token{Type: token_type, Value: string(ch), Location: location}
	}

	switch {
	case isDigit(ch):
		value, err := l.readInteger()
		if err != nil {
			return &Token{Type: TOKEN_ERROR, Value: err.Error(), Location: location}
		}
		return &Token{Type: TOKEN_INT, Value: value, Location: location}
	case isSymbolFirstCharacter(ch):
		value := l.readSymbol()
		if keywordType, ok := keywords[value]; ok {
			return &Token{Type: keywordType, Value: value, Location: location}
		} else {
			return &Token{Type: TOKEN_SYMBOL, Value: value, Location: location}
		}
	case ch == '"':
		value, err := l.readString('"')
		if err != nil {
			return &Token{
				Type:     TOKEN_ERROR,
				Value:    "invalid string literal",
				Location: location,
			}
		}
		return &Token{Type: TOKEN_STRING, Value: value, Location: location}
	case ch == '\'':
		value, err := l.readString('\'')
		if err != nil {
			return &Token{
				Type:     TOKEN_ERROR,
				Value:    "invalid character literal",
				Location: location,
			}
		}
		return &Token{Type: TOKEN_CHARACTER, Value: value, Location: location}
	default:
		l.advance()
		return &Token{Type: TOKEN_UNKNOWN, Value: string(ch), Location: location}
	}
}

func (l *Lexer) skipCommentsAndAllWhitespace() {
	inComment := false
	for l.index < len(l.program) {
		ch := l.program[l.index]

		if inComment {
			if ch == '\n' {
				inComment = false
			}
		} else {
			if ch == '#' {
				inComment = true
			} else if !isWhitespace(ch) {
				break
			}
		}

		l.advance()
	}
}

func (l *Lexer) skipCommentsAndWhitespaceExceptNewlines() {
	inLineComment := false
	inBlockComment := false
	for l.index < len(l.program) {
		ch := l.program[l.index]

		if inLineComment {
			if ch == '\n' {
				inLineComment = false
			}
		} else if inBlockComment {
			if strings.HasPrefix(l.program[l.index:], "###") {
				inBlockComment = false
				l.advance()
				l.advance()
			}
		} else {
			if ch == '#' {
				if strings.HasPrefix(l.program[l.index:], "###") {
					inBlockComment = true
					l.advance()
					l.advance()
				} else {
					inLineComment = true
				}
			} else if !isWhitespace(ch) || ch == '\n' {
				break
			}
		}

		l.advance()
	}
}

const binaryDigits = "01"
const octalDigits = "01234567"
const decimalDigits = "0123456789"
const hexadecimalDigits = "0123456789abcdefABCDEF"

func (l *Lexer) readInteger() (string, error) {
	start := l.index

	if l.program[l.index] == '-' {
		l.advance()
	}

	if l.startsWith("0b") {
		l.advance()
		l.advance()
		for l.index < len(l.program) && isSymbolCharacter(l.program[l.index]) {
			ch := l.program[l.index]
			if strings.IndexByte(binaryDigits, ch) == -1 {
				l.advance()
				return "", &LexerError{
					fmt.Sprintf("invalid character in binary integer literal: '%c'", ch),
				}
			}
			l.advance()
		}
		return l.program[start:l.index], nil
	} else if l.startsWith("0o") {
		l.advance()
		l.advance()
		for l.index < len(l.program) && isSymbolCharacter(l.program[l.index]) {
			ch := l.program[l.index]
			if strings.IndexByte(octalDigits, ch) == -1 {
				l.advance()
				return "", &LexerError{
					fmt.Sprintf("invalid character in octal integer literal: '%c'", ch),
				}
			}
			l.advance()
		}
		return l.program[start:l.index], nil
	} else if l.startsWith("0x") {
		l.advance()
		l.advance()
		for l.index < len(l.program) && isSymbolCharacter(l.program[l.index]) {
			ch := l.program[l.index]
			if strings.IndexByte(hexadecimalDigits, ch) == -1 {
				l.advance()
				return "", &LexerError{
					fmt.Sprintf("invalid character in hexadecimal integer literal: '%c'", ch),
				}
			}
			l.advance()
		}
		return l.program[start:l.index], nil
	} else {
		l.advance()
		for l.index < len(l.program) && isSymbolCharacter(l.program[l.index]) {
			ch := l.program[l.index]
			if strings.IndexByte(decimalDigits, ch) == -1 {
				l.advance()
				return "", &LexerError{
					fmt.Sprintf("invalid character in integer literal: '%c'", ch),
				}
			}
			l.advance()
		}

		// Check the first character afterwards so that the whole literal is read as a
		// single token.
		if l.index-start > 1 &&
			(l.program[start] == '0' ||
				(l.program[start] == '-' && l.program[start+1] == '0')) {
			return "", &LexerError{"integer literal cannot start with '0'"}
		}

		return l.program[start:l.index], nil
	}
}

func (l *Lexer) readSymbol() string {
	start := l.index
	l.advance()
	for l.index < len(l.program) && isSymbolCharacter(l.program[l.index]) {
		l.advance()
	}
	return l.program[start:l.index]
}

func (l *Lexer) readString(delimiter byte) (string, error) {
	start := l.index
	l.advance()
	for l.index < len(l.program) {
		if l.program[l.index] == delimiter {
			break
		} else if l.program[l.index] == '\\' {
			l.advance()
			l.advance()
		} else {
			l.advance()
		}
	}
	l.advance()
	return strconv.Unquote(l.program[start:l.index])
}

func (l *Lexer) advance() {
	if l.index < len(l.program) {
		if l.program[l.index] == '\n' {
			l.location.Line += 1
			l.location.Column = 1
		} else {
			l.location.Column += 1
		}
		l.index += 1
	}
}

func (l *Lexer) startsWith(prefix string) bool {
	return strings.HasPrefix(l.program[l.index:], prefix)
}

func (l *Lexer) copyLocation() *Location {
	return &Location{
		Line:     l.location.Line,
		Column:   l.location.Column,
		FilePath: l.location.FilePath,
	}
}

func (token *Token) String() string {
	return fmt.Sprintf(
		"%s (%q) at line %d, column %d",
		token.Type,
		token.Value,
		token.Location.Line,
		token.Location.Column,
	)
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
	TOKEN_AND                    = "TOKEN_AND"
	TOKEN_ARROW                  = "TOKEN_ARROW"
	TOKEN_AS                     = "TOKEN_AS"
	TOKEN_ASSIGN                 = "TOKEN_ASSIGN"
	TOKEN_ASSIGN_ADD             = "TOKEN_ASSIGN_ADD"
	TOKEN_ASSIGN_DIV             = "TOKEN_ASSIGN_DIV"
	TOKEN_ASSIGN_MUL             = "TOKEN_ASSIGN_MUL"
	TOKEN_ASSIGN_SUB             = "TOKEN_ASSIGN_SUB"
	TOKEN_ASTERISK               = "TOKEN_ASTERISK"
	TOKEN_BREAK                  = "TOKEN_BREAK"
	TOKEN_CASE                   = "TOKEN_CASE"
	TOKEN_CHARACTER              = "TOKEN_CHARACTER"
	TOKEN_CLASS                  = "TOKEN_CLASS"
	TOKEN_COLON                  = "TOKEN_COLON"
	TOKEN_COMMA                  = "TOKEN_COMMA"
	TOKEN_CONSTRUCTOR            = "TOKEN_CONSTRUCTOR"
	TOKEN_CONTINUE               = "TOKEN_CONTINUE"
	TOKEN_DEFAULT                = "TOKEN_DEFAULT"
	TOKEN_DOT                    = "TOKEN_DOT"
	TOKEN_DOUBLE_COLON           = "TOKEN_DOUBLE_COLON"
	TOKEN_DOUBLE_PLUS            = "TOKEN_DOUBLE_PLUS"
	TOKEN_ELLIPSIS               = "TOKEN_ELLIPSIS"
	TOKEN_ELSE                   = "TOKEN_ELSE"
	TOKEN_ENUM                   = "TOKEN_ENUM"
	TOKEN_EQUALS                 = "TOKEN_EQUALS"
	TOKEN_ERROR                  = "TOKEN_ERROR"
	TOKEN_FALSE                  = "TOKEN_FALSE"
	TOKEN_FN                     = "TOKEN_FN"
	TOKEN_FOR                    = "TOKEN_FOR"
	TOKEN_GREATER_THAN           = "TOKEN_GREATER_THAN"
	TOKEN_GREATER_THAN_OR_EQUALS = "TOKEN_GREATER_THAN_OR_EQUALS"
	TOKEN_IF                     = "TOKEN_IF"
	TOKEN_IMPORT                 = "TOKEN_IMPORT"
	TOKEN_IN                     = "TOKEN_IN"
	TOKEN_INT                    = "TOKEN_INT"
	TOKEN_EOF                    = "TOKEN_EOF"
	TOKEN_LEFT_CURLY             = "TOKEN_LEFT_CURLY"
	TOKEN_LEFT_PAREN             = "TOKEN_LEFT_PAREN"
	TOKEN_LEFT_SQUARE            = "TOKEN_LEFT_SQUARE"
	TOKEN_LESS_THAN              = "TOKEN_LESS_THAN"
	TOKEN_LESS_THAN_OR_EQUALS    = "TOKEN_LESS_THAN_OR_EQUALS"
	TOKEN_LET                    = "TOKEN_LET"
	TOKEN_MATCH                  = "TOKEN_MATCH"
	TOKEN_MINUS                  = "TOKEN_MINUS"
	TOKEN_NEWLINE                = "TOKEN_NEWLINE"
	TOKEN_NO                     = "TOKEN_NO"
	TOKEN_NOT                    = "TOKEN_NOT"
	TOKEN_NOT_EQUALS             = "TOKEN_NOT_EQUALS"
	TOKEN_OR                     = "TOKEN_OR"
	TOKEN_PRIVATE                = "TOKEN_PRIVATE"
	TOKEN_PLUS                   = "TOKEN_PLUS"
	TOKEN_PUBLIC                 = "TOKEN_PUBLIC"
	TOKEN_RETURN                 = "TOKEN_RETURN"
	TOKEN_RIGHT_CURLY            = "TOKEN_RIGHT_CURLY"
	TOKEN_RIGHT_PAREN            = "TOKEN_RIGHT_PAREN"
	TOKEN_RIGHT_SQUARE           = "TOKEN_RIGHT_SQUARE"
	TOKEN_SELF                   = "TOKEN_SELF"
	TOKEN_SEMICOLON              = "TOKEN_SEMICOLON"
	TOKEN_SLASH                  = "TOKEN_SLASH"
	TOKEN_STRING                 = "TOKEN_STRING"
	TOKEN_SYMBOL                 = "TOKEN_SYMBOL"
	TOKEN_TRUE                   = "TOKEN_TRUE"
	TOKEN_UNKNOWN                = "TOKEN_UNKNOWN"
	TOKEN_VAR                    = "TOKEN_VAR"
	TOKEN_VOID                   = "TOKEN_VOID"
	TOKEN_WHILE                  = "TOKEN_WHILE"
)

var keywords = map[string]string{
	"and":         TOKEN_AND,
	"as":          TOKEN_AS,
	"break":       TOKEN_BREAK,
	"case":        TOKEN_CASE,
	"class":       TOKEN_CLASS,
	"constructor": TOKEN_CONSTRUCTOR,
	"continue":    TOKEN_CONTINUE,
	"default":     TOKEN_DEFAULT,
	"else":        TOKEN_ELSE,
	"enum":        TOKEN_ENUM,
	"false":       TOKEN_FALSE,
	"fn":          TOKEN_FN,
	"for":         TOKEN_FOR,
	"if":          TOKEN_IF,
	"import":      TOKEN_IMPORT,
	"in":          TOKEN_IN,
	"let":         TOKEN_LET,
	"match":       TOKEN_MATCH,
	"no":          TOKEN_NO,
	"not":         TOKEN_NOT,
	"or":          TOKEN_OR,
	"private":     TOKEN_PRIVATE,
	"public":      TOKEN_PUBLIC,
	"return":      TOKEN_RETURN,
	"self":        TOKEN_SELF,
	"true":        TOKEN_TRUE,
	"var":         TOKEN_VAR,
	"void":        TOKEN_VOID,
	"while":       TOKEN_WHILE,
}

var one_char_tokens = map[byte]string{
	'=':  TOKEN_ASSIGN,
	'*':  TOKEN_ASTERISK,
	':':  TOKEN_COLON,
	',':  TOKEN_COMMA,
	'.':  TOKEN_DOT,
	'>':  TOKEN_GREATER_THAN,
	'{':  TOKEN_LEFT_CURLY,
	'(':  TOKEN_LEFT_PAREN,
	'[':  TOKEN_LEFT_SQUARE,
	'<':  TOKEN_LESS_THAN,
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
	"+=": TOKEN_ASSIGN_ADD,
	"/=": TOKEN_ASSIGN_DIV,
	"*=": TOKEN_ASSIGN_MUL,
	"-=": TOKEN_ASSIGN_SUB,
	"::": TOKEN_DOUBLE_COLON,
	"++": TOKEN_DOUBLE_PLUS,
	"==": TOKEN_EQUALS,
	">=": TOKEN_GREATER_THAN_OR_EQUALS,
	"<=": TOKEN_LESS_THAN_OR_EQUALS,
	"!=": TOKEN_NOT_EQUALS,
}

var three_char_tokens = map[string]string{
	"...": TOKEN_ELLIPSIS,
}

func (l *Location) String() string {
	var filePath string
	if l.FilePath != "" {
		filePath = l.FilePath
	} else {
		filePath = "<string>"
	}
	return fmt.Sprintf("line %d, column %d of %s", l.Line, l.Column, filePath)
}

type LexerError struct {
	Message string
}

func (e *LexerError) Error() string {
	return e.Message
}
