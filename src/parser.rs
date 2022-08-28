// The parser transforms the stream of tokens emitted by the lexer into an abstract
// syntax tree.

use super::common;
use super::errors;
use super::lexer;
use super::lexer::TokenType;
use super::ptree;
use std::collections::HashMap;

pub fn parse(lexer: lexer::Lexer) -> Result<ptree::Program, Vec<errors::VeniceError>> {
    let mut parser = Parser::new(lexer);
    let ptree = parser.parse();
    if !parser.errors.is_empty() {
        Err(parser.errors.clone())
    } else {
        Ok(ptree)
    }
}

struct Parser {
    lexer: lexer::Lexer,
    errors: Vec<errors::VeniceError>,
}

impl Parser {
    fn new(lexer: lexer::Lexer) -> Self {
        Parser {
            lexer,
            errors: Vec::new(),
        }
    }

    fn parse(&mut self) -> ptree::Program {
        let mut declarations = Vec::new();
        while !self.lexer.done() {
            if let Ok(declaration) = self.match_declaration() {
                declarations.push(declaration);
            }
        }
        ptree::Program { declarations }
    }

    fn match_declaration(&mut self) -> Result<ptree::Declaration, ()> {
        let token = self.lexer.token();
        match token.type_ {
            TokenType::Func => self
                .match_function_declaration()
                .map(ptree::Declaration::Function),
            // TODO: handle const and record declarations
            _ => {
                let msg = format!(
                    "expected const, func, or record declaration, got {}",
                    token.value
                );
                self.errors
                    .push(errors::VeniceError::new(&msg, token.location));

                self.skip_past(TokenType::CurlyClose);
                Err(())
            }
        }
    }

    fn match_function_declaration(&mut self) -> Result<ptree::FunctionDeclaration, ()> {
        let location = self.lexer.token().location;
        self.expect_token(&self.lexer.token(), TokenType::Func, "func keyword")?;

        let mut token = self.lexer.next();
        self.expect_token(&token, TokenType::Symbol, "function name")?;
        let name = token.value;

        token = self.lexer.next();
        self.expect_token(&token, TokenType::ParenOpen, "(")?;

        self.lexer.next();
        let mut parameters = Vec::new();
        loop {
            token = self.lexer.token();
            if token.type_ == TokenType::ParenClose {
                break;
            }

            self.expect_token(&token, TokenType::Symbol, "parameter name")?;
            let parameter_name = token.value;
            let _parameter_location = token.location.clone();

            token = self.lexer.next();
            self.expect_token(&token, TokenType::Colon, ":")?;

            self.lexer.next();
            let type_ = self.match_type()?;
            parameters.push(ptree::FunctionParameter {
                name: parameter_name,
                type_,
            });

            token = self.lexer.token();
            if token.type_ == TokenType::Comma {
                self.lexer.next();
            } else if token.type_ == TokenType::ParenClose {
                break;
            } else {
                self.unexpected(&token, "comma or )");
                return Err(());
            }
        }

        token = self.lexer.next();
        self.expect_token(&token, TokenType::Arrow, "->")?;

        self.lexer.next();
        let return_type = self.match_type()?;

        let body = self.match_block()?;
        Ok(ptree::FunctionDeclaration {
            name,
            parameters,
            return_type,
            body,
            location,
        })
    }

    fn match_block(&mut self) -> Result<Vec<ptree::Statement>, ()> {
        let mut token = self.lexer.token();
        self.expect_token(&token, TokenType::CurlyOpen, "{")?;
        self.lexer.next();

        let mut statements = Vec::new();
        loop {
            token = self.lexer.token();
            if token.type_ == TokenType::CurlyClose {
                self.lexer.next();
                break;
            } else if token.type_ == TokenType::End {
                self.unexpected(&token, "statement or end of block");
                return Err(());
            }

            if let Ok(statement) = self.match_statement() {
                statements.push(statement);
            }
        }
        Ok(statements)
    }

