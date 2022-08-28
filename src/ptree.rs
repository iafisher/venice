// Copyright 2022 The Venice Authors. All rights reserved.
// Use of this source code is governed by an MIT-style license that can be
// found in the LICENSE file.

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
    pub name: String,
    pub parameters: Vec<FunctionParameter>,
    pub return_type: Type,
    pub body: Vec<Statement>,
    pub location: common::Location,
}

#[derive(Clone, Debug)]
pub struct FunctionInfo {
    pub stack_frame_size: usize,
}

#[derive(Debug)]
pub struct FunctionParameter {
    pub name: String,
    pub type_: Type,
}

#[derive(Debug)]
pub struct ConstDeclaration {
    pub symbol: String,
    pub type_: Type,
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
}

#[derive(Debug)]
pub struct LetStatement {
    pub symbol: String,
    pub type_: Type,
    pub value: Expression,
    pub location: common::Location,
}

#[derive(Debug)]
pub struct AssignStatement {
    pub symbol: String,
    pub value: Expression,
    pub location: common::Location,
}

#[derive(Debug)]
pub struct IfStatement {
    pub if_clause: IfClause,
    pub elif_clauses: Vec<IfClause>,
    pub else_body: Vec<Statement>,
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
    pub location: common::Location,
}

#[derive(Debug)]
pub enum ExpressionKind {
    Boolean(bool),
    Integer(i64),
    String(String),
    Symbol(String),
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
pub struct BinaryExpression {
    pub op: common::BinaryOpType,
    pub left: Box<Expression>,
    pub right: Box<Expression>,
    pub location: common::Location,
}

#[derive(Debug)]
pub struct ComparisonExpression {
    pub op: common::ComparisonOpType,
    pub left: Box<Expression>,
    pub right: Box<Expression>,
    pub location: common::Location,
}

#[derive(Debug)]
pub struct UnaryExpression {
    pub op: common::UnaryOpType,
    pub operand: Box<Expression>,
    pub location: common::Location,
}

#[derive(Debug)]
pub struct CallExpression {
    pub function: String,
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
pub struct Type {
    pub kind: TypeKind,
    pub location: common::Location,
}

#[derive(Debug)]
pub enum TypeKind {
    Literal(String),
    Parameterized(ParameterizedType),
}

#[derive(Debug)]
pub struct ParameterizedType {
    pub symbol: String,
    pub parameters: Vec<Type>,
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

            write!(f, "({} {})", parameter.name, parameter.type_)?;
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
        }
    }
}

impl fmt::Display for LetStatement {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "(let {} {} {})", self.symbol, self.type_, self.value)
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
        write!(f, "{}", self.kind)
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
        match &self.kind {
            TypeKind::Literal(s) => write!(f, "(type {})", s),
            TypeKind::Parameterized(ptype) => {
                write!(f, "(type {}", ptype.symbol)?;
                for parameter in &ptype.parameters {
                    write!(f, " {}", parameter)?;
                }
                write!(f, ")")
            }
        }
    }
}

fn format_block(f: &mut fmt::Formatter<'_>, block: &[Statement]) -> fmt::Result {
    write!(f, "(block")?;
    for stmt in block {
        write!(f, " {}", stmt)?;
    }
    write!(f, ")")
}
