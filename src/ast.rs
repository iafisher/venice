// Venice abstract syntax trees
//
// Many nodes have both `type_` and `semantic_type` fields; the former represents
// a syntactic type while the latter represents a type that has been resolved concretely
// by the type-checker. The AST emitted by the parser has only syntactic types; semantic
// types are only added by the analyzer.

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
}

#[derive(Debug)]
pub struct FunctionDeclaration {
    pub name: SymbolExpression,
    pub parameters: Vec<FunctionParameter>,
    pub return_type: SyntacticType,
    pub semantic_return_type: Type,
    pub body: Vec<Statement>,
    pub info: Option<FunctionInfo>,
    pub location: common::Location,
}

#[derive(Clone, Debug)]
pub struct FunctionInfo {
    pub stack_frame_size: usize,
}

#[derive(Debug)]
pub struct FunctionParameter {
    pub name: SymbolExpression,
    pub type_: SyntacticType,
    pub semantic_type: Type,
}

#[derive(Debug)]
pub struct ConstDeclaration {
    pub symbol: String,
    pub type_: SyntacticType,
    pub semantic_type: Type,
    pub value: Expression,
    pub location: common::Location,
}

#[derive(Debug)]
pub struct RecordDeclaration {
    pub name: String,
    pub fields: Vec<RecordField>,
    pub location: common::Location,
}

#[derive(Debug)]
pub struct RecordField {
    pub name: String,
    pub type_: SyntacticType,
    pub semantic_type: Type,
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
}

#[derive(Debug)]
pub struct LetStatement {
    pub symbol: SymbolExpression,
    pub type_: SyntacticType,
    pub semantic_type: Type,
    pub value: Expression,
    pub location: common::Location,
}

#[derive(Debug)]
pub struct AssignStatement {
    pub symbol: SymbolExpression,
    pub value: Expression,
    pub location: common::Location,
}

#[derive(Debug)]
pub struct IfStatement {
    pub if_clause: IfClause,
    pub elif_clauses: Vec<IfClause>,
    pub else_clause: Vec<Statement>,
    pub location: common::Location,
}

#[derive(Debug)]
pub struct IfClause {
    pub condition: Expression,
    pub body: Vec<Statement>,
}

#[derive(Debug)]
pub struct WhileStatement {
    pub condition: Expression,
    pub body: Vec<Statement>,
    pub location: common::Location,
}

#[derive(Debug)]
pub struct ForStatement {
    pub symbol: String,
    pub symbol2: Option<String>,
    pub iterator: Expression,
    pub body: Vec<Statement>,
    pub location: common::Location,
}

#[derive(Debug)]
pub struct ReturnStatement {
    pub value: Expression,
    pub location: common::Location,
}

#[derive(Debug)]
pub struct AssertStatement {
    pub condition: Expression,
    pub location: common::Location,
}

#[derive(Debug)]
pub struct Expression {
    pub kind: ExpressionKind,
    pub semantic_type: Type,
    pub location: common::Location,
}

#[derive(Debug)]
pub enum ExpressionKind {
    Boolean(bool),
    Integer(i64),
    String(String),
    Symbol(SymbolExpression),
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
}

#[derive(Debug)]
pub struct SymbolExpression {
    pub name: String,
    pub entry: Option<SymbolEntry>,
    pub location: common::Location,
}

#[derive(Clone, Copy, Debug)]
pub enum BinaryOpType {
    Add,
    And,
    Concat,
    Divide,
    Modulo,
    Multiply,
    Or,
    Subtract,
}

#[derive(Debug)]
pub struct BinaryExpression {
    pub op: BinaryOpType,
    pub left: Box<Expression>,
    pub right: Box<Expression>,
    pub location: common::Location,
}

#[derive(Clone, Copy, Debug)]
pub enum ComparisonOpType {
    Equals,
    GreaterThan,
    GreaterThanEquals,
    LessThan,
    LessThanEquals,
    NotEquals,
}

#[derive(Debug)]
pub struct ComparisonExpression {
    pub op: ComparisonOpType,
    pub left: Box<Expression>,
    pub right: Box<Expression>,
    pub location: common::Location,
}

#[derive(Debug)]
pub enum UnaryOpType {
    Negate,
    Not,
}

#[derive(Debug)]
pub struct UnaryExpression {
    pub op: UnaryOpType,
    pub operand: Box<Expression>,
    pub location: common::Location,
}

#[derive(Debug)]
pub struct CallExpression {
    pub function: SymbolExpression,
    pub arguments: Vec<Expression>,
    pub location: common::Location,
}

