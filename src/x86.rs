// Copyright 2022 The Venice Authors. All rights reserved.
// Use of this source code is governed by an MIT-style license that can be
// found in the LICENSE file.
//
// Compiles a VIL program into concrete x86 machine code.

use super::vil;
use std::fmt;

pub fn generate(vil: &vil::Program) -> Result<Program, String> {
    let mut generator = Generator::new();
    generator.generate_program(vil);
    Ok(generator.program)
}

pub struct Program {
    externs: Vec<String>,
    blocks: Vec<Block>,
    data: Vec<Data>,
}

pub struct Block {
    global: bool,
    label: String,
    instructions: Vec<Instruction>,
}

pub enum Instruction {
    Add(Value, Value),
    Call(String),
    Cmp(Value, Value),
    IDiv(Value),
    IMul(Value, Value),
    Je(String),
    Jg(String),
    Jge(String),
    Jl(String),
    Jle(String),
    Jmp(String),
    Jne(String),
    Mov(Value, Value),
    Neg(Value),
    Pop(Value),
    Push(Value),
    Ret,
    SetE(Value),
    Sub(Value, Value),
    Test(Value, Value),
    Xor(Value, Value),
    ToDo(String),
}

pub enum Value {
    Immediate(i64),
    Register(Register),
    /// Directly holds a register's assembly-language name for special cases, e.g. for byte
    /// registers like AL.
    SpecialRegister(String),
    Label(String),
    Memory {
        scale: u8,
        displacement: i32,
        base: Register,
        index: Option<Register>,
    },
}

impl Value {
    /// Constructs an x86 register from a VIL register.
    fn r(r: &vil::Register) -> Self {
        Value::Register(Register(r.absolute_index()))
    }
}

#[derive(Clone, Copy)]
pub struct Register(u8);

pub struct Data {
    name: String,
    value: DataValue,
}

pub enum DataValue {
    Str(String),
}

struct Generator {
    program: Program,
}

impl Generator {
    fn new() -> Self {
        Generator {
            program: Program {
                externs: Vec::new(),
                blocks: Vec::new(),
                data: Vec::new(),
            },
        }
    }

    fn generate_program(&mut self, vil: &vil::Program) {
        self.program.externs = vil.externs.clone();

        for declaration in &vil.declarations {
            self.generate_declaration(declaration);
        }

        for (string_name, string_value) in &vil.strings {
            self.program.data.push(Data {
                name: string_name.clone(),
                value: DataValue::Str(string_value.clone()),
            });
        }
    }

    fn generate_declaration(&mut self, declaration: &vil::FunctionDeclaration) {
        let block = Block {
            // TODO: replace this with more robust logic
            global: declaration.name == "main",
            label: declaration.name.clone(),
            instructions: Vec::new(),
        };

        self.program.blocks.push(block);
        for block in &declaration.blocks {
            self.generate_block(declaration, block);
        }
    }

    fn generate_block(&mut self, declaration: &vil::FunctionDeclaration, block: &vil::Block) {
        let mut instructions = Vec::new();
        for instruction in &block.instructions {
            self.generate_instruction(declaration, &mut instructions, instruction);
        }
        self.program.blocks.push(Block {
            global: false,
            label: block.name.clone(),
            instructions,
        });
    }

