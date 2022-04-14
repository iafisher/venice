import enum
from dataclasses import dataclass
from typing import List, Tuple

from .common import Location


class TokenType(enum.Enum):
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
        raise NotImplementedError

    def next(self) -> Token:
        raise NotImplementedError

    def done(self) -> bool:
        raise NotImplementedError


def debug(program: str) -> List[Tuple[TokenType, str]]:
    pairs = []
    lexer = Lexer(program)
    while not lexer.done():
        token = lexer.next()
        pairs.append((token.type, token.value))

    return pairs
