#!/usr/bin/env python3
import sys
import textwrap
from collections import namedtuple
from io import StringIO


def main(args):
    # path = args[0]
    # output_path = os.path.splitext(path)[0] + ".vn"
    # with open(path, "r", encoding="utf8") as infile:
    #     with open(output_path, "w", encoding="utf8") as outfile:
    #         vcompile(infile, outfile)

    if len(args) == 2:
        if args[0] == "run":
            outfile = StringIO()
            with open(args[1], "r", encoding="utf8") as infile:
                vcompile(infile, outfile)

            program = outfile.getvalue()
            exec(program, {}, {})
        elif args[0] == "parse":
            with open(args[1], "r", encoding="utf8") as infile:
                ast = vparse(infile)

            pretty_print_tree(ast)
        elif args[0] == "tokenize":
            with open(args[1], "r", encoding="utf8") as infile:
                lexer = Lexer(infile)
                while True:
                    token = lexer.next()
                    if token.type == "TOKEN_EOF":
                        break
                    else:
                        if token.type in (
                            "TOKEN_STRING",
                            "TOKEN_NEWLINE",
                            "TOKEN_UNKNOWN",
                        ):
                            print(token.type.ljust(20), repr(token.value))
                        else:
                            print(token.type.ljust(20), token.value)
        else:
            print(f"Error: unknown subcommand {args[0]!r}", file=sys.stderr)
            sys.exit(1)
    elif len(args) == 1:
        if args[0] in ("-h", "--help"):
            print(
                textwrap.dedent(
                    """\
                v.py <path>           Compile the program and print the output.
                v.py run <path>       Compile and run the program.
                v.py parse <path>     Print the AST of the program.
                v.py tokenize<path>   Print the lexical tokens of the program.
                v.py --help           Print this help message."""
                )
            )
        else:
            path = args[0]
            with open(path, "r", encoding="utf8") as infile:
                vcompile(infile, sys.stdout)
    else:
        print(f"Error: expected 1 or 2 arguments, got {len(args)}", file=sys.stderr)
        sys.exit(1)


def vcompile(infile, outfile):
    ast = vparse(infile)
    vcheck(ast)
    vgenerate(outfile, ast)


def vcompile_string(program):
    infile = StringIO(program)
    outfile = StringIO()
    vcompile(infile, outfile)
    return outfile.getvalue()


def vgenerate(outfile, ast):
    if isinstance(ast, AstProgram):
        vgenerate_block(outfile, ast.statements)
    else:
        raise VeniceError("argument to vgenerate must be an AstProgram")


def vgenerate_block(outfile, statements, *, indent=0):
    for statement in statements:
        vgenerate_statement(outfile, statement, indent=indent)


def vgenerate_statement(outfile, ast, *, indent=0):
    if isinstance(ast, AstFunction):
        outfile.write(("  " * indent) + f"def {ast.label}(")
        outfile.write(", ".join(parameter.label for parameter in ast.parameters))
        outfile.write("):\n")
        vgenerate_block(outfile, ast.statements, indent=indent + 1)
    elif isinstance(ast, AstReturn):
        outfile.write(("  " * indent) + "return ")
        vgenerate_expression(outfile, ast.value, bracketed=False)
        outfile.write("\n")
    elif isinstance(ast, AstIf):
        for i, clause in enumerate(ast.if_clauses):
            outfile.write("  " * indent)
            if i == 0:
                outfile.write("if ")
            else:
                outfile.write("elif ")

            vgenerate_expression(outfile, clause.condition, bracketed=False)
            outfile.write(":\n")
            vgenerate_block(outfile, clause.statements, indent=indent + 1)

        if ast.else_clause:
            outfile.write(("  " * indent) + "else:\n")
            vgenerate_block(outfile, ast.else_clause, indent=indent + 1)
    elif isinstance(ast, (AstLet, AstAssign)):
        outfile.write("  " * indent)
        if isinstance(ast.label, str):
            outfile.write(ast.label)
        else:
            vgenerate_expression(outfile, ast.label, bracketed=False)
        outfile.write(" = ")
        vgenerate_expression(outfile, ast.value, bracketed=False)
        outfile.write("\n")
    elif isinstance(ast, AstWhile):
        outfile.write(("  " * indent) + "while ")
        vgenerate_expression(outfile, ast.condition, bracketed=False)
        outfile.write(":\n")
        vgenerate_block(outfile, ast.statements, indent=indent + 1)
    elif isinstance(ast, AstFor):
        outfile.write(("  " * indent) + "for " + ast.loop_variable + " in ")
        vgenerate_expression(outfile, ast.iterator, bracketed=False)
        outfile.write(":\n")
        vgenerate_block(outfile, ast.statements, indent=indent + 1)
    elif isinstance(ast, AstExpressionStatement):
        outfile.write("  " * indent)
        vgenerate_expression(outfile, ast.value, bracketed=False)
        outfile.write("\n")
    elif isinstance(ast, AstStructDeclaration):
        vgenerate_struct_declaration(outfile, ast, indent=indent)
    else:
        raise VeniceError(f"unknown AST statement type: {ast.__class__.__name__}")


