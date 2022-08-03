pub struct Program {
    declarations: Vec<Declaration>,
}

pub enum Declaration {
    Function(FunctionDeclaration),
    Const(ConstDeclaration),
}

pub struct FunctionDeclaration {
    name: String,
    parameters: Vec<FunctionParameter>,
    return_type: Type,
    blocks: Vec<Block>,
}

pub struct FunctionParameter {
    name: String,
    type_: Type,
}

pub struct ConstDeclaration {
    symbol: String,
    value: TypedExpression,
}

pub struct Block {
    name: String,
    instructions: Vec<Instruction>,
    exit: ExitInstruction,
}

pub enum Instruction {
    Alloca {
        symbol: String,
        type_: Type,
        size: usize,
    },
    Store {
        symbol: String,
        expression: TypedExpression,
    },
    Load {
        symbol: String,
        expression: TypedExpression,
    },
    Add {
        symbol: String,
        left: TypedExpression,
        right: TypedExpression,
    },
    Sub {
        symbol: String,
        left: TypedExpression,
        right: TypedExpression,
    },
    Mul {
        symbol: String,
        left: TypedExpression,
        right: TypedExpression,
    },
    Div {
        symbol: String,
        left: TypedExpression,
        right: TypedExpression,
    },
    Call {
        function: String,
        arguments: Vec<TypedExpression>,
    },
}

pub enum ExitInstruction {
    Ret(Expression),
    Break {
        condition: Expression,
        label_true: String,
        label_false: String,
    },
}

pub struct TypedExpression {
    type_: Type,
    value: Expression,
}

pub enum Expression {
    Integer(i64),
}

pub enum Type {
    I64,
    Pointer(Box<Type>),
}