    fn match_statement(&mut self) -> Result<ptree::Statement, ()> {
        let mut token = self.lexer.token();
        match token.type_ {
            TokenType::Assert => self.match_assert_statement().map(ptree::Statement::Assert),
            TokenType::If => self.match_if_statement().map(ptree::Statement::If),
            TokenType::Let => self.match_let_statement().map(ptree::Statement::Let),
            TokenType::Return => self.match_return_statement().map(ptree::Statement::Return),
            TokenType::While => self.match_while_statement().map(ptree::Statement::While),
            _ => {
                let expr = self.match_expression()?;
                token = self.lexer.token();
                if token.type_ == TokenType::Assign {
                    self.match_assign_statement(expr)
                        .map(ptree::Statement::Assign)
                } else if token.type_ == TokenType::Semicolon {
                    self.lexer.next();
                    Ok(ptree::Statement::Expression(expr))
                } else {
                    self.lexer.next();
                    self.unexpected(&token, "start of statement");
                    Err(())
                }
            }
        }
    }

    fn match_assert_statement(&mut self) -> Result<ptree::AssertStatement, ()> {
        let mut token = self.lexer.token();
        let location = token.location.clone();
        self.expect_token(&token, TokenType::Assert, "assert")?;

        self.lexer.next();
        let condition = self.match_expression()?;

        token = self.lexer.token();
        self.expect_token(&token, TokenType::Semicolon, ";")?;
        self.lexer.next();

        Ok(ptree::AssertStatement {
            condition,
            location,
        })
    }

    fn match_assign_statement(
        &mut self,
        expr: ptree::Expression,
    ) -> Result<ptree::AssignStatement, ()> {
        let symbol = if let ptree::ExpressionKind::Symbol(symbol) = expr.kind {
            symbol
        } else {
            self.error("can only assign to symbols", expr.location.clone());
            return Err(());
        };

        let mut token = self.lexer.token();
        let location = token.location.clone();
        self.expect_token(&token, TokenType::Assign, "=")?;

        self.lexer.next();
        let value = self.match_expression()?;

        token = self.lexer.token();
        self.expect_token(&token, TokenType::Semicolon, ";")?;
        self.lexer.next();

        Ok(ptree::AssignStatement {
            symbol,
            value,
            location,
        })
    }

    fn match_if_statement(&mut self) -> Result<ptree::IfStatement, ()> {
        let mut token = self.lexer.token();
        let location = token.location.clone();
        self.expect_token(&token, TokenType::If, "if")?;

        self.lexer.next();
        let condition = self.match_expression()?;
        let body = self.match_block()?;

        let mut elif_clauses = Vec::new();
        loop {
            token = self.lexer.token();
            if token.type_ == TokenType::Else {
                token = self.lexer.next();
                if token.type_ == TokenType::If {
                    self.lexer.next();
                    let elif_condition = self.match_expression()?;
                    let elif_body = self.match_block()?;
                    elif_clauses.push(ptree::IfClause {
                        condition: elif_condition,
                        body: elif_body,
                    });
                } else {
                    let else_body = self.match_block()?;
                    return Ok(ptree::IfStatement {
                        if_clause: ptree::IfClause { condition, body },
                        elif_clauses,
                        else_body,
                        location,
                    });
                }
            } else {
                return Ok(ptree::IfStatement {
                    if_clause: ptree::IfClause { condition, body },
                    elif_clauses,
                    else_body: Vec::new(),
                    location,
                });
            }
        }
    }

    fn match_let_statement(&mut self) -> Result<ptree::LetStatement, ()> {
        let mut token = self.lexer.token();
        let location = token.location.clone();
        self.expect_token(&token, TokenType::Let, "let")?;

        token = self.lexer.next();
        self.expect_token(&token, TokenType::Symbol, "symbol")?;
        let symbol = token.value;

        token = self.lexer.next();
        self.expect_token(&token, TokenType::Colon, ":")?;

        self.lexer.next();
        let type_ = self.match_type()?;

        token = self.lexer.token();
        self.expect_token(&token, TokenType::Assign, "=")?;

        self.lexer.next();
        let value = self.match_expression()?;

        token = self.lexer.token();
        self.expect_token(&token, TokenType::Semicolon, ";")?;
        self.lexer.next();

        Ok(ptree::LetStatement {
            symbol,
            type_,
            value,
            location,
        })
    }

    fn match_return_statement(&mut self) -> Result<ptree::ReturnStatement, ()> {
        let mut token = self.lexer.token();
        let location = token.location.clone();
        // TODO: remove all of these? they are redundant unless I've made a programming
        // error since match_XYZ_statement will only be called when XYZ is the current
        // token.
        self.expect_token(&token, TokenType::Return, "return")?;

        self.lexer.next();
        let value = self.match_expression()?;

        token = self.lexer.token();
        self.expect_token(&token, TokenType::Semicolon, ";")?;
        self.lexer.next();

        Ok(ptree::ReturnStatement { value, location })
    }

