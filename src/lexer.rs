use super::common;

pub enum TokenKind {
    EOF,
    Unknown,
}

pub struct Token {
    kind: TokenKind,
    value: String,
    location: common::Location,
}

pub struct Lexer {}
