#!/usr/bin/env python3
import argparse
import glob
import os
import subprocess
import sys
import textwrap
import unittest

from main import get_native_output_path
from main import main as venice_main


def run_e2e_tests() -> None:
    tests_run = 0
    tests_failed = 0

    for path in glob.glob("tests/e2e/*.vn"):
        print(blue(path))
        expected_output = get_expected_output(path)

        tests_run += 1
        try:
            venice_main(path, native=True)
        except SystemExit:
            print("Compilation failed.")
            tests_failed += 1
        else:
            executable = get_native_output_path(path)

            result = subprocess.run(
                [executable], stdout=subprocess.PIPE, encoding="utf8"
            )
            if result.stdout != expected_output:
                print("Expected output:")
                print(textwrap.indent(expected_output, prefix=">  "))

                print("Actual output:")
                print(textwrap.indent(result.stdout, prefix=">  "))

                tests_failed += 1

    print()
    print()
    if tests_failed > 0:
        print(f"{tests_run} run, {tests_failed} {red('failed')}.")
        return False
    else:
        print(f"{tests_run} run, all {green('passed')}.")
        return True


def get_expected_output(path: str) -> str:
    with open(path, "r", encoding="utf8") as f:
        lines = []
        for line in f:
            if not line.startswith("//"):
                break
            else:
                lines.append(line[2:].lstrip())

        return "\n".join(lines)


def blue(s: str) -> str:
    return "\033[94m" + s + "\033[0m"


def green(s: str) -> str:
    return "\033[92m" + s + "\033[0m"


def red(s: str) -> str:
    return "\033[91m" + s + "\033[0m"


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Run Venice's test suite.")
    parser.add_argument("--e2e", action="store_true", default=False)
    args = parser.parse_args()

    test_program = unittest.main(module=None, argv=sys.argv[:1], exit=False)
    unit_tests_result = test_program.result
    unit_tests_passed = (
        len(unit_tests_result.errors) == 0 and len(unit_tests_result.failures) == 0
    )

    if not unit_tests_passed:
        sys.exit(1)

    if args.e2e:
        print()
        print()
        print("== End-to-end tests==")
        e2e_tests_passed = run_e2e_tests()
        if not e2e_tests_passed:
            sys.exit(1)
