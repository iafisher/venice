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
    And(Value, Value),
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
        Value::Register(Register(r.index()))
    }

    /// Constructs the x86 register for a function's i'th parameter (starting at 0).
    fn param(i: u8) -> Self {
        Value::Register(Register(i + 7))
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
    stack_alignment: i64,
}

impl Generator {
    fn new() -> Self {
        Generator {
            program: Program {
                externs: Vec::new(),
                blocks: Vec::new(),
                data: Vec::new(),
            },
            stack_alignment: 0,
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
        self.stack_alignment = 8;

        let block = Block {
            // TODO: replace this with more robust logic
            global: declaration.name == "venice_main",
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
            self.generate_instruction(&mut instructions, instruction);
        }
        self.program.blocks.push(Block {
            global: false,
            label: block.name.clone(),
            instructions,
        });
    }

    fn generate_instruction(
        &mut self,
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
                instructions.push(Instruction::Add(Value::r(r2), Value::r(r3)));
                instructions.push(Instruction::Mov(Value::r(r1), Value::r(r2)));
            }
            Sub(r1, r2, r3) => {
                instructions.push(Instruction::Sub(Value::r(r2), Value::r(r3)));
                instructions.push(Instruction::Mov(Value::r(r1), Value::r(r2)));
            }
            Mul(r1, r2, r3) => {
                instructions.push(Instruction::IMul(Value::r(r2), Value::r(r3)));
                instructions.push(Instruction::Mov(Value::r(r1), Value::r(r2)));
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
                instructions.push(Instruction::Neg(Value::r(r2)));
                instructions.push(Instruction::Mov(Value::r(r1), Value::r(r2)));
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
            Call(label, registers) => {
                for (i, register) in registers.iter().enumerate() {
                    instructions.push(Instruction::Mov(
                        Value::param(u8::try_from(i).unwrap()),
                        Value::r(register),
                    ));
                }

                self.align_stack(instructions);
                instructions.push(Instruction::Call(label.0.clone()));
                self.unalign_stack(instructions);
            }
            CallVariadic(label, registers) => {
                for (i, register) in registers.iter().enumerate() {
                    instructions.push(Instruction::Mov(
                        Value::param(u8::try_from(i).unwrap()),
                        Value::r(register),
                    ));
                }

                // The System V ABI requires setting AL to the number of vector registers when
                // calling a variadic function.
                instructions.push(Instruction::Mov(
                    Value::SpecialRegister(String::from("al")),
                    Value::Immediate(0),
                ));

                self.align_stack(instructions);
                instructions.push(Instruction::Call(label.0.clone()));
                self.unalign_stack(instructions);
            }
            FrameSetUp(size) => {
                instructions.push(Instruction::Push(RBP));
                self.stack_alignment += 8;
                instructions.push(Instruction::Mov(RBP, RSP));
                let size_as_i64 = i64::try_from(*size).unwrap();
                instructions.push(Instruction::Sub(RSP, Value::Immediate(size_as_i64)));
                self.stack_alignment += size_as_i64;
            }
            FrameTearDown(size) => {
                let size_as_i64 = i64::try_from(*size).unwrap();
                instructions.push(Instruction::Add(RSP, Value::Immediate(size_as_i64)));
                self.stack_alignment -= size_as_i64;
                instructions.push(Instruction::Pop(RBP));
                self.stack_alignment -= 8;
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
                self.stack_alignment -= 8;
            }
            CalleeSave(r) => {
                instructions.push(Instruction::Push(Value::r(r)));
                self.stack_alignment += 8;
            }
            CallerRestore(r) => {
                instructions.push(Instruction::Pop(Value::r(r)));
                self.stack_alignment -= 8;
            }
            CallerSave(r) => {
                instructions.push(Instruction::Push(Value::r(r)));
                self.stack_alignment += 8;
            }
        }
    }

    fn align_stack(&mut self, instructions: &mut Vec<Instruction>) {
        let diff = self.stack_alignment % 16;
        if diff > 0 {
            instructions.push(Instruction::Sub(RSP, Value::Immediate(diff)));
        }
    }