    fn match_while_statement(&mut self) -> Result<ptree::WhileStatement, ()> {
        let token = self.lexer.token();
        let location = token.location.clone();
        self.expect_token(&token, TokenType::While, "while")?;

        self.lexer.next();
        let condition = self.match_expression()?;
        let body = self.match_block()?;
        Ok(ptree::WhileStatement {
            condition,
            body,
            location,
        })
    }

    fn match_expression(&mut self) -> Result<ptree::Expression, ()> {
        self.match_expression_with_precedence(PRECEDENCE_LOWEST)
    }

    fn match_expression_with_precedence(
        &mut self,
        precedence: u32,
    ) -> Result<ptree::Expression, ()> {
        let mut expr = self.match_literal()?;

        let mut token = self.lexer.token();
        loop {
            if let Some(other_precedence) = PRECEDENCE.get(&token.type_) {
                if precedence < *other_precedence {
                    if token.type_ == TokenType::ParenOpen {
                        self.lexer.next();
                        let call = self.match_function_call(&expr, token.location.clone())?;
                        expr = ptree::Expression {
                            kind: ptree::ExpressionKind::Call(call),
                            location: token.location.clone(),
                        };
                    } else if token.type_ == TokenType::SquareOpen {
                        self.lexer.next();
                        let index = self.match_expression()?;
                        self.expect_token(&self.lexer.token(), TokenType::SquareClose, "]")?;
                        self.lexer.next();
                        expr = ptree::Expression {
                            kind: ptree::ExpressionKind::Index(ptree::IndexExpression {
                                value: Box::new(expr),
                                index: Box::new(index),
                                location: token.location.clone(),
                            }),
                            location: token.location.clone(),
                        };
                    } else {
                        self.lexer.next();
                        let right = self.match_expression_with_precedence(*other_precedence)?;
                        if is_binary_comparison_op(token.type_) {
                            expr = ptree::Expression {
                                kind: ptree::ExpressionKind::Comparison(
                                    ptree::ComparisonExpression {
                                        op: token_type_to_comparison_op_type(token.type_),
                                        left: Box::new(expr),
                                        right: Box::new(right),
                                        location: token.location.clone(),
                                    },
                                ),
                                location: token.location.clone(),
                            };
                        } else {
                            expr = ptree::Expression {
                                kind: ptree::ExpressionKind::Binary(ptree::BinaryExpression {
                                    op: token_type_to_binary_op_type(token.type_),
                                    left: Box::new(expr),
                                    right: Box::new(right),
                                    location: token.location.clone(),
                                }),
                                location: token.location.clone(),
                            };
                        }
                    }
                    token = self.lexer.token();
                    continue;
                }
            }
            break;
        }
        Ok(expr)
    }

    fn match_function_call(
        &mut self,
        expr: &ptree::Expression,
        location: common::Location,
    ) -> Result<ptree::CallExpression, ()> {
        if let ptree::ExpressionKind::Symbol(name) = &expr.kind {
            let arguments = self.match_expression_list()?;
            let token = self.lexer.token();
            self.expect_token(&token, TokenType::ParenClose, ")")?;
            self.lexer.next();
            Ok(ptree::CallExpression {
                function: name.clone(),
                arguments,
                location,
            })
        } else {
            self.error("function must be a symbol", expr.location.clone());
            Err(())
        }
    }

    fn match_expression_list(&mut self) -> Result<Vec<ptree::Expression>, ()> {
        let mut items = Vec::new();
        loop {
            let mut token = self.lexer.token();
            if token.type_ == TokenType::ParenClose || token.type_ == TokenType::SquareClose {
                break;
            }

            let expr = self.match_expression()?;
            items.push(expr);

            token = self.lexer.token();
            if token.type_ == TokenType::ParenClose || token.type_ == TokenType::SquareClose {
                break;
            } else if token.type_ == TokenType::Comma {
                self.lexer.next();
            } else {
                self.unexpected(&token, "comma or closing bracket");
                return Err(());
            }
        }
        Ok(items)
    }

