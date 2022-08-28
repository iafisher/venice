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
        use Register::*;
        match self {
            Param(i) => i,
            General(i) => i,
            Return => RETURN_REGISTER_INDEX,
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
            panic!(
                "internal error: tried to use a parameter register but all {} are taken",
                PARAM_REGISTER_COUNT
            );
        }
        Register::Param(i)
    }

    pub fn gp(i: u8) -> Self {
        if i >= GP_REGISTER_COUNT {
            panic!(
                "internal error: tried to use a general-purpose register but all {} are taken",
                GP_REGISTER_COUNT
            );
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
        use InstructionKind::*;
        match self {
            Add(r1, r2, r3) => write!(f, "{} = add {}, {}", r1, r2, r3),
            Call(func) => write!(f, "call {}", func),
            CallVariadic(func) => write!(f, "call_variadic {}", func),
            CalleeSave(r) => write!(f, "callee_save {}", r),
            CalleeRestore(r) => write!(f, "{} = callee_restore", r),
            CallerSave(r) => write!(f, "caller_save {}", r),
            CallerRestore(r) => write!(f, "{} = caller_restore", r),
            Cmp(r1, r2) => write!(f, "cmp {}, {}", r1, r2),
            Div(r1, r2, r3) => write!(f, "{} = div {}, {}", r1, r2, r3),
            FrameSetUp(size) => write!(f, "frame_set_up {}", size),
            FrameTearDown(size) => write!(f, "frame_tear_down {}", size),
            Load(r, offset) => write!(f, "{} = load {}", r, offset),
            Jump(label) => write!(f, "jump {}", label),
            JumpEq(l1, l2) => write!(f, "jump_eq {} {}", l1, l2),
            JumpGt(l1, l2) => write!(f, "jump_gt {} {}", l1, l2),
            JumpGte(l1, l2) => write!(f, "jump_gte {} {}", l1, l2),
            JumpLt(l1, l2) => write!(f, "jump_lt {} {}", l1, l2),
            JumpLte(l1, l2) => write!(f, "jump_lte {} {}", l1, l2),
            JumpNeq(l1, l2) => write!(f, "jump_neq {} {}", l1, l2),
            LogicalNot(r1, r2) => write!(f, "{} = logical_not {}", r1, r2),
            Move(r1, r2) => write!(f, "{} = move {}", r1, r2),
            Mul(r1, r2, r3) => write!(f, "{} = mul {}, {}", r1, r2, r3),
            Negate(r1, r2) => write!(f, "{} = negate {}", r1, r2),
            Ret => write!(f, "ret"),
            Set(r, x) => write!(f, "{} = set {}", r, x),
            Store(r, offset) => write!(f, "store {}, {}", r, offset),
            Sub(r1, r2, r3) => write!(f, "{} = sub {}, {}", r1, r2, r3),
            ToDo(s) => write!(f, "<todo: {}>", s),
        }
    }
}

impl fmt::Display for Register {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        use Register::*;
        match self {
            Param(i) => write!(f, "%rp{}", i),
            General(i) => write!(f, "%rg{}", i),
            Return => write!(f, "%rt"),
        }
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
        write!(f, "%{}", self.0)
    }
}

impl fmt::Display for FunctionLabel {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "%{}", self.0)
    }
}
