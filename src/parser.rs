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
            name: name,
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
            TokenType::Symbol => self
                .match_assign_statement()
                .map(|stmt| ast::Statement::Assign(stmt)),
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
                self.unexpected(&token, "start of statement");
                Err(())
            }
        }
    }

    fn match_assign_statement(&mut self) -> Result<ast::AssignStatement, ()> {
        let mut token = self.lexer.token();
        let location = token.location.clone();
        self.expect_token(&token, TokenType::Symbol, "symbol")?;
        let symbol = token.value;

        token = self.lexer.next();
        self.expect_token(&token, TokenType::Assign, "=")?;

        self.lexer.next();
        let value = self.match_expression()?;

        token = self.lexer.token();
        self.expect_token(&token, TokenType::Semicolon, ";")?;
        self.lexer.next();

        Ok(ast::AssignStatement {
            symbol: symbol,
            value: value,
            location: location,
        })
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
            symbol: symbol,
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
        let mut token = self.lexer.token();
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
                    token = self.lexer.token();
                    continue;
                }
            }
            break;
        }
        Ok(expr)
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
            TokenType::Str => {
                self.lexer.next();
                if let Ok(s) = parse_string_literal(&token.value) {
                    Ok(ast::Expression {
                        kind: ast::ExpressionKind::Str(s),
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
                    kind: ast::ExpressionKind::Symbol(token.value),
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
}

const PRECEDENCE_LOWEST: u32 = 0;
const PRECEDENCE_COMPARISON: u32 = 1;
const PRECEDENCE_ADDITION: u32 = 2;
const PRECEDENCE_MULTIPLICATION: u32 = 3;
const PRECEDENCE_CALL: u32 = 4;

lazy_static! {
    static ref PRECEDENCE: HashMap<TokenType, u32> = {
        let mut m = HashMap::new();
        m.insert(TokenType::LessThan, PRECEDENCE_COMPARISON);
        m.insert(TokenType::Plus, PRECEDENCE_ADDITION);
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
