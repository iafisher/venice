// Copyright 2022 The Venice Authors. All rights reserved.
// Use of this source code is governed by an MIT-style license that can be
// found in the LICENSE file.

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

// The parse tree and abstract syntax tree use the same op types, so they are defined here.

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
