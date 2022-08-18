use std::fmt;

pub struct Program {
    pub declarations: Vec<FunctionDeclaration>,
}

pub struct FunctionDeclaration {
    pub name: String,
    pub parameters: Vec<FunctionParameter>,
    pub return_type: Type,
    pub blocks: Vec<Block>,
}

pub struct FunctionParameter {
    pub name: String,
    pub type_: Type,
}

pub struct ConstDeclaration {
    pub symbol: String,
    pub type_: Type,
    pub value: Immediate,
}

pub struct Block {
    pub name: String,
    pub instructions: Vec<Instruction>,
    pub exit: ExitInstruction,
}

pub enum Instruction {
    Set(Register, Immediate),
    Move(Register, Register),
    Alloca(Memory, u64),
    Load(Register, Memory, u64),
    Store(Memory, Register, u64),
    Add(Register, Register, Register),
    Sub(Register, Register, Register),
    Mul(Register, Register, Register),
    Div(Register, Register, Register),
    Call(Register, FunctionLabel, Vec<Register>),
    CmpEq(Register, Register),
    CmpNeq(Register, Register),
    CmpLt(Register, Register),
    CmpLte(Register, Register),
    CmpGt(Register, Register),
    CmpGte(Register, Register),
    SetCmp(Register),
}

pub enum ExitInstruction {
    Ret(Register),
    Jump(Label),
    JumpIf(Label, Label),
    Placeholder,
}

#[derive(Clone)]
pub struct Register(pub u32);
#[derive(Clone)]
pub struct Immediate(pub i64);
#[derive(Clone)]
pub struct Memory(pub String);
#[derive(Clone)]
pub struct Label(pub String);
#[derive(Clone)]
pub struct FunctionLabel(pub String);

pub enum Type {
    I64,
    Pointer(Box<Type>),
}

impl fmt::Display for Program {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        for declaration in &self.declarations {
            write!(f, "{}", declaration)?;
        }
        Ok(())
    }
}

impl fmt::Display for FunctionDeclaration {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "func {}(", self.name)?;
        for (i, param) in self.parameters.iter().enumerate() {
            write!(f, "{}: {}", param.name, param.type_)?;
            if i != self.parameters.len() - 1 {
                write!(f, ", ")?;
            }
        }
        write!(f, ") -> {} {{\n", self.return_type)?;
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
        write!(f, "{}\n", self.exit)
    }
}

impl fmt::Display for Instruction {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Instruction::Set(r, x) => write!(f, "  {} = set {}", r, x),
            Instruction::Move(r1, r2) => write!(f, "  {} = move {}", r1, r2),
            Instruction::Alloca(mem, size) => write!(f, "  {} = alloca {}", mem, size),
            Instruction::Load(r, mem, offset) => write!(f, "  {} = load {}, {}", r, mem, offset),
            Instruction::Store(mem, r, offset) => write!(f, "  {} = store {}, {}", mem, r, offset),
            Instruction::Add(r1, r2, r3) => write!(f, "  {} = add {}, {}", r1, r2, r3),
            Instruction::Sub(r1, r2, r3) => write!(f, "  {} = sub {}, {}", r1, r2, r3),
            Instruction::Mul(r1, r2, r3) => write!(f, "  {} = mul {}, {}", r1, r2, r3),
            Instruction::Div(r1, r2, r3) => write!(f, "  {} = div {}, {}", r1, r2, r3),
            Instruction::Call(r, func, rs) => {
                write!(f, "  {} = call {}", r, func)?;
                for r in rs {
                    write!(f, ", {}", r)?;
                }
                fmt::Result::Ok(())
            }
            Instruction::CmpEq(r1, r2) => write!(f, "  cmpeq {}, {}", r1, r2),
            Instruction::CmpNeq(r1, r2) => write!(f, "  cmpneq {}, {}", r1, r2),
            Instruction::CmpLt(r1, r2) => write!(f, "  cmplt {}, {}", r1, r2),
            Instruction::CmpLte(r1, r2) => write!(f, "  cmplte {}, {}", r1, r2),
            Instruction::CmpGt(r1, r2) => write!(f, "  cmpgt {}, {}", r1, r2),
            Instruction::CmpGte(r1, r2) => write!(f, "  cmpgte {}, {}", r1, r2),
            Instruction::SetCmp(r) => write!(f, "  {} = setcmp", r),
        }
    }
}

impl fmt::Display for Register {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "%{}", self.0)
    }
}

impl fmt::Display for Immediate {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "{}", self.0)
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

impl fmt::Display for ExitInstruction {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            ExitInstruction::Ret(expr) => write!(f, "  ret {}", expr),
            ExitInstruction::Jump(label) => write!(f, "  jump {}", label),
            ExitInstruction::JumpIf(label1, label2) => write!(f, "  jumpif {}, {}", label1, label2),
            ExitInstruction::Placeholder => write!(f, "  <placeholder>"),
        }
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
