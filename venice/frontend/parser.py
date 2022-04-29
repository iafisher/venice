from typing import List, Optional, Union

from venice import ast
from venice.common import VeniceInternalError, VeniceSyntaxError

from .lexer import Lexer, Token, TokenType


def parse_module(program: str) -> ast.Module:
    parser = Parser(program)
    return parser.match_module()


# Operator precedence; higher values for operators that bind "tighter".
PRECEDENCE_LOWEST = 0
PRECEDENCE_ADDITION = 1
PRECEDENCE_MULTIPLICATION = 2
PRECEDENCE_CALL = 3


class Parser:
    lexer: Lexer
    pushed_back = Optional[Token]

    def __init__(self, program: str) -> None:
        self.lexer = Lexer(program)
        self.pushed_back = None

    def match_module(self) -> ast.Module:
        functions = []
        while True:
            token = self._next()
            if token.type == TokenType.FUNC:
                function = self.match_function()
                functions.append(function)
            elif token.type == TokenType.END:
                break
            else:
                self._error("function declaration", token)

        return ast.Module(functions=functions)

    def match_function(self) -> ast.Function:
        name = self._expect(TokenType.SYMBOL).value
        self._expect(TokenType.LEFT_PAREN)
        parameters = self.match_parameter_list()

        token = self._next()
        if token.type == TokenType.ARROW:
            return_type_node = self.match_type()
        elif token.type == TokenType.LEFT_CURLY:
            self._push_back(token)
            return_type_node = None
        else:
            self._error("arrow or start of block", token)

        body = self.match_body()
        return ast.Function(
            name=name, parameters=parameters, body=body, return_type=return_type_node
        )

    def match_body(self) -> List[ast.Statement]:
        self._expect(TokenType.LEFT_CURLY)

        body = []
        while True:
            token = self._next()
            if token.type == TokenType.RIGHT_CURLY:
                break
            else:
                self._push_back(token)

            stmt = self.match_statement()
            body.append(stmt)

        return body

    def match_statement(self) -> ast.Statement:
        token = self._next()
        if token.type == TokenType.LET:
            return self.match_let()
        elif token.type == TokenType.RETURN:
            return self.match_return()
        else:
            self._push_back(token)
            expr = self.match_expression()
            self._expect(TokenType.SEMICOLON)
            return ast.ExpressionStatement(expr)

    def match_let(self) -> ast.Let:
        symbol = self._expect(TokenType.SYMBOL).value
        self._expect(TokenType.ASSIGN)
        value = self.match_expression()
        self._expect(TokenType.SEMICOLON)
        return ast.Let(symbol=symbol, value=value)

    def match_return(self) -> ast.Return:
        value = self.match_expression()
        self._expect(TokenType.SEMICOLON)
        return ast.Return(value=value)

    def match_parameter_list(self) -> List[ast.Parameter]:
        parameters = []

        while True:
            token = self._next()
            if token.type == TokenType.RIGHT_PAREN:
                break
            elif token.type == TokenType.SYMBOL:
                name = token.value
                self._expect(TokenType.COLON)
                type_node = self.match_type()
                parameters.append(ast.Parameter(name=name, type=type_node))

                token = self._next()
                if token.type == TokenType.COMMA:
                    continue
                elif token.type == TokenType.RIGHT_PAREN:
                    break
                else:
                    self._error("comma or right parenthesis", token)
            else:
                self._error("parameter", token)

        return parameters

    def match_expression(self, precedence: int = PRECEDENCE_LOWEST) -> ast.Expression:
        expr = self.match_prefix()

        token = self._next()
        while token.type in PRECEDENCE_MAP:
            operator_precedence = PRECEDENCE_MAP[token.type]
            if precedence < operator_precedence:
                if token.type == TokenType.LEFT_PAREN:
                    # Function calls
                    arguments = self.match_argument_list()

                    # TODO: Parse non-symbol function calls.
                    if not isinstance(expr, ast.Symbol):
                        raise VeniceSyntaxError("function must be symbol")

                    expr = ast.FunctionCall(function=expr.value, arguments=arguments)
                    token = self._next()
                else:
                    # Regular infix expressions
                    right = self.match_expression(operator_precedence)
                    expr = ast.Infix(token.value, expr, right)
                    token = self._next()
            else:
                break

        self._push_back(token)
        return expr

    def match_prefix(self) -> ast.Expression:
        token = self._next()
        if token.type == TokenType.INTEGER:
            return ast.Integer(int(token.value))
        elif token.type == TokenType.STRING:
            return ast.String(token.value)
        elif token.type == TokenType.SYMBOL:
            return ast.Symbol(token.value)
        else:
            self._error("start of expression", token)

    def match_argument_list(self) -> List[ast.Expression]:
        arguments = []

        while True:
            token = self._next()
            if token.type == TokenType.RIGHT_PAREN:
                break
            else:
                self._push_back(token)
                argument = self.match_expression()
                arguments.append(argument)

                token = self._next()
                if token.type == TokenType.COMMA:
                    continue
                elif token.type == TokenType.RIGHT_PAREN:
                    break
                else:
                    self._error("comma or right parenthesis", token)

        return arguments

    def match_type(self) -> ast.Type:
        token = self._next()
        if token.type == TokenType.SYMBOL:
            return ast.SymbolType(token.value)
        else:
            self._error("Venice type", token)

    def _expect(self, type_: TokenType) -> Token:
        token = self._next()
        if token.type != type_:
            self._error(type_, token)
        return token

    def _next(self) -> Token:
        if self.pushed_back is not None:
            r = self.pushed_back
            self.pushed_back = None
            return r
        else:
            return self.lexer.next()

    def _push_back(self, token: Token) -> None:
        if self.pushed_back is not None:
            raise VeniceInternalError("cannot push back multiple tokens")

        self.pushed_back = token

    def _error(self, expected: Union[str, TokenType], token: Token) -> None:
        if isinstance(expected, TokenType):
            expected = expected.name

        raise VeniceSyntaxError(
            f"expected {expected}, got {token.type.name} at {token.location}"
        )


PRECEDENCE_MAP = {
    TokenType.MINUS: PRECEDENCE_ADDITION,
    TokenType.PLUS: PRECEDENCE_ADDITION,
    TokenType.SLASH: PRECEDENCE_MULTIPLICATION,
    TokenType.STAR: PRECEDENCE_MULTIPLICATION,
    TokenType.LEFT_PAREN: PRECEDENCE_CALL,
}