def vgenerate_expression(outfile, ast, *, bracketed):
    if isinstance(ast, AstSymbol):
        outfile.write(ast.label)
    elif isinstance(ast, AstInfix):
        if bracketed:
            outfile.write("(")

        vgenerate_expression(outfile, ast.left, bracketed=True)
        outfile.write(" " + ast.operator + " ")
        vgenerate_expression(outfile, ast.right, bracketed=True)

        if bracketed:
            outfile.write(")")
    elif isinstance(ast, AstPrefix):
        if bracketed:
            outfile.write("(")

        outfile.write(ast.operator + " ")
        vgenerate_expression(outfile, ast.value, bracketed=True)

        if bracketed:
            outfile.write(")")
    elif isinstance(ast, AstCall):
        vgenerate_expression(outfile, ast.function, bracketed=True)
        outfile.write("(")
        for i, argument in enumerate(ast.arguments):
            if isinstance(argument, AstKeywordArgument):
                outfile.write(argument.label + "=")
                vgenerate_expression(outfile, argument.value, bracketed=True)
            else:
                vgenerate_expression(outfile, argument, bracketed=True)

            if i != len(ast.arguments) - 1:
                outfile.write(", ")
        outfile.write(")")
    elif isinstance(ast, AstList):
        outfile.write("[")
        for i, value in enumerate(ast.values):
            vgenerate_expression(outfile, value, bracketed=False)
            if i != len(ast.values) - 1:
                outfile.write(", ")
        outfile.write("]")
    elif isinstance(ast, AstLiteral):
        outfile.write(repr(ast.value))
    elif isinstance(ast, AstIndex):
        vgenerate_expression(outfile, ast.list, bracketed=True)
        outfile.write("[")
        vgenerate_expression(outfile, ast.index, bracketed=False)
        outfile.write("]")
    elif isinstance(ast, AstMap):
        outfile.write("{")
        for i, pair in enumerate(ast.pairs):
            vgenerate_expression(outfile, pair.key, bracketed=False)
            outfile.write(": ")
            vgenerate_expression(outfile, pair.value, bracketed=False)

            if i != len(ast.pairs) - 1:
                outfile.write(", ")
        outfile.write("}")
    else:
        raise VeniceError(f"unknown AST expression type: {ast.__class__.__name__}")


def vgenerate_struct_declaration(outfile, ast, *, indent):
    outfile.write(("  " * indent) + "class " + ast.label + ":\n")
    parameters = ", ".join(field.label for field in ast.fields)
    outfile.write(("  " * (indent + 1)) + f"def __init__(self, *, {parameters}):\n")
    for field in ast.fields:
        outfile.write(
            ("  " * (indent + 2)) + "self." + field.label + " = " + field.label + "\n"
        )

    outfile.write("\n")

    fields = repr([field.label for field in ast.fields])
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


def vcheck(ast):
    vcheck_block(ast.statements, SymbolTable.with_globals())


def vcheck_block(statements, symbol_table):
    for statement in statements:
        vcheck_statement(statement, symbol_table)