    fn match_literal(&mut self) -> Result<ptree::Expression, ()> {
        let token = self.lexer.token();
        match token.type_ {
            TokenType::Integer => {
                self.lexer.next();
                if let Ok(x) = token.value.parse::<i64>() {
                    Ok(ptree::Expression {
                        kind: ptree::ExpressionKind::Integer(x),
                        location: token.location.clone(),
                    })
                } else {
                    self.error("could not parse integer literal", token.location.clone());
                    Err(())
                }
            }
            TokenType::True => {
                self.lexer.next();
                Ok(ptree::Expression {
                    kind: ptree::ExpressionKind::Boolean(true),
                    location: token.location.clone(),
                })
            }
            TokenType::False => {
                self.lexer.next();
                Ok(ptree::Expression {
                    kind: ptree::ExpressionKind::Boolean(false),
                    location: token.location.clone(),
                })
            }
            TokenType::String => {
                self.lexer.next();
                if let Ok(s) = parse_string_literal(&token.value) {
                    Ok(ptree::Expression {
                        kind: ptree::ExpressionKind::String(s),
                        location: token.location.clone(),
                    })
                } else {
                    self.error("could not parse string literal", token.location.clone());
                    Err(())
                }
            }
            TokenType::Symbol => {
                self.lexer.next();
                Ok(ptree::Expression {
                    kind: ptree::ExpressionKind::Symbol(token.value),
                    location: token.location.clone(),
                })
            }
            TokenType::ParenOpen => {
                self.lexer.next();
                let expr = self.match_expression()?;
                self.expect_token(&self.lexer.token(), TokenType::ParenClose, ")")?;
                self.lexer.next();
                Ok(expr)
            }
            TokenType::SquareOpen => {
                self.lexer.next();
                let items = self.match_expression_list()?;
                self.expect_token(&self.lexer.token(), TokenType::SquareClose, "]")?;
                self.lexer.next();
                Ok(ptree::Expression {
                    kind: ptree::ExpressionKind::List(ptree::ListLiteral {
                        items,
                        location: token.location.clone(),
                    }),
                    location: token.location.clone(),
                })
            }
            TokenType::Minus => {
                self.lexer.next();
                let operand = self.match_expression()?;
                Ok(ptree::Expression {
                    kind: ptree::ExpressionKind::Unary(ptree::UnaryExpression {
                        op: common::UnaryOpType::Negate,
                        operand: Box::new(operand),
                        location: token.location.clone(),
                    }),
                    location: token.location.clone(),
                })
            }
            _ => {
                self.unexpected(&token, "start of expression");
                self.lexer.next();
                Err(())
            }
        }
    }

    fn match_type(&mut self) -> Result<ptree::Type, ()> {
        let mut token = self.lexer.token();
        let location = token.location.clone();
        self.expect_token(&token, TokenType::Symbol, "type")?;
        let symbol = token.value;
        token = self.lexer.next();

        if token.type_ == TokenType::LessThan {
            let mut parameters = Vec::new();
            self.lexer.next();
            loop {
                token = self.lexer.token();
                if token.type_ == TokenType::GreaterThan {
                    break;
                } else {
                    let type_ = self.match_type()?;
                    parameters.push(type_);
                    token = self.lexer.token();
                    if token.type_ == TokenType::Comma {
                        self.lexer.next();
                    } else if token.type_ == TokenType::GreaterThan {
                        break;
                    } else {
                        self.unexpected(&token, "comma or >");
                        self.lexer.next();
                        return Err(());
                    }
                }
            }
            self.lexer.next();
            Ok(ptree::Type {
                kind: ptree::TypeKind::Parameterized(ptree::ParameterizedType {
                    symbol,
                    parameters,
                }),
                location,
            })
        } else {
            Ok(ptree::Type {
                kind: ptree::TypeKind::Literal(symbol),
                location: location,
            })
        }
    }

    fn expect_token(
        &mut self,
        token: &lexer::Token,
        type_: TokenType,
        message: &str,
    ) -> Result<(), ()> {
        if token.type_ == type_ {
            Ok(())
        } else {
            self.unexpected(token, message);
            Err(())
        }
    }

    fn unexpected(&mut self, token: &lexer::Token, message: &str) {
        let msg = if token.type_ == TokenType::End {
            format!("expected {}, got end of file", message)
        } else {
            format!("expected {}, got {}", message, token.value)
        };

        self.error(&msg, token.location.clone());
    }

    fn error(&mut self, message: &str, location: common::Location) {
        self.errors
            .push(errors::VeniceError::new(message, location));
    }

