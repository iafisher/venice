// Copyright 2022 The Venice Authors. All rights reserved.
// Use of this source code is governed by an MIT-style license that can be
// found in the LICENSE file.
//
// End-to-end tests for the Venice compiler.
//
// Each test executes the compiler on a Venice program and checks the output (both stdout and
// stderr) and intermediate representations (VIL and x86 assembly) against a snapshot. Snapshotting
// is handled by the `insta` crate.

use std::fs;
use std::fs::File;
use std::io::prelude::*;
use std::io::BufReader;
use std::path::PathBuf;
use std::process::{Command, Stdio};
use std::str;

extern crate insta;

#[test]
fn test_00_hello() {
    test_e2e("00_hello");
}

#[test]
fn test_01_simple_if() {
    test_e2e("01_simple_if");
}

#[test]
fn test_02_countdown() {
    test_e2e("02_countdown");
}

#[test]
fn test_03_simple_function() {
    test_e2e("03_simple_function");
}

#[test]
fn test_04_simple_function_with_args() {
    test_e2e("04_simple_function_with_args");
}

#[test]
fn test_05_multiply_divide() {
    test_e2e("05_multiply_divide");
}

#[test]
fn test_06_panic() {
    test_e2e_expect_error("06_panic");
}

#[test]
fn test_07_bools() {
    test_e2e("07_bools");
}

#[test]
fn test_08_nested_function_calls() {
    test_e2e("08_nested_function_calls");
}

#[ignore] // TODO(#148): Re-enable this test once compiler can handle register allocation.
#[test]
fn test_09_more_nested_function_calls() {
    test_e2e("09_more_nested_function_calls");
}

#[test]
fn test_10_fibonacci() {
    test_e2e("10_fibonacci");
}

#[test]
fn test_11_fibonacci_recursive() {
    test_e2e("11_fibonacci_recursive");
}

#[test]
fn test_12_list_literal() {
    test_e2e("12_list_literal");
}

#[test]
fn test_13_argv() {
    test_e2e_with_args("13_argv", &["a", "b", "c"]);
}

#[test]
fn test_14_file_io() {
    test_e2e("14_file_io");
}

#[test]
fn test_15_register_overflow() {
    test_e2e("15_register_overflow");
}

#[test]
fn test_16_tricky_function_calls() {
    test_e2e("16_tricky_function_calls");
}

fn test_e2e(base_name: &str) {
    test_e2e_full_options(base_name, &[], /* expect_error= */ false);
}

fn test_e2e_with_args(base_name: &str, args: &[&str]) {
    test_e2e_full_options(base_name, args, /* expect_error= */ false);
}

fn test_e2e_expect_error(base_name: &str) {
    test_e2e_full_options(base_name, &[], /* expect_error= */ true);
}

fn test_e2e_full_options(base_name: &str, args: &[&str], expect_error: bool) {
    let bin_path = build_path(base_name, "");
    let obj_path = build_path(base_name, "o");
    let vil_path = build_path(base_name, "vil");
    let x86_path = build_path(base_name, "x86.s");
    let input_path = build_path(base_name, "vn");

    // Ensure that intermediate files are removed at the end of the test.
    let _cleanup = CleanupFile(vec![
        bin_path.clone(),
        obj_path,
        vil_path.clone(),
        x86_path.clone(),
    ]);

    // Run the compiler.
    let status = Command::new("target/debug/venice")
        .arg(&input_path)
        .arg("--debug")
        .arg("--keep-intermediate")
        .stdout(Stdio::null())
        .spawn()
        .unwrap()
        .wait()
        .unwrap();
    assert!(status.success());

    // Run the binary itself, under the `timeout` utility so it doesn't run forever.
    let output = Command::new("timeout")
        .arg("5s")
        .arg(&bin_path)
        .args(args)
        .output()
        .unwrap();

    if expect_error {
        assert!(!output.status.success());
    } else {
        assert!(output.status.success());
    }

    // Check the output.
    let stdout = str::from_utf8(&output.stdout).unwrap();
    insta::assert_display_snapshot!(format!("{}-stdout", base_name), stdout);
    let stderr = str::from_utf8(&output.stderr).unwrap();
    insta::assert_display_snapshot!(format!("{}-stderr", base_name), stderr);

    // Check the intermediate files.
    let vil_output = read_file(&vil_path);
    insta::assert_display_snapshot!(format!("{}-vil", base_name), vil_output);
    let x86_output = read_file(&x86_path);
    insta::assert_display_snapshot!(format!("{}-x86", base_name), x86_output);
}

struct CleanupFile(Vec<String>);

impl Drop for CleanupFile {
    fn drop(&mut self) {
        for path in &self.0 {
            let _ = fs::remove_file(path);
        }
    }
}

fn build_path(base_name: &str, ext: &str) -> String {
    let mut buf = PathBuf::new();
    buf.push("tests");
    buf.push(base_name);
    buf.set_extension(ext);
    buf.into_os_string().into_string().unwrap()
}

fn read_file(path: &str) -> String {
    let f = File::open(path).unwrap();
    let mut buf_reader = BufReader::new(f);
    let mut s = String::new();
    buf_reader.read_to_string(&mut s).unwrap();
    s
}
