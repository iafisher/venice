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

fn main() {
    let fibonacci_program = r"
    func fibonacci(n: i64) -> i64 {
      let fib_i: i64 = 1;
      let fib_i_minus_1: i64 = 0;
      let i: i64 = 1;

      while i < n {
        let tmp: i64 = fib_i;
        fib_i = fib_i + fib_i_minus_1;
        fib_i_minus_1 = tmp;
        i = i + 1;
      }

      return fib_i;
    }
    ";

    let mut lexer = lexer::Lexer::new("<string>", fibonacci_program);
    let ast_result = parser::parse(lexer);
    if let Err(errors) = ast_result {
        for error in errors {
            println!("error: {} ({})", error.message, error.location);
        }
        std::process::exit(1);
    }

    let mut ast = ast_result.unwrap();
    println!("Parse tree:\n");
    println!("  {}", ast);

    let typecheck_result = analyzer::analyze(&mut ast);
    if let Err(errors) = typecheck_result {
        for error in errors {
            println!("error: {} ({})", error.message, error.location);
        }
        std::process::exit(1);
    }

    println!("\nTyped AST:\n");
    println!("  {}", ast);
}
