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

    if len(args) == 2 and args[0] == "run":
        vrun(args[1])
    else:
        path = args[0]
        with open(path, "r", encoding="utf8") as infile:
            vcompile(infile, sys.stdout)


def vrun(path):
    outfile = StringIO()
    with open(path, "r", encoding="utf8") as infile:
        vcompile(infile, outfile)

    program = outfile.getvalue()
    exec(program, {}, {})


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
            outfile.write("\n")
    elif isinstance(ast, AstReturn):
        write_with_indent(outfile, "return ", indent=indent)
        vgenerate_expression(outfile, ast.value, bracketed=False)
    else:
        write_with_indent(outfile, "", indent=indent)
        vgenerate_expression(outfile, ast, bracketed=False)


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
        raise VeniceError(f"unknown AST type: {ast.__class__.__name__}")


def vcheck(ast):
    # TODO(2020-12-24): Real type-checker.
    pass


def vparse(infile):
    return Parser(Lexer(infile)).parse()


AstFunction = namedtuple(
    "AstFunction", ["label", "parameters", "return_type", "statements"]
)
AstParameter = namedtuple("AstParameter", ["label", "type"])

AstLet = namedtuple("AstLet", ["label", "value"])
AstReturn = namedtuple("AstReturn", ["value"])

AstCall = namedtuple("AstCall", ["function", "arguments"])
AstInfix = namedtuple("AstInfix", ["operator", "left", "right"])
AstPrefix = namedtuple("AstPrefix", ["operator", "value"])
AstSymbol = namedtuple("AstSymbol", ["label"])
AstLiteral = namedtuple("AstLiteral", ["value"])


class Parser:
    def __init__(self, lexer):
        self.lexer = lexer
        self.pushed_back = None

    def parse(self):
        token = self.next()
        if token.type == "TOKEN_FN":
            ast = self.match_function()
        else:
            self.push_back(token)
            ast = self.match_statement()

        self.expect("TOKEN_EOF")
        return ast

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
        else:
            self.push_back(token)
            return self.match_expression(PRECEDENCE_LOWEST)

    def match_let(self):
        symbol_token = self.expect("TOKEN_SYMBOL")
        self.expect("TOKEN_ASSIGN")
        value = self.match_expression(PRECEDENCE_LOWEST)
        return AstLet(label=symbol_token.value, value=value)

    def match_return(self):
        value = self.match_expression(PRECEDENCE_LOWEST)
        return AstReturn(value)

    def match_expression(self, precedence):
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
        elif token.type == "TOKEN_LPAREN":
            left = self.match_expression(PRECEDENCE_LOWEST)
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
            argument = self.match_expression(PRECEDENCE_LOWEST)
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
        else:
            return Token(self.special.get(c, "TOKEN_UNKNOWN"), c)

    def skip_whitespace(self):
        self.read_while(str.isspace)

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

    def push_back(self, c):
        self.pushed_back = c

    def read_symbol(self):
        return self.read_while(is_symbol_char)

    def read_int(self):
        return self.read_while(str.isdigit)

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


Token = namedtuple("Token", ["type", "value"])


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


def is_symbol_char(c):
    return c.isdigit() or c.isalpha() or c == "_"


def write_with_indent(outfile, contents, *, indent):
    if indent != 0:
        outfile.write("    " * indent)

    outfile.write(contents)


class VeniceError(Exception):
    pass


if __name__ == "__main__":
    main(sys.argv[1:])
