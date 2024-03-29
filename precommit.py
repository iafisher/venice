"""
Pre-commit configuration for git.

This file was created by precommit (https://github.com/iafisher/precommit).
You are welcome to edit it yourself to customize your pre-commit hook.
"""
from iprecommit import checks


def init(precommit):
    precommit.check(checks.NoStagedAndUnstagedChanges())
    precommit.check(checks.NoWhitespaceInFilePath())
    precommit.check(checks.DoNotSubmit())

    pyinclude = ["pycompiler/test"]

    # Check Python format with black.
    precommit.check(checks.PythonFormat(include=pyinclude))

    # Lint Python code with flake8.
    precommit.check(checks.PythonLint(include=pyinclude))

    # Check the order of Python imports with isort.
    precommit.check(checks.PythonImportOrder(include=pyinclude))

    # Check that requirements.txt matches pip freeze.
    # precommit.check(checks.PipFreeze(venv=".venv"))

    # Check Python static type annotations with mypy.
    # precommit.check(checks.PythonTypes())

    precommit.check(checks.ClangFormat())

    # Check Rust format with rustfmt.
    precommit.check(checks.RustFormat())
    precommit.check(
        checks.Command(
            "RustClippy",
            ["cargo", "clippy"],
            fix=["cargo", "clippy", "--fix", "--allow-staged", "--allow-dirty"],
            include=["*.rs"],
        )
    )

    precommit.check(checks.Command("UnitTests", ["cargo", "test"], include=["*.rs"]))
    precommit.check(
        checks.Command(
            "RuntimeTests",
            ["./tools/test_runtime"],
            include=["runtime/*.h", "runtime/*.c"],
        )
    )
