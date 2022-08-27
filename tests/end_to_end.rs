use std::fs;
use std::fs::File;
use std::io::prelude::*;
use std::io::BufReader;
use std::path::PathBuf;
use std::process::{Command, Stdio};
use std::str;

extern crate insta;

#[test]
fn test_hello() {
    test_e2e("00_hello");
}

#[test]
fn test_simple_if() {
    test_e2e("01_simple_if");
}

#[test]
fn test_countdown() {
    test_e2e("02_countdown");
}

#[test]
fn test_simple_function() {
    test_e2e("03_simple_function");
}

#[test]
fn test_simple_function_with_args() {
    test_e2e("04_simple_function_with_args");
}

#[test]
fn test_fibonacci() {
    test_e2e("10_fibonacci");
}

#[test]
fn test_fibonacci_recursive() {
    test_e2e("11_fibonacci_recursive");
}

fn test_e2e(folder: &str) {
    let bin_path = build_path(folder, "program");
    let obj_path = build_path(folder, "program.o");
    let vil_path = build_path(folder, "program.vil");
    let x86_path = build_path(folder, "program.x86.s");
    let input_path = build_path(folder, "program.vn");

    // Ensure that intermediate files are removed at the end of the test.
    let _cleanup = CleanupFile(vec![
        bin_path.clone(),
        obj_path.clone(),
        vil_path.clone(),
        x86_path.clone(),
    ]);

    // Run the compiler.
    let status = Command::new("target/debug/venice")
        .arg(&input_path)
        .arg("--debug")
        .stdout(Stdio::null())
        .spawn()
        .unwrap()
        .wait()
        .unwrap();
    assert!(status.success());

    // Check the intermediate files.
    let vil_output = read_file(&vil_path);
    insta::assert_display_snapshot!(format!("{}-vil", folder), vil_output);
    let x86_output = read_file(&x86_path);
    insta::assert_display_snapshot!(format!("{}-x86", folder), x86_output);

    // Run the binary itself, under the `timeout` utility so it doesn't run forever.
    let output = Command::new("timeout")
        .arg("5s")
        .arg(&bin_path)
        .output()
        .unwrap();
    assert!(output.status.success());

    // Check the output.
    let stdout = str::from_utf8(&output.stdout).unwrap();
    insta::assert_display_snapshot!(format!("{}-stdout", folder), stdout);
    let stderr = str::from_utf8(&output.stderr).unwrap();
    insta::assert_display_snapshot!(format!("{}-stderr", folder), stderr);
}

struct CleanupFile(Vec<String>);

impl Drop for CleanupFile {
    fn drop(&mut self) {
        for path in &self.0 {
            let _ = fs::remove_file(path);
        }
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
