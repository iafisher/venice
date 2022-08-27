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
    pub return_type: Type,
    pub blocks: Vec<Block>,
    pub max_register_count: u8,
}

pub struct ConstDeclaration {
    pub symbol: String,
    pub type_: Type,
    pub value: Immediate,
}

pub struct Block {
    pub name: String,
    pub instructions: Vec<Instruction>,
}

#[derive(Debug)]
pub enum Instruction {
    Add(Register, Register, Register),
    Alloca(Memory, u64),
    Call(FunctionLabel),
    Cmp(Register, Register),
    Div(Register, Register, Register),
    FrameSetUp(u32),
    FrameTearDown(u32),
    Jump(Label),
    JumpEq(Label, Label),
    JumpGt(Label, Label),
    JumpGte(Label, Label),
    JumpLt(Label, Label),
    JumpLte(Label, Label),
    JumpNeq(Label, Label),
    Load(Register, Memory, u64),
    Move(Register, Register),
    Mul(Register, Register, Register),
    Ret,
    Set(Register, Immediate),
    Store(Memory, Register, u64),
    Sub(Register, Register, Register),
    ToDo(String),
}

#[derive(Copy, Clone, Debug)]
pub enum Register {
    Param(u8),
    General(u8),
    Return,
    Stack,
    Base,
}

// TODO: make this private
pub const PARAM_REGISTER_COUNT: u8 = 6;
const GP_REGISTER_COUNT: u8 = 7;
const RETURN_REGISTER_INDEX: u8 = 13;
const STACK_REGISTER_INDEX: u8 = 14;
const BASE_REGISTER_INDEX: u8 = 15;

impl Register {
    pub fn index(self) -> u8 {
        match self {
            Register::Param(i) => i,
            Register::General(i) => i,
            Register::Return => RETURN_REGISTER_INDEX,
            Register::Stack => STACK_REGISTER_INDEX,
            Register::Base => BASE_REGISTER_INDEX,
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
pub struct Memory(pub String);
#[derive(Clone, Debug)]
pub struct Label(pub String);
#[derive(Clone, Debug)]
pub struct FunctionLabel(pub String);

pub enum Type {
    I64,
    Pointer(Box<Type>),
}

impl fmt::Display for Program {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        if self.externs.len() > 0 {
            for extern_ in &self.externs {
                write!(f, "extern {};\n", extern_)?;
            }
            write!(f, "\n\n")?;
        }

        for declaration in &self.declarations {
            write!(f, "{}", declaration)?;
        }

        if self.strings.len() > 0 {
            write!(f, "\n\n")?;
            for (string_name, string_value) in &self.strings {
                write!(f, "data {} = {:?};\n", string_name, string_value)?;
            }
        }

        Ok(())
    }
}

impl fmt::Display for FunctionDeclaration {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "\n")?;
        write!(f, "// max_register_count = {}\n", self.max_register_count)?;
        write!(f, "func {} -> {} {{\n", self.name, self.return_type)?;
        for block in &self.blocks {
            write!(f, "{}", block)?;
        }
        write!(f, "}}\n")
    }
}

impl fmt::Display for Block {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "\n{}:\n", self.name)?;
        for instruction in &self.instructions {
            write!(f, "{}\n", instruction)?;
        }
        fmt::Result::Ok(())
    }
}

impl fmt::Display for Instruction {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Instruction::Add(r1, r2, r3) => write!(f, "  {} = add {}, {}", r1, r2, r3),
            Instruction::Alloca(mem, size) => write!(f, "  {} = alloca {}", mem, size),
            Instruction::Call(func) => write!(f, "  call {}", func),
            Instruction::Cmp(r1, r2) => write!(f, "  cmp {}, {}", r1, r2),
            Instruction::Div(r1, r2, r3) => write!(f, "  {} = div {}, {}", r1, r2, r3),
            Instruction::FrameSetUp(size) => write!(f, "  frame_set_up {}", size),
            Instruction::FrameTearDown(size) => write!(f, "  frame_tear_down {}", size),
            Instruction::Load(r, mem, offset) => write!(f, "  {} = load {}, {}", r, mem, offset),
            Instruction::Jump(label) => write!(f, "  jump {}", label),
            Instruction::JumpEq(l1, l2) => write!(f, "  jump_eq {} {}", l1, l2),
            Instruction::JumpGt(l1, l2) => write!(f, "  jump_gt {} {}", l1, l2),
            Instruction::JumpGte(l1, l2) => write!(f, "  jump_gte {} {}", l1, l2),
            Instruction::JumpLt(l1, l2) => write!(f, "  jump_lt {} {}", l1, l2),
            Instruction::JumpLte(l1, l2) => write!(f, "  jump_lte {} {}", l1, l2),
            Instruction::JumpNeq(l1, l2) => write!(f, "  jump_neq {} {}", l1, l2),
            Instruction::Move(r1, r2) => write!(f, "  {} = move {}", r1, r2),
            Instruction::Mul(r1, r2, r3) => write!(f, "  {} = mul {}, {}", r1, r2, r3),
            Instruction::Ret => write!(f, "  ret"),
            Instruction::Set(r, x) => write!(f, "  {} = set {}", r, x),
            Instruction::Store(mem, r, offset) => write!(f, "  {} = store {}, {}", mem, r, offset),
            Instruction::Sub(r1, r2, r3) => write!(f, "  {} = sub {}, {}", r1, r2, r3),
            Instruction::ToDo(s) => write!(f, "  <todo: {}>", s),
        }
    }
}

impl fmt::Display for Register {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Register::Param(i) => write!(f, "%rp{}", i),
            Register::General(i) => write!(f, "%rg{}", i),
            Register::Return => write!(f, "%rt"),
            Register::Stack => write!(f, "%rsp"),
            Register::Base => write!(f, "%rbp"),
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

impl fmt::Display for Memory {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "%{}", self.0)
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

impl fmt::Display for Type {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Type::I64 => write!(f, "i64"),
            Type::Pointer(t) => write!(f, "ptr<{}>", t),
        }
    }
}