#[derive(Debug)]
pub struct IndexExpression {
    pub value: Box<Expression>,
    pub index: Box<Expression>,
    pub location: common::Location,
}

#[derive(Debug)]
pub struct TupleIndexExpression {
    pub value: Box<Expression>,
    pub index: usize,
    pub location: common::Location,
}

#[derive(Debug)]
pub struct AttributeExpression {
    pub value: Box<Expression>,
    pub attribute: String,
    pub location: common::Location,
}

#[derive(Debug)]
pub struct ListLiteral {
    pub items: Vec<Expression>,
    pub location: common::Location,
}

#[derive(Debug)]
pub struct TupleLiteral {
    pub items: Vec<Expression>,
    pub location: common::Location,
}

#[derive(Debug)]
pub struct MapLiteral {
    pub items: Vec<(Expression, Expression)>,
    pub location: common::Location,
}

#[derive(Debug)]
pub struct RecordLiteral {
    pub name: String,
    pub items: Vec<(String, Expression)>,
    pub location: common::Location,
}

#[derive(Debug)]
pub struct SyntacticType {
    pub kind: SyntacticTypeKind,
    pub location: common::Location,
}

#[derive(Debug)]
pub enum SyntacticTypeKind {
    Literal(String),
    Parameterized(SyntacticParameterizedType),
}

#[derive(Debug)]
pub struct SyntacticParameterizedType {
    pub symbol: String,
    pub parameters: Vec<SyntacticType>,
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

            match parameter.semantic_type {
                Type::Unknown => {
                    write!(f, "({} {})", parameter.name, parameter.type_)?;
                }
                _ => {
                    write!(f, "{}:{}", parameter.name, parameter.semantic_type)?;
                }
            }
        }

        write!(f, ") ")?;
        match self.semantic_return_type {
            Type::Unknown => {
                write!(f, "{}", self.return_type)?;
            }
            _ => {
                write!(f, "{}", self.semantic_return_type)?;
            }
        }

        for statement in &self.body {
            write!(f, " {}", statement)?;
        }
        write!(f, ")")
    }
}

impl fmt::Display for ConstDeclaration {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self.semantic_type {
            Type::Unknown => {
                write!(f, "(const {} {} {})", self.symbol, self.type_, self.value)
            }
            _ => {
                write!(
                    f,
                    "(const {} {} {})",
                    self.symbol, self.semantic_type, self.value
                )
            }
        }
    }
}

impl fmt::Display for RecordDeclaration {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "(record-decl {}", self.name)?;
        for field in &self.fields {
            match field.semantic_type {
                Type::Unknown => {
                    write!(f, "({} {})", field.name, field.type_)?;
                }
                _ => {
                    write!(f, "({} {})", field.name, field.semantic_type)?;
                }
            }
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
        }
    }
}

impl fmt::Display for LetStatement {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self.semantic_type {
            Type::Unknown => {
                write!(f, "(let {} {} {})", self.symbol, self.type_, self.value)
            }
            _ => {
                write!(
                    f,
                    "(let {}:{} {})",
                    self.symbol, self.semantic_type, self.value
                )
            }
        }
    }
}

impl fmt::Display for AssignStatement {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "(assign {} {})", self.symbol, self.value)
    }
}

impl fmt::Display for IfStatement {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "(if {} ", self.if_clause.condition)?;
        format_block(f, &self.if_clause.body)?;

        for elif_clause in &self.elif_clauses {
            write!(f, " (elif {} ", elif_clause.condition)?;
            format_block(f, &elif_clause.body)?;
            write!(f, ")")?;
        }

        if self.else_clause.len() > 0 {
            write!(f, " (else ")?;
            format_block(f, &self.else_clause)?;
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
        match self.semantic_type {
            Type::Unknown => {
                write!(f, "{}", self.kind)
            }
            _ => {
                write!(f, "({}):{}", self.kind, self.semantic_type)
            }
        }
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
        }
    }
}

impl fmt::Display for SymbolExpression {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "{}", self.name)
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

impl fmt::Display for SyntacticType {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match &self.kind {
            SyntacticTypeKind::Literal(s) => write!(f, "(type {})", s),
            SyntacticTypeKind::Parameterized(ptype) => {
                write!(f, "(type {}", ptype.symbol)?;
                for parameter in &ptype.parameters {
                    write!(f, " {}", parameter)?;
                }
                write!(f, ")")
            }
        }
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

fn format_block(f: &mut fmt::Formatter<'_>, block: &Vec<Statement>) -> fmt::Result {
    write!(f, "(block")?;
    for stmt in block {
        write!(f, " {}", stmt)?;
    }
    write!(f, ")")
}
