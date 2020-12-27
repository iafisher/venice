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
        else:
            print(f"Error: unknown subcommand {args[0]!r}", file=sys.stderr)
            sys.exit(1)
    elif len(args) == 1:
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
    if indent > 0:
        outfile.write("  " * indent)

    if isinstance(ast, AstFunction):
        outfile.write(f"def {ast.label}(")
        outfile.write(", ".join(parameter.label for parameter in ast.parameters))
        outfile.write("):\n")
        vgenerate_block(outfile, ast.statements, indent=indent + 1)
    elif isinstance(ast, AstReturn):
        outfile.write("return ")
        vgenerate_expression(outfile, ast.value, bracketed=False)
        outfile.write("\n")
    elif isinstance(ast, AstIf):
        outfile.write("if ")
        vgenerate_expression(outfile, ast.condition, bracketed=False)
        outfile.write(":\n")
        vgenerate_block(outfile, ast.true_clause, indent=indent + 1)
        outfile.write(("  " * indent) + "else:\n")
        vgenerate_block(outfile, ast.false_clause, indent=indent + 1)
    elif isinstance(ast, AstExpressionStatement):
        vgenerate_expression(outfile, ast.value, bracketed=False)
        outfile.write("\n")
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
    elif isinstance(ast, AstCall):
        vgenerate_expression(outfile, ast.function, bracketed=True)
        outfile.write("(")
        for i, argument in enumerate(ast.arguments):
            vgenerate_expression(outfile, argument, bracketed=True)
            if i != len(ast.arguments) - 1:
                outfile.write(", ")
        outfile.write(")")
    elif isinstance(ast, AstLiteral):
        outfile.write(repr(ast.value))
    else:
        raise VeniceError(f"unknown AST expression type: {ast.__class__.__name__}")


def vcheck(ast):
    # TODO(2020-12-24): Real type-checker.
    pass


def vparse(infile):
    return Parser(Lexer(infile)).parse()


AstProgram = namedtuple("AstProgram", ["statements"])

AstFunction = namedtuple(
    "AstFunction", ["label", "parameters", "return_type", "statements"]
)
AstParameter = namedtuple("AstParameter", ["label", "type"])

AstLet = namedtuple("AstLet", ["label", "value"])
AstReturn = namedtuple("AstReturn", ["value"])
AstIf = namedtuple("AstIf", ["condition", "true_clause", "false_clause"])
AstExpressionStatement = namedtuple("AstExpressionStatement", ["value"])

AstCall = namedtuple("AstCall", ["function", "arguments"])
AstInfix = namedtuple("AstInfix", ["operator", "left", "right"])
AstPrefix = namedtuple("AstPrefix", ["operator", "value"])
AstSymbol = namedtuple("AstSymbol", ["label"])
AstLiteral = namedtuple("AstLiteral", ["value"])


PRECEDENCE_LOWEST = 0
PRECEDENCE_ADD_SUB = 1
PRECEDENCE_MUL_DIV = 2
PRECEDENCE_PREFIX = 3
PRECEDENCE_CALL = 4

PRECEDENCE_MAP = {
    "TOKEN_PLUS": PRECEDENCE_ADD_SUB,
    "TOKEN_MINUS": PRECEDENCE_ADD_SUB,
    "TOKEN_ASTERISK": PRECEDENCE_MUL_DIV,
    "TOKEN_SLASH": PRECEDENCE_MUL_DIV,
    # The left parenthesis is the "infix operator" for function-call expressions.
    "TOKEN_LPAREN": PRECEDENCE_CALL,
}


class Parser:
    def __init__(self, lexer):
        self.lexer = lexer
        self.pushed_back = None

    def parse(self):
        statements = []

        while True:
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
        statements = []
        while True:
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
        else:
            self.push_back(token)
            value = self.match_expression()
            return AstExpressionStatement(value)

    def match_let(self):
        symbol_token = self.expect("TOKEN_SYMBOL")
        self.expect("TOKEN_ASSIGN")
        value = self.match_expression()
        return AstLet(label=symbol_token.value, value=value)

    def match_return(self):
        value = self.match_expression()
        return AstReturn(value)

    def match_if(self):
        condition = self.match_expression()
        true_clause = self.match_block()
        self.expect("TOKEN_ELSE")
        false_clause = self.match_block()
        return AstIf(
            condition=condition, true_clause=true_clause, false_clause=false_clause
        )

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
        else:
            if token.type == "TOKEN_EOF":
                raise VeniceError("premature end of input")
            else:
                raise VeniceError(f"unexpected token {token.value!r}")

        return left

    def match_infix(self, left, token, precedence):
        if token.type == "TOKEN_LPAREN":
            args = self.match_arguments()
            self.expect("TOKEN_RPAREN")
            return AstCall(left, args)
        else:
            right = self.match_expression(precedence)
            return AstInfix(token.value, left, right)

    def match_arguments(self):
        arguments = []
        while True:
            argument = self.match_expression()
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

    def next(self):
        if self.pushed_back is not None:
            token = self.pushed_back
            self.pushed_back = None
            return token
        else:
            return self.lexer.next()

    def push_back(self, token):
        self.pushed_back = token

    def expect(self, type):
        token = self.next()
        if token.type != type:
            if type == "TOKEN_EOF":
                raise VeniceError("trailing input")
            elif token.type == "TOKEN_EOF":
                raise VeniceError("premature end of input")
            else:
                raise VeniceError(f"unexpected token {token!r}")

        return token


class Lexer:
    keywords = frozenset(["fn", "let", "if", "then", "else", "true", "false", "return"])
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
        "=": "TOKEN_ASSIGN",
        ":": "TOKEN_COLON",
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

    def next(self):
        self.skip_whitespace()

        if self.done:
            return Token("TOKEN_EOF", "")

        c = self.read()
        if c.isalpha() or c == "_":
            self.push_back(c)
            value = self.read_symbol()
            if value in self.keywords:
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
        self.read_while(str.isspace)

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
