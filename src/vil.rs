// Copyright 2022 The Venice Authors. All rights reserved.
// Use of this source code is governed by an MIT-style license that can be
// found in the LICENSE file.
//
// The Venice Intermediate Language (VIL) is the intermediate representation used inside the Venice
// compiler. See docs/vil.md for details.

use std::collections::BTreeMap;
use std::fmt;

pub struct Program {
    pub externs: Vec<String>,
    pub declarations: Vec<FunctionDeclaration>,
    // BTreeMap so that the output is sorted.
    pub strings: BTreeMap<String, String>,
}

pub struct FunctionDeclaration {
    pub name: String,
    pub blocks: Vec<Block>,
    pub stack_frame_size: i32,
    pub parameters: Vec<FunctionParameter>,
}

pub struct FunctionParameter {
    pub stack_offset: i32,
}

pub struct Block {
    pub name: String,
    pub instructions: Vec<Instruction>,
}

#[derive(Debug)]
pub struct Instruction {
    pub kind: InstructionKind,
    pub comment: String,
}

type MemoryOffset = i32;

#[derive(Debug)]
pub enum InstructionKind {
    Binary(BinaryOp, Register, Register, Register),
    Unary(UnaryOp, Register, Register),
    Call {
        label: FunctionLabel,
        offsets: Vec<MemoryOffset>,
        variadic: bool,
    },
    Cmp(Register, Register),
    Jump(Label),
    JumpIf(JumpCondition, Label, Label),
    Load(Register, MemoryOffset),
    Move(Register, Register),
    Set(Register, Immediate),
    Store(Register, MemoryOffset),
}

#[derive(Debug)]
pub enum BinaryOp {
    Add,
    Div,
    Mul,
    Sub,
}

#[derive(Debug)]
pub enum UnaryOp {
    LogicalNot,
    Negate,
}

#[derive(Debug)]
pub enum JumpCondition {
    Eq,
    Gt,
    Gte,
    Lt,
    Lte,
    Neq,
}

#[derive(Clone, Copy, Debug)]
pub struct Register(u8);

const GP_REGISTER_COUNT: u8 = 7;
const RETURN_REGISTER_INDEX: u8 = 13;

impl Register {
    pub fn index(self) -> u8 {
        self.0
    }

    pub fn new(i: u8) -> Self {
        if i >= GP_REGISTER_COUNT {
            panic!(
                "internal error: tried to use a general-purpose register but all {} are taken",
                GP_REGISTER_COUNT
            );
        }
        Register(i)
    }

    pub fn scratch() -> Self {
        Register(RETURN_REGISTER_INDEX)
    }

    pub fn ret() -> Self {
        Register(RETURN_REGISTER_INDEX)
    }
}

#[derive(Clone, Debug)]
pub enum Immediate {
    Integer(i64),
    Label(String),
}
#[derive(Clone, Debug)]
pub struct Label(pub String);
#[derive(Clone, Debug)]
pub struct FunctionLabel(pub String);

impl fmt::Display for Program {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        if !self.externs.is_empty() {
            for extern_ in &self.externs {
                writeln!(f, "extern {};", extern_)?;
            }
            write!(f, "\n\n")?;
        }

        for declaration in &self.declarations {
            write!(f, "{}", declaration)?;
        }

        if !self.strings.is_empty() {
            write!(f, "\n\n")?;
            for (string_name, string_value) in &self.strings {
                writeln!(f, "data {} = {:?};", string_name, string_value)?;
            }
        }

        Ok(())
    }
}

impl fmt::Display for FunctionDeclaration {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        writeln!(f)?;
        writeln!(f, "func {} {{", self.name)?;
        writeln!(f, "  // stack_frame_size = {}", self.stack_frame_size)?;
        for block in &self.blocks {
            write!(f, "{}", block)?;
        }
        writeln!(f, "}}")
    }
}

impl fmt::Display for Block {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "\n{}:\n", self.name)?;
        for instruction in &self.instructions {
            writeln!(f, "  {}", instruction)?;
        }
        fmt::Result::Ok(())
    }
}

impl fmt::Display for Instruction {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "{}", self.kind)?;
        if !self.comment.is_empty() {
            write!(f, "  // {}", self.comment)?;
        }
        fmt::Result::Ok(())
    }
}

impl fmt::Display for InstructionKind {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        use InstructionKind::*;
        match self {
            Binary(op, r1, r2, r3) => {
                let opstr = match op {
                    BinaryOp::Add => "add",
                    BinaryOp::Div => "div",
                    BinaryOp::Mul => "mul",
                    BinaryOp::Sub => "sub",
                };
                write!(f, "{} = {} {}, {}", r1, opstr, r2, r3)
            }
            Unary(UnaryOp::LogicalNot, r1, r2) => write!(f, "{} = logical_not {}", r1, r2),
            Unary(UnaryOp::Negate, r1, r2) => write!(f, "{} = negate {}", r1, r2),
            Call {
                label,
                offsets,
                variadic,
            } => {
                if *variadic {
                    write!(f, "call_variadic {}", label)?;
                } else {
                    write!(f, "call {}", label)?;
                }

                for offset in offsets {
                    write!(f, ", mem[{}]", offset)?;
                }
                fmt::Result::Ok(())
            }
            Cmp(r1, r2) => write!(f, "cmp {}, {}", r1, r2),
            Load(r, offset) => write!(f, "{} = load {}", r, offset),
            Jump(label) => write!(f, "jump {}", label),
            JumpIf(cond, l1, l2) => {
                let suffix = match cond {
                    JumpCondition::Eq => "eq",
                    JumpCondition::Gt => "gt",
                    JumpCondition::Gte => "gte",
                    JumpCondition::Lt => "lt",
                    JumpCondition::Lte => "gte",
                    JumpCondition::Neq => "neq",
                };
                write!(f, "jump_{} {}, {}", suffix, l1, l2)
            }
            Move(r1, r2) => write!(f, "{} = move {}", r1, r2),
            Set(r, x) => write!(f, "{} = set {}", r, x),
            Store(r, offset) => write!(f, "store {}, {}", r, offset),
        }
    }
}

impl fmt::Display for Register {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "R{}", self.0)
    }
}

impl fmt::Display for Immediate {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        use Immediate::*;
        match self {
            Integer(x) => write!(f, "{}", x),
            Label(s) => write!(f, "{}", s),
        }
    }
}

impl fmt::Display for Label {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "{}", self.0)
    }
}

impl fmt::Display for FunctionLabel {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "{}", self.0)
    }
}
