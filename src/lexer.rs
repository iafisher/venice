use super::common;
use std::collections::HashMap;

#[derive(Clone, Copy, Debug, Eq, PartialEq)]
pub enum TokenType {
    // Literals
    Integer,
    Symbol,
    // Operators
    Minus,
    Percent,
    Plus,
    Slash,
    Star,
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
    While,
    // Miscellaneous
    EOF,
    Unknown,
}

#[derive(Clone, Debug)]
pub struct Token {
    type_: TokenType,
    value: String,
    location: common::Location,
}

impl Token {
    pub fn new(type_: TokenType, value: &str, location: common::Location) -> Self {
        Token {
            type_: type_,
            value: String::from(value),
            location: common::Location::empty(),
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
        m.insert('-', TokenType::Minus);
        m.insert('%', TokenType::Percent);
        m.insert('+', TokenType::Plus);
        m.insert('/', TokenType::Slash);
        m.insert('*', TokenType::Star);
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
                value: String::from(""),
                location: location,
            },
        };
        // "Prime the pump" so that we can immediately call token() to retrieve the
        // first token.
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

        if self.index >= self.program.len() {
            return self.make_token(TokenType::EOF);
        }

        self.start = self.index;
        self.start_location = self.location.clone();

        let c = self.ch();
        if let Some(type_) = ONE_CHAR_TOKENS.get(&c) {
            self.increment_index();
            self.make_token(*type_)
        } else if c.is_ascii_digit() {
            self.read_number()
        } else if is_symbol_first_character(c) {
            self.read_symbol()
        } else {
            self.make_token(TokenType::Unknown)
        }
    }

    fn read_number(&mut self) -> Token {
        while self.index < self.program.len() && self.ch().is_numeric() {
            self.increment_index();
        }
        self.make_token(TokenType::Integer)
    }

    fn read_symbol(&mut self) -> Token {
        while self.index < self.program.len() && is_symbol_character(self.ch()) {
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
        while self.index < self.program.len() && self.ch().is_whitespace() {
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
        if self.index >= self.program.len() {
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
        let mut lexer = Lexer::new("<string>", "+-*/%");
        assert_eq!(lexer.token(), token(TokenType::Plus, "+"));
        assert_eq!(lexer.next(), token(TokenType::Minus, "-"));
        assert_eq!(lexer.next(), token(TokenType::Star, "*"));
        assert_eq!(lexer.next(), token(TokenType::Slash, "/"));
        assert_eq!(lexer.next(), token(TokenType::Percent, "%"));
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
            "let assert record new and or not if else while for in const func",
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
    }
}
