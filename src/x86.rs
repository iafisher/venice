// Compiles a VIL program into concrete x86 machine code.

use super::vil;
use std::collections::HashMap;
use std::fmt;

pub fn generate(vil: &vil::Program) -> Result<Program, String> {
    let mut generator = Generator::new();
    generator.generate_program(&vil);
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
    Je(String),
    Jg(String),
    Jge(String),
    Jl(String),
    Jle(String),
    Jmp(String),
    Jne(String),
    Mov(Value, Value),
    Pop(Value),
    Push(Value),
    Ret,
    Sub(Value, Value),
    ToDo(String),
}

pub enum Value {
    Immediate(i64),
    Register(Register),
    Label(String),
    Memory {
        scale: u8,
        displacement: i32,
        base: Register,
        index: Option<Register>,
    },
}

impl Value {
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
    offsets: HashMap<String, u32>,
    total_offset: u32,
}

impl Generator {
    fn new() -> Self {
        Generator {
            program: Program {
                externs: Vec::new(),
                blocks: Vec::new(),
                data: Vec::new(),
            },
            offsets: HashMap::new(),
            total_offset: 0,
        }
    }

    fn generate_program(&mut self, vil: &vil::Program) {
        self.program.externs = vil.externs.clone();

        for declaration in &vil.declarations {
            self.generate_declaration(&declaration);
        }

        for (string_name, string_value) in &vil.strings {
            self.program.data.push(Data {
                name: string_name.clone(),
                value: DataValue::Str(string_value.clone()),
            });
        }
    }

    fn generate_declaration(&mut self, declaration: &vil::FunctionDeclaration) {
        self.offsets.clear();
        self.total_offset = 8;
        let mut block = Block {
            // TODO: replace this with more robust logic
            global: declaration.name == "main",
            label: declaration.name.clone(),
            instructions: Vec::new(),
        };

        // Starting at 1 because 0 is RAX, which is the function's return value and is always
        // clobbered.
        for register_number in 1..declaration.max_register_count {
            block
                .instructions
                .push(Instruction::Push(Value::Register(Register(
                    (register_number as u8) + vil::PARAM_REGISTER_COUNT,
                ))));
        }

        self.program.blocks.push(block);
        for block in &declaration.blocks {
            self.generate_block(&declaration, &block);
        }
    }

    fn generate_block(&mut self, declaration: &vil::FunctionDeclaration, block: &vil::Block) {
        let mut instructions = Vec::new();
        for instruction in &block.instructions {
            self.generate_instruction(&declaration, &mut instructions, &instruction);
        }
        self.program.blocks.push(Block {
            global: false,
            label: block.name.clone(),
            instructions: instructions,
        });
    }