    fn skip_past(&mut self, token_type: lexer::TokenType) {
        self.lexer.next();
        loop {
            let token = self.lexer.token();
            if token.type_ == TokenType::End {
                break;
            }
            if token.type_ == token_type {
                self.lexer.next();
                break;
            }
            self.lexer.next();
        }
    }
}

const PRECEDENCE_LOWEST: u32 = 0;
const PRECEDENCE_COMPARISON: u32 = 1;
const PRECEDENCE_ADDITION: u32 = 2;
const PRECEDENCE_MULTIPLICATION: u32 = 3;
const PRECEDENCE_CALL: u32 = 4;

lazy_static! {
    static ref PRECEDENCE: HashMap<TokenType, u32> = {
        let mut m = HashMap::new();
        m.insert(TokenType::GreaterThan, PRECEDENCE_COMPARISON);
        m.insert(TokenType::GreaterThanEquals, PRECEDENCE_COMPARISON);
        m.insert(TokenType::LessThan, PRECEDENCE_COMPARISON);
        m.insert(TokenType::LessThanEquals, PRECEDENCE_COMPARISON);
        m.insert(TokenType::Equals, PRECEDENCE_COMPARISON);
        m.insert(TokenType::NotEquals, PRECEDENCE_COMPARISON);
        m.insert(TokenType::Minus, PRECEDENCE_ADDITION);
        m.insert(TokenType::Plus, PRECEDENCE_ADDITION);
        m.insert(TokenType::Slash, PRECEDENCE_MULTIPLICATION);
        m.insert(TokenType::Star, PRECEDENCE_MULTIPLICATION);
        // '(' is the "operator" for function calls.
        m.insert(TokenType::ParenOpen, PRECEDENCE_CALL);
        // '[' is the "operator" for indexing.
        m.insert(TokenType::SquareOpen, PRECEDENCE_CALL);
        m
    };
}

fn is_binary_comparison_op(type_: TokenType) -> bool {
    matches!(
        type_,
        TokenType::Equals
            | TokenType::GreaterThan
            | TokenType::GreaterThanEquals
            | TokenType::LessThan
            | TokenType::LessThanEquals
            | TokenType::NotEquals
    )
}

fn token_type_to_binary_op_type(type_: TokenType) -> common::BinaryOpType {
    match type_ {
        TokenType::And => common::BinaryOpType::And,
        TokenType::Concat => common::BinaryOpType::Concat,
        TokenType::Minus => common::BinaryOpType::Subtract,
        TokenType::Or => common::BinaryOpType::Or,
        TokenType::Percent => common::BinaryOpType::Modulo,
        TokenType::Plus => common::BinaryOpType::Add,
        TokenType::Slash => common::BinaryOpType::Divide,
        TokenType::Star => common::BinaryOpType::Multiply,
        _ => {
            panic!("token type does not correspond to binary op type");
        }
    }
}

fn token_type_to_comparison_op_type(type_: TokenType) -> common::ComparisonOpType {
    match type_ {
        TokenType::Equals => common::ComparisonOpType::Equals,
        TokenType::GreaterThan => common::ComparisonOpType::GreaterThan,
        TokenType::GreaterThanEquals => common::ComparisonOpType::GreaterThanEquals,
        TokenType::LessThan => common::ComparisonOpType::LessThan,
        TokenType::LessThanEquals => common::ComparisonOpType::LessThanEquals,
        TokenType::NotEquals => common::ComparisonOpType::NotEquals,
        _ => {
            panic!("token type does not correspond to comparison op type");
        }
    }
}

