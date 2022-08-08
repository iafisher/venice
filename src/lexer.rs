use super::common;

#[derive(Clone, Debug, Eq, PartialEq)]
pub enum TokenType {
    // Literals
    Integer,
    // Operators
    Plus,
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

    pub fn without_location(type_: TokenType, value: &str) -> Self {
        Token::new(type_, value, common::Location::empty())
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
        if c == '+' {
            self.increment_index();
            self.make_token(TokenType::Plus)
        } else if c.is_numeric() {
            self.read_number()
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

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn simple_expressions() {
        let mut lexer = Lexer::new("<string>", "1 + 1");
        assert_eq!(
            lexer.token(),
            Token::without_location(TokenType::Integer, "1")
        );
        assert_eq!(lexer.next(), Token::without_location(TokenType::Plus, "+"),);
        assert_eq!(
            lexer.next(),
            Token::without_location(TokenType::Integer, "1")
        );
        assert_eq!(lexer.next(), Token::without_location(TokenType::EOF, ""));

        // Make sure that multiple calls to lexer.next() at the end of the token stream
        // continue to return the EOF token.
        assert_eq!(lexer.next(), Token::without_location(TokenType::EOF, ""));
    }
}
