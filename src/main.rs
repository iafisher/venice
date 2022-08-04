mod analyzer;
mod ast;
mod codegen;
mod common;
mod lexer;
mod ptree;
mod vil;
mod x86;

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

    let fibonacci_ptree = ptree::Program {
        declarations: vec![ptree::Declaration::Function(ptree::FunctionDeclaration {
            name: String::from("fibonacci"),
            parameters: vec![ptree::FunctionParameter {
                name: String::from("n"),
                type_: ptree::Type::Literal(String::from("i64")),
            }],
            return_type: ptree::Type::Literal(String::from("i64")),
            body: vec![
                ptree::Statement::Let(ptree::LetStatement {
                    symbol: String::from("fib_i"),
                    type_: ptree::Type::Literal(String::from("i64")),
                    value: ptree::Expression::Integer(1),
                }),
                ptree::Statement::Let(ptree::LetStatement {
                    symbol: String::from("fib_i_minus_1"),
                    type_: ptree::Type::Literal(String::from("i64")),
                    value: ptree::Expression::Integer(0),
                }),
                ptree::Statement::Let(ptree::LetStatement {
                    symbol: String::from("i"),
                    type_: ptree::Type::Literal(String::from("i64")),
                    value: ptree::Expression::Integer(1),
                }),
                ptree::Statement::While(ptree::WhileStatement {
                    condition: ptree::Expression::Binary(ptree::BinaryExpression {
                        op: ptree::BinaryOpType::LessThan,
                        left: Box::new(ptree::Expression::Symbol(String::from("i"))),
                        right: Box::new(ptree::Expression::Symbol(String::from("n"))),
                    }),
                    body: vec![
                        ptree::Statement::Let(ptree::LetStatement {
                            symbol: String::from("tmp"),
                            type_: ptree::Type::Literal(String::from("i64")),
                            value: ptree::Expression::Symbol(String::from("fib_i")),
                        }),
                        ptree::Statement::Assign(ptree::AssignStatement {
                            symbol: String::from("fib_i"),
                            value: ptree::Expression::Binary(ptree::BinaryExpression {
                                op: ptree::BinaryOpType::Add,
                                left: Box::new(ptree::Expression::Symbol(String::from("fib_i"))),
                                right: Box::new(ptree::Expression::Symbol(String::from(
                                    "fib_i_minus_1",
                                ))),
                            }),
                        }),
                        ptree::Statement::Assign(ptree::AssignStatement {
                            symbol: String::from("fib_i_minus_1"),
                            value: ptree::Expression::Symbol(String::from("tmp")),
                        }),
                        ptree::Statement::Assign(ptree::AssignStatement {
                            symbol: String::from("i"),
                            value: ptree::Expression::Binary(ptree::BinaryExpression {
                                op: ptree::BinaryOpType::Add,
                                left: Box::new(ptree::Expression::Symbol(String::from("i"))),
                                right: Box::new(ptree::Expression::Integer(1)),
                            }),
                        }),
                    ],
                }),
                ptree::Statement::Return(ptree::ReturnStatement {
                    value: ptree::Expression::Symbol(String::from("fib_i")),
                }),
            ],
        })],
    };

    println!("{}", fibonacci_ptree);
}
