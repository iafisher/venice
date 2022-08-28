// Copyright 2022 The Venice Authors. All rights reserved.
// Use of this source code is governed by an MIT-style license that can be
// found in the LICENSE file.
//
// The Venice Intermediate Language (VIL) is the intermediate representation used inside
// the Venice compiler. It is a high-level, machine-independent assembly language.

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

#[derive(Debug)]
pub enum InstructionKind {
    Add(Register, Register, Register),
    Call(FunctionLabel),
    CallVariadic(FunctionLabel),
    CalleeRestore(Register),
    CalleeSave(Register),
    CallerRestore(Register),
    CallerSave(Register),
    Cmp(Register, Register),
    Div(Register, Register, Register),
    FrameSetUp(usize),
    FrameTearDown(usize),
    Jump(Label),
    JumpEq(Label, Label),
    JumpGt(Label, Label),
    JumpGte(Label, Label),
    JumpLt(Label, Label),
    JumpLte(Label, Label),
    JumpNeq(Label, Label),
    Load(Register, i32),
    LogicalNot(Register, Register),
    Move(Register, Register),
    Mul(Register, Register, Register),
    Negate(Register, Register),
    Ret,
    Set(Register, Immediate),
    Store(Register, i32),
    Sub(Register, Register, Register),
    ToDo(String),
}

#[derive(Copy, Clone, Debug)]
pub enum Register {
    Param(u8),
    General(u8),
    Return,
}

// TODO: make these private
pub const PARAM_REGISTER_COUNT: u8 = 6;
pub const GP_REGISTER_COUNT: u8 = 7;
const RETURN_REGISTER_INDEX: u8 = 13;

impl Register {
    pub fn index(self) -> u8 {
        match self {
            Register::Param(i) => i,
            Register::General(i) => i,
            Register::Return => RETURN_REGISTER_INDEX,
        }
    }

    pub fn absolute_index(self) -> u8 {
        if let Register::General(i) = self {
            i + PARAM_REGISTER_COUNT
        } else {
            self.index()
        }
    }

    pub fn param(i: u8) -> Self {
        if i >= PARAM_REGISTER_COUNT {
            panic!("not enough registers");
        }
        Register::Param(i)
    }

    pub fn gp(i: u8) -> Self {
        if i >= GP_REGISTER_COUNT {
            panic!("not enough registers");
        }
        Register::General(i)
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
        match self {
            InstructionKind::Add(r1, r2, r3) => write!(f, "{} = add {}, {}", r1, r2, r3),
            InstructionKind::Call(func) => write!(f, "call {}", func),
            InstructionKind::CallVariadic(func) => write!(f, "call_variadic {}", func),
            InstructionKind::CalleeSave(r) => write!(f, "callee_save {}", r),
            InstructionKind::CalleeRestore(r) => write!(f, "{} = callee_restore", r),
            InstructionKind::CallerSave(r) => write!(f, "caller_save {}", r),
            InstructionKind::CallerRestore(r) => write!(f, "{} = caller_restore", r),
            InstructionKind::Cmp(r1, r2) => write!(f, "cmp {}, {}", r1, r2),
            InstructionKind::Div(r1, r2, r3) => write!(f, "{} = div {}, {}", r1, r2, r3),
            InstructionKind::FrameSetUp(size) => write!(f, "frame_set_up {}", size),
            InstructionKind::FrameTearDown(size) => write!(f, "frame_tear_down {}", size),
            InstructionKind::Load(r, offset) => write!(f, "{} = load {}", r, offset),
            InstructionKind::Jump(label) => write!(f, "jump {}", label),
            InstructionKind::JumpEq(l1, l2) => write!(f, "jump_eq {} {}", l1, l2),
            InstructionKind::JumpGt(l1, l2) => write!(f, "jump_gt {} {}", l1, l2),
            InstructionKind::JumpGte(l1, l2) => write!(f, "jump_gte {} {}", l1, l2),
            InstructionKind::JumpLt(l1, l2) => write!(f, "jump_lt {} {}", l1, l2),
            InstructionKind::JumpLte(l1, l2) => write!(f, "jump_lte {} {}", l1, l2),
            InstructionKind::JumpNeq(l1, l2) => write!(f, "jump_neq {} {}", l1, l2),
            InstructionKind::LogicalNot(r1, r2) => write!(f, "{} = logical_not {}", r1, r2),
            InstructionKind::Move(r1, r2) => write!(f, "{} = move {}", r1, r2),
            InstructionKind::Mul(r1, r2, r3) => write!(f, "{} = mul {}, {}", r1, r2, r3),
            InstructionKind::Negate(r1, r2) => write!(f, "{} = negate {}", r1, r2),
            InstructionKind::Ret => write!(f, "ret"),
            InstructionKind::Set(r, x) => write!(f, "{} = set {}", r, x),
            InstructionKind::Store(r, offset) => write!(f, "store {}, {}", r, offset),
            InstructionKind::Sub(r1, r2, r3) => write!(f, "{} = sub {}, {}", r1, r2, r3),
            InstructionKind::ToDo(s) => write!(f, "<todo: {}>", s),
        }
    }
}

impl fmt::Display for Register {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Register::Param(i) => write!(f, "%rp{}", i),
            Register::General(i) => write!(f, "%rg{}", i),
            Register::Return => write!(f, "%rt"),
        }
    }
}

impl fmt::Display for Immediate {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Immediate::Integer(x) => write!(f, "{}", x),
            Immediate::Label(s) => write!(f, "{}", s),
        }
    }
}

impl fmt::Display for Label {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "%{}", self.0)
    }
}

impl fmt::Display for FunctionLabel {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "%{}", self.0)
    }
}
