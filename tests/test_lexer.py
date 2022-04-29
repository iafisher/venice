import unittest

from venice.frontend import lexer
from venice.frontend.lexer import TokenType


class TestLexer(unittest.TestCase):
    def test_function_call(self):
        self.assertEqual(
            lexer.debug("add(2, 3)"),
            [
                (TokenType.SYMBOL, "add"),
                (TokenType.LEFT_PAREN, "("),
                (TokenType.INTEGER, "2"),
                (TokenType.COMMA, ","),
                (TokenType.INTEGER, "3"),
                (TokenType.RIGHT_PAREN, ")"),
            ],
        )

    def test_arithmetic(self):
        self.assertEqual(
            lexer.debug("1 + 2 - 3 * 4 / 5"),
            [
                (TokenType.INTEGER, "1"),
                (TokenType.PLUS, "+"),
                (TokenType.INTEGER, "2"),
                (TokenType.MINUS, "-"),
                (TokenType.INTEGER, "3"),
                (TokenType.STAR, "*"),
                (TokenType.INTEGER, "4"),
                (TokenType.SLASH, "/"),
                (TokenType.INTEGER, "5"),
            ],
        )

    def test_object_and_array_access(self):
        self.assertEqual(
            lexer.debug("_o.a1[10]"),
            [
                (TokenType.SYMBOL, "_o"),
                (TokenType.DOT, "."),
                (TokenType.SYMBOL, "a1"),
                (TokenType.LEFT_SQUARE, "["),
                (TokenType.INTEGER, "10"),
                (TokenType.RIGHT_SQUARE, "]"),
            ],
        )

    def test_map(self):
        self.assertEqual(
            lexer.debug('{1: "one"}'),
            [
                (TokenType.LEFT_CURLY, "{"),
                (TokenType.INTEGER, "1"),
                (TokenType.COLON, ":"),
                (TokenType.STRING, "one"),
                (TokenType.RIGHT_CURLY, "}"),
            ],
        )

    def test_start_of_functino(self):
        self.assertEqual(
            lexer.debug("func add(x: int, y: int) -> int {"),
            [
                (TokenType.FUNC, "func"),
                (TokenType.SYMBOL, "add"),
                (TokenType.LEFT_PAREN, "("),
                (TokenType.SYMBOL, "x"),
                (TokenType.COLON, ":"),
                (TokenType.SYMBOL, "int"),
                (TokenType.COMMA, ","),
                (TokenType.SYMBOL, "y"),
                (TokenType.COLON, ":"),
                (TokenType.SYMBOL, "int"),
                (TokenType.RIGHT_PAREN, ")"),
                (TokenType.ARROW, "->"),
                (TokenType.SYMBOL, "int"),
                (TokenType.LEFT_CURLY, "{"),
            ],
        )

    def test_let_statement(self):
        self.assertEqual(
            lexer.debug("let x = y;"),
            [
                (TokenType.LET, "let"),
                (TokenType.SYMBOL, "x"),
                (TokenType.ASSIGN, "="),
                (TokenType.SYMBOL, "y"),
                (TokenType.SEMICOLON, ";"),
            ],
        )

    def test_comment(self):
        self.assertEqual(
            lexer.debug("// Comment\n1 + 3"),
            [
                (TokenType.INTEGER, "1"),
                (TokenType.PLUS, "+"),
                (TokenType.INTEGER, "3"),
            ],
        )
