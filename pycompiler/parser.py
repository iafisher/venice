from attr import attrib, attrs

from pycompiler import ast
from pycompiler.common import VeniceError


def vparse(infile, *, debug=False):
    return Parser(Lexer(infile), debug=debug).parse()


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
    "TOKEN_LTE": PRECEDENCE_CMP,
    "TOKEN_GTE": PRECEDENCE_CMP,
    "TOKEN_LANGLE": PRECEDENCE_CMP,
    "TOKEN_RANGLE": PRECEDENCE_CMP,
    "TOKEN_EQ": PRECEDENCE_CMP,
    "TOKEN_PLUS": PRECEDENCE_ADD_SUB,
    "TOKEN_MINUS": PRECEDENCE_ADD_SUB,
    "TOKEN_ASTERISK": PRECEDENCE_MUL_DIV,
    "TOKEN_SLASH": PRECEDENCE_MUL_DIV,
    # The left parenthesis is the "infix operator" for function-call expressions.
    "TOKEN_LPAREN": PRECEDENCE_CALL,
    "TOKEN_LSQUARE": PRECEDENCE_CALL,
    "TOKEN_PERIOD": PRECEDENCE_CALL,
}


def debuggable(f):
    def wrapped(self, *args, **kwargs):
        name = f.__name__
        if self.debug:
            indent = "  " * (self.debug_indent * 2)
            self.debug_indent += 1
            print(f"{indent}{name}")

        r = f(self, *args, **kwargs)

        if self.debug:
            print(f"{indent}{name} returned (value = {r!r})")
            self.debug_indent -= 1

        return r

    return wrapped


