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
    let lexer = lexer::Lexer::new(&cli.path, &program);
    let ptree_result = parser::parse(lexer);
    if let Err(errors) = ptree_result {
        for error in errors {
            println!("error: {} ({})", error.message, error.location);
        }
        std::process::exit(1);
    }

    // Type-check the program.
    let ptree = ptree_result.unwrap();
    let ast_result = analyzer::analyze(&ptree);
    if let Err(errors) = ast_result {
        for error in errors {
            println!("error: {} ({})", error.message, error.location);
        }
        std::process::exit(1);
    }

    // Generate a VIL program.
    let ast = ast_result.unwrap();
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

    // Generate an x86 program.
    let x86_program = x86::generate(&vil_program).unwrap();

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

    // Invoke nasm to turn the textual assembly program into a binary object file.
    let mut cmd = Command::new("nasm");
    if cli.debug {
        cmd.arg("-g");
    }
    let mut child = cmd
        .arg("-F")
        .arg("dwarf")
        .arg("-f")
        .arg("elf64")
        .arg("-o")
        .arg(&object_output_path)
        .arg(&x86_output_path)
        .spawn()
        .expect("failed to execute nasm");
    let mut error_code = child.wait().expect("failed to wait on child");
    if !error_code.success() {
        if let Some(error_code) = error_code.code() {
            panic!("nasm returned non-zero exit code: {}", error_code);
        } else {
            panic!("nasm returned non-zero exit code");
        }
    }

    // Invoke ld to link the binary object file into an ELF executable. It is necessary
    // to supply the various glibc libraries and initialization code because the Venice
    // runtime relies on libc. See https://stackoverflow.com/questions/3577922/ for more
    // information.
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
    child = Command::new("ld")
        .arg("-dynamic-linker")
        // TODO: don't hard-code these values
        .arg("/lib64/ld-linux-x86-64.so.2")
        .arg("/usr/lib/x86_64-linux-gnu/crt1.o")
        .arg("/usr/lib/x86_64-linux-gnu/crti.o")
        .arg(runtime_library)
        .arg("-lc")
        .arg(&object_output_path)
        .arg("/usr/lib/x86_64-linux-gnu/crtn.o")
        .arg("-o")
        .arg(&output_path)
        .arg(main_wrapper)
        .spawn()
        .expect("failed to execute ld");
    error_code = child.wait().expect("failed to wait on child");
    if !error_code.success() {
        if let Some(error_code) = error_code.code() {
            panic!("ld returned non-zero exit code: {}", error_code);
        } else {
            panic!("ld returned non-zero exit code");
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
