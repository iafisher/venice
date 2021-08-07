#!/usr/bin/env python3
import glob
import subprocess
import sys


def main():
    total = 0
    failures = 0
    for path in glob.glob("tests/**/*.vn", recursive=True):
        total += 1
        print(path)
        result = check_path(path)
        if not result:
            failures += 1
            print("  FAILURE!")

    if total == 0:
        print("No tests found.")
        sys.exit(2)

    print()
    if failures > 0:
        print(f"Tests FAILED: {failures} failure(s)!")
        sys.exit(1)
    else:
        print("Tests passed.")
        sys.exit(0)


def check_path(path, *, expect_failure=False):
    result = subprocess.run(
        ["./src/venice", "execute", path],
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        encoding="utf8",
    )

    expected_output = get_expected_output(path)
    if expect_failure:
        if expected_output and expected_output != result.stderr.rstrip("\n"):
            return False

        return result.returncode != 0
    else:
        if expected_output and expected_output != result.stdout.rstrip("\n"):
            return False

        return result.returncode == 0


def get_expected_output(path):
    output = []
    with open(path, "r", encoding="utf8") as f:
        for line in f:
            line = line.strip()
            if line.startswith("//"):
                line = line[2:].lstrip()
                output.append(line)
            else:
                break

    return "\n".join(output)


if __name__ == "__main__":
    main()