def vcheck_statement(ast, symbol_table):
    if isinstance(ast, AstFunction):
        parameter_types = [
            VeniceKeywordArgumentType(p.label, resolve_type(p.type))
            for p in ast.parameters
        ]
        symbol_table.put(
            ast.label,
            VeniceFunctionType(
                parameter_types=parameter_types,
                return_type=resolve_type(ast.return_type),
            ),
        )
    elif isinstance(ast, AstReturn):
        # TODO
        pass
    elif isinstance(ast, AstIf):
        for clause in ast.if_clauses:
            vassert(clause.condition, symbol_table, VENICE_TYPE_BOOLEAN)
            vcheck_block(clause.statements, symbol_table)

        if ast.else_clause:
            vcheck_block(ast.else_clause, symbol_table)
    elif isinstance(ast, AstLet):
        symbol_table.put(ast.label, vcheck_expression(ast.value, symbol_table))
    elif isinstance(ast, AstAssign):
        original_type = symbol_table.get(ast.label.label)
        if original_type is None:
            raise VeniceError(f"assignment to undefined variable: {ast.label}")

        vassert(ast.value, symbol_table, original_type)
    elif isinstance(ast, AstWhile):
        vassert(ast.condition, symbol_table, VENICE_TYPE_BOOLEAN)
        vcheck_block(ast.statements, symbol_table)
    elif isinstance(ast, AstFor):
        iterator_type = vcheck_expression(ast.iterator, symbol_table)
        if not isinstance(iterator_type, VeniceListType):
            raise VeniceError("loop iterator must be list")

        loop_variable_type = iterator_type.item_type
        loop_symbol_table = SymbolTable(parent=symbol_table)
        loop_symbol_table.put(ast.loop_variable, loop_variable_type)

        vcheck_block(ast.statements, loop_symbol_table)
    elif isinstance(ast, AstExpressionStatement):
        vcheck_expression(ast.value, symbol_table)
    elif isinstance(ast, AstStructDeclaration):
        field_types = [
            VeniceKeywordArgumentType(p.label, resolve_type(p.type)) for p in ast.fields
        ]
        symbol_table.put(
            ast.label, VeniceStructType(field_types=field_types),
        )
    else:
        raise VeniceError(f"unknown AST statement type: {ast.__class__.__name__}")


def vcheck_expression(ast, symbol_table):
    if isinstance(ast, AstSymbol):
        symbol_type = symbol_table.get(ast.label)
        if symbol_type is not None:
            return symbol_type
        else:
            raise VeniceError(f"undefined symbol: {ast.label}")
    elif isinstance(ast, AstInfix):
        vassert(ast.left, symbol_table, VENICE_TYPE_INTEGER)
        vassert(ast.right, symbol_table, VENICE_TYPE_INTEGER)
        if ast.operator in [">=", "<=", ">", "<", "==", "!="]:
            return VENICE_TYPE_BOOLEAN
        else:
            return VENICE_TYPE_INTEGER
    elif isinstance(ast, AstPrefix):
        if ast.operator == "not":
            vassert(ast.value, symbol_table, VENICE_TYPE_BOOLEAN)
            return VENICE_TYPE_BOOLEAN
        else:
            vassert(ast.value, symbol_table, VENICE_TYPE_INTEGER)
            return VENICE_TYPE_INTEGER
    elif isinstance(ast, AstCall):
        function_type = vcheck_expression(ast.function, symbol_table)
        if isinstance(function_type, VeniceFunctionType):
            if len(function_type.parameter_types) != len(ast.arguments):
                raise VeniceError(
                    f"expected {len(function_type.parameters)} arguments, "
                    + "got {len(ast.arguments)}"
                )

            for parameter, argument in zip(
                function_type.parameter_types, ast.arguments
            ):
                if isinstance(argument, AstKeywordArgument):
                    vassert(argument.value, symbol_table, parameter.type)
                else:
                    vassert(argument, symbol_table, parameter.type)

            return function_type.return_type
        elif isinstance(function_type, VeniceStructType):
            for parameter, argument in zip(function_type.field_types, ast.arguments):
                if not isinstance(argument, AstKeywordArgument):
                    raise VeniceError(
                        "struct constructor only accepts keyword arguments"
                    )

                if not argument.label == parameter.label:
                    raise VeniceError(
                        f"expected keyword argument {parameter.label}, "
                        + "got {argument.label}"
                    )

                vassert(argument.value, symbol_table, parameter.type)

            return function_type
        else:
            raise VeniceError(f"{function_type!r} is not a function type")
    elif isinstance(ast, AstList):
        # TODO: empty list
        item_type = vcheck_expression(ast.values[0], symbol_table)
        for value in ast.values[1:]:
            # TODO: Probably need a more robust way of checking item types, e.g.
            # collecting all types and seeing if there's a common super-type.
            another_item_type = vcheck_expression(value, symbol_table)
            if not are_types_compatible(item_type, another_item_type):
                raise VeniceError(
                    "list contains items of multiple types: "
                    + f"{item_type!r} and {another_item_type!r}"
                )

        return VeniceListType(item_type)
    elif isinstance(ast, AstLiteral):
        if isinstance(ast.value, str):
            return VENICE_TYPE_STRING
        elif isinstance(ast.value, bool):
            # This must come before `int` because bools are ints in Python.
            return VENICE_TYPE_BOOLEAN
        elif isinstance(ast.value, int):
            return VENICE_TYPE_INTEGER
        else:
            raise VeniceError(
                f"unknown AstLiteral type: {ast.value.__class__.__name__}"
            )
    elif isinstance(ast, AstIndex):
        list_type = vcheck_expression(ast.list, symbol_table)
        index_type = vcheck_expression(ast.index, symbol_table)

        if isinstance(list_type, VeniceListType):
            if index_type != VENICE_TYPE_INTEGER:
                raise VeniceError(
                    f"index expression must be of integer type, not {index_type!r}"
                )

            return list_type.item_type
        elif isinstance(list_type, VeniceMapType):
            if not are_types_compatible(list_type.key_type, index_type):
                raise VeniceError(
                    f"expected {list_type.key_type!r} for map key, "
                    + f"got {index_type!r}"
                )

            return list_type.value_type
        else:
            raise VeniceError(f"{list_type!r} is not a list type")
    elif isinstance(ast, AstMap):
        key_type = vcheck_expression(ast.pairs[0].key, symbol_table)
        value_type = vcheck_expression(ast.pairs[0].value, symbol_table)
        for pair in ast.pairs[1:]:
            another_key_type = vcheck_expression(pair.key, symbol_table)
            another_value_type = vcheck_expression(pair.value, symbol_table)
            if not are_types_compatible(key_type, another_key_type):
                raise VeniceError(
                    "map contains keys of multiple types: "
                    + f"{key_type!r} and {another_key_type!r}"
                )

            if not are_types_compatible(value_type, another_value_type):
                raise VeniceError(
                    "map contains values of multiple types: "
                    + f"{value_type!r} and {another_value_type!r}"
                )

        return VeniceMapType(key_type, value_type)
    else:
        raise VeniceError(f"unknown AST expression type: {ast.__class__.__name__}")


