from pycompiler import ast
from pycompiler.common import VeniceError


def vgenerate_javascript(outfile, tree):
    if isinstance(tree, ast.ProgramNode):
        vgenerate_block(outfile, tree.statements)
    else:
        raise VeniceError("argument to vgenerate must be an ast.ProgramNode")


def vgenerate_block(outfile, statements, *, indent=0):
    for statement in statements:
        vgenerate_statement(outfile, statement, indent=indent)


def vgenerate_statement(outfile, tree, *, indent=0):
    if isinstance(tree, ast.FunctionNode):
        outfile.write(("  " * indent) + f"function {tree.label}(")
        outfile.write(", ".join(parameter.label for parameter in tree.parameters))
        outfile.write(") {\n")
        vgenerate_block(outfile, tree.statements, indent=indent + 1)
        outfile.write(("  " * indent) + "}\n")
    elif isinstance(tree, ast.ReturnNode):
        outfile.write(("  " * indent) + "return ")
        vgenerate_expression(outfile, tree.value, bracketed=False)
        outfile.write(";\n")
    elif isinstance(tree, ast.IfNode):
        for i, clause in enumerate(tree.if_clauses):
            outfile.write("  " * indent)
            if i == 0:
                outfile.write("if (")
            else:
                outfile.write("else if (")

            vgenerate_expression(outfile, clause.condition, bracketed=False)
            outfile.write(") {\n")
            vgenerate_block(outfile, clause.statements, indent=indent + 1)

        if tree.else_clause:
            outfile.write(("  " * indent) + "} else {\n")
            vgenerate_block(outfile, tree.else_clause, indent=indent + 1)

        outfile.write(("  " * indent) + "}\n")
    elif isinstance(tree, ast.LetNode):
        outfile.write(("  " * indent) + "var ")
        if isinstance(tree.label, str):
            outfile.write(tree.label)
        else:
            vgenerate_expression(outfile, tree.label, bracketed=False)
        outfile.write(" = ")
        vgenerate_expression(outfile, tree.value, bracketed=False)
        outfile.write(";\n")
    elif isinstance(tree, ast.AssignNode):
        outfile.write("  " * indent)
        if isinstance(tree.label, str):
            outfile.write(tree.label)
        else:
            vgenerate_expression(outfile, tree.label, bracketed=False)
        outfile.write(" = ")
        vgenerate_expression(outfile, tree.value, bracketed=False)
        outfile.write(";\n")
    elif isinstance(tree, ast.WhileNode):
        outfile.write(("  " * indent) + "while (")
        vgenerate_expression(outfile, tree.condition, bracketed=False)
        outfile.write(") {\n")
        vgenerate_block(outfile, tree.statements, indent=indent + 1)
        outfile.write(("  " * indent) + "}\n")
    elif isinstance(tree, ast.ForNode):
        outfile.write(("  " * indent) + "for (var " + tree.loop_variable + " of ")
        vgenerate_expression(outfile, tree.iterator, bracketed=False)
        outfile.write(") { \n")
        vgenerate_block(outfile, tree.statements, indent=indent + 1)
        outfile.write(("  " * indent) + "}\n")
    elif isinstance(tree, ast.ExpressionStatementNode):
        outfile.write("  " * indent)
        vgenerate_expression(outfile, tree.value, bracketed=False)
        outfile.write(";\n")
    elif isinstance(tree, ast.StructDeclarationNode):
        vgenerate_struct_declaration(outfile, tree, indent=indent)
    else:
        raise VeniceError(f"unknown AST statement type: {tree.__class__.__name__}")


def vgenerate_expression(outfile, tree, *, bracketed):
    if isinstance(tree, ast.SymbolNode):
        outfile.write(tree.label)
    elif isinstance(tree, ast.InfixNode):
        if bracketed:
            outfile.write("(")

        vgenerate_expression(outfile, tree.left, bracketed=True)
        outfile.write(" " + tree.operator + " ")
        vgenerate_expression(outfile, tree.right, bracketed=True)

        if bracketed:
            outfile.write(")")
    elif isinstance(tree, ast.PrefixNode):
        if bracketed:
            outfile.write("(")

        outfile.write(tree.operator + " ")
        vgenerate_expression(outfile, tree.value, bracketed=True)

        if bracketed:
            outfile.write(")")
    elif isinstance(tree, ast.CallNode):
        vgenerate_expression(outfile, tree.function, bracketed=True)
        outfile.write("(")
        for i, argument in enumerate(tree.arguments):
            if isinstance(argument, ast.KeywordArgumentNode):
                outfile.write(argument.label + "=")
                vgenerate_expression(outfile, argument.value, bracketed=True)
            else:
                vgenerate_expression(outfile, argument, bracketed=True)

            if i != len(tree.arguments) - 1:
                outfile.write(", ")
        outfile.write(")")
    elif isinstance(tree, ast.ListNode):
        outfile.write("[")
        for i, value in enumerate(tree.values):
            vgenerate_expression(outfile, value, bracketed=False)
            if i != len(tree.values) - 1:
                outfile.write(", ")
        outfile.write("]")
    elif isinstance(tree, ast.LiteralNode):
        outfile.write(repr(tree.value))
    elif isinstance(tree, ast.IndexNode):
        vgenerate_expression(outfile, tree.list, bracketed=True)
        outfile.write("[")
        vgenerate_expression(outfile, tree.index, bracketed=False)
        outfile.write("]")
    elif isinstance(tree, ast.MapNode):
        outfile.write("{")
        for i, pair in enumerate(tree.pairs):
            vgenerate_expression(outfile, pair.key, bracketed=False)
            outfile.write(": ")
            vgenerate_expression(outfile, pair.value, bracketed=False)

            if i != len(tree.pairs) - 1:
                outfile.write(", ")
        outfile.write("}")
    elif isinstance(tree, ast.FieldAccessNode):
        vgenerate_expression(outfile, tree.value, bracketed=True)
        outfile.write(".")
        outfile.write(tree.field.value)
    else:
        raise VeniceError(f"unknown AST expression type: {tree.__class__.__name__}")


def vgenerate_struct_declaration(outfile, tree, *, indent):
    return
