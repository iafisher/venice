#!/usr/bin/env python3
import contextlib
import glob
import os
import subprocess
import sys


def main():
    print("=== Building Venice binaries ===")
    result = subprocess.run(["go", "build"], cwd="src")
    if result.returncode != 0:
        sys.exit(1)
    else:
        print("Build succeeded.")

    print()
    print()
    print("=== Running tests ===")
    total = 0
    failures = 0
    for path in sorted(glob.glob("tests/**/*.vn", recursive=True)):
        total += 1
        print(path)
        result = check_path(path)
        if not result:
            failures += 1
            print("  FAILURE!")
            print()

        # Remove the bytecode file when done.
        with contextlib.suppress(FileNotFoundError):
            os.remove(path + "b")

    if total == 0:
        print("No tests found.")
        sys.exit(2)

    print()
    if failures > 0:
        print(f"Tests FAILED: {failures} failure(s)!")
        sys.exit(3)
    else:
        print("Tests passed.")
        sys.exit(0)


def check_path(path):
    result = subprocess.run(
        ["./src/venice", "execute", path],
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        encoding="utf8",
    )

    expected_output, expect_failure = get_expected_output(path)
    if expect_failure:
        if expected_output and expected_output != result.stderr.rstrip("\n"):
            passed = False
        else:
            passed =  result.returncode != 0
    else:
        if expected_output and expected_output != result.stdout.rstrip("\n"):
            passed = False
        else:
            passed = result.returncode == 0

    if not passed:
        print("--- STDERR ---")
        print(result.stderr, end="")
        print("---  END   ---")
        print()

    return passed


def get_expected_output(path):
    output_builder = []
    expect_failure = False
    in_output = False
    with open(path, "r", encoding="utf8") as f:
        for line in f:
            line = line.strip()
            if line.startswith("#"):
                if line == "# FAIL":
                    expect_failure = True
                elif line == "# OUTPUT":
                    in_output = True
                elif line == "# END OUTPUT":
                    in_output = False
                else:
                    if in_output:
                        line = line[1:].lstrip()
                        output_builder.append(line)
            else:
                break

    if not expect_failure and not output_builder:
        raise Exception(f"could not parse expected output for {path}")

    output = "\n".join(output_builder)
    return output, expect_failure


if __name__ == "__main__":
    main()
