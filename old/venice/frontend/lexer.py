import enum
from dataclasses import dataclass
from typing import List, Tuple

from venice import common
from venice.common import Location


class TokenType(enum.Enum):
    # Literals
    INTEGER = enum.auto()
    STRING = enum.auto()
    SYMBOL = enum.auto()

    # Punctuation
    ARROW = enum.auto()
    COLON = enum.auto()
    COMMA = enum.auto()
    DOT = enum.auto()
    LEFT_CURLY = enum.auto()
    LEFT_PAREN = enum.auto()
    LEFT_SQUARE = enum.auto()
    RIGHT_CURLY = enum.auto()
    RIGHT_PAREN = enum.auto()
    RIGHT_SQUARE = enum.auto()
    SEMICOLON = enum.auto()

    # Operators
    ASSIGN = enum.auto()
    MINUS = enum.auto()
    PLUS = enum.auto()
    SLASH = enum.auto()
    STAR = enum.auto()

    # Keywords
    FUNC = enum.auto()
    LET = enum.auto()
    RETURN = enum.auto()

    # Special token types
    END = enum.auto()
    UNKNOWN = enum.auto()


@dataclass
class Token:
    type: TokenType
    value: str
    location: Location


class Lexer:
    def __init__(self, program: str) -> None:
        self.program = program
        self.index = 0
        self.file = "<string>"
        self.line = 1
        self.column = 1

    def next(self) -> Token:
        self._consume_whitespace_and_comments()

        if self.done():
            return Token(type=TokenType.END, value="", location=self.location())

        location = self.location()
        c = self.program[self.index]

        if c in DOUBLE_CHAR_TOKENS:
            match, type_ = DOUBLE_CHAR_TOKENS[c]
            if self._peek() == match:
                self._advance()
                self._advance()
                return Token(type=type_, value=c + match, location=location)

        if c in SINGLE_CHAR_TOKENS:
            self._advance()
            type_ = SINGLE_CHAR_TOKENS[c]
            return Token(type=type_, value=c, location=location)
        elif is_symbol_first_char(c):
            value = self._consume_symbol()
            type_ = KEYWORDS.get(value, TokenType.SYMBOL)
            return Token(type=type_, value=value, location=location)
        elif c.isdigit():
            (value, type_) = self._consume_number()
            return Token(type=type_, value=value, location=location)
        elif c == '"':
            value = self._consume_string()
            return Token(type=TokenType.STRING, value=value, location=location)
        else:
            self._advance()
            return Token(type=TokenType.UNKNOWN, value=c, location=location)

    def done(self) -> bool:
        return self.index == len(self.program)

    def location(self) -> Location:
        return Location(file=self.file, line=self.line, column=self.column)

    def _consume_number(self) -> Tuple[str, TokenType]:
        start = self.index
        while not self.done() and self.program[self.index].isdigit():
            self._advance()

        return (self.program[start : self.index], TokenType.INTEGER)

    def _consume_string(self) -> str:
        # TODO: backslash escapes.

        # Move past the opening quotation mark.
        self._advance()
        start = self.index
        while not self.done() and self.program[self.index] != '"':
            self._advance()

        if self.done():
            raise common.VeniceSyntaxError("unclosed string literal")

        value = self.program[start : self.index]
        # Move past the closing quotation mark.
        self._advance()
        return value

    def _consume_symbol(self) -> str:
        start = self.index
        while not self.done() and is_symbol_char(self.program[self.index]):
            self._advance()

        return self.program[start : self.index]

    def _consume_whitespace_and_comments(self) -> None:
        while not self.done():
            while not self.done() and self.program[self.index].isspace():
                self._advance()

            if self.done():
                return

            if self.program[self.index] == "/" and self._peek() == "/":
                while not self.done() and self.program[self.index] != "\n":
                    self._advance()
            else:
                break

    def _advance(self) -> None:
        if self.done():
            return

        if self.program[self.index] == "\n":
            self.line += 1
            self.column = 1
        else:
            self.column += 1

        self.index += 1

    def _peek(self) -> str:
        if self.index + 1 < len(self.program):
            return self.program[self.index + 1]
        else:
            return ""


def is_symbol_first_char(c: str) -> bool:
    return c.isalpha() or c == "_"


def is_symbol_char(c: str) -> bool:
    return is_symbol_first_char(c) or c.isdigit()


def debug(program: str) -> List[Tuple[TokenType, str]]:
    pairs = []
    lexer = Lexer(program)
    while not lexer.done():
        token = lexer.next()
        pairs.append((token.type, token.value))

    return pairs


KEYWORDS = {
    "func": TokenType.FUNC,
    "let": TokenType.LET,
    "return": TokenType.RETURN,
}


DOUBLE_CHAR_TOKENS = {
    "-": (">", TokenType.ARROW),
}


SINGLE_CHAR_TOKENS = {
    "=": TokenType.ASSIGN,
    ":": TokenType.COLON,
    ",": TokenType.COMMA,
    ".": TokenType.DOT,
    "{": TokenType.LEFT_CURLY,
    "(": TokenType.LEFT_PAREN,
    "[": TokenType.LEFT_SQUARE,
    "-": TokenType.MINUS,
    "+": TokenType.PLUS,
    "}": TokenType.RIGHT_CURLY,
    ")": TokenType.RIGHT_PAREN,
    "]": TokenType.RIGHT_SQUARE,
    ";": TokenType.SEMICOLON,
    "/": TokenType.SLASH,
    "*": TokenType.STAR,
}
