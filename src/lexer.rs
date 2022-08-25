use super::common;
use std::collections::HashMap;

#[derive(Clone, Copy, Debug, Hash, Eq, PartialEq)]
pub enum TokenType {
    // Literals
    Integer,
    // TODO: rename to String
    Str,
    Symbol,
    // Operators
    Assign,
    Concat,
    GreaterThan,
    GreaterThanEquals,
    Equals,
    LessThan,
    LessThanEquals,
    Minus,
    NotEquals,
    Percent,
    Plus,
    Slash,
    Star,
    // Punctuation
    Arrow,
    Colon,
    Comma,
    CurlyClose,
    CurlyOpen,
    Dot,
    ParenClose,
    ParenOpen,
    SquareClose,
    Semicolon,
    SquareOpen,
    // Keywords
    And,
    Assert,
    Const,
    Else,
    For,
    Func,
    If,
    In,
    Let,
    New,
    Not,
    Or,
    Record,
    Return,
    While,
    // Miscellaneous
    EOF,
    Unknown,
}

#[derive(Clone, Debug)]
pub struct Token {
    pub type_: TokenType,
    pub value: String,
    pub location: common::Location,
}

impl Token {
    pub fn new(type_: TokenType, value: &str, location: common::Location) -> Self {
        Token {
            type_: type_,
            value: String::from(value),
            location: location,
        }
    }
}

impl PartialEq for Token {
    fn eq(&self, other: &Self) -> bool {
        self.type_ == other.type_ && self.value == other.value
    }
}
impl Eq for Token {}

pub struct Lexer {
    program: String,
    index: usize,
    start: usize,
    location: common::Location,
    start_location: common::Location,
    token: Token,
}

lazy_static! {
    static ref ONE_CHAR_TOKENS: HashMap<char, TokenType> = {
        let mut m = HashMap::new();
        m.insert('=', TokenType::Assign);
        m.insert(':', TokenType::Colon);
        m.insert(',', TokenType::Comma);
        m.insert('}', TokenType::CurlyClose);
        m.insert('{', TokenType::CurlyOpen);
        m.insert('.', TokenType::Dot);
        m.insert('>', TokenType::GreaterThan);
        m.insert('<', TokenType::LessThan);
        m.insert('-', TokenType::Minus);
        m.insert(')', TokenType::ParenClose);
        m.insert('(', TokenType::ParenOpen);
        m.insert('%', TokenType::Percent);
        m.insert('+', TokenType::Plus);
        m.insert(';', TokenType::Semicolon);
        m.insert('/', TokenType::Slash);
        m.insert(']', TokenType::SquareClose);
        m.insert('[', TokenType::SquareOpen);
        m.insert('*', TokenType::Star);
        m
    };
    static ref TWO_CHAR_TOKENS: HashMap<(char, char), TokenType> = {
        let mut m = HashMap::new();
        m.insert(('-', '>'), TokenType::Arrow);
        m.insert(('=', '='), TokenType::Equals);
        m.insert(('>', '='), TokenType::GreaterThanEquals);
        m.insert(('<', '='), TokenType::LessThanEquals);
        m.insert(('!', '='), TokenType::NotEquals);
        m.insert(('+', '+'), TokenType::Concat);
        m
    };
    static ref KEYWORDS: HashMap<&'static str, TokenType> = {
        let mut m = HashMap::new();
        m.insert("and", TokenType::And);
        m.insert("assert", TokenType::Assert);
        m.insert("const", TokenType::Const);
        m.insert("else", TokenType::Else);
        m.insert("for", TokenType::For);
        m.insert("func", TokenType::Func);
        m.insert("if", TokenType::If);
        m.insert("in", TokenType::In);
        m.insert("let", TokenType::Let);
        m.insert("new", TokenType::New);
        m.insert("not", TokenType::Not);
        m.insert("or", TokenType::Or);
        m.insert("record", TokenType::Record);
        m.insert("return", TokenType::Return);
        m.insert("while", TokenType::While);
        m
    };
}

impl Lexer {
    /// Constructs a new lexer.
    pub fn new(file: &str, program: &str) -> Self {
        let location = common::Location {
            file: String::from(file),
            column: 1,
            line: 1,
        };
        let mut lexer = Lexer {
            program: String::from(program),
            index: 0,
            start: 0,
            location: location.clone(),
            start_location: location.clone(),
            token: Token {
                type_: TokenType::Unknown,
                value: String::new(),
                location: location,
            },
        };
        // "Prime the pump" so that we can immediately call token() to retrieve the
        // first token.
        //
        // TODO: remove this?
        lexer.next();
        lexer
    }

