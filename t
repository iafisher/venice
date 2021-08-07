#!/usr/bin/env python3
import glob
import subprocess
import sys


def main():
    print("=== BUILDING VENICE ===")
    subprocess.run(["go", "build"], cwd="src")


    print()
    print()
    print("=== RUNNING TESTS ===")
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
            return False

        return result.returncode != 0
    else:
        if expected_output and expected_output != result.stdout.rstrip("\n"):
            return False

        return result.returncode == 0


def get_expected_output(path):
    output_builder = []
    with open(path, "r", encoding="utf8") as f:
        for line in f:
            line = line.strip()
            if line.startswith("#"):
                line = line[1:].lstrip()
                output_builder.append(line)
            else:
                break

    output = "\n".join(output_builder)
    if output.startswith("FAIL"):
        if output.startswith("FAIL:"):
            return output[5:].lstrip(), True
        else:
            return "", True
    else:
        return output, False


if __name__ == "__main__":
    main()
