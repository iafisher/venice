pub struct Program {
    pub declarations: Vec<Declaration>,
}

pub enum Declaration {
    Function(FunctionDeclaration),
    Const(ConstDeclaration),
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
    pub value: TypedExpression,
}

pub struct Block {
    pub name: String,
    pub instructions: Vec<Instruction>,
    pub exit: ExitInstruction,
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
    CmpLt {
        symbol: String,
        left: TypedExpression,
        right: TypedExpression,
    },
}

pub enum ExitInstruction {
    Ret(TypedExpression),
    Jump(String),
    JumpCond {
        condition: TypedExpression,
        label_true: String,
        label_false: String,
    },
}

pub struct TypedExpression {
    pub type_: Type,
    pub value: Expression,
}

pub enum Expression {
    Integer(i64),
    Symbol(String),
}

pub enum Type {
    I64,
    Pointer(Box<Type>),
}
