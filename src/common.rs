use std::fmt;

#[derive(Clone, Debug)]
pub struct Location {
    pub file: String,
    pub column: u32,
    pub line: u32,
}

impl Location {
    pub fn empty() -> Self {
        Location {
            file: String::new(),
            column: 0,
            line: 0,
        }
    }
}

impl fmt::Display for Location {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(
            f,
            "line {}, column {} of {}",
            self.line, self.column, self.file
        )
    }
}

#[derive(Clone, Copy, Debug)]
pub enum BinaryOpType {
    Add,
    And,
    Concat,
    Divide,
    Modulo,
    Multiply,
    Or,
    Subtract,
}

#[derive(Clone, Copy, Debug)]
pub enum ComparisonOpType {
    Equals,
    GreaterThan,
    GreaterThanEquals,
    LessThan,
    LessThanEquals,
    NotEquals,
}

#[derive(Clone, Copy, Debug)]
pub enum UnaryOpType {
    Negate,
    Not,
}
