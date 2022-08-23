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
    globals: Vec<String>,
    blocks: Vec<Block>,
    data: Vec<Data>,
}

pub struct Block {
    label: String,
    instructions: Vec<Instruction>,
}

pub enum Instruction {
    Add(Value, Value),
    Jmp(String),
    Movq(Value, Value),
    Ret,
    Sub(Value, Value),
    ToDo(String),
}

pub enum Value {
    Immediate(i64),
    Register(Register),
    Memory {
        scale: u8,
        displacement: u32,
        base: Register,
        index: Option<Register>,
    },
}

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
                globals: Vec::new(),
                blocks: Vec::new(),
                data: Vec::new(),
            },
            offsets: HashMap::new(),
            total_offset: 0,
        }
    }

    fn generate_program(&mut self, vil: &vil::Program) {
        for declaration in &vil.declarations {
            self.generate_declaration(&declaration);
        }
    }

    fn generate_declaration(&mut self, declaration: &vil::FunctionDeclaration) {
        self.offsets.clear();
        self.total_offset = 0;
        self.program.blocks.push(Block {
            label: declaration.name.clone(),
            instructions: Vec::new(),
        });

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
                self.total_offset += *size as u32;
                self.offsets.insert(mem.0.clone(), self.total_offset);
            }
            vil::Instruction::Set(r, imm) => {
                instructions.push(Instruction::Movq(R(r), Value::Immediate(imm.0)));
            }
            vil::Instruction::Move(r1, r2) => {
                instructions.push(Instruction::Movq(R(r1), R(r2)));
            }
            vil::Instruction::Add(r1, r2, r3) => {
                instructions.push(Instruction::Movq(R(r1), R(r3)));
                instructions.push(Instruction::Add(R(r1), R(r2)));
            }
            vil::Instruction::Sub(r1, r2, r3) => {
                instructions.push(Instruction::Movq(R(r1), R(r3)));
                instructions.push(Instruction::Sub(R(r1), R(r2)));
            }
            vil::Instruction::Load(r, mem, offset) => {
                let real_offset = *self.offsets.get(&mem.0).unwrap() + *offset as u32;
                instructions.push(Instruction::Movq(
                    R(r),
                    Value::Memory {
                        scale: 1,
                        displacement: real_offset,
                        base: RSP_reg,
                        index: None,
                    },
                ));
            }
            vil::Instruction::Store(mem, r, offset) => {}
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
                instructions.push(Instruction::Movq(RAX, R(r)));
                instructions.push(Instruction::Ret);
            }
            vil::ExitInstruction::Jump(l) => {
                instructions.push(Instruction::Jmp(l.0.clone()));
            }
            x => {
                // TODO
                instructions.push(Instruction::ToDo(format!("{:?}", x)));
            }
        }
    }
}

fn R(r: &vil::Register) -> Value {
    Value::Register(Register(r.0.try_into().unwrap()))
}

const RAX_reg: Register = Register(0);
const RAX: Value = Value::Register(RAX_reg);
const RBP_reg: Register = Register(15);
const RBP: Value = Value::Register(RBP_reg);
const RSP_reg: Register = Register(16);
const RSP: Value = Value::Register(RSP_reg);

impl fmt::Display for Program {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        // TODO: externs
        // TODO: globals
        for block in &self.blocks {
            write!(f, "{}\n", block)?;
        }
        // TODO: data
        Ok(())
    }
}

impl fmt::Display for Block {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
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
            Instruction::Jmp(l) => write!(f, "jmp {}", l),
            Instruction::Movq(x, y) => write!(f, "movq {}, {}", x, y),
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
            Value::Memory {
                scale,
                displacement,
                base,
                index,
            } => {
                if *displacement != 0 {
                    write!(f, "{}", displacement)?;
                }
                write!(f, "[")?;
                if *scale != 1 {
                    write!(f, "{}*", scale)?;
                }
                write!(f, "{}", base)?;
                if let Some(index) = index {
                    write!(f, "+{}", index)?;
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
            6 => write!(f, "rdi"),
            7 => write!(f, "r8"),
            8 => write!(f, "r9"),
            9 => write!(f, "r10"),
            10 => write!(f, "r11"),
            11 => write!(f, "r12"),
            12 => write!(f, "r13"),
            13 => write!(f, "r14"),
            14 => write!(f, "r15"),
            15 => write!(f, "rbp"),
            16 => write!(f, "rsp"),
            x => write!(f, "r{}", x),
        }
    }
}
