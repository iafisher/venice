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
        Value::Register(Register(r.0.try_into().unwrap()))
    }
}

#[derive(Clone)]
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
        self.total_offset = 0;
        let mut block = Block {
            // TODO: replace this with more robust logic
            global: declaration.name == "main",
            label: declaration.name.clone(),
            instructions: Vec::new(),
        };

        block.instructions.push(Instruction::Push(RBP));
        block.instructions.push(Instruction::Mov(RBP, RSP));
        self.total_offset += 8;

        if declaration.parameters.len() > 6 {
            panic!("too many parameter");
        }

        // Move the function's parameters from registers onto the stack, where the rest
        // of the function's body expects them to be.
        for (i, parameter) in declaration.parameters.iter().enumerate() {
            // TODO: calculate size accurately
            let size: u32 = 8;
            block
                .instructions
                .push(Instruction::Sub(RSP, Value::Immediate(size as i64)));
            block.instructions.push(Instruction::Mov(
                Value::Memory {
                    scale: 1,
                    displacement: -(self.total_offset as i32),
                    base: RBP_REGISTER,
                    index: None,
                },
                Value::Register(CALL_REGISTERS[i].clone()),
            ));
            self.offsets
                .insert(parameter.unique_name.clone(), self.total_offset);
            self.total_offset += size;
        }

        self.program.blocks.push(block);
        for block in &declaration.blocks {
            self.generate_block(&block);
        }
    }

    fn generate_block(&mut self, block: &vil::Block) {
        let mut instructions = Vec::new();
        for instruction in &block.instructions {
            self.generate_instruction(&mut instructions, &instruction);
        }
        self.generate_exit_instruction(&mut instructions, &block.exit);
        self.program.blocks.push(Block {
            global: false,
            label: block.name.clone(),
            instructions: instructions,
        });
    }

    fn generate_instruction(
        &mut self,
        instructions: &mut Vec<Instruction>,
        instruction: &vil::Instruction,
    ) {
        match instruction {
            vil::Instruction::Alloca(mem, size) => {
                instructions.push(Instruction::Sub(
                    RSP,
                    Value::Immediate((*size).try_into().unwrap()),
                ));
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
            vil::Instruction::Call(r, label, args) => {
                // TODO: handle additional arguments on the stack.
                if args.len() > 6 {
                    panic!("too many arguments");
                }

                for (i, arg) in args.iter().enumerate() {
                    let call_register = Value::Register(CALL_REGISTERS[i].clone());
                    instructions.push(Instruction::Mov(call_register, Value::r(arg)));
                }
                instructions.push(Instruction::Call(label.0.clone()));
                instructions.push(Instruction::Mov(Value::r(r), RAX));
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

    fn generate_exit_instruction(
        &mut self,
        instructions: &mut Vec<Instruction>,
        instruction: &vil::ExitInstruction,
    ) {
        match instruction {
            vil::ExitInstruction::Ret(r) => {
                instructions.push(Instruction::Mov(RAX, Value::r(r)));
                // TODO: The total offset is always at least 8 because RBP starts at
                // the value of RSP before the function was called, but this check is
                // really janky.
                if self.total_offset > 8 {
                    // TODO: does this always work?
                    instructions.push(Instruction::Add(
                        RSP,
                        Value::Immediate((self.total_offset - 8) as i64),
                    ));
                }
                instructions.push(Instruction::Pop(RBP));
                instructions.push(Instruction::Ret);
            }
            vil::ExitInstruction::Jump(l) => {
                instructions.push(Instruction::Jmp(l.0.clone()));
            }
            vil::ExitInstruction::JumpEq(true_label, false_label) => {
                instructions.push(Instruction::Je(true_label.0.clone()));
                instructions.push(Instruction::Jmp(false_label.0.clone()));
            }
            vil::ExitInstruction::JumpGt(true_label, false_label) => {
                instructions.push(Instruction::Jg(true_label.0.clone()));
                instructions.push(Instruction::Jmp(false_label.0.clone()));
            }
            vil::ExitInstruction::JumpGte(true_label, false_label) => {
                instructions.push(Instruction::Jge(true_label.0.clone()));
                instructions.push(Instruction::Jmp(false_label.0.clone()));
            }
            vil::ExitInstruction::JumpLt(true_label, false_label) => {
                instructions.push(Instruction::Jl(true_label.0.clone()));
                instructions.push(Instruction::Jmp(false_label.0.clone()));
            }
            vil::ExitInstruction::JumpLte(true_label, false_label) => {
                instructions.push(Instruction::Jle(true_label.0.clone()));
                instructions.push(Instruction::Jmp(false_label.0.clone()));
            }
            vil::ExitInstruction::JumpNeq(true_label, false_label) => {
                instructions.push(Instruction::Jne(true_label.0.clone()));
                instructions.push(Instruction::Jmp(false_label.0.clone()));
            }
            x => {
                // TODO
                instructions.push(Instruction::ToDo(format!("{:?}", x)));
            }
        }
    }
}

const RAX_REGISTER: Register = Register(0);
const RAX: Value = Value::Register(RAX_REGISTER);
const RBP_REGISTER: Register = Register(14);
const RBP: Value = Value::Register(RBP_REGISTER);
const RSP_REGISTER: Register = Register(15);
const RSP: Value = Value::Register(RSP_REGISTER);

const RDI_REGISTER: Register = Register(5);
const RSI_REGISTER: Register = Register(4);
const RDX_REGISTER: Register = Register(3);
const RCX_REGISTER: Register = Register(2);
const R8_REGISTER: Register = Register(6);
const R9_REGISTER: Register = Register(7);
const CALL_REGISTERS: [Register; 6] = [
    RDI_REGISTER,
    RSI_REGISTER,
    RDX_REGISTER,
    RCX_REGISTER,
    R8_REGISTER,
    R9_REGISTER,
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
            0 => write!(f, "rax"),
            1 => write!(f, "rbx"),
            2 => write!(f, "rcx"),
            3 => write!(f, "rdx"),
            4 => write!(f, "rsi"),
            5 => write!(f, "rdi"),
            6 => write!(f, "r8"),
            7 => write!(f, "r9"),
            8 => write!(f, "r10"),
            9 => write!(f, "r11"),
            10 => write!(f, "r12"),
            11 => write!(f, "r13"),
            12 => write!(f, "r14"),
            13 => write!(f, "r15"),
            14 => write!(f, "rbp"),
            15 => write!(f, "rsp"),
            x => write!(f, "r{}", x),
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
