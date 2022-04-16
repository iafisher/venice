import textwrap
import unittest

from venice import ast, codegen


class TestCodegen(unittest.TestCase):
    def test_let(self):
        tree = ast.Let(symbol="x", value=ast.Integer(42))
        self.assertEqual(codegen.codegen(tree), "venice_int_t x = 42;")

    def test_function(self):
        tree = ast.Module(
            functions=[
                ast.Function(
                    name="donothing",
                    parameters=[ast.Parameter(name="x", type=ast.SymbolType("int"))],
                    return_type=None,
                    body=[
                        ast.Let(
                            symbol="y",
                            value=ast.Infix("+", ast.Symbol("x"), ast.Integer(1)),
                        ),
                    ],
                )
            ]
        )
        self.assertEqual(
            codegen.codegen(tree),
            S(
                """
                #include <venice.h>

                void donothing(venice_int_t x) {
                  venice_int_t y = x + 1;
                }
                """
            ),
        )

    def test_print_statement(self):
        tree = make_statement(
            ast.ExpressionStatement(
                ast.FunctionCall(
                    function="print", arguments=[ast.String("Hello, world")]
                )
            )
        )

        self.assertEqual(
            codegen.codegen(tree),
            S(
                """
                #include <venice.h>

                void dummy() {
                  venice_print(venice_string_new("Hello, world"));
                }
                """
            ),
        )


def make_statement(statement: ast.Statement):
    return ast.Module(
        functions=[
            ast.Function(
                name="dummy",
                parameters=[],
                return_type=None,
                body=[statement],
            )
        ]
    )


def S(s: str) -> str:
    """
    Dedents and strips a string so that it can be used in an assertEquals call.
    """
    return textwrap.dedent(s).strip() + "\n"
