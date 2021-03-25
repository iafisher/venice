"""
Pre-commit configuration for git.

This file was created by precommit (https://github.com/iafisher/precommit).
You are welcome to edit it yourself to customize your pre-commit hook.
"""
from precommitlib import checks


def init(precommit):
    precommit.check(checks.NoStagedAndUnstagedChanges())
    precommit.check(checks.NoWhitespaceInFilePath())
    precommit.check(checks.DoNotSubmit())

    pyinclude = ["test"]

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

    # Lint JavaScript code with ESLint.
    precommit.check(checks.JavaScriptLint())

    # Check Rust format with rustfmt.
    precommit.check(checks.RustFormat())

    # Run a custom command.
    # precommit.check(checks.Command("UnitTests", ["./test"]))

    # Run a custom command on each file.
    # precommit.check(checks.Command("FileCheck", ["check_file"], pass_files=True))

    precommit.check(checks.Command("UnitTests", ["./test"]))
