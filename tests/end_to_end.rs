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
    test_e2e_expect_runtime_error("06_panic");
}

#[test]
fn test_07_bools() {
    test_e2e("07_bools");
}

#[test]
fn test_08_nested_function_calls() {
    test_e2e("08_nested_function_calls");
}

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
    test_e2e_with_args("13_argv", vec!["a", "b", "c"]);
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

#[test]
fn test_17_register_overflow_2() {
    test_e2e("17_register_overflow_2");
}

#[test]
fn test_18_register_overflow_3() {
    test_e2e("18_register_overflow_3");
}

#[test]
fn test_19_concat() {
    test_e2e("19_concat");
}

#[test]
fn test_20_else_if() {
    test_e2e_simple("20_else_if");
}

/// Runs a full end-to-end test, checking the VIL and x86 code and the output of the program
/// against stored snapshots.
fn test_e2e(base_name: &str) {
    test_e2e_with_options(
        base_name,
        TestOptions {
            args: Vec::new(),
            expect_error: false,
            snapshot_vil: true,
            snapshot_x86: true,
        },
    );
}

/// Runs an end-to-end test only checking the output of the program, not the VIL and x86 snapshots.
fn test_e2e_simple(base_name: &str) {
    test_e2e_with_options(
        base_name,
        TestOptions {
            args: Vec::new(),
            expect_error: false,
            snapshot_vil: false,
            snapshot_x86: false,
        },
    );
}

/// Like `test_e2e`, except that the arguments are passed to the program.
fn test_e2e_with_args(base_name: &str, args: Vec<&'static str>) {
    test_e2e_with_options(
        base_name,
        TestOptions {
            args,
            expect_error: false,
            snapshot_vil: true,
            snapshot_x86: true,
        },
    );
}

/// Like `test_e2e`, except that the program is expected to return an error code.
fn test_e2e_expect_runtime_error(base_name: &str) {
    test_e2e_with_options(
        base_name,
        TestOptions {
            args: Vec::new(),
            expect_error: true,
            snapshot_vil: true,
            snapshot_x86: true,
        },
    );
}

struct TestOptions {
    args: Vec<&'static str>,
    expect_error: bool,
    snapshot_vil: bool,
    snapshot_x86: bool,
}

fn test_e2e_with_options(base_name: &str, options: TestOptions) {
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
        .args(options.args)
        .output()
        .unwrap();

    // Check the output.
    let stdout = str::from_utf8(&output.stdout).unwrap();
    insta::assert_display_snapshot!(format!("{}-stdout", base_name), stdout);
    let stderr = str::from_utf8(&output.stderr).unwrap();
    insta::assert_display_snapshot!(format!("{}-stderr", base_name), stderr);

    if options.expect_error {
        assert!(!output.status.success());
    } else {
        assert!(output.status.success());
    }

    // Check the intermediate files.
    if options.snapshot_vil {
        let vil_output = read_file(&vil_path);
        insta::assert_display_snapshot!(format!("{}-vil", base_name), vil_output);
    }

    if options.snapshot_x86 {
        let x86_output = read_file(&x86_path);
        insta::assert_display_snapshot!(format!("{}-x86", base_name), x86_output);
    }
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
