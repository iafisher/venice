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
use std::process::Command;
use std::str;

extern crate insta;

#[test]
fn test_00_hello() {
    test_e2e("00_hello", TestOptions::full());
}

#[test]
fn test_01_simple_if() {
    test_e2e("01_simple_if", TestOptions::full());
}

#[test]
fn test_02_countdown() {
    test_e2e("02_countdown", TestOptions::full());
}

#[test]
fn test_03_simple_function() {
    test_e2e("03_simple_function", TestOptions::full());
}

#[test]
fn test_04_simple_function_with_args() {
    test_e2e("04_simple_function_with_args", TestOptions::full());
}

#[test]
fn test_05_multiply_divide() {
    test_e2e("05_multiply_divide", TestOptions::full());
}

#[test]
fn test_06_panic() {
    test_e2e("06_panic", TestOptions::runtime_error());
}

#[test]
fn test_07_bools() {
    test_e2e("07_bools", TestOptions::full());
}

#[test]
fn test_08_nested_function_calls() {
    test_e2e("08_nested_function_calls", TestOptions::full());
}

#[test]
fn test_09_more_nested_function_calls() {
    test_e2e("09_more_nested_function_calls", TestOptions::full());
}

#[test]
fn test_10_fibonacci() {
    test_e2e("10_fibonacci", TestOptions::full());
}

#[test]
fn test_11_fibonacci_recursive() {
    test_e2e("11_fibonacci_recursive", TestOptions::full());
}

#[test]
fn test_12_list_literal() {
    test_e2e("12_list_literal", TestOptions::full());
}

#[test]
fn test_13_argv() {
    test_e2e("13_argv", TestOptions::full_with_args(vec!["a", "b", "c"]));
}

#[test]
fn test_14_file_io() {
    test_e2e("14_file_io", TestOptions::full());
}

#[test]
fn test_15_register_overflow() {
    test_e2e("15_register_overflow", TestOptions::full());
}

#[test]
fn test_16_tricky_function_calls() {
    test_e2e("16_tricky_function_calls", TestOptions::full());
}

#[test]
fn test_17_register_overflow_2() {
    test_e2e("17_register_overflow_2", TestOptions::full());
}

#[test]
fn test_18_register_overflow_3() {
    test_e2e("18_register_overflow_3", TestOptions::full());
}

#[test]
fn test_19_concat() {
    test_e2e("19_concat", TestOptions::full());
}

#[test]
fn test_20_else_if() {
    test_e2e("20_else_if", TestOptions::simple());
}

#[test]
fn test_error_00_bad_addition() {
    test_e2e("error_00_bad_addition", TestOptions::compile_error());
}

#[test]
fn test_error_01_bad_printint() {
    test_e2e("error_01_bad_printint", TestOptions::compile_error());
}

#[test]
fn test_error_02_too_many_arguments() {
    test_e2e("error_02_too_many_arguments", TestOptions::compile_error());
}

#[test]
fn test_error_03_too_few_arguments() {
    test_e2e("error_03_too_few_arguments", TestOptions::compile_error());
}

#[test]
fn test_error_04_bad_parameter() {
    test_e2e("error_04_bad_parameter", TestOptions::compile_error());
}

#[test]
fn test_error_05_undefined_symbol() {
    test_e2e("error_05_undefined_symbol", TestOptions::compile_error());
}

#[test]
fn test_error_06_symbol_of_wrong_type() {
    test_e2e(
        "error_06_symbol_of_wrong_type",
        TestOptions::compile_error(),
    );
}

#[test]
fn test_error_07_assign_to_unknown_symbol() {
    test_e2e(
        "error_07_assign_to_unknown_symbol",
        TestOptions::compile_error(),
    );
}

#[test]
fn test_error_08_assign_of_wrong_type() {
    test_e2e(
        "error_08_assign_of_wrong_type",
        TestOptions::compile_error(),
    );
}

#[test]
fn test_error_09_bad_if_conditions() {
    test_e2e("error_09_bad_if_conditions", TestOptions::compile_error());
}

#[test]
fn test_error_10_bad_while_condition() {
    test_e2e("error_10_bad_while_condition", TestOptions::compile_error());
}

#[test]
fn test_error_11_calling_not_a_function() {
    test_e2e(
        "error_11_calling_not_a_function",
        TestOptions::compile_error(),
    );
}

#[test]
fn test_error_12_bad_list_indices() {
    test_e2e("error_12_bad_list_indices", TestOptions::compile_error());
}

#[test]
fn test_error_13_bad_return_type() {
    test_e2e("error_13_bad_return_type", TestOptions::compile_error());
}

#[test]
fn test_error_14_no_return() {
    test_e2e("error_14_no_return", TestOptions::compile_error());
}

struct TestOptions {
    args: Vec<&'static str>,
    expect_compile_error: bool,
    expect_error: bool,
    snapshot_vil: bool,
    snapshot_x86: bool,
}

impl TestOptions {
    /// A full end-to-end test that checks the VIL and x86 code and the output of the Venice program
    /// against stored snapshots.
    fn full() -> Self {
        TestOptions {
            args: Vec::new(),
            expect_compile_error: false,
            expect_error: false,
            snapshot_vil: true,
            snapshot_x86: true,
        }
    }

    /// Simple test that only checks the output of the Venice program, not the VIL and x86
    /// snapshots.
    fn simple() -> Self {
        TestOptions {
            args: Vec::new(),
            expect_compile_error: false,
            expect_error: false,
            snapshot_vil: false,
            snapshot_x86: false,
        }
    }

    /// Like `full`, except that arguments can be specified to pass to the Venice program.
    fn full_with_args(args: Vec<&'static str>) -> Self {
        TestOptions {
            args,
            expect_compile_error: false,
            expect_error: false,
            snapshot_vil: true,
            snapshot_x86: true,
        }
    }

    /// Like `full`, except that the Venice program is expected to return an error code.
    fn runtime_error() -> Self {
        TestOptions {
            args: Vec::new(),
            expect_compile_error: false,
            expect_error: true,
            snapshot_vil: true,
            snapshot_x86: true,
        }
    }

    /// An end-to-end test that expects the Venice program to fail to compile and checks the error
    /// message against a stored snapshot.
    fn compile_error() -> Self {
        TestOptions {
            args: Vec::new(),
            expect_compile_error: true,
            expect_error: false,
            snapshot_vil: false,
            snapshot_x86: false,
        }
    }
}

fn test_e2e(base_name: &str, options: TestOptions) {
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
    let compiler_output = Command::new("target/debug/venice")
        .arg(&input_path)
        .arg("--debug")
        .arg("--keep-intermediate")
        .output()
        .unwrap();

    if options.expect_compile_error {
        assert!(!compiler_output.status.success());
        let stdout = str::from_utf8(&compiler_output.stdout).unwrap();
        insta::assert_display_snapshot!(format!("{}-compiler-stdout", base_name), stdout);
        return;
    } else {
        assert!(compiler_output.status.success());
    }

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
