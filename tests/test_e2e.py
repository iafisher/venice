import os
import subprocess
from pathlib import Path

import pytest


@pytest.mark.parametrize("case_dir", list(Path("tests/snapshots").iterdir()))
def test_snapshots(case_dir, snapshot):
    compile_result = subprocess.run(["./v", "--native", case_dir / "input.vn"])
    assert compile_result.returncode == 0

    exe = case_dir / "input"
    exec_result = subprocess.run([exe], capture_output=True, encoding="utf8")
    os.remove(exe)

    snapshot.snapshot_dir = case_dir
    snapshot.assert_match(exec_result.stdout, "stdout.txt")
    snapshot.assert_match(exec_result.stderr, "stderr.txt")
