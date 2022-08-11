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
    Placeholder,
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
            Instruction::Alloca {
                symbol,
                type_,
                size,
            } => {
                write!(f, "  {} = alloca:{} {}", symbol, type_, size)
            }
            Instruction::Add {
                symbol,
                left,
                right,
            } => format_binary_op(f, "add", symbol, left, right),
            Instruction::Store { symbol, expression } => {
                write!(f, "  store {}, {}", symbol, expression)
            }
            _ => {
                // TODO
                write!(f, "  <unknown>")
            }
        }
    }
}

impl fmt::Display for ExitInstruction {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            ExitInstruction::Ret(expr) => write!(f, "  ret {}", expr),
            ExitInstruction::Jump(label) => write!(f, "  jmp {}", label),
            ExitInstruction::JumpCond {
                condition,
                label_true,
                label_false,
            } => {
                write!(f, "  jmp {}, {} {}", condition, label_true, label_false)
            }
            ExitInstruction::Placeholder => write!(f, "  <placeholder>"),
        }
    }
}

impl fmt::Display for TypedExpression {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "{}:{}", self.value, self.type_)
    }
}

impl fmt::Display for Expression {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Expression::Integer(x) => write!(f, "{}", x),
            Expression::Symbol(s) => write!(f, "{}", s),
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

fn format_binary_op(
    f: &mut fmt::Formatter<'_>,
    op: &str,
    destination: &str,
    left: &TypedExpression,
    right: &TypedExpression,
) -> fmt::Result {
    write!(f, "  {} = {} {} {}", destination, op, left, right)
}