def vassert(ast, symbol_table, expected):
    actual = vcheck_expression(ast, symbol_table)
    if not are_types_compatible(expected, actual):
        raise VeniceError(f"expected {expected!r}, got {actual!r}")


def resolve_type(type_tree):
    if isinstance(type_tree, AstSymbol):
        if type_tree.label in {"boolean", "integer", "string"}:
            return VeniceType(type_tree.label)
        else:
            raise VeniceError(f"unknown type: {type_tree.label}")
    else:
        raise VeniceError(f"{type_tree!r} cannot be interpreted as a type")


def are_types_compatible(expected_type, actual_type):
    if expected_type == VENICE_TYPE_ANY:
        return True

    return expected_type == actual_type


VeniceType = namedtuple("VeniceType", ["label"])
VeniceListType = namedtuple("VeniceListType", ["item_type"])
VeniceFunctionType = namedtuple(
    "VeniceFunctionType", ["parameter_types", "return_type"]
)
VeniceStructType = namedtuple("VeniceStructType", ["field_types"])
VeniceKeywordArgumentType = namedtuple("VeniceKeywordArgumentType", ["label", "type"])
VeniceMapType = namedtuple("VeniceMapType", ["key_type", "value_type"])

VENICE_TYPE_BOOLEAN = VeniceType("boolean")
VENICE_TYPE_INTEGER = VeniceType("integer")
VENICE_TYPE_STRING = VeniceType("string")
VENICE_TYPE_VOID = VeniceType("void")
VENICE_TYPE_ANY = VeniceType("any")


