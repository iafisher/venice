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

pub struct Instruction {
    name: String,
    operands: Vec<Expression>,
}

pub enum Expression {
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
