use std::fs;
use std::fs::File;
use std::io::prelude::*;
use std::io::BufReader;
use std::path::Path;
use std::path::PathBuf;
use std::process::Command;
use std::str;

#[test]
fn test_hello() {
    test_e2e("00_hello");
}

#[test]
fn test_simple_if() {
    test_e2e("00_simple_if");
}

#[test]
fn test_countdown() {
    test_e2e("00_countdown");
}

fn test_e2e(folder: &str) {
    let bin_path = build_path(folder, "program");
    let input_path = build_path(folder, "program.vn");
    let stdout_path = build_path(folder, "stdout.txt");
    let stderr_path = build_path(folder, "stderr.txt");

    // Ensure that the binary file is removed at the end of the test.
    let _cleanup = CleanupFile::new(&bin_path);

    // Read the expected output from the test file.
    let expected_stdout = read_file(&stdout_path);
    let expected_stderr = if Path::new(&stderr_path).exists() {
        read_file(&stderr_path)
    } else {
        String::new()
    };

    // Run the compiler.
    let status = Command::new("target/debug/venice")
        .arg(&input_path)
        .spawn()
        .unwrap()
        .wait()
        .unwrap();
    assert!(status.success());

    // Run the binary itself.
    let output = Command::new(&bin_path).output().unwrap();
    assert!(output.status.success());

    // Check the output.
    let stdout = str::from_utf8(&output.stdout).unwrap();
    assert_eq!(stdout, &expected_stdout);
    let stderr = str::from_utf8(&output.stderr).unwrap();
    assert_eq!(stderr, &expected_stderr);
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

fn build_path(folder: &str, file: &str) -> String {
    let mut buf = PathBuf::new();
    buf.push("tests");
    buf.push(folder);
    buf.push(file);
    buf.into_os_string().into_string().unwrap()
}

fn read_file(path: &str) -> String {
    let f = File::open(path).unwrap();
    let mut buf_reader = BufReader::new(f);
    let mut s = String::new();
    buf_reader.read_to_string(&mut s).unwrap();
    s
}
