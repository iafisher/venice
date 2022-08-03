use super::ptree;

pub struct Program {
    declarations: Vec<Declaration>,
}

pub enum Declaration {
    Function(FunctionDeclaration),
    Const(ConstDeclaration),
    Record(RecordDeclaration),
}

pub struct FunctionDeclaration {
    name: String,
    parameters: Vec<FunctionParameter>,
    return_type: Type,
    body: Vec<Statement>,
}

pub struct FunctionParameter {
    name: String,
    type_: Type,
}

pub struct ConstDeclaration {
    symbol: String,
    value: TypedExpression,
}

pub struct RecordDeclaration {
    name: String,
    fields: Vec<RecordField>,
}

pub struct RecordField {
    name: String,
    type_: Type,
}

pub enum Statement {
    Let(LetStatement),
    Assign(AssignStatement),
    If(IfStatement),
    While(WhileStatement),
    For(ForStatement),
    Assert(AssertStatement),
}

pub struct LetStatement {
    symbol: String,
    type_: Type,
    value: TypedExpression,
}

pub struct AssignStatement {
    symbol: String,
    value: TypedExpression,
}

pub struct IfStatement {
    condition: TypedExpression,
    body: Vec<Statement>,
    else_clause: Vec<Statement>,
}

pub struct WhileStatement {
    condition: TypedExpression,
    body: Vec<Statement>,
}

pub struct ForStatement {
    symbol: String,
    symbol2: Option<String>,
    iterator: TypedExpression,
    body: Vec<Statement>,
}

pub struct AssertStatement {
    condition: TypedExpression,
}

pub struct TypedExpression {
    type_: Type,
    value: ptree::Expression,
}

pub enum Type {
    Boolean,
    I64,
    Str,
    Tuple(Vec<Type>),
    List(Box<Type>),
    Map { key: Box<Type>, value: Box<Type> },
    Record(String),
}

pub struct ParameterizedType {
    symbol: String,
    parameters: Vec<Type>,
}