    fn generate_instruction(
        &mut self,
        // TODO: better way of exposing this information
        _declaration: &vil::FunctionDeclaration,
        instructions: &mut Vec<Instruction>,
        instruction: &vil::Instruction,
    ) {
        use vil::InstructionKind::*;
        match &instruction.kind {
            Set(r, imm) => match imm {
                vil::Immediate::Integer(x) => {
                    instructions.push(Instruction::Mov(Value::r(r), Value::Immediate(*x)));
                }
                vil::Immediate::Label(s) => {
                    instructions.push(Instruction::Mov(Value::r(r), Value::Label(s.clone())));
                }
            },
            Move(r1, r2) => {
                instructions.push(Instruction::Mov(Value::r(r1), Value::r(r2)));
            }
            Add(r1, r2, r3) => {
                instructions.push(Instruction::Mov(Value::r(r1), Value::r(r3)));
                instructions.push(Instruction::Add(Value::r(r1), Value::r(r2)));
            }
            Sub(r1, r2, r3) => {
                instructions.push(Instruction::Mov(Value::r(r1), Value::r(r2)));
                instructions.push(Instruction::Sub(Value::r(r1), Value::r(r3)));
            }
            Mul(r1, r2, r3) => {
                instructions.push(Instruction::Mov(Value::r(r1), Value::r(r2)));
                instructions.push(Instruction::IMul(Value::r(r1), Value::r(r3)));
            }
            Div(r1, r2, r3) => {
                // In x86, `div RXX` computes RDX:RAX / RXX and stores the quotient in RAX and the
                // remainder in RDX.
                //
                // The compiler will never use RAX or RDX for regular expressions, so we don't have
                // to worry about the case where r1, r2, or r3 is RAX or RDX.

                // First, we zero out RDX since we are only doing 64-bit division, not 128-bit.
                instructions.push(Instruction::Xor(RDX, RDX));

                // Move the dividend into RAX.
                instructions.push(Instruction::Mov(RAX, Value::r(r2)));

                // Divide by the divisor.
                instructions.push(Instruction::IDiv(Value::r(r3)));

                // Move RAX into the destination register.
                instructions.push(Instruction::Mov(Value::r(r1), RAX));
            }
            Negate(r1, r2) => {
                instructions.push(Instruction::Mov(Value::r(r1), Value::r(r2)));
                instructions.push(Instruction::Neg(Value::r(r1)));
            }
            LogicalNot(r1, r2) => {
                // XOR RAX with itself to produce 0, then test it against the source register and
                // set AL to the ZF flag. Since we already zeroed out RAX, all the high bits will
                // also be 0.
                instructions.push(Instruction::Xor(RAX, RAX));
                instructions.push(Instruction::Test(RAX, Value::r(r2)));
                instructions.push(Instruction::SetE(Value::SpecialRegister(String::from(
                    "al",
                ))));
                instructions.push(Instruction::Mov(Value::r(r1), RAX));
            }
            Load(r, offset) => {
                instructions.push(Instruction::Mov(
                    Value::r(r),
                    Value::Memory {
                        scale: 1,
                        displacement: *offset,
                        base: RBP_REGISTER,
                        index: None,
                    },
                ));
            }
            Store(r, offset) => {
                instructions.push(Instruction::Mov(
                    Value::Memory {
                        scale: 1,
                        displacement: *offset,
                        base: RBP_REGISTER,
                        index: None,
                    },
                    Value::r(r),
                ));
            }
            Cmp(r1, r2) => {
                instructions.push(Instruction::Cmp(Value::r(r1), Value::r(r2)));
            }
            Call(label) => {
                instructions.push(Instruction::Call(label.0.clone()));
            }
            CallVariadic(label) => {
                // The System V ABI requires setting AL to the number of vector registers when
                // calling a variadic function.
                instructions.push(Instruction::Mov(
                    Value::SpecialRegister(String::from("al")),
                    Value::Immediate(0),
                ));
                instructions.push(Instruction::Call(label.0.clone()));
            }
            FrameSetUp(size) => {
                instructions.push(Instruction::Push(RBP));
                instructions.push(Instruction::Mov(RBP, RSP));
                instructions.push(Instruction::Sub(RSP, Value::Immediate(*size as i64)));
            }
            FrameTearDown(size) => {
                instructions.push(Instruction::Add(RSP, Value::Immediate(*size as i64)));
                instructions.push(Instruction::Pop(RBP));
            }
            Ret => {
                instructions.push(Instruction::Ret);
            }
            Jump(l) => {
                instructions.push(Instruction::Jmp(l.0.clone()));
            }
            JumpEq(true_label, false_label) => {
                instructions.push(Instruction::Je(true_label.0.clone()));
                instructions.push(Instruction::Jmp(false_label.0.clone()));
            }
            JumpGt(true_label, false_label) => {
                instructions.push(Instruction::Jg(true_label.0.clone()));
                instructions.push(Instruction::Jmp(false_label.0.clone()));
            }
            JumpGte(true_label, false_label) => {
                instructions.push(Instruction::Jge(true_label.0.clone()));
                instructions.push(Instruction::Jmp(false_label.0.clone()));
            }
            JumpLt(true_label, false_label) => {
                instructions.push(Instruction::Jl(true_label.0.clone()));
                instructions.push(Instruction::Jmp(false_label.0.clone()));
            }
            JumpLte(true_label, false_label) => {
                instructions.push(Instruction::Jle(true_label.0.clone()));
                instructions.push(Instruction::Jmp(false_label.0.clone()));
            }
            JumpNeq(true_label, false_label) => {
                instructions.push(Instruction::Jne(true_label.0.clone()));
                instructions.push(Instruction::Jmp(false_label.0.clone()));
            }
            CalleeRestore(r) => {
                instructions.push(Instruction::Pop(Value::r(r)));
            }
            CalleeSave(r) => {
                instructions.push(Instruction::Push(Value::r(r)));
            }
            CallerRestore(r) => {
                instructions.push(Instruction::Pop(Value::r(r)));
            }
            CallerSave(r) => {
                instructions.push(Instruction::Push(Value::r(r)));
            }
            ToDo(s) => {
                // TODO
                instructions.push(Instruction::ToDo(s.clone()));
            }
        }
    }
}

