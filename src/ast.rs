use super::common;
use std::fmt;

#[derive(Debug)]
pub struct Program {
    pub declarations: Vec<Declaration>,
}

#[derive(Debug)]
pub enum Declaration {
    Function(FunctionDeclaration),
    Const(ConstDeclaration),
    Record(RecordDeclaration),
    Error,
}

#[derive(Debug)]
pub struct FunctionDeclaration {
    pub name: SymbolEntry,
    pub parameters: Vec<FunctionParameter>,
    pub return_type: Type,
    pub body: Vec<Statement>,
    pub info: FunctionInfo,
}

#[derive(Clone, Debug)]
pub struct FunctionInfo {
    pub stack_frame_size: usize,
}

#[derive(Debug)]
pub struct FunctionParameter {
    pub name: SymbolEntry,
    pub type_: Type,
}

#[derive(Debug)]
pub struct ConstDeclaration {
    pub symbol: SymbolEntry,
    pub type_: Type,
    pub value: Expression,
}

#[derive(Debug)]
pub struct RecordDeclaration {
    pub name: SymbolEntry,
    pub fields: Vec<RecordField>,
}

#[derive(Debug)]
pub struct RecordField {
    pub name: String,
    pub type_: Type,
}

#[derive(Debug)]
pub enum Statement {
    Assert(AssertStatement),
    Assign(AssignStatement),
    Expression(Expression),
    For(ForStatement),
    If(IfStatement),
    Let(LetStatement),
    Return(ReturnStatement),
    While(WhileStatement),
    Error,
}

#[derive(Debug)]
pub struct LetStatement {
    pub symbol: SymbolEntry,
    pub type_: Type,
    pub value: Expression,
}

#[derive(Debug)]
pub struct AssignStatement {
    pub symbol: SymbolEntry,
    pub value: Expression,
}

#[derive(Debug)]
pub struct IfStatement {
    pub condition: Expression,
    pub body: Vec<Statement>,
    pub else_body: Vec<Statement>,
}

#[derive(Debug)]
pub struct WhileStatement {
    pub condition: Expression,
    pub body: Vec<Statement>,
}

#[derive(Debug)]
pub struct ForStatement {
    pub symbol: SymbolEntry,
    pub symbol2: Option<SymbolEntry>,
    pub iterator: Expression,
    pub body: Vec<Statement>,
}

#[derive(Debug)]
pub struct ReturnStatement {
    pub value: Expression,
}

#[derive(Debug)]
pub struct AssertStatement {
    pub condition: Expression,
}

#[derive(Clone, Debug)]
pub struct Expression {
    pub kind: ExpressionKind,
    pub type_: Type,
}

#[derive(Clone, Debug)]
pub enum ExpressionKind {
    Boolean(bool),
    Integer(i64),
    String(String),
    Symbol(SymbolEntry),
    Binary(BinaryExpression),
    Comparison(ComparisonExpression),
    Unary(UnaryExpression),
    Call(CallExpression),
    Index(IndexExpression),
    TupleIndex(TupleIndexExpression),
    Attribute(AttributeExpression),
    List(ListLiteral),
    Tuple(TupleLiteral),
    Map(MapLiteral),
    Record(RecordLiteral),
    Error,
}

pub const EXPRESSION_ERROR: Expression = Expression {
    kind: ExpressionKind::Error,
    type_: Type::Error,
};

#[derive(Clone, Debug)]
pub struct BinaryExpression {
    pub op: common::BinaryOpType,
    pub left: Box<Expression>,
    pub right: Box<Expression>,
}

#[derive(Clone, Debug)]
pub struct ComparisonExpression {
    pub op: common::ComparisonOpType,
    pub left: Box<Expression>,
    pub right: Box<Expression>,
}

#[derive(Clone, Debug)]
pub struct UnaryExpression {
    pub op: common::UnaryOpType,
    pub operand: Box<Expression>,
}

#[derive(Clone, Debug)]
pub struct CallExpression {
    pub function: SymbolEntry,
    pub arguments: Vec<Expression>,
}

#[derive(Clone, Debug)]
pub struct IndexExpression {
    pub value: Box<Expression>,
    pub index: Box<Expression>,
}

#[derive(Clone, Debug)]
pub struct TupleIndexExpression {
    pub value: Box<Expression>,
    pub index: usize,
}

#[derive(Clone, Debug)]
pub struct AttributeExpression {
    pub value: Box<Expression>,
    pub attribute: String,
}

#[derive(Clone, Debug)]
pub struct ListLiteral {
    pub items: Vec<Expression>,
}

#[derive(Clone, Debug)]
pub struct TupleLiteral {
    pub items: Vec<Expression>,
}