    fn unalign_stack(&mut self, instructions: &mut Vec<Instruction>) {
        let diff = self.stack_alignment % 16;
        if diff > 0 {
            instructions.push(Instruction::Add(RSP, Value::Immediate(diff)));
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
        for block in &self.blocks {
            writeln!(f, "{}", block)?;
        }

        for datum in &self.data {
            writeln!(f, "{}", datum)?;
        }

        Ok(())
    }
}

impl fmt::Display for Block {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        if self.global {
            writeln!(f, ".globl {}", self.label)?;
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
            Add(x, y) => write!(f, "addq {}, {}", y, x),
            And(x, y) => write!(f, "andq {}, {}", y, x),
            Call(l) => write!(f, "call {}", l),
            Cmp(x, y) => write!(f, "cmpq {}, {}", y, x),
            IDiv(x) => write!(f, "divq {}", x),
            IMul(x, y) => write!(f, "imulq {}, {}", y, x),
            Je(l) => write!(f, "je {}", l),
            Jg(l) => write!(f, "jg {}", l),
            Jge(l) => write!(f, "jge {}", l),
            Jl(l) => write!(f, "jl {}", l),
            Jle(l) => write!(f, "jle {}", l),
            Jmp(l) => write!(f, "jmp {}", l),
            Jne(l) => write!(f, "jne {}", l),
            Mov(x, y) => {
                // TODO: this logic is brittle, and also needs to be applied to all the other
                // instructions.
                if matches!(x, Value::SpecialRegister(_)) || matches!(y, Value::SpecialRegister(_))
                {
                    write!(f, "movb {}, {}", y, x)
                } else {
                    write!(f, "movq {}, {}", y, x)
                }
            }
            Neg(x) => write!(f, "negq {}", x),
            Pop(x) => write!(f, "popq {}", x),
            Push(x) => write!(f, "pushq {}", x),
            Ret => write!(f, "retq"),
            SetE(x) => write!(f, "sete {}", x),
            Sub(x, y) => write!(f, "subq {}, {}", y, x),
            Test(x, y) => write!(f, "testq {}, {}", y, x),
            Xor(x, y) => write!(f, "xorq {}, {}", y, x),
        }
    }
}

impl fmt::Display for Value {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        use Value::*;
        match self {
            Immediate(x) => write!(f, "${}", x),
            Register(r) => write!(f, "%{}", r),
            SpecialRegister(s) => write!(f, "%{}", s),
            Label(s) => write!(f, "$.{}", s),
            Memory {
                scale,
                displacement,
                base,
                index,
            } => {
                if *scale != 1 || index.is_some() {
                    // TODO
                    panic!("internal error: memory operand not supported");
                }

                if *displacement != 0 {
                    write!(f, "{}(%{})", displacement, base)
                } else {
                    write!(f, "(%{})", base)
                }
            }
        }
    }
}

impl fmt::Display for Register {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self.0 {
            0 => write!(f, "r10"),
            1 => write!(f, "r11"),
            2 => write!(f, "r12"),
            3 => write!(f, "r13"),
            4 => write!(f, "r14"),
            5 => write!(f, "r15"),
            6 => write!(f, "rbx"),
            7 => write!(f, "rdi"),
            8 => write!(f, "rsi"),
            9 => write!(f, "rdx"),
            10 => write!(f, "rcx"),
            11 => write!(f, "r8"),
            12 => write!(f, "r9"),
            13 => write!(f, "rax"),
            14 => write!(f, "rsp"),
            15 => write!(f, "rbp"),
            _x => {
                panic!("internal error: register out of range: {}", self.0);
            }
        }
    }
}

impl fmt::Display for Data {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match &self.value {
            DataValue::Str(s) => {
                // TODO: does this handle backslash escapes correctly? i.e., is the debug format
                // that Rust generates compatible with the assembler?
                write!(f, ".{}:\n  .string {:?}", self.name, s)
            }
        }
    }
}