fn parse_string_literal(s: &str) -> Result<String, ()> {
    // TODO
    Ok(String::from(&s[1..s.len() - 1]))
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn simple_expression() {
        let expr = parse_expression("12 + 34");
        assert_eq!(format!("{}", expr), "(binary Add 12 34)");
    }

    #[test]
    fn negative_number() {
        let expr = parse_expression("-1");
        assert_eq!(format!("{}", expr), "(unary Negate 1)");
    }

    #[test]
    fn list_index() {
        let expr = parse_expression("a + b[0]");
        assert_eq!(format!("{}", expr), "(binary Add a (index b 0))");
    }

    #[test]
    fn let_statement() {
        let stmt = parse_statement("let x: i64 = 0;");
        assert_eq!(format!("{}", stmt), "(let x (type i64) 0)");
    }

    #[test]
    fn let_statement_with_list_literal() {
        let stmt = parse_statement("let xs: list<i64> = [1, 2, 3];");
        assert_eq!(
            format!("{}", stmt),
            "(let xs (type list (type i64)) (list 1 2 3))"
        );
    }

    #[test]
    fn assign_statement() {
        let stmt = parse_statement("x = 42;");
        assert_eq!(format!("{}", stmt), "(assign x 42)");
    }

    #[test]
    fn assert_statement() {
        let stmt = parse_statement("assert false;");
        assert_eq!(format!("{}", stmt), "(assert false)");
    }

    #[test]
    fn return_statement() {
        let stmt = parse_statement("return 42;");
        assert_eq!(format!("{}", stmt), "(return 42)");
    }

    #[test]
    fn if_statement() {
        let stmt = parse_statement("if true {\n  x = 42;\n} else {\n  x = 0;\n}\n");
        assert_eq!(
            format!("{}", stmt),
            "(if true (block (assign x 42)) (else (block (assign x 0)))"
        );
    }

    #[test]
    fn if_elif_statement() {
        let stmt = parse_statement(
            r#"
if x == 0 {
    return 0;
} else if x == 1 {
    return 1;
} else {
    return recursive(x - 1);
}
"#,
        );
        assert_eq!(format!("{}", stmt), "(if (cmp Equals x 0) (block (return 0)) (elif (cmp Equals x 1) (block (return 1))) (else (block (return (call recursive ((binary Subtract x 1))))))");
    }

    #[test]
    fn precedence() {
        let expr = parse_expression("1 * 2 + 3");
        assert_eq!(format!("{}", expr), "(binary Add (binary Multiply 1 2) 3)");
    }

    #[test]
    fn function_call() {
        let expr = parse_expression("2 * f(1, 2 + x, 3)");
        assert_eq!(
            format!("{}", expr),
            "(binary Multiply 2 (call f (1 (binary Add 2 x) 3)))"
        );
    }

    #[test]
    fn equals() {
        let expr = parse_expression("n == 0");
        assert_eq!(format!("{}", expr), "(cmp Equals n 0)");
    }

    #[test]
    fn function_declaration() {
        let decl = parse_function_declaration("func inc(x: i64) -> i64 {\n  return x + 1;\n}\n");
        assert_eq!(
            format!("{}", decl),
            "(func inc ((x (type i64))) (type i64) (return (binary Add x 1)))"
        );
    }

    #[test]
    fn function_declaration_with_two_parameters() {
        let decl =
            parse_function_declaration("func add(x: i64, y: i64) -> i64 {\n  return x + y;\n}\n");
        assert_eq!(
            format!("{}", decl),
            "(func add ((x (type i64)) (y (type i64))) (type i64) (return (binary Add x y)))"
        );
    }

    fn parse_function_declaration(program: &str) -> ptree::FunctionDeclaration {
        let mut parser = Parser::new(lexer::Lexer::new("<string>", program));
        let r = parser.match_function_declaration();

        let mut message = String::new();
        for (i, error) in parser.errors.iter().enumerate() {
            message.push_str(&error.message.to_string());
            if i != parser.errors.len() - 1 {
                message.push('\n');
            }
        }

        assert!(r.is_ok(), "{}", message);
        assert!(parser.lexer.done());
        r.unwrap()
    }

    fn parse_statement(program: &str) -> ptree::Statement {
        let mut parser = Parser::new(lexer::Lexer::new("<string>", program));
        let r = parser.match_statement();

        let mut message = String::new();
        for (i, error) in parser.errors.iter().enumerate() {
            message.push_str(&error.message.to_string());
            if i != parser.errors.len() - 1 {
                message.push('\n');
            }
        }

        assert!(r.is_ok(), "{}", message);
        assert!(parser.lexer.done());
        r.unwrap()
    }

    fn parse_expression(program: &str) -> ptree::Expression {
        let mut parser = Parser::new(lexer::Lexer::new("<string>", program));
        let r = parser.match_expression();

        let mut message = String::new();
        for (i, error) in parser.errors.iter().enumerate() {
            message.push_str(&error.message.to_string());
            if i != parser.errors.len() - 1 {
                message.push('\n');
            }
        }

        assert!(r.is_ok(), "{}", message);
        assert!(
            parser.lexer.done(),
            "trailing input: {}",
            parser.lexer.token().value
        );
        r.unwrap()
    }
}