    /// Returns the current token without advancing.
    pub fn token(&self) -> Token {
        self.token.clone()
    }

    /// Advances to the next token and returns it.
    pub fn next(&mut self) -> Token {
        self.skip_whitespace();

        if self.done() {
            return self.make_token(TokenType::EOF);
        }

        self.start = self.index;
        self.start_location = self.location.clone();

        let c = self.ch();

        if self.index + 1 < self.program.len() {
            let c2 = self.peek(1);
            if let Some(type_) = TWO_CHAR_TOKENS.get(&(c, c2)) {
                self.increment_index();
                self.increment_index();
                return self.make_token(*type_);
            }
        }

        if let Some(type_) = ONE_CHAR_TOKENS.get(&c) {
            self.increment_index();
            self.make_token(*type_)
        } else if c.is_ascii_digit() {
            self.read_number()
        } else if c == '"' {
            self.read_string()
        } else if is_symbol_first_character(c) {
            self.read_symbol()
        } else {
            self.make_token(TokenType::Unknown)
        }
    }

    pub fn done(&self) -> bool {
        self.index >= self.program.len()
    }

    fn read_number(&mut self) -> Token {
        while !self.done() && self.ch().is_numeric() {
            self.increment_index();
        }
        self.make_token(TokenType::Integer)
    }

    fn read_string(&mut self) -> Token {
        // Move past the opening quotation mark.
        self.increment_index();
        while !self.done() {
            let c = self.ch();
            if c == '"' {
                self.increment_index();
                break;
            } else if c == '\\' {
                // TODO: what if backslash is last character in program?
                self.increment_index();
                self.increment_index();
            } else {
                self.increment_index();
            }
        }

        // TODO: handle unclosed string literals (newlines and EOF)
        self.make_token(TokenType::Str)
    }

    fn read_symbol(&mut self) -> Token {
        while !self.done() && is_symbol_character(self.ch()) {
            self.increment_index()
        }
        let value = &self.program[self.start..self.index];
        if let Some(type_) = KEYWORDS.get(value) {
            self.make_token(*type_)
        } else {
            self.make_token(TokenType::Symbol)
        }
    }

    fn skip_whitespace(&mut self) {
        while !self.done() && self.ch().is_whitespace() {
            self.increment_index();
        }
    }

    fn make_token(&mut self, type_: TokenType) -> Token {
        let token = Token::new(
            type_,
            &self.program[self.start..self.index],
            self.start_location.clone(),
        );
        self.start = self.index;
        self.start_location = self.location.clone();
        self.token = token;
        self.token.clone()
    }

    fn increment_index(&mut self) {
        if self.done() {
            return;
        }

        if self.ch() == '\n' {
            self.location.column = 1;
            self.location.line += 1;
        } else {
            self.location.column += 1;
        }
        self.index += 1;
    }

    fn ch(&self) -> char {
        // TODO: more efficient way to do this?
        self.program.chars().nth(self.index).unwrap()
    }

    fn peek(&self, n: usize) -> char {
        self.program.chars().nth(self.index + n).unwrap()
    }
}

fn is_symbol_first_character(c: char) -> bool {
    c.is_ascii_alphabetic() || c == '_'
}

fn is_symbol_character(c: char) -> bool {
    c.is_ascii_alphanumeric() || c == '_'
}

#[cfg(test)]
mod tests {
    use super::*;

    fn token(type_: TokenType, value: &str) -> Token {
        Token::new(type_, value, common::Location::empty())
    }

    #[test]
    fn simple_expressions() {
        let mut lexer = Lexer::new("<string>", "1 + 1");
        assert_eq!(lexer.token(), token(TokenType::Integer, "1"));
        assert_eq!(lexer.next(), token(TokenType::Plus, "+"),);
        assert_eq!(lexer.next(), token(TokenType::Integer, "1"));
        assert_eq!(lexer.next(), token(TokenType::EOF, ""));

        // Make sure that multiple calls to lexer.next() at the end of the token stream
        // continue to return the EOF token.
        assert_eq!(lexer.next(), token(TokenType::EOF, ""));
    }