    fn generate_instruction(
        &mut self,
        // TODO: better way of exposing this information
        declaration: &vil::FunctionDeclaration,
        instructions: &mut Vec<Instruction>,
        instruction: &vil::Instruction,
    ) {
        match instruction {
            vil::Instruction::Alloca(mem, size) => {
                self.offsets.insert(mem.0.clone(), self.total_offset);
                self.total_offset += *size as u32;
            }
            vil::Instruction::Set(r, imm) => match imm {
                vil::Immediate::Integer(x) => {
                    instructions.push(Instruction::Mov(Value::r(r), Value::Immediate(*x)));
                }
                vil::Immediate::Label(s) => {
                    instructions.push(Instruction::Mov(Value::r(r), Value::Label(s.clone())));
                }
            },
            vil::Instruction::Move(r1, r2) => {
                instructions.push(Instruction::Mov(Value::r(r1), Value::r(r2)));
            }
            vil::Instruction::Add(r1, r2, r3) => {
                instructions.push(Instruction::Mov(Value::r(r1), Value::r(r3)));
                instructions.push(Instruction::Add(Value::r(r1), Value::r(r2)));
            }
            vil::Instruction::Sub(r1, r2, r3) => {
                instructions.push(Instruction::Mov(Value::r(r1), Value::r(r2)));
                instructions.push(Instruction::Sub(Value::r(r1), Value::r(r3)));
            }
            vil::Instruction::Load(r, mem, offset) => {
                let real_offset = *self.offsets.get(&mem.0).unwrap() + *offset as u32;
                instructions.push(Instruction::Mov(
                    Value::r(r),
                    Value::Memory {
                        scale: 1,
                        displacement: -(real_offset as i32),
                        base: RBP_REGISTER,
                        index: None,
                    },
                ));
            }
            vil::Instruction::Store(mem, r, offset) => {
                let real_offset = *self.offsets.get(&mem.0).unwrap() + *offset as u32;
                instructions.push(Instruction::Mov(
                    Value::Memory {
                        scale: 1,
                        displacement: -(real_offset as i32),
                        base: RBP_REGISTER,
                        index: None,
                    },
                    Value::r(r),
                ));
            }
            vil::Instruction::Cmp(r1, r2) => {
                instructions.push(Instruction::Cmp(Value::r(r1), Value::r(r2)));
            }
            vil::Instruction::Call(label) => {
                instructions.push(Instruction::Call(label.0.clone()));
            }
            vil::Instruction::FrameSetUp(size) => {
                instructions.push(Instruction::Push(RBP));
                instructions.push(Instruction::Mov(RBP, RSP));
                instructions.push(Instruction::Sub(RSP, Value::Immediate(*size as i64)));
            }
            vil::Instruction::FrameTearDown(size) => {
                instructions.push(Instruction::Add(RSP, Value::Immediate(*size as i64)));
                instructions.push(Instruction::Pop(RBP));
            }
            vil::Instruction::Ret => {
                // Starting at 1 because 0 is RAX, which is the function's return value and is always
                // clobbered.
                for register_number in (1..declaration.max_register_count).rev() {
                    instructions.push(Instruction::Pop(Value::Register(Register(
                        (register_number as u8) + vil::PARAM_REGISTER_COUNT,
                    ))));
                }

                instructions.push(Instruction::Ret);
            }
            vil::Instruction::Jump(l) => {
                instructions.push(Instruction::Jmp(l.0.clone()));
            }
            vil::Instruction::JumpEq(true_label, false_label) => {
                instructions.push(Instruction::Je(true_label.0.clone()));
                instructions.push(Instruction::Jmp(false_label.0.clone()));
            }
            vil::Instruction::JumpGt(true_label, false_label) => {
                instructions.push(Instruction::Jg(true_label.0.clone()));
                instructions.push(Instruction::Jmp(false_label.0.clone()));
            }
            vil::Instruction::JumpGte(true_label, false_label) => {
                instructions.push(Instruction::Jge(true_label.0.clone()));
                instructions.push(Instruction::Jmp(false_label.0.clone()));
            }
            vil::Instruction::JumpLt(true_label, false_label) => {
                instructions.push(Instruction::Jl(true_label.0.clone()));
                instructions.push(Instruction::Jmp(false_label.0.clone()));
            }
            vil::Instruction::JumpLte(true_label, false_label) => {
                instructions.push(Instruction::Jle(true_label.0.clone()));
                instructions.push(Instruction::Jmp(false_label.0.clone()));
            }
            vil::Instruction::JumpNeq(true_label, false_label) => {
                instructions.push(Instruction::Jne(true_label.0.clone()));
                instructions.push(Instruction::Jmp(false_label.0.clone()));
            }
            vil::Instruction::ToDo(s) => {
                // TODO
                instructions.push(Instruction::ToDo(s.clone()));
            }
            x => {
                // TODO
                instructions.push(Instruction::ToDo(format!("{:?}", x)));
            }
        }
    }
}

