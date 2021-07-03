import argparse
import subprocess
import sys
import tempfile
from io import StringIO

import attr

from pycompiler import ast
from pycompiler.analyzer import vcheck
from pycompiler.common import VeniceError
from pycompiler.generator_javascript import vgenerate_javascript
from pycompiler.generator_python import vgenerate_python
from pycompiler.parser import Lexer, vparse


def main():
    parser = argparse.ArgumentParser(description="The Venice programming language.")
    subparsers = parser.add_subparsers()

    parser_compile = subparsers.add_parser("compile")
    parser_compile.add_argument("path")
    parser_compile.add_argument("--python", action="store_true")
    parser_compile.set_defaults(func=main_compile)

    parser_repl = subparsers.add_parser("repl")
    parser_repl.set_defaults(func=main_repl)

    parser_run = subparsers.add_parser("run")
    parser_run.add_argument("--quiet", action="store_true")
    parser_run.add_argument("--python", action="store_true")
    parser_run.add_argument("path")
    parser_run.set_defaults(func=main_run)

    parser_parse = subparsers.add_parser("parse")
    parser_parse.add_argument("path")
    parser_parse.set_defaults(func=main_parse)

    parser_tokenize = subparsers.add_parser("tokenize")
    parser_tokenize.add_argument("path")
    parser_tokenize.set_defaults(func=main_tokenize)

    args = parser.parse_args()
    args.func(args)


def main_compile(args):
    with open(args.path, "r", encoding="utf8") as infile:
        vcompile(infile, sys.stdout, python=args.python)


def main_repl(args):
    while True:
        try:
            line = input(">>> ")
        except (EOFError, KeyboardInterrupt):
            print()
            break

        infile = StringIO(line)
        outfile = StringIO()

        vcompile(infile, outfile, python=True)

        globals_map = {}
        locals_map = {}
        result = eval(outfile.getvalue(), globals_map, locals_map)
        print(repr(result))


def main_run(args):
    with tempfile.NamedTemporaryFile("w", encoding="utf8") as outfile:
        with open(args.path, "r", encoding="utf8") as infile:
            try:
                vcompile(infile, outfile, python=args.python)
            except VeniceError as e:
                if args.quiet:
                    print(f"ERROR: {e}", file=sys.stderr)
                    sys.exit(1)
                else:
                    raise e

        outfile.flush()
        if args.python:
            subprocess.run(["python3", outfile.name])
        else:
            subprocess.run(["node", outfile.name])


def main_parse(args):
    with open(args.path, "r", encoding="utf8") as infile:
        ast = vparse(infile)

    pretty_print_tree(ast)


def main_tokenize(args):
    with open(args.path, "r", encoding="utf8") as infile:
        lexer = Lexer(infile)
        while True:
            token = lexer.next()
            if token.type == "TOKEN_EOF":
                break
            else:
                if token.type in ("TOKEN_STRING", "TOKEN_NEWLINE", "TOKEN_UNKNOWN",):
                    print(token.type.ljust(20), repr(token.value))
                else:
                    print(token.type.ljust(20), token.value)


def vcompile(infile, outfile, *, python=False):
    ast = vparse(infile)
    vcheck(ast)

    if python:
        vgenerate_python(outfile, ast)
    else:
        vgenerate_javascript(outfile, ast)


def vcompile_string(program, *, python=False):
    infile = StringIO(program)
    outfile = StringIO()
    vcompile(infile, outfile, python=python)
    return outfile.getvalue()


def pretty_print_tree(tree, indent=0):
    print(("  " * indent) + tree.__class__.__name__)
    for key, value in attr.asdict(tree, recurse=False).items():
        print(("  " * (indent + 1)) + key + ":", end="")
        if isinstance(value, ast.SymbolNode):
            print(f" ast.SymbolNode({value.label!r})")
        elif isinstance(value, ast.LiteralNode):
            print(f" ast.LiteralNode({value.value!r})")
        else:
            print()
            if not isinstance(value, list):
                value = [value]

            for subvalue in value:
                if isinstance(subvalue, ast.AbstractNode):
                    pretty_print_tree(subvalue, indent + 2)
                else:
                    print(("  " * (indent + 2)) + repr(subvalue))
