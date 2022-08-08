#[macro_use]
extern crate lazy_static;

mod analyzer;
mod ast;
mod codegen;
mod common;
mod errors;
mod lexer;
mod vil;
mod x86;

fn Integer(x: i64) -> ast::Expression {
    ast::Expression {
        kind: ast::ExpressionKind::Integer(x),
        semantic_type: ast::Type::Unknown,
        location: common::Location::empty(),
    }
}

fn String(s: &str) -> ast::Expression {
    ast::Expression {
        kind: ast::ExpressionKind::Str(String::from(s)),
        semantic_type: ast::Type::Unknown,
        location: common::Location::empty(),
    }
}

fn Symbol(s: &str) -> ast::Expression {
    ast::Expression {
        kind: ast::ExpressionKind::Symbol(String::from(s)),
        semantic_type: ast::Type::Unknown,
        location: common::Location::empty(),
    }
}

fn I64() -> ast::SyntacticType {
    ast::SyntacticType {
        kind: ast::SyntacticTypeKind::Literal(String::from("i64")),
        location: common::Location::empty(),
    }
}

fn location(line: u32, column: u32) -> common::Location {
    common::Location {
        file: String::from("<string>"),
        line: line,
        column: column,
    }
}

fn main() {
    let _fibonacci_program = r"
    func fibonacci(n: i64) -> i64 {
      let fib_i: i64 = 1;
      let fib_i_minus_1 : i64 = 0;
      let i: i64 = 1;

      while (i < n) {
        let tmp: i64 = fib_i;
        fib_i = fib_i + fib_i_minus_1;
        fib_i_minus_1 = tmp;
        i = i + 1;
      }

      return fib_i;
    }
    ";

    let mut fibonacci_ast = ast::Program {
        declarations: vec![ast::Declaration::Function(ast::FunctionDeclaration {
            name: String::from("fibonacci"),
            parameters: vec![ast::FunctionParameter {
                name: String::from("n"),
                type_: I64(),
                semantic_type: ast::Type::Unknown,
            }],
            return_type: I64(),
            semantic_return_type: ast::Type::Unknown,
            location: location(1, 1),
            body: vec![
                ast::Statement::Let(ast::LetStatement {
                    symbol: String::from("fib_i"),
                    type_: I64(),
                    semantic_type: ast::Type::Unknown,
                    value: Integer(1),
                    location: location(2, 3),
                }),
                ast::Statement::Let(ast::LetStatement {
                    symbol: String::from("fib_i_minus_1"),
                    type_: I64(),
                    semantic_type: ast::Type::Unknown,
                    value: Integer(0),
                    location: location(3, 3),
                }),
                ast::Statement::Let(ast::LetStatement {
                    symbol: String::from("i"),
                    type_: I64(),
                    semantic_type: ast::Type::Unknown,
                    value: Integer(1),
                    location: location(4, 3),
                }),
                ast::Statement::While(ast::WhileStatement {
                    condition: ast::Expression {
                        kind: ast::ExpressionKind::Binary(ast::BinaryExpression {
                            op: ast::BinaryOpType::LessThan,
                            left: Box::new(Symbol("i")),
                            right: Box::new(Symbol("n")),
                            location: location(6, 9),
                        }),
                        semantic_type: ast::Type::Unknown,
                        location: location(6, 9),
                    },
                    location: location(6, 3),
                    body: vec![
                        ast::Statement::Let(ast::LetStatement {
                            symbol: String::from("tmp"),
                            type_: I64(),
                            semantic_type: ast::Type::Unknown,
                            value: Symbol("fib_i"),
                            location: location(7, 5),
                        }),
                        ast::Statement::Assign(ast::AssignStatement {
                            symbol: String::from("fib_i"),
                            value: ast::Expression {
                                kind: ast::ExpressionKind::Binary(ast::BinaryExpression {
                                    op: ast::BinaryOpType::Add,
                                    left: Box::new(Symbol("fib_i")),
                                    right: Box::new(Symbol("fib_i_minus_1")),
                                    location: location(8, 13),
                                }),
                                semantic_type: ast::Type::Unknown,
                                location: location(8, 13),
                            },
                            location: location(8, 5),
                        }),
                        ast::Statement::Assign(ast::AssignStatement {
                            symbol: String::from("fib_i_minus_1"),
                            value: Symbol("tmp"),
                            location: location(9, 5),
                        }),
                        ast::Statement::Assign(ast::AssignStatement {
                            symbol: String::from("i"),
                            value: ast::Expression {
                                kind: ast::ExpressionKind::Binary(ast::BinaryExpression {
                                    op: ast::BinaryOpType::Add,
                                    left: Box::new(Symbol("i")),
                                    right: Box::new(Integer(1)),
                                    location: location(10, 9),
                                }),
                                semantic_type: ast::Type::Unknown,
                                location: location(10, 9),
                            },
                            location: location(10, 5),
                        }),
                    ],
                }),
                ast::Statement::Return(ast::ReturnStatement {
                    value: Symbol("fib_i"),
                    location: location(13, 3),
                }),
            ],
        })],
    };

    println!("{}", fibonacci_ast);
    let r = analyzer::analyze(&mut fibonacci_ast);
    if let Err(errors) = r {
        for error in errors {
            println!("error: {} ({})", error.message, error.location);
        }
    } else {
        println!("\nNo errors!");
    }
}