const RDI_REGISTER: Register = Register(0);
const RSI_REGISTER: Register = Register(1);
const RDX_REGISTER: Register = Register(2);
const RCX_REGISTER: Register = Register(3);
const R8_REGISTER: Register = Register(4);
const R9_REGISTER: Register = Register(5);
const R10_REGISTER: Register = Register(6);
const R11_REGISTER: Register = Register(7);
const R12_REGISTER: Register = Register(8);
const R13_REGISTER: Register = Register(9);
const R14_REGISTER: Register = Register(10);
const R15_REGISTER: Register = Register(11);
const RBX_REGISTER: Register = Register(12);
const RAX_REGISTER: Register = Register(13);
const RAX: Value = Value::Register(RAX_REGISTER);
const RSP_REGISTER: Register = Register(14);
const RSP: Value = Value::Register(RSP_REGISTER);
const RBP_REGISTER: Register = Register(15);
const RBP: Value = Value::Register(RBP_REGISTER);

const CALL_REGISTERS: &[Register] = &[
    RDI_REGISTER,
    RSI_REGISTER,
    RDX_REGISTER,
    RCX_REGISTER,
    R8_REGISTER,
    R9_REGISTER,
];

const CALLEE_SAVE_REGISTERS: &[Register] = &[
    RBP_REGISTER,
    RBX_REGISTER,
    R12_REGISTER,
    R13_REGISTER,
    R14_REGISTER,
    R15_REGISTER,
];

const CALLER_SAVE_REGISTERS: &[Register] = &[
    RAX_REGISTER,
    RCX_REGISTER,
    RDX_REGISTER,
    RDI_REGISTER,
    RSI_REGISTER,
    R8_REGISTER,
    R9_REGISTER,
    R10_REGISTER,
    R11_REGISTER,
];

impl fmt::Display for Program {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        if self.externs.len() > 0 {
            for extern_ in &self.externs {
                write!(f, "extern {}\n", extern_)?;
            }
            write!(f, "\n\n")?;
        }

        // TODO: globals

        write!(f, "section .text\n\n")?;
        for block in &self.blocks {
            write!(f, "{}\n", block)?;
        }

        write!(f, "section .data\n\n")?;
        for datum in &self.data {
            write!(f, "  {}\n", datum)?;
        }

        Ok(())
    }
}

impl fmt::Display for Block {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        if self.global {
            write!(f, "global {}\n", self.label)?;
        }

        write!(f, "{}:\n", self.label)?;
        for instruction in &self.instructions {
            write!(f, "  {}\n", instruction)?;
        }
        Ok(())
    }
}

impl fmt::Display for Instruction {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Instruction::Add(x, y) => write!(f, "add {}, {}", x, y),
            Instruction::Call(l) => write!(f, "call {}", l),
            Instruction::Cmp(x, y) => write!(f, "cmp {}, {}", x, y),
            Instruction::Je(l) => write!(f, "je {}", l),
            Instruction::Jg(l) => write!(f, "jg {}", l),
            Instruction::Jge(l) => write!(f, "jge {}", l),
            Instruction::Jl(l) => write!(f, "jl {}", l),
            Instruction::Jle(l) => write!(f, "jle {}", l),
            Instruction::Jmp(l) => write!(f, "jmp {}", l),
            Instruction::Jne(l) => write!(f, "jne {}", l),
            Instruction::Mov(x, y) => write!(f, "mov {}, {}", x, y),
            Instruction::Pop(x) => write!(f, "pop {}", x),
            Instruction::Push(x) => write!(f, "push {}", x),
            Instruction::Ret => write!(f, "ret"),
            Instruction::Sub(x, y) => write!(f, "sub {}, {}", x, y),
            Instruction::ToDo(s) => write!(f, "<todo: {}>", s),
        }
    }
}

impl fmt::Display for Value {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Value::Immediate(x) => write!(f, "{}", x),
            Value::Register(r) => write!(f, "{}", r),
            Value::Label(s) => write!(f, "{}", s),
            Value::Memory {
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
            x => {
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
