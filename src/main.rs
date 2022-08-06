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
    }
}

fn String(s: &str) -> ast::Expression {
    ast::Expression {
        kind: ast::ExpressionKind::Str(String::from(s)),
        semantic_type: ast::Type::Unknown,
    }
}

fn Symbol(s: &str) -> ast::Expression {
    ast::Expression {
        kind: ast::ExpressionKind::Symbol(String::from(s)),
        semantic_type: ast::Type::Unknown,
    }
}

fn I64() -> ast::SyntacticType {
    ast::SyntacticType::Literal(String::from("i64"))
}

fn main() {
    let _fibonacci_program = r"
    func fibonacci(n: i64) -> i64 {
      let fib_i: i64 = 1;
      let fib_i_minus_1 : i64 = 0;
      let i: i64 = 1;

      while (i < n) {
        let tmp: i64 = fib_i;
        fib_i += fib_i_minus_1;
        fib_i_minus_1 = tmp;
        i += 1;
      }

      return fib_i;
    }
    ";

    let fibonacci_ast = ast::Program {
        declarations: vec![ast::Declaration::Function(ast::FunctionDeclaration {
            name: String::from("fibonacci"),
            parameters: vec![ast::FunctionParameter {
                name: String::from("n"),
                type_: I64(),
                semantic_type: ast::Type::Unknown,
            }],
            return_type: I64(),
            semantic_return_type: ast::Type::Unknown,
            body: vec![
                ast::Statement::Let(ast::LetStatement {
                    symbol: String::from("fib_i"),
                    type_: I64(),
                    semantic_type: ast::Type::Unknown,
                    value: Integer(1),
                }),
                ast::Statement::Let(ast::LetStatement {
                    symbol: String::from("fib_i_minus_1"),
                    type_: I64(),
                    semantic_type: ast::Type::Unknown,
                    value: Integer(0),
                }),
                ast::Statement::Let(ast::LetStatement {
                    symbol: String::from("i"),
                    type_: I64(),
                    semantic_type: ast::Type::Unknown,
                    value: Integer(1),
                }),
                ast::Statement::While(ast::WhileStatement {
                    condition: ast::Expression {
                        kind: ast::ExpressionKind::Binary(ast::BinaryExpression {
                            op: ast::BinaryOpType::LessThan,
                            left: Box::new(Symbol("i")),
                            right: Box::new(Symbol("n")),
                        }),
                        semantic_type: ast::Type::Unknown,
                    },
                    body: vec![
                        ast::Statement::Let(ast::LetStatement {
                            symbol: String::from("tmp"),
                            type_: I64(),
                            semantic_type: ast::Type::Unknown,
                            value: Symbol("fib_i"),
                        }),
                        ast::Statement::Assign(ast::AssignStatement {
                            symbol: String::from("fib_i"),
                            value: ast::Expression {
                                kind: ast::ExpressionKind::Binary(ast::BinaryExpression {
                                    op: ast::BinaryOpType::Add,
                                    left: Box::new(Symbol("fib_i")),
                                    right: Box::new(Symbol("fib_i_minus_1")),
                                }),
                                semantic_type: ast::Type::Unknown,
                            },
                        }),
                        ast::Statement::Assign(ast::AssignStatement {
                            symbol: String::from("fib_i_minus_1"),
                            value: Symbol("tmp"),
                        }),
                        ast::Statement::Assign(ast::AssignStatement {
                            symbol: String::from("i"),
                            value: ast::Expression {
                                kind: ast::ExpressionKind::Binary(ast::BinaryExpression {
                                    op: ast::BinaryOpType::Add,
                                    left: Box::new(Symbol("i")),
                                    right: Box::new(Integer(1)),
                                }),
                                semantic_type: ast::Type::Unknown,
                            },
                        }),
                    ],
                }),
                ast::Statement::Return(ast::ReturnStatement {
                    value: Symbol("fib_i"),
                }),
            ],
        })],
    };

    println!("{}", fibonacci_ast);
}
