use super::common;

pub enum TokenType {
    EOF,
    Unknown,
}

pub struct Token {
    type_: TokenType,
    value: String,
    location: common::Location,
}

pub struct Lexer {}