class SymbolTable:
    def __init__(self, parent=None):
        self.parent = parent
        self.symbols = {}

    @classmethod
    def with_globals(cls):
        symbol_table = cls(parent=None)
        symbol_table.put(
            "print",
            VeniceFunctionType(
                [VeniceKeywordArgumentType(label="x", type=VENICE_TYPE_ANY)],
                return_type=VENICE_TYPE_VOID,
            ),
        )
        return symbol_table

    def has(self, symbol):
        if symbol in self.symbols:
            return True
        elif self.parent:
            return self.parent.has(symbol)
        else:
            return False

    def get(self, symbol):
        if symbol in self.symbols:
            return self.symbols[symbol]
        elif self.parent:
            return self.parent.get(symbol)
        else:
            return None

    def put(self, symbol, type):
        self.symbols[symbol] = type


def vparse(infile):
    return Parser(Lexer(infile)).parse()


AstProgram = namedtuple("AstProgram", ["statements"])

AstFunction = namedtuple(
    "AstFunction", ["label", "parameters", "return_type", "statements"]
)
AstParameter = namedtuple("AstParameter", ["label", "type"])

AstLet = namedtuple("AstLet", ["label", "value"])
AstReturn = namedtuple("AstReturn", ["value"])
AstIf = namedtuple("AstIf", ["if_clauses", "else_clause"])
AstIfClause = namedtuple("AstElifClause", ["condition", "statements"])
AstWhile = namedtuple("AstWhile", ["condition", "statements"])
AstFor = namedtuple("AstFor", ["loop_variable", "iterator", "statements"])
AstAssign = namedtuple("AstAssign", ["label", "value"])
AstStructDeclaration = namedtuple("AstStructDeclaration", ["label", "fields"])
AstStructDeclarationField = namedtuple("AstStructDeclarationField", ["label", "type"])
AstExpressionStatement = namedtuple("AstExpressionStatement", ["value"])

AstCall = namedtuple("AstCall", ["function", "arguments"])
AstIndex = namedtuple("AstIndex", ["list", "index"])
AstKeywordArgument = namedtuple("AstKeywordArgument", ["label", "value"])
AstInfix = namedtuple("AstInfix", ["operator", "left", "right"])
AstPrefix = namedtuple("AstPrefix", ["operator", "value"])
AstSymbol = namedtuple("AstSymbol", ["label"])
AstLiteral = namedtuple("AstLiteral", ["value"])
AstList = namedtuple("AstList", ["values"])
AstMap = namedtuple("AstMap", ["pairs"])
AstMapLiteralPair = namedtuple("AstMapLiteralPair", ["key", "value"])


# Based on https://docs.python.org/3.6/reference/expressions.html#operator-precedence
# Higher precedence means tighter-binding.
PRECEDENCE_LOWEST = 0
PRECEDENCE_ASSIGN = 1
PRECEDENCE_CMP = 2
PRECEDENCE_ADD_SUB = 3
PRECEDENCE_MUL_DIV = 4
PRECEDENCE_PREFIX = 5
PRECEDENCE_CALL = 6

PRECEDENCE_MAP = {
    "TOKEN_ASSIGN": PRECEDENCE_ASSIGN,
    "TOKEN_CMP": PRECEDENCE_CMP,
    "TOKEN_PLUS": PRECEDENCE_ADD_SUB,
    "TOKEN_MINUS": PRECEDENCE_ADD_SUB,
    "TOKEN_ASTERISK": PRECEDENCE_MUL_DIV,
    "TOKEN_SLASH": PRECEDENCE_MUL_DIV,
    # The left parenthesis is the "infix operator" for function-call expressions.
    "TOKEN_LPAREN": PRECEDENCE_CALL,
    "TOKEN_LSQUARE": PRECEDENCE_CALL,
}


