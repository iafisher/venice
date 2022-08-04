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
}

#[derive(Debug)]
pub struct RecordDeclaration {
    pub name: String,
    pub fields: Vec<RecordField>,
}

#[derive(Debug)]
pub struct RecordField {
    pub name: String,
    pub type_: Type,
}

#[derive(Debug)]
pub enum Statement {
    Let(LetStatement),
    Assign(AssignStatement),
    If(IfStatement),
    While(WhileStatement),
    For(ForStatement),
    Return(ReturnStatement),
    Assert(AssertStatement),
}

#[derive(Debug)]
pub struct LetStatement {
    pub symbol: String,
    pub type_: Type,
    pub value: Expression,
}

#[derive(Debug)]
pub struct AssignStatement {
    pub symbol: String,
    pub value: Expression,
}

#[derive(Debug)]
pub struct IfStatement {
    pub if_clause: IfClause,
    pub elif_clauses: Vec<IfClause>,
    pub else_clause: Vec<Statement>,
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
}

#[derive(Debug)]
pub struct ForStatement {
    pub symbol: String,
    pub symbol2: Option<String>,
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

#[derive(Debug)]
pub enum Expression {
    Boolean(bool),
    Integer(i64),
    Str(String),
    Symbol(String),
    Binary(BinaryExpression),
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
pub enum BinaryOpType {
    Add,
    Subtract,
    Divide,
    Modulo,
    Multiply,
    Concat,
    Or,
    And,
    LessThan,
    LessThanEquals,
    GreaterThan,
    GreaterThanEquals,
    Equals,
    NotEquals,
}

#[derive(Debug)]
pub struct BinaryExpression {
    pub op: BinaryOpType,
    pub left: Box<Expression>,
    pub right: Box<Expression>,
}

#[derive(Debug)]
pub enum UnaryOpType {
    Negate,
}

#[derive(Debug)]
pub struct UnaryExpression {
    pub op: UnaryOpType,
    pub operand: Box<Expression>,
}

#[derive(Debug)]
pub struct CallExpression {
    pub function: String,
    pub arguments: Vec<Expression>,
}

#[derive(Debug)]
pub struct IndexExpression {
    pub value: Box<Expression>,
    pub index: Box<Expression>,
}

#[derive(Debug)]
pub struct TupleIndexExpression {
    pub value: Box<Expression>,
    pub index: u64,
}

#[derive(Debug)]
pub struct AttributeExpression {
    pub value: Box<Expression>,
    pub attribute: String,
}

#[derive(Debug)]
pub struct ListLiteral {
    pub items: Vec<Expression>,
}

#[derive(Debug)]
pub struct TupleLiteral {
    pub items: Vec<Expression>,
}

#[derive(Debug)]
pub struct MapLiteral {
    pub items: Vec<(Expression, Expression)>,
}

#[derive(Debug)]
pub struct RecordLiteral {
    pub name: String,
    pub items: Vec<(String, Expression)>,
}

#[derive(Debug)]
pub enum Type {
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
        for parameter in &self.parameters {
            write!(f, " ({} {})", parameter.name, parameter.type_)?;
        }
        write!(f, ") {}", self.return_type)?;
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

        if self.else_clause.len() > 0 {
            write!(f, "(else ")?;
            format_block(f, &self.else_clause)?;
            write!(f, ")")?;
        }

        fmt::Result::Ok(())
    }
}

impl fmt::Display for WhileStatement {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "(while {}", self.condition)?;
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
        match self {
            Expression::Boolean(e) => write!(f, "{}", e),
            Expression::Integer(e) => write!(f, "{}", e),
            Expression::Str(e) => write!(f, "{}", e),
            Expression::Symbol(e) => write!(f, "{}", e),
            Expression::Binary(e) => write!(f, "{}", e),
            Expression::Unary(e) => write!(f, "{}", e),
            Expression::Call(e) => write!(f, "{}", e),
            Expression::Index(e) => write!(f, "{}", e),
            Expression::TupleIndex(e) => write!(f, "{}", e),
            Expression::Attribute(e) => write!(f, "{}", e),
            Expression::List(e) => write!(f, "{}", e),
            Expression::Tuple(e) => write!(f, "{}", e),
            Expression::Map(e) => write!(f, "{}", e),
            Expression::Record(e) => write!(f, "{}", e),
        }
    }
}

impl fmt::Display for BinaryExpression {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "(binary {:?} {} {})", self.op, self.left, self.right)
    }
}

impl fmt::Display for UnaryExpression {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "(unary {:?} {})", self.op, self.operand)
    }
}

impl fmt::Display for CallExpression {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "(call {}", self.function)?;
        for argument in &self.arguments {
            write!(f, " {}", argument)?;
        }
        write!(f, ")")
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
            Type::Literal(s) => write!(f, "(type {})", s),
            Type::Parameterized(ptype) => {
                write!(f, "(type {}", ptype.symbol)?;
                for parameter in &ptype.parameters {
                    write!(f, " {}", parameter)?;
                }
                write!(f, ")")
            }
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
