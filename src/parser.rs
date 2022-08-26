// The parser transforms the stream of tokens emitted by the lexer into an abstract
// syntax tree.

use super::ast;
use super::common;
use super::errors;
use super::lexer;
use super::lexer::TokenType;
use std::collections::HashMap;

pub fn parse(lexer: lexer::Lexer) -> Result<ast::Program, Vec<errors::VeniceError>> {
    let mut parser = Parser::new(lexer);
    let ast = parser.parse();
    if parser.errors.len() > 0 {
        Err(parser.errors.clone())
    } else {
        Ok(ast)
    }
}

struct Parser {
    lexer: lexer::Lexer,
    errors: Vec<errors::VeniceError>,
}

impl Parser {
    fn new(lexer: lexer::Lexer) -> Self {
        Parser {
            lexer: lexer,
            errors: Vec::new(),
        }
    }

    fn parse(&mut self) -> ast::Program {
        let mut declarations = Vec::new();
        while !self.lexer.done() {
            if let Ok(declaration) = self.match_declaration() {
                declarations.push(declaration);
            }
        }
        ast::Program {
            declarations: declarations,
        }
    }

    fn match_declaration(&mut self) -> Result<ast::Declaration, ()> {
        let token = self.lexer.token();
        match token.type_ {
            TokenType::Func => self
                .match_function_declaration()
                .map(|d| ast::Declaration::Function(d)),
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

    fn match_function_declaration(&mut self) -> Result<ast::FunctionDeclaration, ()> {
        let location = self.lexer.token().location.clone();
        self.expect_token(&self.lexer.token(), TokenType::Func, "func keyword")?;

        let mut token = self.lexer.next();
        self.expect_token(&token, TokenType::Symbol, "function name")?;
        let name = token.value;

        token = self.lexer.next();
        self.expect_token(&token, TokenType::ParenOpen, "(")?;

        let mut parameters = Vec::new();
        loop {
            token = self.lexer.next();
            if token.type_ == TokenType::ParenClose {
                break;
            }

            self.expect_token(&token, TokenType::Symbol, "parameter name")?;
            let parameter_name = token.value;

            token = self.lexer.next();
            self.expect_token(&token, TokenType::Colon, ":")?;

            self.lexer.next();
            let type_ = self.match_type()?;
            parameters.push(ast::FunctionParameter {
                name: parameter_name,
                type_: type_,
                semantic_type: ast::Type::Unknown,
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
        Ok(ast::FunctionDeclaration {
            name: ast::SymbolExpression {
                name: name.clone(),
                entry: None,
                location: location.clone(),
            },
            parameters: parameters,
            return_type: return_type,
            semantic_return_type: ast::Type::Unknown,
            body: body,
            location: location,
        })
    }

    fn match_block(&mut self) -> Result<Vec<ast::Statement>, ()> {
        let mut token = self.lexer.token();
        self.expect_token(&token, TokenType::CurlyOpen, "{")?;
        self.lexer.next();

        let mut statements = Vec::new();
        loop {
            token = self.lexer.token();
            if token.type_ == TokenType::CurlyClose {
                self.lexer.next();
                break;
            } else if token.type_ == TokenType::EOF {
                self.unexpected(&token, "statement or end of block");
                return Err(());
            }

            if let Ok(statement) = self.match_statement() {
                statements.push(statement);
            }
        }
        Ok(statements)
    }

    fn match_statement(&mut self) -> Result<ast::Statement, ()> {
        let mut token = self.lexer.token();
        match token.type_ {
            TokenType::Assert => self
                .match_assert_statement()
                .map(|stmt| ast::Statement::Assert(stmt)),
            TokenType::If => self
                .match_if_statement()
                .map(|stmt| ast::Statement::If(stmt)),
            TokenType::Let => self
                .match_let_statement()
                .map(|stmt| ast::Statement::Let(stmt)),
            TokenType::Return => self
                .match_return_statement()
                .map(|stmt| ast::Statement::Return(stmt)),
            TokenType::While => self
                .match_while_statement()
                .map(|stmt| ast::Statement::While(stmt)),
            _ => {
                let expr = self.match_expression()?;
                token = self.lexer.token();
                if token.type_ == TokenType::Assign {
                    self.match_assign_statement(expr)
                        .map(|stmt| ast::Statement::Assign(stmt))
                } else if token.type_ == TokenType::Semicolon {
                    self.lexer.next();
                    Ok(ast::Statement::Expression(expr))
                } else {
                    self.lexer.next();
                    self.unexpected(&token, "start of statement");
                    Err(())
                }
            }
        }
    }

    fn match_assert_statement(&mut self) -> Result<ast::AssertStatement, ()> {
        let mut token = self.lexer.token();
        let location = token.location.clone();
        self.expect_token(&token, TokenType::Assert, "assert")?;

        self.lexer.next();
        let condition = self.match_expression()?;

        token = self.lexer.token();
        self.expect_token(&token, TokenType::Semicolon, ";")?;
        self.lexer.next();

        Ok(ast::AssertStatement {
            condition: condition,
            location: location,
        })
    }

    fn match_assign_statement(
        &mut self,
        expr: ast::Expression,
    ) -> Result<ast::AssignStatement, ()> {
        let symbol = if let ast::ExpressionKind::Symbol(symbol_expr) = expr.kind {
            symbol_expr.name
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

        Ok(ast::AssignStatement {
            symbol: ast::SymbolExpression {
                name: symbol,
                entry: None,
                location: location.clone(),
            },
            value: value,
            location: location,
        })
    }

    fn match_if_statement(&mut self) -> Result<ast::IfStatement, ()> {
        let token = self.lexer.token();
        let location = token.location.clone();
        self.expect_token(&token, TokenType::If, "if")?;

        self.lexer.next();
        let condition = self.match_expression()?;
        let body = self.match_block()?;

        if self.lexer.token().type_ == TokenType::Else {
            self.lexer.next();
            // TODO: handle else-ifs
            let else_body = self.match_block()?;
            Ok(ast::IfStatement {
                if_clause: ast::IfClause {
                    condition: condition,
                    body: body,
                },
                elif_clauses: Vec::new(),
                else_clause: else_body,
                location: location,
            })
        } else {
            Ok(ast::IfStatement {
                if_clause: ast::IfClause {
                    condition: condition,
                    body: body,
                },
                elif_clauses: Vec::new(),
                else_clause: Vec::new(),
                location: location,
            })
        }
    }

    fn match_let_statement(&mut self) -> Result<ast::LetStatement, ()> {
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

        Ok(ast::LetStatement {
            symbol: ast::SymbolExpression {
                name: symbol,
                entry: None,
                location: location.clone(),
            },
            type_: type_,
            semantic_type: ast::Type::Unknown,
            value: value,
            location: location,
        })
    }

    fn match_return_statement(&mut self) -> Result<ast::ReturnStatement, ()> {
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

        Ok(ast::ReturnStatement {
            value: value,
            location: location,
        })
    }

    fn match_while_statement(&mut self) -> Result<ast::WhileStatement, ()> {
        let token = self.lexer.token();
        let location = token.location.clone();
        self.expect_token(&token, TokenType::While, "while")?;

        self.lexer.next();
        let condition = self.match_expression()?;
        let body = self.match_block()?;
        Ok(ast::WhileStatement {
            condition: condition,
            body: body,
            location: location,
        })
    }

    fn match_expression(&mut self) -> Result<ast::Expression, ()> {
        self.match_expression_with_precedence(PRECEDENCE_LOWEST)
    }

    fn match_expression_with_precedence(&mut self, precedence: u32) -> Result<ast::Expression, ()> {
        let mut expr = self.match_literal()?;

        let mut token = self.lexer.token();
        loop {
            if let Some(other_precedence) = PRECEDENCE.get(&token.type_) {
                if precedence < *other_precedence {
                    if token.type_ == TokenType::ParenOpen {
                        self.lexer.next();
                        let call = self.match_function_call(&expr, token.location.clone())?;
                        expr = ast::Expression {
                            kind: ast::ExpressionKind::Call(call),
                            semantic_type: ast::Type::Unknown,
                            location: token.location.clone(),
                        };
                    } else {
                        self.lexer.next();
                        let right = self.match_expression_with_precedence(*other_precedence)?;
                        expr = ast::Expression {
                            kind: ast::ExpressionKind::Binary(ast::BinaryExpression {
                                op: token_type_to_binary_op_type(token.type_),
                                left: Box::new(expr),
                                right: Box::new(right),
                                location: token.location.clone(),
                            }),
                            semantic_type: ast::Type::Unknown,
                            location: token.location.clone(),
                        };
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
        expr: &ast::Expression,
        location: common::Location,
    ) -> Result<ast::CallExpression, ()> {
        if let ast::ExpressionKind::Symbol(ast::SymbolExpression { name, .. }) = &expr.kind {
            let arguments = self.match_expression_list()?;
            let token = self.lexer.token();
            self.expect_token(&token, TokenType::ParenClose, ")")?;
            self.lexer.next();
            Ok(ast::CallExpression {
                function: ast::SymbolExpression {
                    name: name.clone(),
                    entry: None,
                    location: location.clone(),
                },
                arguments: arguments,
                location: location,
            })
        } else {
            self.error("function must be a symbol", expr.location.clone());
            Err(())
        }
    }

    fn match_expression_list(&mut self) -> Result<Vec<ast::Expression>, ()> {
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

    fn match_literal(&mut self) -> Result<ast::Expression, ()> {
        let token = self.lexer.token();
        match token.type_ {
            TokenType::Integer => {
                self.lexer.next();
                if let Ok(x) = i64::from_str_radix(&token.value, 10) {
                    Ok(ast::Expression {
                        kind: ast::ExpressionKind::Integer(x),
                        semantic_type: ast::Type::Unknown,
                        location: token.location.clone(),
                    })
                } else {
                    self.error("could not parse integer literal", token.location.clone());
                    Err(())
                }
            }
            TokenType::True => {
                self.lexer.next();
                Ok(ast::Expression {
                    kind: ast::ExpressionKind::Boolean(true),
                    semantic_type: ast::Type::Unknown,
                    location: token.location.clone(),
                })
            }
            TokenType::False => {
                self.lexer.next();
                Ok(ast::Expression {
                    kind: ast::ExpressionKind::Boolean(false),
                    semantic_type: ast::Type::Unknown,
                    location: token.location.clone(),
                })
            }
            TokenType::String => {
                self.lexer.next();
                if let Ok(s) = parse_string_literal(&token.value) {
                    Ok(ast::Expression {
                        kind: ast::ExpressionKind::String(s),
                        semantic_type: ast::Type::Unknown,
                        location: token.location.clone(),
                    })
                } else {
                    self.error("could not parse string literal", token.location.clone());
                    Err(())
                }
            }
            TokenType::Symbol => {
                self.lexer.next();
                Ok(ast::Expression {
                    kind: ast::ExpressionKind::Symbol(ast::SymbolExpression {
                        name: token.value,
                        entry: None,
                        location: token.location.clone(),
                    }),
                    semantic_type: ast::Type::Unknown,
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
            _ => {
                self.unexpected(&token, "start of expression");
                self.lexer.next();
                Err(())
            }
        }
    }

    fn match_type(&mut self) -> Result<ast::SyntacticType, ()> {
        let token = self.lexer.token();
        self.expect_token(&token, TokenType::Symbol, "type")?;
        self.lexer.next();
        // TODO: handle parameterized types
        Ok(ast::SyntacticType {
            kind: ast::SyntacticTypeKind::Literal(token.value),
            location: token.location.clone(),
        })
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
        let msg = if token.type_ == TokenType::EOF {
            format!("expected {}, got end of file", message)
        } else {
            format!("expected {}, got {}", message, token.value)
        };

        self.error(&msg, token.location.clone());
    }

    fn error(&mut self, message: &str, location: common::Location) {
        self.errors
            .push(errors::VeniceError::new(&message, location));
    }

    fn skip_past(&mut self, token_type: lexer::TokenType) {
        self.lexer.next();
        loop {
            let token = self.lexer.token();
            if token.type_ == TokenType::EOF {
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
        m.insert(TokenType::Minus, PRECEDENCE_ADDITION);
        m.insert(TokenType::Plus, PRECEDENCE_ADDITION);
        m.insert(TokenType::Slash, PRECEDENCE_MULTIPLICATION);
        m.insert(TokenType::Star, PRECEDENCE_MULTIPLICATION);
        // '(' is the "operator" for function calls.
        m.insert(TokenType::ParenOpen, PRECEDENCE_CALL);
        m
    };
}

fn token_type_to_binary_op_type(type_: TokenType) -> ast::BinaryOpType {
    match type_ {
        TokenType::And => ast::BinaryOpType::And,
        TokenType::Concat => ast::BinaryOpType::Concat,
        TokenType::Equals => ast::BinaryOpType::Equals,
        TokenType::GreaterThan => ast::BinaryOpType::GreaterThan,
        TokenType::GreaterThanEquals => ast::BinaryOpType::GreaterThanEquals,
        TokenType::LessThan => ast::BinaryOpType::LessThan,
        TokenType::LessThanEquals => ast::BinaryOpType::LessThanEquals,
        TokenType::Minus => ast::BinaryOpType::Subtract,
        TokenType::NotEquals => ast::BinaryOpType::NotEquals,
        TokenType::Or => ast::BinaryOpType::Or,
        TokenType::Percent => ast::BinaryOpType::Modulo,
        TokenType::Plus => ast::BinaryOpType::Add,
        TokenType::Slash => ast::BinaryOpType::Divide,
        TokenType::Star => ast::BinaryOpType::Multiply,
        _ => {
            panic!("token type does not correspond to binary op type");
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
    fn let_statement() {
        let stmt = parse_statement("let x: i64 = 0;");
        assert_eq!(format!("{}", stmt), "(let x (type i64) 0)");
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

    fn parse_statement(program: &str) -> ast::Statement {
        let mut parser = Parser::new(lexer::Lexer::new("<string>", &program));
        let r = parser.match_statement();

        let mut message = String::new();
        for (i, error) in parser.errors.iter().enumerate() {
            message.push_str(&format!("{}", error.message));
            if i != parser.errors.len() - 1 {
                message.push('\n');
            }
        }

        assert!(r.is_ok(), "{}", message);
        assert!(parser.lexer.done());
        r.unwrap()
    }

    fn parse_expression(program: &str) -> ast::Expression {
        let mut parser = Parser::new(lexer::Lexer::new("<string>", &program));
        let r = parser.match_expression();

        let mut message = String::new();
        for (i, error) in parser.errors.iter().enumerate() {
            message.push_str(&format!("{}", error.message));
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
