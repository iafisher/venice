import argparse
import os

from venice import codegen, parser


def main(path: str) -> None:
    with open(path, "r", encoding="utf8") as f:
        program = f.read()

    ptree = parser.parse(program)
    code = codegen.codegen(ptree)

    output_path = get_output_path(path)
    with open(output_path, "w", encoding="utf8") as f:
        f.write(code)


def get_output_path(input_path: str) -> str:
    return os.path.splitext(input_path)[0] + ".c"


if __name__ == "__main__":
    argparser = argparse.ArgumentParser()
    argparser.add_argument("path")
    args = argparser.parse_args()

    main(args.path)