class Parser:
    def __init__(self, lexer):
        self.lexer = lexer
        self.pushed_back = None

    def parse(self):
        statements = []

        while True:
            self.skip_newlines()

            token = self.next()
            if token.type == "TOKEN_EOF":
                break
            elif token.type == "TOKEN_FN":
                statements.append(self.match_function())
            else:
                self.push_back(token)
                statements.append(self.match_statement())

        return AstProgram(statements)

    def match_function(self):
        symbol_token = self.expect("TOKEN_SYMBOL")
        self.expect("TOKEN_LPAREN")
        parameters = self.match_parameters()
        self.expect("TOKEN_RPAREN")
        self.expect("TOKEN_COLON")
        return_type = self.match_type()
        statements = self.match_block()
        return AstFunction(
            label=symbol_token.value,
            parameters=parameters,
            return_type=return_type,
            statements=statements,
        )

    def match_block(self):
        self.expect("TOKEN_LCURLY")
        self.expect("TOKEN_NEWLINE")
        statements = []
        while True:
            self.skip_newlines()

            token = self.next()
            if token.type == "TOKEN_RCURLY":
                break
            else:
                self.push_back(token)
                statements.append(self.match_statement())

        return statements

    def match_statement(self):
        token = self.next()
        if token.type == "TOKEN_LET":
            return self.match_let()
        elif token.type == "TOKEN_RETURN":
            return self.match_return()
        elif token.type == "TOKEN_IF":
            return self.match_if()
        elif token.type == "TOKEN_WHILE":
            return self.match_while()
        elif token.type == "TOKEN_FOR":
            return self.match_for()
        elif token.type == "TOKEN_STRUCT":
            return self.match_struct_declaration()
        else:
            self.push_back(token)
            value = self.match_expression()
            self.expect("TOKEN_NEWLINE")
            if isinstance(value, AstAssign):
                return value
            else:
                return AstExpressionStatement(value)

    def match_let(self):
        symbol_token = self.expect("TOKEN_SYMBOL")
        self.expect("TOKEN_ASSIGN")
        value = self.match_expression()
        self.expect("TOKEN_NEWLINE")
        return AstLet(label=symbol_token.value, value=value)

    def match_return(self):
        value = self.match_expression()
        self.expect("TOKEN_NEWLINE")
        return AstReturn(value)

    def match_if(self):
        clauses = []
        condition = self.match_expression()
        statements = self.match_block()
        clauses.append(AstIfClause(condition, statements))
        else_clause = None
        while True:
            token = self.expect(
                ["TOKEN_ELIF", "TOKEN_ELSE", "TOKEN_NEWLINE", "TOKEN_EOF"]
            )
            if token.type == "TOKEN_ELIF":
                condition = self.match_expression()
                statements = self.match_block()
                clauses.append(AstIfClause(condition, statements))
            elif token.type == "TOKEN_ELSE":
                else_clause = self.match_block()
            else:
                break

        return AstIf(if_clauses=clauses, else_clause=else_clause)

    def match_while(self):
        condition = self.match_expression()
        statements = self.match_block()
        return AstWhile(condition, statements)

    def match_for(self):
        symbol_token = self.expect("TOKEN_SYMBOL")
        self.expect("TOKEN_IN")
        iterator = self.match_expression()
        statements = self.match_block()
        return AstFor(
            loop_variable=symbol_token.value, iterator=iterator, statements=statements
        )

    def match_struct_declaration(self):
        symbol_token = self.expect("TOKEN_SYMBOL")
        self.expect("TOKEN_LCURLY")
        fields = []
        while True:
            field_token = self.expect("TOKEN_SYMBOL")
            self.expect("TOKEN_COLON")
            type_tree = self.match_type()
            fields.append(AstStructDeclarationField(field_token.value, type_tree))
            token = self.expect(["TOKEN_COMMA", "TOKEN_RCURLY"])
            if token.type == "TOKEN_COMMA":
                continue
            else:
                break

        self.expect("TOKEN_NEWLINE")
        return AstStructDeclaration(symbol_token.value, fields)

    def match_expression(self, precedence=PRECEDENCE_LOWEST):
        left = self.match_prefix()

        token = self.next()
        while token.type in PRECEDENCE_MAP and precedence < PRECEDENCE_MAP[token.type]:
            left = self.match_infix(left, token, PRECEDENCE_MAP[token.type])
            token = self.next()

        self.push_back(token)
        return left

    def match_prefix(self):
        token = self.next()
        if token.type == "TOKEN_INT":
            left = AstLiteral(int(token.value))
        elif token.type == "TOKEN_TRUE":
            left = AstLiteral(True)
        elif token.type == "TOKEN_FALSE":
            left = AstLiteral(False)
        elif token.type == "TOKEN_SYMBOL":
            left = AstSymbol(token.value)
        elif token.type == "TOKEN_STRING":
            left = AstLiteral(token.value)
        elif token.type == "TOKEN_LPAREN":
            left = self.match_expression()
            self.expect("TOKEN_RPAREN")
        elif token.type == "TOKEN_MINUS":
            left = AstPrefix("-", self.match_expression(PRECEDENCE_PREFIX))
        elif token.type == "TOKEN_NOT":
            left = AstPrefix("not", self.match_expression(PRECEDENCE_PREFIX))
        elif token.type == "TOKEN_LSQUARE":
            values = self.match_sequence()
            self.expect("TOKEN_RSQUARE")
            return AstList(values)
        elif token.type == "TOKEN_LCURLY":
            key_value_pairs = self.match_map_literal()
            self.expect("TOKEN_RCURLY")
            return AstMap(key_value_pairs)
        else:
            if token.type == "TOKEN_EOF":
                raise VeniceError("premature end of input")
            else:
                raise VeniceError(f"unexpected token {token!r}")

        return left

    def match_infix(self, left, token, precedence):
        if token.type == "TOKEN_LPAREN":
            args = self.match_arguments()
            self.expect("TOKEN_RPAREN")
            return AstCall(left, args)
        elif token.type == "TOKEN_LSQUARE":
            index = self.match_expression()
            self.expect("TOKEN_RSQUARE")
            return AstIndex(left, index)
        elif token.type == "TOKEN_ASSIGN":
            right = self.match_expression(precedence)
            return AstAssign(left, right)
        else:
            right = self.match_expression(precedence)
            return AstInfix(token.value, left, right)

    def match_sequence(self):
        arguments = []
        while True:
            argument = self.match_expression()
            arguments.append(argument)

            token = self.next()
            if token.type != "TOKEN_COMMA":
                self.push_back(token)
                break

        return arguments

    def match_map_literal(self):
        pairs = []
        while True:
            key = self.match_expression()
            self.expect("TOKEN_COLON")
            value = self.match_expression()
            pairs.append(AstMapLiteralPair(key, value))

            token = self.next()
            if token.type != "TOKEN_COMMA":
                self.push_back(token)
                break

        return pairs

    def match_arguments(self):
        arguments = []
        while True:
            argument = self.match_expression()
            if isinstance(argument, AstSymbol):
                token = self.next()
                if token.type == "TOKEN_COLON":
                    label = argument.label
                    argument = self.match_expression()
                    arguments.append(AstKeywordArgument(label=label, value=argument))
                else:
                    self.push_back(token)
                    arguments.append(argument)
            else:
                arguments.append(argument)

            token = self.next()
            if token.type != "TOKEN_COMMA":
                self.push_back(token)
                break

        return arguments

    def match_parameters(self):
        parameters = []
        while True:
            symbol_token = self.expect("TOKEN_SYMBOL")
            self.expect("TOKEN_COLON")
            symbol_type = self.match_type()
            parameters.append(AstParameter(label=symbol_token.value, type=symbol_type))

            token = self.next()
            if token.type != "TOKEN_COMMA":
                self.push_back(token)
                break
        return parameters

    def match_type(self):
        symbol_token = self.expect("TOKEN_SYMBOL")
        return AstSymbol(symbol_token.value)

    def skip_newlines(self):
        while True:
            token = self.next()
            if token.type == "TOKEN_NEWLINE":
                continue
            else:
                self.push_back(token)
                break

    def next(self):
        if self.pushed_back is not None:
            token = self.pushed_back
            self.pushed_back = None
            return token
        else:
            return self.lexer.next()

    def push_back(self, token):
        self.pushed_back = token

    def expect(self, type_or_types):
        if isinstance(type_or_types, str):
            type_or_types = {type_or_types}
        else:
            type_or_types = set(type_or_types)

        token = self.next()
        if token.type not in type_or_types:
            if token.type == "TOKEN_EOF":
                raise VeniceError("premature end of input")
            else:
                raise VeniceError(f"unexpected token {token!r}")

        return token