const RDX_REGISTER: Register = Register(2);
const RDX: Value = Value::Register(RDX_REGISTER);
const RAX_REGISTER: Register = Register(13);
const RAX: Value = Value::Register(RAX_REGISTER);
const RSP_REGISTER: Register = Register(14);
const RSP: Value = Value::Register(RSP_REGISTER);
const RBP_REGISTER: Register = Register(15);
const RBP: Value = Value::Register(RBP_REGISTER);

impl fmt::Display for Program {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        if !self.externs.is_empty() {
            for extern_ in &self.externs {
                writeln!(f, "extern {}", extern_)?;
            }
            write!(f, "\n\n")?;
        }

        // TODO: globals

        write!(f, "section .text\n\n")?;
        for block in &self.blocks {
            writeln!(f, "{}", block)?;
        }

        write!(f, "section .data\n\n")?;
        for datum in &self.data {
            writeln!(f, "  {}", datum)?;
        }

        Ok(())
    }
}

impl fmt::Display for Block {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        if self.global {
            writeln!(f, "global {}", self.label)?;
        }

        writeln!(f, "{}:", self.label)?;
        for instruction in &self.instructions {
            writeln!(f, "  {}", instruction)?;
        }
        Ok(())
    }
}

impl fmt::Display for Instruction {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        use Instruction::*;
        match self {
            Add(x, y) => write!(f, "add {}, {}", x, y),
            Call(l) => write!(f, "call {}", l),
            Cmp(x, y) => write!(f, "cmp {}, {}", x, y),
            IDiv(x) => write!(f, "div {}", x),
            IMul(x, y) => write!(f, "imul {}, {}", x, y),
            Je(l) => write!(f, "je {}", l),
            Jg(l) => write!(f, "jg {}", l),
            Jge(l) => write!(f, "jge {}", l),
            Jl(l) => write!(f, "jl {}", l),
            Jle(l) => write!(f, "jle {}", l),
            Jmp(l) => write!(f, "jmp {}", l),
            Jne(l) => write!(f, "jne {}", l),
            Mov(x, y) => write!(f, "mov {}, {}", x, y),
            Neg(x) => write!(f, "neg {}", x),
            Pop(x) => write!(f, "pop {}", x),
            Push(x) => write!(f, "push {}", x),
            Ret => write!(f, "ret"),
            SetE(x) => write!(f, "sete {}", x),
            Sub(x, y) => write!(f, "sub {}, {}", x, y),
            Test(x, y) => write!(f, "test {}, {}", x, y),
            Xor(x, y) => write!(f, "xor {}, {}", x, y),
            ToDo(s) => write!(f, "<todo: {}>", s),
        }
    }
}

impl fmt::Display for Value {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        use Value::*;
        match self {
            Immediate(x) => write!(f, "{}", x),
            Register(r) => write!(f, "{}", r),
            SpecialRegister(s) => write!(f, "{}", s),
            Label(s) => write!(f, "{}", s),
            Memory {
                scale,
                displacement,
                base,
                index,
            } => {
                write!(f, "[")?;
                if *scale != 1 {
                    write!(f, "{}*", scale)?;
                }
                write!(f, "{}", base)?;
                if let Some(index) = index {
                    write!(f, "+{}", index)?;
                }
                if *displacement != 0 {
                    if *displacement < 0 {
                        write!(f, "-{}", -displacement)?;
                    } else {
                        write!(f, "+{}", displacement)?;
                    }
                }
                write!(f, "]")
            }
        }
    }
}

impl fmt::Display for Register {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self.0 {
            0 => write!(f, "rdi"),
            1 => write!(f, "rsi"),
            2 => write!(f, "rdx"),
            3 => write!(f, "rcx"),
            4 => write!(f, "r8"),
            5 => write!(f, "r9"),
            6 => write!(f, "r10"),
            7 => write!(f, "r11"),
            8 => write!(f, "r12"),
            9 => write!(f, "r13"),
            10 => write!(f, "r14"),
            11 => write!(f, "r15"),
            12 => write!(f, "rbx"),
            13 => write!(f, "rax"),
            14 => write!(f, "rsp"),
            15 => write!(f, "rbp"),
            _x => {
                panic!("register out of range");
            }
        }
    }
}

impl fmt::Display for Data {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match &self.value {
            DataValue::Str(s) => {
                write!(f, "{} db {:?}, 0", self.name, s)
            }
        }
    }
}
