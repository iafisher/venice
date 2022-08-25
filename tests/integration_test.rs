use assert_cmd::prelude::*;
use predicates::prelude::*;
use std::fs;
use std::process::Command;

#[test]
fn test_hello() {
    let output_path = "examples/hello";
    let _cleanup = CleanupFile::new(&output_path);

    let mut cmd = Command::cargo_bin("venice").unwrap();
    cmd.arg("examples/hello.vn");
    cmd.assert().success();

    cmd = Command::new(&output_path);
    cmd.assert().stdout(predicate::str::diff("Hello, world!\n"));
}

struct CleanupFile(String);

impl CleanupFile {
    fn new(path: &str) -> Self {
        CleanupFile(String::from(path))
    }
}

impl Drop for CleanupFile {
    fn drop(&mut self) {
        let _ = fs::remove_file(&self.0);
    }
}
