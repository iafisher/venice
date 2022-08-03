import argparse
import contextlib
import os
import subprocess
import sys

from venice.analysis import analyze
from venice.backend import codegen
from venice.common import VeniceSyntaxError, VeniceTypeError
from venice.frontend import parser


def main(path: str, *, native: bool = False) -> None:
    with open(path, "r", encoding="utf8") as f:
        program = f.read()

    try:
        tree = parser.parse_module(program)
    except VeniceSyntaxError as e:
        print(f"Syntax error: {e}", file=sys.stderr)
        sys.exit(2)

    analyze(tree)

    try:
        code = codegen.codegen(tree)
    except VeniceTypeError as e:
        print(f"Type error: {e}", file=sys.stderr)
        sys.exit(3)

    output_path = get_output_path(path)
    with open(output_path, "w", encoding="utf8") as f:
        f.write(code)

    if native:
        native_output_path = get_native_output_path(path)
        r = subprocess.run(
            [
                "gcc",
                "-Wall",
                "-Werror",
                "-Iruntime",
                "-o",
                native_output_path,
                output_path,
                "runtime/libvenice.so",
            ]
        )

        with contextlib.suppress(FileNotFoundError):
            os.remove(output_path)

        if r.returncode != 0:
            sys.exit(r.returncode)


def get_output_path(input_path: str) -> str:
    return os.path.splitext(input_path)[0] + ".c"


def get_native_output_path(input_path: str) -> str:
    return os.path.splitext(input_path)[0]


if __name__ == "__main__":
    argparser = argparse.ArgumentParser()
    argparser.add_argument("path")
    argparser.add_argument("--native", action="store_true", default=False)
    args = argparser.parse_args()

    main(args.path, native=args.native)
