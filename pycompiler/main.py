import argparse
import sys
from io import StringIO

from pycompiler import ast
from pycompiler.analyzer import vcheck
from pycompiler.common import VeniceError
from pycompiler.generator_javascript import vgenerate_javascript
from pycompiler.generator_python import vgenerate_python
from pycompiler.parser import Lexer, vparse


def main():
    parser = argparse.ArgumentParser(description="The Venice programming language.")
    subparsers = parser.add_subparsers()

    parser_run = subparsers.add_parser("run")
    parser_run.add_argument("--quiet", action="store_true")
    parser_run.add_argument("--js", action="store_true")
    parser_run.add_argument("path")
    parser_run.set_defaults(func=main_run)

    parser_parse = subparsers.add_parser("parse")
    parser_parse.add_argument("path")
    parser_parse.set_defaults(func=main_parse)

    parser_tokenize = subparsers.add_parser("tokenize")
    parser_tokenize.add_argument("path")
    parser_tokenize.set_defaults(func=main_tokenize)

    parser_compile = subparsers.add_parser("compile")
    parser_compile.add_argument("path")
    parser_compile.add_argument("--js", action="store_true")
    parser_compile.set_defaults(func=main_compile)

    args = parser.parse_args()
    args.func(args)


def main_run(args):
    outfile = StringIO()
    with open(args.path, "r", encoding="utf8") as infile:
        try:
            vcompile(infile, outfile, js=args.js)
        except VeniceError as e:
            if args.quiet:
                print(f"ERROR: {e}", file=sys.stderr)
                sys.exit(1)
            else:
                raise e

    program = outfile.getvalue()

    if not args.js:
        env = {}
        exec(program, env, env)
    else:
        raise NotImplementedError


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


def main_compile(args):
    with open(args.path, "r", encoding="utf8") as infile:
        vcompile(infile, sys.stdout, js=args.js)


def vcompile(infile, outfile, *, js=False):
    ast = vparse(infile)
    vcheck(ast)

    if js:
        vgenerate_javascript(outfile, ast)
    else:
        vgenerate_python(outfile, ast)


def vcompile_string(program, *, js=False):
    infile = StringIO(program)
    outfile = StringIO()
    vcompile(infile, outfile, js=js)
    return outfile.getvalue()


def pretty_print_tree(tree, indent=0):
    print(("  " * indent) + tree.__class__.__name__)
    for key, value in tree._asdict().items():
        print(("  " * (indent + 1)) + key + ":", end="")
        if isinstance(value, ast.Symbol):
            print(f" ast.Symbol({value.label!r})")
        elif isinstance(value, ast.Literal):
            print(f" ast.Literal({value.value!r})")
        else:
            print()
            if not isinstance(value, list):
                value = [value]

            for subvalue in value:
                # TODO(2021-03-24): Clean up this hacky logic.
                if getattr(subvalue.__class__, "__module__", "").endswith("ast"):
                    pretty_print_tree(subvalue, indent + 2)
                else:
                    print(("  " * (indent + 2)) + repr(subvalue))


if __name__ == "__main__":
    main()
