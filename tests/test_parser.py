from venice import ast
from venice.common import VeniceSyntaxError
from venice.frontend import parser


def test_parse_literals():
    assert parse_expression("123") == ast.Integer(123)
    assert parse_expression('"words"') == ast.String("words")
    assert parse_expression("abc") == ast.Symbol("abc")


def test_parse_function_call():
    assert parse_expression("add(1, 2)") == ast.FunctionCall(
        "add", [ast.Integer(1), ast.Integer(2)]
    )

    assert parse_statement("add(1, 2);") == ast.ExpressionStatement(
        ast.FunctionCall("add", [ast.Integer(1), ast.Integer(2)])
    )


def test_parse_simple_functions():
    assert parse_function(
        """
            func empty() {}
            """
    ) == ast.Function(
        name="empty",
        parameters=[],
        body=[],
        return_type=None,
    )

    assert parse_function(
        """
            func add(x: int, y: int) -> int {
              return x + y;
            }
            """
    ) == ast.Function(
        name="add",
        parameters=[
            ast.Parameter(name="x", type=ast.SymbolType("int")),
            ast.Parameter(name="y", type=ast.SymbolType("int")),
        ],
        body=[
            ast.Return(ast.Infix("+", ast.Symbol("x"), ast.Symbol("y"))),
        ],
        return_type=ast.SymbolType("int"),
    )

    assert parse_function(
        """
            func add_and_print(x: int, y: int) -> void {
              print(x + y);
            }
            """
    ) == ast.Function(
        name="add_and_print",
        parameters=[
            ast.Parameter(name="x", type=ast.SymbolType("int")),
            ast.Parameter(name="y", type=ast.SymbolType("int")),
        ],
        body=[
            ast.ExpressionStatement(
                ast.FunctionCall(
                    function="print",
                    arguments=[ast.Infix("+", ast.Symbol("x"), ast.Symbol("y"))],
                )
            )
        ],
        return_type=ast.SymbolType("void"),
    )


def test_parse_expression_statement():
    assert parse_statement("1 + 2;") == ast.ExpressionStatement(
        ast.Infix("+", ast.Integer(1), ast.Integer(2))
    )


def parse_expression(program: str) -> ast.Expression:
    p = parser.Parser(program)
    e = p.match_expression()
    if not p.lexer.done():
        raise VeniceSyntaxError("trailing input")
    return e


def parse_statement(program: str) -> ast.Statement:
    p = parser.Parser(program)
    stmt = p.match_statement()
    if not p.lexer.done():
        raise VeniceSyntaxError("trailing input")
    return stmt


def parse_function(program: str) -> ast.Function:
    p = parser.Parser(program)
    mod = p.match_module()
    if not p.lexer.done():
        raise VeniceSyntaxError("trailing input")

    if len(mod.functions) != 1:
        raise VeniceSyntaxError(f"expected 1 function, got {len(mod.functions)}")

    return mod.functions[0]