#[derive(Clone, Debug)]
pub struct MapLiteral {
    pub items: Vec<(Expression, Expression)>,
}

#[derive(Clone, Debug)]
pub struct RecordLiteral {
    pub name: SymbolEntry,
    pub items: Vec<(String, Expression)>,
}

#[derive(Clone, Debug)]
pub enum Type {
    Boolean,
    I64,
    String,
    Void,
    Tuple(Vec<Type>),
    List(Box<Type>),
    Map {
        key: Box<Type>,
        value: Box<Type>,
    },
    Function {
        parameters: Vec<Type>,
        return_type: Box<Type>,
    },
    Record(String),
    Unknown,
    Error,
}

#[derive(Clone, Debug)]
pub struct ParameterizedType {
    symbol: String,
    parameters: Vec<Type>,
}

#[derive(Clone, Debug)]
pub struct SymbolEntry {
    pub unique_name: String,
    pub type_: Type,
    pub constant: bool,
    pub external: bool,
}

impl Type {
    pub fn matches(&self, other: &Type) -> bool {
        match (self, other) {
            (Type::Boolean, Type::Boolean) => true,
            (Type::I64, Type::I64) => true,
            (Type::String, Type::String) => true,
            (Type::Tuple(ts1), Type::Tuple(ts2)) => {
                if ts1.len() != ts2.len() {
                    return false;
                }
                for (t1, t2) in ts1.iter().zip(ts2.iter()) {
                    if !t1.matches(t2) {
                        return false;
                    }
                }
                true
            }
            _ => false,
        }
    }

    pub fn stack_size(&self) -> usize {
        match self {
            Type::Void | Type::Unknown | Type::Error => 0,
            // Everything else is either a primitive value or a pointer.
            _ => 8,
        }
    }
}

impl fmt::Display for Program {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "(program")?;
        for declaration in &self.declarations {
            write!(f, " {}", declaration)?;
        }
        write!(f, ")")
    }
}

impl fmt::Display for Declaration {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Declaration::Function(declaration) => write!(f, "{}", declaration),
            Declaration::Const(declaration) => write!(f, "{}", declaration),
            Declaration::Record(declaration) => write!(f, "{}", declaration),
            Declaration::Error => write!(f, "error"),
        }
    }
}

impl fmt::Display for FunctionDeclaration {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "(func {} (", self.name)?;
        for (i, parameter) in self.parameters.iter().enumerate() {
            if i != 0 {
                write!(f, " ")?;
            }

            write!(f, "{}:{}", parameter.name, parameter.type_)?;
        }

        write!(f, ") ")?;
        write!(f, "{}", self.return_type)?;

        for statement in &self.body {
            write!(f, " {}", statement)?;
        }
        write!(f, ")")
    }
}

impl fmt::Display for ConstDeclaration {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "(const {} {} {})", self.symbol, self.type_, self.value)
    }
}

impl fmt::Display for RecordDeclaration {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "(record-decl {}", self.name)?;
        for field in &self.fields {
            write!(f, "({} {})", field.name, field.type_)?;
        }
        write!(f, ")")
    }
}

impl fmt::Display for Statement {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Statement::Let(stmt) => write!(f, "{}", stmt),
            Statement::Assign(stmt) => write!(f, "{}", stmt),
            Statement::If(stmt) => write!(f, "{}", stmt),
            Statement::While(stmt) => write!(f, "{}", stmt),
            Statement::For(stmt) => write!(f, "{}", stmt),
            Statement::Return(stmt) => write!(f, "{}", stmt),
            Statement::Assert(stmt) => write!(f, "{}", stmt),
            Statement::Expression(stmt) => write!(f, "{}", stmt),
            Statement::Error => write!(f, "error"),
        }
    }
}

impl fmt::Display for LetStatement {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "(let {}:{} {})", self.symbol, self.type_, self.value)
    }
}

impl fmt::Display for AssignStatement {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "(assign {} {})", self.symbol, self.value)
    }
}

impl fmt::Display for IfStatement {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "(if {} ", self.condition)?;
        format_block(f, &self.body)?;
        if !self.else_body.is_empty() {
            write!(f, " (else ")?;
            format_block(f, &self.else_body)?;
            write!(f, ")")?;
        }

        fmt::Result::Ok(())
    }
}

impl fmt::Display for WhileStatement {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "(while {} ", self.condition)?;
        format_block(f, &self.body)?;
        write!(f, ")")
    }
}

impl fmt::Display for ForStatement {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "(for ")?;
        if let Some(symbol2) = &self.symbol2 {
            write!(f, "({} {})", self.symbol, symbol2)?;
        } else {
            write!(f, "{}", self.symbol)?;
        }
        write!(f, " {}", self.iterator)?;
        format_block(f, &self.body)
    }
}