class Lexer:
    keywords = frozenset(
        [
            "elif",
            "else",
            "false",
            "fn",
            "for",
            "if",
            "in",
            "let",
            "return",
            "struct",
            "true",
            "while",
        ]
    )
    special = {
        "(": "TOKEN_LPAREN",
        ")": "TOKEN_RPAREN",
        "{": "TOKEN_LCURLY",
        "}": "TOKEN_RCURLY",
        ",": "TOKEN_COMMA",
        "+": "TOKEN_PLUS",
        "-": "TOKEN_MINUS",
        "*": "TOKEN_ASTERISK",
        "/": "TOKEN_SLASH",
        ":": "TOKEN_COLON",
        "\n": "TOKEN_NEWLINE",
        "[": "TOKEN_LSQUARE",
        "]": "TOKEN_RSQUARE",
    }
    escapes = {
        '"': '"',
        "'": "'",
        "n": "\n",
        "r": "\r",
        "\\": "\\",
        "t": "\t",
        "b": "\b",
        "f": "\f",
        "v": "\v",
        "0": "\0",
    }

    def __init__(self, infile):
        self.infile = infile
        self.done = False
        self.pushed_back = ""
        self.line = 1
        self.column = 1

    def next(self):
        self.skip_whitespace()

        if self.done:
            return Token("TOKEN_EOF", "")

        c = self.read()
        if c.isalpha() or c == "_":
            self.push_back(c)
            value = self.read_symbol()
            if value == "not":
                return Token("TOKEN_NOT", value)
            elif value in self.keywords:
                return Token("TOKEN_" + value.upper(), value)
            else:
                return Token("TOKEN_SYMBOL", value)
        elif c.isdigit():
            self.push_back(c)
            value = self.read_int()
            return Token("TOKEN_INT", value)
        elif c == '"':
            value = self.read_string()
            return Token("TOKEN_STRING", value)
        elif c == ">":
            c2 = self.read()
            if c2 == "=":
                return Token("TOKEN_CMP", ">=")
            else:
                self.push_back(c2)
                return Token("TOKEN_CMP", ">")
        elif c == "<":
            c2 = self.read()
            if c2 == "=":
                return Token("TOKEN_CMP", ">=")
            else:
                self.push_back(c2)
                return Token("TOKEN_CMP", "<")
        elif c == "=":
            c2 = self.read()
            if c2 == "=":
                return Token("TOKEN_CMP", "==")
            else:
                self.push_back(c2)
                return Token("TOKEN_ASSIGN", "=")
        else:
            return Token(self.special.get(c, "TOKEN_UNKNOWN"), c)

    def read_symbol(self):
        return self.read_while(is_symbol_char)

    def read_int(self):
        return self.read_while(str.isdigit)

    def read_string(self):
        chars = []
        while True:
            c = self.read()
            if self.done:
                # TODO(2020-12-27): Better error
                return Token("TOKEN_UNKNOWN", "".join(chars))
            elif c == '"':
                break
            elif c == "\\":
                c2 = self.read()
                if self.done:
                    # TODO(2020-12-27): Better error
                    return Token("TOKEN_UNKNOWN", "".join(chars))
                else:
                    chars.append(self.get_backslash_escape(c2))
            else:
                chars.append(c)

        return "".join(chars)

    def get_backslash_escape(self, c):
        # TODO(2020-12-27): Warning for unrecognized escapes
        # TODO(2020-12-27): Hex and octal escape codes
        return self.escapes.get(c, "\\" + c)

    def push_back(self, c):
        self.pushed_back = c

    def skip_whitespace(self):
        self.read_while(lambda c: c.isspace() and c != "\n")

    def read_while(self, pred):
        chars = []
        while True:
            c = self.read()
            if pred(c):
                chars.append(c)
                continue
            else:
                self.push_back(c)
                break

        return "".join(chars)

    def read(self):
        if self.pushed_back:
            c = self.pushed_back
            self.pushed_back = ""
            return c
        else:
            c = self.infile.read(1)
            if c == "":
                self.done = True
            return c


Token = namedtuple("Token", ["type", "value"])


def is_symbol_char(c):
    return c.isdigit() or c.isalpha() or c == "_"


def pretty_print_tree(ast, indent=0):
    print(("  " * indent) + ast.__class__.__name__)
    for key, value in ast._asdict().items():
        print(("  " * (indent + 1)) + key + ":", end="")
        if isinstance(value, AstSymbol):
            print(f" AstSymbol({value.label!r})")
        elif isinstance(value, AstLiteral):
            print(f" AstLiteral({value.value!r})")
        else:
            print()
            if not isinstance(value, list):
                value = [value]

            for subvalue in value:
                if subvalue.__class__.__name__.startswith("Ast"):
                    pretty_print_tree(subvalue, indent + 2)
                else:
                    print(("  " * (indent + 2)) + repr(subvalue))


class VeniceError(Exception):
    pass


if __name__ == "__main__":
    main(sys.argv[1:])
