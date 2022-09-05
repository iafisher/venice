// Copyright 2022 The Venice Authors. All rights reserved.
// Use of this source code is governed by an MIT-style license that can be
// found in the LICENSE file.

use clap::Parser;
use std::fs;
use std::fs::File;
use std::io::prelude::*;
use std::io::{BufReader, BufWriter};
use std::path::PathBuf;
use std::process::Command;
use std::time::Instant;

#[macro_use]
extern crate lazy_static;

mod analyzer;
mod ast;
mod codegen;
mod common;
mod errors;
mod lexer;
mod parser;
mod ptree;
mod vil;
mod x86;

/// The compiler for the Venice programming language
#[derive(Parser, Debug)]
struct Cli {
    /// Path to the Venice program to execute.
    #[clap(value_parser)]
    path: String,

    /// Preserves intermediate files.
    #[clap(long)]
    keep_intermediate: bool,

    /// Includes debugging symbols in the executable.
    #[clap(long)]
    debug: bool,

    /// Prints the AST and exits.
    #[clap(long)]
    ast: bool,

    /// Prints execution time of the different stages of compilation.
    #[clap(long)]
    profile: bool,
}

fn main() {
    let cli = Cli::parse();

    // Open the input file.
    let file = File::open(&cli.path).expect("could not open file");
    let mut buf_reader = BufReader::new(file);
    let mut program = String::new();
    buf_reader
        .read_to_string(&mut program)
        .expect("could not read from file");

    // Lex and parse the program.
    let mut now = Instant::now();
    let lexer = lexer::Lexer::new(&cli.path, &program);
    let ptree_result = parser::parse(lexer);
    if let Err(errors) = ptree_result {
        for error in errors {
            println!("error: {} ({})", error.message, error.location);
        }
        std::process::exit(1);
    }

    if cli.profile {
        let elapsed = now.elapsed();
        println!("Parsing: {:.2?}", elapsed);
    }

    // Type-check the program.
    now = Instant::now();
    let ptree = ptree_result.unwrap();
    let ast_result = analyzer::analyze(&ptree);
    if let Err(errors) = ast_result {
        for error in errors {
            println!("error: {} ({})", error.message, error.location);
        }
        std::process::exit(1);
    }

    if cli.profile {
        let elapsed = now.elapsed();
        println!("Analysis: {:.2?}", elapsed);
    }

    let ast = ast_result.unwrap();
    if cli.ast {
        println!("{}", ast);
        std::process::exit(0);
    }

    // Generate a VIL program.
    now = Instant::now();
    let vil_program = codegen::generate(&ast).unwrap();
    if cli.keep_intermediate {
        let mut vil_output_path = PathBuf::from(&cli.path);
        vil_output_path.set_extension("vil");
        {
            let f = File::create(&vil_output_path).expect("could not create file");
            let mut writer = BufWriter::new(f);
            writer
                .write_all(&format!("{}", vil_program).into_bytes())
                .expect("could not write to file");
        }
    }

    if cli.profile {
        let elapsed = now.elapsed();
        println!("Code generation (VIL): {:.2?}", elapsed);
    }

    // Generate an x86 program.
    now = Instant::now();
    let x86_program = x86::generate(&vil_program).unwrap();

    if cli.profile {
        let elapsed = now.elapsed();
        println!("Code generation (x86): {:.2?}", elapsed);
    }

    // Write the assembly program to disk.
    let mut x86_output_path = PathBuf::from(&cli.path);
    x86_output_path.set_extension("x86.s");
    {
        let f = File::create(&x86_output_path).expect("could not create file");
        let mut writer = BufWriter::new(f);
        writer
            .write_all(&format!("{}", x86_program).into_bytes())
            .expect("could not write to file");
    }

    let mut object_output_path = PathBuf::from(&cli.path);
    object_output_path.set_extension("o");
    let mut output_path = PathBuf::from(&cli.path);
    output_path.set_extension("");

    // Invoke gcc to turn the textual assembly program into a binary executable.
    let mut cmd = Command::new("gcc");
    if cli.debug {
        cmd.arg("-g");
    }

    let runtime_library = if cli.debug {
        "runtime/libvenice-debug.so"
    } else {
        "runtime/libvenice.so"
    };
    let main_wrapper = if cli.debug {
        "runtime/main-debug.o"
    } else {
        "runtime/main.o"
    };
    let mut child = cmd
        .arg("-no-pie")
        .arg("-o")
        .arg(&output_path)
        .arg(&main_wrapper)
        .arg(&x86_output_path)
        .arg(&runtime_library)
        .spawn()
        .expect("failed to execute gcc");
    let error_code = child.wait().expect("failed to wait on child");
    if !error_code.success() {
        if let Some(error_code) = error_code.code() {
            panic!("gcc returned non-zero exit code: {}", error_code);
        } else {
            panic!("gcc returned non-zero exit code");
        }
    }

    // Clean up the intermediate files.
    if !cli.keep_intermediate {
        // Assembly file is needed for debugging.
        if !cli.debug {
            let _ = fs::remove_file(&x86_output_path);
        }
        let _ = fs::remove_file(&object_output_path);
    }
}