impl fmt::Display for ReturnStatement {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "(return {})", self.value)
    }
}

impl fmt::Display for AssertStatement {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "(assert {})", self.condition)
    }
}

impl fmt::Display for Expression {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "({}):{}", self.kind, self.type_)
    }
}

impl fmt::Display for ExpressionKind {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            ExpressionKind::Boolean(e) => write!(f, "{}", e),
            ExpressionKind::Integer(e) => write!(f, "{}", e),
            ExpressionKind::String(e) => write!(f, "{:?}", e),
            ExpressionKind::Symbol(e) => write!(f, "{}", e),
            ExpressionKind::Binary(e) => write!(f, "{}", e),
            ExpressionKind::Comparison(e) => write!(f, "{}", e),
            ExpressionKind::Unary(e) => write!(f, "{}", e),
            ExpressionKind::Call(e) => write!(f, "{}", e),
            ExpressionKind::Index(e) => write!(f, "{}", e),
            ExpressionKind::TupleIndex(e) => write!(f, "{}", e),
            ExpressionKind::Attribute(e) => write!(f, "{}", e),
            ExpressionKind::List(e) => write!(f, "{}", e),
            ExpressionKind::Tuple(e) => write!(f, "{}", e),
            ExpressionKind::Map(e) => write!(f, "{}", e),
            ExpressionKind::Record(e) => write!(f, "{}", e),
            ExpressionKind::Error => write!(f, "error"),
        }
    }
}

impl fmt::Display for BinaryExpression {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "(binary {:?} {} {})", self.op, self.left, self.right)
    }
}

impl fmt::Display for ComparisonExpression {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "(cmp {:?} {} {})", self.op, self.left, self.right)
    }
}

impl fmt::Display for UnaryExpression {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "(unary {:?} {})", self.op, self.operand)
    }
}

impl fmt::Display for CallExpression {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "(call {} (", self.function)?;
        for (i, argument) in self.arguments.iter().enumerate() {
            if i != 0 {
                write!(f, " ")?;
            }
            write!(f, "{}", argument)?;
        }
        write!(f, "))")
    }
}

impl fmt::Display for IndexExpression {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "(index {} {})", self.value, self.index)
    }
}

impl fmt::Display for TupleIndexExpression {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "(tuple-index {} {})", self.value, self.index)
    }
}

impl fmt::Display for AttributeExpression {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "(attrib {} {})", self.value, self.attribute)
    }
}

impl fmt::Display for ListLiteral {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "(list")?;
        for item in &self.items {
            write!(f, " {}", item)?;
        }
        write!(f, ")")
    }
}

impl fmt::Display for TupleLiteral {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "(list")?;
        for item in &self.items {
            write!(f, " {}", item)?;
        }
        write!(f, ")")
    }
}

impl fmt::Display for MapLiteral {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "(list")?;
        for item in &self.items {
            write!(f, " ({} {})", item.0, item.1)?;
        }
        write!(f, ")")
    }
}

impl fmt::Display for RecordLiteral {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "(record {}", self.name)?;
        for item in &self.items {
            write!(f, " ({} {})", item.0, item.1)?;
        }
        write!(f, ")")
    }
}

impl fmt::Display for Type {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Type::I64 => write!(f, "i64"),
            Type::Boolean => write!(f, "bool"),
            Type::String => write!(f, "string"),
            Type::Void => write!(f, "void"),
            Type::Tuple(ts) => {
                write!(f, "(")?;
                for (i, t) in ts.iter().enumerate() {
                    write!(f, "{}", t)?;
                    if i != ts.len() - 1 {
                        write!(f, ", ")?;
                    }
                }
                write!(f, ")")
            }
            Type::List(t) => {
                write!(f, "list<{}>", t)
            }
            Type::Map { key, value } => {
                write!(f, "map<{}, {}>", key, value)
            }
            Type::Function {
                parameters,
                return_type,
            } => {
                write!(f, "func<")?;
                for t in parameters {
                    write!(f, "{}, ", t)?;
                }
                write!(f, "{}>", return_type)
            }
            Type::Record(name) => write!(f, "{}", name),
            Type::Unknown => write!(f, "unknown"),
            Type::Error => write!(f, "unknown"),
        }
    }
}

impl fmt::Display for SymbolEntry {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "{}", self.unique_name)
    }
}

fn format_block(f: &mut fmt::Formatter<'_>, block: &[Statement]) -> fmt::Result {
    write!(f, "(block")?;
    for stmt in block {
        write!(f, " {}", stmt)?;
    }
    write!(f, ")")
}
