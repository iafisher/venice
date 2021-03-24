import textwrap

from pycompiler import ast
from pycompiler.common import VeniceError


def vgenerate_python(outfile, tree):
    if isinstance(tree, ast.Program):
        vgenerate_block(outfile, tree.statements)
    else:
        raise VeniceError("argument to vgenerate must be an ast.Program")


def vgenerate_block(outfile, statements, *, indent=0):
    for statement in statements:
        vgenerate_statement(outfile, statement, indent=indent)


def vgenerate_statement(outfile, tree, *, indent=0):
    if isinstance(tree, ast.Function):
        outfile.write(("  " * indent) + f"def {tree.label}(")
        outfile.write(", ".join(parameter.label for parameter in tree.parameters))
        outfile.write("):\n")
        vgenerate_block(outfile, tree.statements, indent=indent + 1)
    elif isinstance(tree, ast.Return):
        outfile.write(("  " * indent) + "return ")
        vgenerate_expression(outfile, tree.value, bracketed=False)
        outfile.write("\n")
    elif isinstance(tree, ast.If):
        for i, clause in enumerate(tree.if_clauses):
            outfile.write("  " * indent)
            if i == 0:
                outfile.write("if ")
            else:
                outfile.write("elif ")

            vgenerate_expression(outfile, clause.condition, bracketed=False)
            outfile.write(":\n")
            vgenerate_block(outfile, clause.statements, indent=indent + 1)

        if tree.else_clause:
            outfile.write(("  " * indent) + "else:\n")
            vgenerate_block(outfile, tree.else_clause, indent=indent + 1)
    elif isinstance(tree, (ast.Let, ast.Assign)):
        outfile.write("  " * indent)
        if isinstance(tree.label, str):
            outfile.write(tree.label)
        else:
            vgenerate_expression(outfile, tree.label, bracketed=False)
        outfile.write(" = ")
        vgenerate_expression(outfile, tree.value, bracketed=False)
        outfile.write("\n")
    elif isinstance(tree, ast.While):
        outfile.write(("  " * indent) + "while ")
        vgenerate_expression(outfile, tree.condition, bracketed=False)
        outfile.write(":\n")
        vgenerate_block(outfile, tree.statements, indent=indent + 1)
    elif isinstance(tree, ast.For):
        outfile.write(("  " * indent) + "for " + tree.loop_variable + " in ")
        vgenerate_expression(outfile, tree.iterator, bracketed=False)
        outfile.write(":\n")
        vgenerate_block(outfile, tree.statements, indent=indent + 1)
    elif isinstance(tree, ast.ExpressionStatement):
        outfile.write("  " * indent)
        vgenerate_expression(outfile, tree.value, bracketed=False)
        outfile.write("\n")
    elif isinstance(tree, ast.StructDeclaration):
        vgenerate_struct_declaration(outfile, tree, indent=indent)
    else:
        raise VeniceError(f"unknown AST statement type: {tree.__class__.__name__}")


def vgenerate_expression(outfile, tree, *, bracketed):
    if isinstance(tree, ast.Symbol):
        outfile.write(tree.label)
    elif isinstance(tree, ast.Infix):
        if bracketed:
            outfile.write("(")

        vgenerate_expression(outfile, tree.left, bracketed=True)
        outfile.write(" " + tree.operator + " ")
        vgenerate_expression(outfile, tree.right, bracketed=True)

        if bracketed:
            outfile.write(")")
    elif isinstance(tree, ast.Prefix):
        if bracketed:
            outfile.write("(")

        outfile.write(tree.operator + " ")
        vgenerate_expression(outfile, tree.value, bracketed=True)

        if bracketed:
            outfile.write(")")
    elif isinstance(tree, ast.Call):
        vgenerate_expression(outfile, tree.function, bracketed=True)
        outfile.write("(")
        for i, argument in enumerate(tree.arguments):
            if isinstance(argument, ast.KeywordArgument):
                outfile.write(argument.label + "=")
                vgenerate_expression(outfile, argument.value, bracketed=True)
            else:
                vgenerate_expression(outfile, argument, bracketed=True)

            if i != len(tree.arguments) - 1:
                outfile.write(", ")
        outfile.write(")")
    elif isinstance(tree, ast.List):
        outfile.write("[")
        for i, value in enumerate(tree.values):
            vgenerate_expression(outfile, value, bracketed=False)
            if i != len(tree.values) - 1:
                outfile.write(", ")
        outfile.write("]")
    elif isinstance(tree, ast.Literal):
        outfile.write(repr(tree.value))
    elif isinstance(tree, ast.Index):
        vgenerate_expression(outfile, tree.list, bracketed=True)
        outfile.write("[")
        vgenerate_expression(outfile, tree.index, bracketed=False)
        outfile.write("]")
    elif isinstance(tree, ast.Map):
        outfile.write("{")
        for i, pair in enumerate(tree.pairs):
            vgenerate_expression(outfile, pair.key, bracketed=False)
            outfile.write(": ")
            vgenerate_expression(outfile, pair.value, bracketed=False)

            if i != len(tree.pairs) - 1:
                outfile.write(", ")
        outfile.write("}")
    elif isinstance(tree, ast.FieldAccess):
        vgenerate_expression(outfile, tree.value, bracketed=True)
        outfile.write(".")
        outfile.write(tree.field.value)
    else:
        raise VeniceError(f"unknown AST expression type: {tree.__class__.__name__}")


def vgenerate_struct_declaration(outfile, tree, *, indent):
    outfile.write(("  " * indent) + "class " + tree.label + ":\n")
    parameters = ", ".join(field.label for field in tree.fields)
    outfile.write(("  " * (indent + 1)) + f"def __init__(self, *, {parameters}):\n")
    for field in tree.fields:
        outfile.write(
            ("  " * (indent + 2)) + "self." + field.label + " = " + field.label + "\n"
        )

    outfile.write("\n")

    fields = repr([field.label for field in tree.fields])
    outfile.write(textwrap.indent(STRUCT_STR_TEMPLATE % fields, "  " * (indent + 1)))


STRUCT_STR_TEMPLATE = """\
def __str__(self):
    builder = [self.__class__.__name__, "("]
    fields = %s
    for i, field in enumerate(fields):
        builder.append(field + ": ")
        builder.append(repr(getattr(self, field)))
        if i != len(fields) - 1:
            builder.append(", ")
    builder.append(")")
    return "".join(builder)
"""
