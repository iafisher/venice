// Copyright 2022 The Venice Authors. All rights reserved.
// Use of this source code is governed by an MIT-style license that can be
// found in the LICENSE file.
//
// The Venice Intermediate Language (VIL) is the intermediate representation used inside the Venice
// compiler. It is a linear, block-based IR, broadly similar to LLVM but simpler and more closely
// tied to x86.
//
// Most VIL instructions operate on registers, of which VIL has an unlimited number. `load` and
// `store` instructions are provided to interface with memory.
//
// In its current state, VIL is not a totally coherent IR. Some operations, like moving a
// function's arguments onto the stack at the beginning of the function body, are implicit and
// handled by the x86 code generator.

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

#[derive(Clone, Debug)]
pub struct Instruction {
    pub kind: InstructionKind,
    pub comment: String,
}

pub type MemoryOffset = i32;

#[derive(Clone, Debug)]
pub enum InstructionKind {
    // Binary(op, r1, r2, r3) computes the binary operation with `r2` as its left operand and `r3`
    // as its right, and places the result in `r1`. The same register can be used multiple times in
    // a single binary instruction.
    //
    // Binary operations are represented by a single instruction kind with an `op` enum to identify
    // the desired operation, instead of separate `Add`, `Div`, etc. variants. This makes it easier
    // for the code generator to process instructions uniformly, e.g. when spilling registers.
    Binary(BinaryOp, Register, Register, Register),
    // Unary(op, r1, r2) computes the unary operation on `r2` and places the result in `r1`. `r1`
    // and `r2` may be the same register.
    Unary(UnaryOp, Register, Register),
    // Calls the function with its arguments as the given memory offsets, and places the return
    // value in the destination register.
    Call {
        destination: Register,
        label: Label,
        offsets: Vec<MemoryOffset>,
        variadic: bool,
    },
    // Compares the two registers and sets flags for a subsequent jump operation.
    Cmp(Register, Register),
    // Unconditionally jumps to the label.
    Jump(Label),
    // Jumps to the first label if the condition is true (according to the flags set  by a previous
    // `Cmp` instruction), to the second label otherwise.
    JumpIf(JumpCondition, Label, Label),
    // Loads the value at the memory offset into the register.
    Load(Register, MemoryOffset),
    // Move(r1, r2) copies the value in `r2` into `r1`.
    Move(Register, Register),
    // Sets the register to the immediate value.
    Set(Register, Immediate),
    // Stores the value in the register into memory at the given offset.
    Store(Register, MemoryOffset),
}

#[derive(Clone, Copy, Debug)]
pub enum BinaryOp {
    Add,
    Div,
    Mul,
    Sub,
}

#[derive(Clone, Copy, Debug)]
pub enum UnaryOp {
    LogicalNot,
    Negate,
}

#[derive(Clone, Copy, Debug)]
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

const RETURN_REGISTER_INDEX: u8 = 13;
// TODO(#157): These scratch register indices may conflict with general-purpose registers.
const SCRATCH_REGISTER_INDEX: u8 = RETURN_REGISTER_INDEX;
const SCRATCH2_REGISTER_INDEX: u8 = 7;

impl Register {
    pub fn index(self) -> u8 {
        self.0
    }

    // Grab a general-purpose register.
    pub fn new(i: u8) -> Self {
        Register(i)
    }

    // Grab a scratch register.
    pub fn scratch() -> Self {
        Register(SCRATCH_REGISTER_INDEX)
    }

    // Grab a second scratch register.
    pub fn scratch2() -> Self {
        Register(SCRATCH2_REGISTER_INDEX)
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
                destination,
                label,
                offsets,
                variadic,
            } => {
                write!(f, "{} = ", destination)?;
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