class Parser:
    def __init__(self, lexer, *, debug=False):
        self.lexer = lexer
        self.pushed_back = None
        self.debug = debug
        self.debug_indent = 0

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

        return ast.ProgramNode(statements)

    @debuggable
    def match_function(self):
        symbol_token = self.expect("TOKEN_SYMBOL")
        self.expect("TOKEN_LPAREN")
        parameters = self.match_comma_separated(self.match_parameter, "TOKEN_RPAREN")
        self.expect("TOKEN_RPAREN")
        self.expect("TOKEN_COLON")
        return_type = self.match_type()
        statements = self.match_block()
        return ast.FunctionNode(
            label=symbol_token.value,
            parameters=parameters,
            return_type=return_type,
            statements=statements,
        )

    @debuggable
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

    @debuggable
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
        elif token.type == "TOKEN_ENUM":
            return self.match_enum_declaration()
        elif token.type == "TOKEN_MATCH":
            return self.match_match()
        else:
            self.push_back(token)
            value = self.match_expression()
            self.expect(("TOKEN_NEWLINE", "TOKEN_EOF"))
            if isinstance(value, ast.AssignNode):
                return value
            else:
                return ast.ExpressionStatementNode(value)

    @debuggable
    def match_let(self):
        symbol_token = self.expect("TOKEN_SYMBOL")
        self.expect("TOKEN_ASSIGN")
        value = self.match_expression()
        self.expect("TOKEN_NEWLINE")
        return ast.LetNode(label=symbol_token.value, value=value)

    @debuggable
    def match_return(self):
        value = self.match_expression()
        self.expect("TOKEN_NEWLINE")
        return ast.ReturnNode(value)

    @debuggable
    def match_if(self):
        clauses = []
        condition = self.match_expression()
        statements = self.match_block()
        clauses.append(ast.IfClauseNode(condition, statements))
        else_clause = None
        while True:
            token = self.expect(
                ["TOKEN_ELIF", "TOKEN_ELSE", "TOKEN_NEWLINE", "TOKEN_EOF"]
            )
            if token.type == "TOKEN_ELIF":
                condition = self.match_expression()
                statements = self.match_block()
                clauses.append(ast.IfClauseNode(condition, statements))
            elif token.type == "TOKEN_ELSE":
                else_clause = self.match_block()
            else:
                break

        return ast.IfNode(if_clauses=clauses, else_clause=else_clause)

    @debuggable
    def match_while(self):
        condition = self.match_expression()
        statements = self.match_block()
        return ast.WhileNode(condition, statements)

    @debuggable
    def match_for(self):
        symbols_list = []
        token = self.expect("TOKEN_SYMBOL")
        symbols_list.append(token.value)
        while True:
            token = self.next()
            if token.type == "TOKEN_COMMA":
                token = self.expect("TOKEN_SYMBOL")
                symbols_list.append(token.value)
            else:
                self.push_back(token)
                break

        self.expect("TOKEN_IN")
        iterator = self.match_expression()
        statements = self.match_block()
        return ast.ForNode(
            loop_variables=symbols_list, iterator=iterator, statements=statements
        )

    @debuggable
    def match_struct_declaration(self):
        symbol_token = self.expect("TOKEN_SYMBOL")
        self.expect("TOKEN_LCURLY")
        fields = []
        while True:
            field_token = self.expect("TOKEN_SYMBOL")
            self.expect("TOKEN_COLON")
            type_tree = self.match_type()
            fields.append(ast.StructDeclarationFieldNode(field_token.value, type_tree))
            token = self.expect(["TOKEN_COMMA", "TOKEN_RCURLY"])
            if token.type == "TOKEN_COMMA":
                continue
            else:
                break

        self.expect("TOKEN_NEWLINE")
        return ast.StructDeclarationNode(symbol_token.value, fields)

    @debuggable
    def match_enum_declaration(self):
        symbol = self.expect("TOKEN_SYMBOL").value
        self.expect("TOKEN_LCURLY")
        self.accept("TOKEN_NEWLINE")
        cases = []
        while True:
            symbol_token = self.expect(("TOKEN_SYMBOL", "TOKEN_RCURLY"))
            if symbol_token.type == "TOKEN_RCURLY":
                break

            token = self.next()
            parameters = []
            if token.type == "TOKEN_LPAREN":
                parameters = self.match_comma_separated(
                    self.match_parameter, "TOKEN_RPAREN"
                )
                self.expect("TOKEN_RPAREN")
            else:
                self.push_back(token)

            cases.append(
                ast.EnumDeclarationCaseNode(
                    label=symbol_token.value, parameters=parameters
                )
            )

            saw_comma = self.accept("TOKEN_COMMA")
            if not saw_comma:
                self.accept("TOKEN_NEWLINE")
                self.expect("TOKEN_RCURLY")
                break
            else:
                self.accept("TOKEN_NEWLINE")

        self.expect("TOKEN_NEWLINE")
        return ast.EnumDeclarationNode(label=symbol, cases=cases)

    @debuggable
    def match_match(self):
        e = self.match_expression()
        self.expect("TOKEN_LCURLY")
        self.expect("TOKEN_NEWLINE")
        cases = []
        while True:
            token = self.expect(("TOKEN_CASE", "TOKEN_RCURLY"))
            if token.type == "TOKEN_RCURLY":
                break

            pattern = self.match_expression()
            statements = self.match_block()
            cases.append(ast.MatchCaseNode(pattern=pattern, statements=statements))
            saw_comma = self.accept("TOKEN_COMMA")
            self.expect("TOKEN_NEWLINE")
            if not saw_comma:
                self.expect("TOKEN_RCURLY")
                break

        self.expect("TOKEN_NEWLINE")
        return ast.MatchNode(value=e, cases=cases)

    @debuggable
    def match_expression(self, precedence=PRECEDENCE_LOWEST):
        left = self.match_prefix()

        token = self.next()
        while token.type in PRECEDENCE_MAP and precedence < PRECEDENCE_MAP[token.type]:
            left = self.match_infix(left, token, PRECEDENCE_MAP[token.type])
            token = self.next()

        self.push_back(token)
        return left

    @debuggable
    def match_prefix(self):
        token = self.next()
        if token.type == "TOKEN_INT":
            left = ast.LiteralNode(int(token.value))
        elif token.type == "TOKEN_TRUE":
            left = ast.LiteralNode(True)
        elif token.type == "TOKEN_FALSE":
            left = ast.LiteralNode(False)
        elif token.type == "TOKEN_SYMBOL":
            left = ast.SymbolNode(token.value)
        elif token.type == "TOKEN_STRING":
            left = ast.LiteralNode(token.value)
        elif token.type == "TOKEN_LPAREN":
            left = self.match_expression()
            self.expect("TOKEN_RPAREN")
        elif token.type == "TOKEN_MINUS":
            left = ast.PrefixNode("-", self.match_expression(PRECEDENCE_PREFIX))
        elif token.type == "TOKEN_NOT":
            left = ast.PrefixNode("not", self.match_expression(PRECEDENCE_PREFIX))
        elif token.type == "TOKEN_LSQUARE":
            values = self.match_comma_separated(self.match_expression, "TOKEN_RSQUARE")
            self.expect("TOKEN_RSQUARE")
            return ast.ListNode(values)
        elif token.type == "TOKEN_LCURLY":
            key_value_pairs = self.match_comma_separated(
                self.match_key_value_pair, "TOKEN_RCURLY"
            )
            self.expect("TOKEN_RCURLY")
            return ast.MapNode(key_value_pairs)
        else:
            if token.type == "TOKEN_EOF":
                raise VeniceError("premature end of input")
            else:
                raise VeniceError(f"unexpected token {token!r}")

        return left

    @debuggable
    def match_infix(self, left, token, precedence):
        if token.type == "TOKEN_LPAREN":
            args = self.match_comma_separated(self.match_argument, "TOKEN_RPAREN")
            self.expect("TOKEN_RPAREN")
            return ast.CallNode(left, args)
        elif token.type == "TOKEN_LSQUARE":
            index = self.match_expression()
            self.expect("TOKEN_RSQUARE")
            return ast.IndexNode(left, index)
        elif token.type == "TOKEN_ASSIGN":
            right = self.match_expression(precedence)
            return ast.AssignNode(left, right)
        elif token.type == "TOKEN_PERIOD":
            symbol_token = self.expect("TOKEN_SYMBOL")
            return ast.FieldAccessNode(left, symbol_token)
        else:
            right = self.match_expression(precedence)
            return ast.InfixNode(token.value, left, right)

    @debuggable
    def match_argument(self):
        argument = self.match_expression()
        if isinstance(argument, ast.SymbolNode):
            token = self.next()
            if token.type == "TOKEN_COLON":
                label = argument.label
                argument = self.match_expression()
                return ast.KeywordArgumentNode(label=label, value=argument)
            else:
                self.push_back(token)
                return argument
        else:
            return argument

    @debuggable
    def match_parameter(self):
        symbol_token = self.expect("TOKEN_SYMBOL")
        self.expect("TOKEN_COLON")
        symbol_type = self.match_type()
        return ast.ParameterNode(label=symbol_token.value, type_label=symbol_type)

    @debuggable
    def match_key_value_pair(self):
        key = self.match_expression()
        self.expect("TOKEN_COLON")
        value = self.match_expression()
        return ast.MapLiteralPairNode(key, value)

    @debuggable
    def match_type(self):
        symbol_token = self.expect("TOKEN_SYMBOL")
        token = self.next()
        if token.type == "TOKEN_LANGLE":
            inner_types = self.match_comma_separated(self.match_type, "TOKEN_RANGLE")
            self.expect("TOKEN_RANGLE")
            return ast.ParameterizedTypeNode(symbol_token, inner_types)
        else:
            self.push_back(token)
            return ast.SymbolNode(symbol_token.value)

    @debuggable
    def match_comma_separated(self, matcher, terminator):
        values = []
        while True:
            if terminator is not None:
                token = self.next()
                self.push_back(token)
                if token.type == terminator:
                    break

            values.append(matcher())

            token = self.next()
            if token.type != "TOKEN_COMMA":
                self.push_back(token)
                break
        return values

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
            suffix = " (pushed back)"
            self.pushed_back = None
        else:
            token = self.lexer.next()
            suffix = ""

        if self.debug:
            indent = "  " * (self.debug_indent * 2)
            print(f"{indent}{token!r}{suffix}")

        return token

    def push_back(self, token):
        self.pushed_back = token

    def accept(self, type_or_types):
        if isinstance(type_or_types, str):
            type_or_types = {type_or_types}
        else:
            type_or_types = set(type_or_types)

        token = self.next()
        if token.type not in type_or_types:
            self.push_back(token)
            return False
        else:
            return True

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
            "case",
            "elif",
            "else",
            "enum",
            "false",
            "fn",
            "for",
            "if",
            "in",
            "let",
            "match",
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
        ".": "TOKEN_PERIOD",
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
        self.pushed_back = []
        self.line = 1
        self.column = 1

    def next(self):
        self.skip_whitespace_and_comments()

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
                return Token("TOKEN_GTE", ">=")
            else:
                self.push_back(c2)
                return Token("TOKEN_RANGLE", ">")
        elif c == "<":
            c2 = self.read()
            if c2 == "=":
                return Token("TOKEN_LTE", "<=")
            else:
                self.push_back(c2)
                return Token("TOKEN_LANGLE", "<")
        elif c == "=":
            c2 = self.read()
            if c2 == "=":
                return Token("TOKEN_EQ", "==")
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
        self.pushed_back.append(c)

    def skip_whitespace_and_comments(self):
        while True:
            c = self.read()
            if c.isspace() and c != "\n":
                self.read_while(lambda c: c.isspace() and c != "\n")
            elif c == "/":
                c2 = self.read()
                if c2 == "/":
                    self.read_while(lambda c: c != "\n")
                    continue
                else:
                    self.push_back("/")
                    self.push_back("/")
                    break
            else:
                self.push_back(c)
                break

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
            return self.pushed_back.pop()
        else:
            c = self.infile.read(1)
            if c == "":
                self.done = True
            return c


@attrs
class Token:
    type = attrib()
    value = attrib()


def is_symbol_char(c):
    return c.isdigit() or c.isalpha() or c == "_"