    #[test]
    fn operators() {
        let mut lexer = Lexer::new("<string>", "+-*/%++ < > != == <= >= =");
        assert_eq!(lexer.token(), token(TokenType::Plus, "+"));
        assert_eq!(lexer.next(), token(TokenType::Minus, "-"));
        assert_eq!(lexer.next(), token(TokenType::Star, "*"));
        assert_eq!(lexer.next(), token(TokenType::Slash, "/"));
        assert_eq!(lexer.next(), token(TokenType::Percent, "%"));
        assert_eq!(lexer.next(), token(TokenType::Concat, "++"));
        assert_eq!(lexer.next(), token(TokenType::LessThan, "<"));
        assert_eq!(lexer.next(), token(TokenType::GreaterThan, ">"));
        assert_eq!(lexer.next(), token(TokenType::NotEquals, "!="));
        assert_eq!(lexer.next(), token(TokenType::Equals, "=="));
        assert_eq!(lexer.next(), token(TokenType::LessThanEquals, "<="));
        assert_eq!(lexer.next(), token(TokenType::GreaterThanEquals, ">="));
        assert_eq!(lexer.next(), token(TokenType::Assign, "="));
        assert_eq!(lexer.next(), token(TokenType::EOF, ""));
    }

    #[test]
    fn symbols() {
        let mut lexer = Lexer::new("<string>", "_ abc0 lorem_ipsum");
        assert_eq!(lexer.token(), token(TokenType::Symbol, "_"));
        assert_eq!(lexer.next(), token(TokenType::Symbol, "abc0"));
        assert_eq!(lexer.next(), token(TokenType::Symbol, "lorem_ipsum"));
    }

    #[test]
    fn keywords() {
        let mut lexer = Lexer::new(
            "<string>",
            "let assert record new and or not if else while for in const func return",
        );
        assert_eq!(lexer.token(), token(TokenType::Let, "let"));
        assert_eq!(lexer.next(), token(TokenType::Assert, "assert"));
        assert_eq!(lexer.next(), token(TokenType::Record, "record"));
        assert_eq!(lexer.next(), token(TokenType::New, "new"));
        assert_eq!(lexer.next(), token(TokenType::And, "and"));
        assert_eq!(lexer.next(), token(TokenType::Or, "or"));
        assert_eq!(lexer.next(), token(TokenType::Not, "not"));
        assert_eq!(lexer.next(), token(TokenType::If, "if"));
        assert_eq!(lexer.next(), token(TokenType::Else, "else"));
        assert_eq!(lexer.next(), token(TokenType::While, "while"));
        assert_eq!(lexer.next(), token(TokenType::For, "for"));
        assert_eq!(lexer.next(), token(TokenType::In, "in"));
        assert_eq!(lexer.next(), token(TokenType::Const, "const"));
        assert_eq!(lexer.next(), token(TokenType::Func, "func"));
        assert_eq!(lexer.next(), token(TokenType::Return, "return"));
    }

    #[test]
    fn punctuation() {
        let mut lexer = Lexer::new("<string>", ".,()[]{}->:;");
        assert_eq!(lexer.token(), token(TokenType::Dot, "."));
        assert_eq!(lexer.next(), token(TokenType::Comma, ","));
        assert_eq!(lexer.next(), token(TokenType::ParenOpen, "("));
        assert_eq!(lexer.next(), token(TokenType::ParenClose, ")"));
        assert_eq!(lexer.next(), token(TokenType::SquareOpen, "["));
        assert_eq!(lexer.next(), token(TokenType::SquareClose, "]"));
        assert_eq!(lexer.next(), token(TokenType::CurlyOpen, "{"));
        assert_eq!(lexer.next(), token(TokenType::CurlyClose, "}"));
        assert_eq!(lexer.next(), token(TokenType::Arrow, "->"));
        assert_eq!(lexer.next(), token(TokenType::Colon, ":"));
        assert_eq!(lexer.next(), token(TokenType::Semicolon, ";"));
    }

    #[test]
    fn simple_string_literal() {
        let lexer = Lexer::new("<string>", "\"abc\"");
        assert_eq!(lexer.token(), token(TokenType::Str, "\"abc\""));
    }

    #[test]
    fn string_literal_with_backslash() {
        // A two-character string literal: a backslash followed by a double quote
        let lexer = Lexer::new("<string>", r#""\"""#);
        assert_eq!(lexer.token(), token(TokenType::Str, r#""\"""#));
    }
}
