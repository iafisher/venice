use super::vil;
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
    Movq(Value, Value),
}

pub enum Value {
    Immediate(i64),
    Register(String),
}

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
                globals: Vec::new(),
                blocks: Vec::new(),
                data: Vec::new(),
            },
        }
    }

    fn generate_program(&mut self, vil: &vil::Program) {
        for declaration in &vil.declarations {
            self.generate_declaration(&declaration);
        }
    }

    fn generate_declaration(&mut self, declaration: &vil::FunctionDeclaration) {
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
            vil::Instruction::Move(vr1, vr2) => {
                let r1 = self.claim_register();
                let r2 = self.claim_register();
                instructions.push(Instruction::Movq(r1, r2));
            }
            _ => {
                // TODO
            }
        }
    }

    fn generate_exit_instruction(
        &mut self,
        instructions: &mut Vec<Instruction>,
        instruction: &vil::ExitInstruction,
    ) {
        // TODO
    }

    fn claim_register(&mut self) -> Value {
        // TODO
        Value::Register(String::from("rax"))
    }
}

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
            Instruction::Movq(x, y) => write!(f, "movq {}, {}", x, y),
        }
    }
}

impl fmt::Display for Value {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Value::Immediate(x) => write!(f, "{}", x),
            Value::Register(r) => write!(f, "{}", r),
        }
    }
}
