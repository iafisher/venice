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
mod vil;
mod x86;

/// The compiler for the Venice programming language
#[derive(Parser, Debug)]
struct Cli {
    /// Path to the Venice program to execute.
    #[clap(value_parser)]
    path: String,

    /// Prints intermediate data structures and preserves intermediate files.
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
    let ast_result = parser::parse(lexer);
    if let Err(errors) = ast_result {
        for error in errors {
            println!("error: {} ({})", error.message, error.location);
        }
        std::process::exit(1);
    }

    let mut ast = ast_result.unwrap();
    if cli.debug {
        println!("Parse tree:\n");
        println!("  {}", ast);
    }

    // Type-check the program.
    let typecheck_result = analyzer::analyze(&mut ast);
    if let Err(errors) = typecheck_result {
        for error in errors {
            println!("error: {} ({})", error.message, error.location);
        }
        std::process::exit(1);
    }

    if cli.debug {
        println!("\nTyped AST:\n");
        println!("  {}", ast);
    }

    // Generate a VIL program.
    let vil_program = codegen::generate(&ast).unwrap();
    if cli.debug {
        println!("\nVIL:\n");
        println!("{}", vil_program);
    }

    // Generate an x86 program.
    let x86_program = x86::generate(&vil_program).unwrap();
    if cli.debug {
        println!("\nx86:\n");
        println!("{}", x86_program);
    }

    // Write the assembly program to disk.
    let mut asm_output_path = PathBuf::from(&cli.path);
    asm_output_path.set_extension("s");
    {
        let f = File::create(&asm_output_path).expect("could not create file");
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
    let mut child = Command::new("nasm")
        // TODO: only use debugging flags when explicitly requested
        .arg("-g")
        .arg("-F")
        .arg("dwarf")
        .arg("-f")
        .arg("elf64")
        .arg("-o")
        .arg(&object_output_path)
        .arg(&asm_output_path)
        .spawn()
        .expect("failed to execute nasm");
    let mut error_code = child.wait().expect("failed to wait on child");
    if !error_code.success() {
        panic!("nasm returned non-zero exit code");
    }

    // Invoke ld to link the binary object file into an ELF executable. It is necessary
    // to supply the various glibc libraries and initialization code because the Venice
    // runtime relies on libc. See https://stackoverflow.com/questions/3577922/ for more
    // information.
    child = Command::new("ld")
        .arg("-dynamic-linker")
        // TODO: don't hard-code these values
        .arg("/lib64/ld-linux-x86-64.so.2")
        .arg("/usr/lib/x86_64-linux-gnu/crt1.o")
        .arg("/usr/lib/x86_64-linux-gnu/crti.o")
        .arg("runtime/libvenice.so")
        .arg("-lc")
        .arg(&object_output_path)
        .arg("/usr/lib/x86_64-linux-gnu/crtn.o")
        .arg("-o")
        .arg(&output_path)
        .spawn()
        .expect("failed to execute ld");
    error_code = child.wait().expect("failed to wait on child");
    if !error_code.success() {
        panic!("ld returned non-zero exit code");
    }

    // Clean up the intermediate files.
    if !cli.debug {
        let _ = fs::remove_file(&asm_output_path);
        let _ = fs::remove_file(&object_output_path);
    }
}
