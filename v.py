#!/usr/bin/env python3
import sys
from collections import namedtuple
from io import StringIO


def main(args):
    # path = args[0]
    # output_path = os.path.splitext(path)[0] + ".vn"
    # with open(path, "r", encoding="utf8") as infile:
    #     with open(output_path, "w", encoding="utf8") as outfile:
    #         vcompile(infile, outfile)

    print(vcompile_string(""))


def vcompile(infile, outfile):
    ast = vparse(infile)
    vcheck(ast)
    vgenerate(outfile, ast)


def vcompile_string(program):
    infile = StringIO(program)
    outfile = StringIO()
    vcompile(infile, outfile)
    return outfile.getvalue()


def vgenerate(outfile, ast, *, indent=0):
    if isinstance(ast, AstFunction):
        write_with_indent(outfile, f"def {ast.label}(", indent=indent)
        outfile.write(", ".join(parameter.label for parameter in ast.parameters))
        outfile.write("):\n")
        for statement in ast.statements:
            vgenerate(outfile, statement, indent=indent + 1)
    elif isinstance(ast, AstReturn):
        write_with_indent(outfile, "return ", indent=indent)
        vgenerate_expression(outfile, ast.value)
    else:
        raise VeniceError(f"unknown AST type: {ast.__class__.__name__}")


def vgenerate_expression(outfile, ast):
    if isinstance(ast, AstSymbol):
        outfile.write(ast.label)
    elif isinstance(ast, AstInfix):
        outfile.write("(")
        vgenerate_expression(outfile, ast.left)
        outfile.write(") ")
        outfile.write(ast.operator)
        outfile.write(" (")
        vgenerate_expression(outfile, ast.right)
        outfile.write(")")
    elif isinstance(ast, AstLiteral):
        outfile.write(repr(ast.value))
    else:
        raise VeniceError(f"unknown AST type: {ast.__class__.__name__}")


def vcheck(ast):
    # TODO(2020-12-24): Real type-checker.
    pass


def vparse(infile):
    # TODO(2020-12-24): Real parser.

    # fn double(x: int): int {
    #   return x * 2
    # }
    return AstFunction(
        label="double",
        parameters=[AstParameter(label="x", type=AstSymbol("int"))],
        return_type=AstSymbol("int"),
        statements=[
            AstReturn(AstInfix(operator="*", left=AstSymbol("x"), right=AstLiteral(2)))
        ],
    )


AstFunction = namedtuple(
    "AstFunction", ["label", "parameters", "return_type", "statements"]
)
AstParameter = namedtuple("AstParameter", ["label", "type"])
AstSymbol = namedtuple("AstSymbol", ["label"])
AstReturn = namedtuple("AstReturn", ["value"])
AstInfix = namedtuple("AstInfix", ["operator", "left", "right"])
AstLiteral = namedtuple("AstLiteral", ["value"])


def write_with_indent(outfile, contents, *, indent):
    if indent != 0:
        outfile.write("    " * indent)

    outfile.write(contents)


class VeniceError(Exception):
    pass


if __name__ == "__main__":
    main(sys.argv[1:])
