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

    let vil_program = codegen::generate_ir(&ast).unwrap();
    println!("\nVIL:\n");
    println!("{}", vil_program);

    /*
    let vil_program = vil::Program {
        declarations: vec![vil::Declaration::Function(vil::FunctionDeclaration {
            name: String::from("fibonacci"),
            parameters: vec![vil::FunctionParameter {
                name: String::from("n"),
                type_: vil::Type::I64,
            }],
            return_type: vil::Type::I64,
            blocks: vec![
                vil::Block {
                    name: String::from("f0"),
                    instructions: vec![
                        vil::Instruction::Alloca {
                            symbol: String::from("fib_i"),
                            type_: vil::Type::I64,
                            size: 8,
                        },
                        vil::Instruction::Store {
                            symbol: String::from("fib_i"),
                            expression: vil::TypedExpression {
                                type_: vil::Type::I64,
                                value: vil::Expression::Integer(1),
                            },
                        },
                        vil::Instruction::Alloca {
                            symbol: String::from("fib_i_minus_1"),
                            type_: vil::Type::I64,
                            size: 8,
                        },
                        vil::Instruction::Store {
                            symbol: String::from("fib_i_minus_1"),
                            expression: vil::TypedExpression {
                                type_: vil::Type::I64,
                                value: vil::Expression::Integer(0),
                            },
                        },
                        vil::Instruction::Alloca {
                            symbol: String::from("i"),
                            type_: vil::Type::I64,
                            size: 8,
                        },
                        vil::Instruction::Store {
                            symbol: String::from("i"),
                            expression: vil::TypedExpression {
                                type_: vil::Type::I64,
                                value: vil::Expression::Integer(1),
                            },
                        },
                    ],
                    exit: vil::ExitInstruction::Jump(String::from("loop0_cmp")),
                },
                vil::Block {
                    name: String::from("loop0_cmp"),
                    instructions: vec![
                        vil::Instruction::Alloca {
                            symbol: String::from("t0"),
                            type_: vil::Type::I64,
                            size: 8,
                        },
                        vil::Instruction::CmpLt {
                            symbol: String::from("t0"),
                            left: vil::TypedExpression {
                                type_: vil::Type::I64,
                                value: vil::Expression::Symbol(String::from("i")),
                            },
                            right: vil::TypedExpression {
                                type_: vil::Type::I64,
                                value: vil::Expression::Symbol(String::from("n")),
                            },
                        },
                    ],
                    exit: vil::ExitInstruction::JumpCond {
                        condition: vil::TypedExpression {
                            type_: vil::Type::I64,
                            value: vil::Expression::Symbol(String::from("t0")),
                        },
                        label_true: String::from("loop0"),
                        label_false: String::from("loop0_end"),
                    },
                },
                vil::Block {
                    name: String::from("loop0"),
                    instructions: vec![
                        vil::Instruction::Alloca {
                            symbol: String::from("tmp"),
                            type_: vil::Type::I64,
                            size: 8,
                        },
                        vil::Instruction::Store {
                            symbol: String::from("tmp"),
                            expression: vil::TypedExpression {
                                type_: vil::Type::I64,
                                value: vil::Expression::Symbol(String::from("fib_i")),
                            },
                        },
                        vil::Instruction::Add {
                            symbol: String::from("fib_i"),
                            left: vil::TypedExpression {
                                type_: vil::Type::I64,
                                value: vil::Expression::Symbol(String::from("fib_i")),
                            },
                            right: vil::TypedExpression {
                                type_: vil::Type::I64,
                                value: vil::Expression::Symbol(String::from("fib_i_minus_1")),
                            },
                        },
                        vil::Instruction::Store {
                            symbol: String::from("fib_i_minus_1"),
                            expression: vil::TypedExpression {
                                type_: vil::Type::I64,
                                value: vil::Expression::Symbol(String::from("tmp")),
                            },
                        },
                        vil::Instruction::Add {
                            symbol: String::from("i"),
                            left: vil::TypedExpression {
                                type_: vil::Type::I64,
                                value: vil::Expression::Symbol(String::from("i")),
                            },
                            right: vil::TypedExpression {
                                type_: vil::Type::I64,
                                value: vil::Expression::Integer(1),
                            },
                        },
                    ],
                    exit: vil::ExitInstruction::Jump(String::from("loop0_cmp")),
                },
                vil::Block {
                    name: String::from("loop0_end"),
                    instructions: vec![],
                    exit: vil::ExitInstruction::Ret(vil::TypedExpression {
                        type_: vil::Type::I64,
                        value: vil::Expression::Symbol(String::from("fib_i")),
                    }),
                },
            ],
        })],
    };
    */
}
