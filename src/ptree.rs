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
    type_: Type,
    value: Expression,
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
    value: Expression,
}

pub struct AssignStatement {
    symbol: String,
    value: Expression,
}

pub struct IfStatement {
    if_clause: IfClause,
    elif_clauses: Vec<IfClause>,
    else_clause: Vec<Statement>,
}

pub struct IfClause {
    condition: Expression,
    body: Vec<Statement>,
}

pub struct WhileStatement {
    condition: Expression,
    body: Vec<Statement>,
}

pub struct ForStatement {
    symbol: String,
    symbol2: Option<String>,
    iterator: Expression,
    body: Vec<Statement>,
}

pub struct AssertStatement {
    condition: Expression,
}

pub enum Expression {
    Boolean(bool),
    Integer(i64),
    Str(String),
    Symbol(String),
    BinaryOp(BinaryExpression),
    UnaryOp(UnaryExpression),
    Call(CallExpression),
    Index(IndexExpression),
    TupleIndex(TupleIndexExpression),
    Attribute(AttributeExpression),
    List(ListLiteral),
    Tuple(TupleLiteral),
    Map(MapLiteral),
    Record(RecordLiteral),
}

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

pub struct BinaryExpression {
    op: BinaryOpType,
    left: Box<Expression>,
    right: Box<Expression>,
}

pub enum UnaryOpType {
    Negate,
}

pub struct UnaryExpression {
    op: UnaryOpType,
    operand: Box<Expression>,
}

pub struct CallExpression {
    function: String,
    arguments: Vec<Expression>,
}

pub struct IndexExpression {
    value: Box<Expression>,
    index: Box<Expression>,
}

pub struct TupleIndexExpression {
    value: Box<Expression>,
    index: u64,
}

pub struct AttributeExpression {
    value: Box<Expression>,
    attribute: String,
}

pub struct ListLiteral {
    items: Vec<Expression>,
}

pub struct TupleLiteral {
    items: Vec<Expression>,
}

pub struct MapLiteral {
    items: Vec<(Expression, Expression)>,
}

pub struct RecordLiteral {
    items: Vec<(String, Expression)>,
}

pub enum Type {
    Literal(String),
    Parameterized(ParameterizedType),
}

pub struct ParameterizedType {
    symbol: String,
    parameters: Vec<Type>,
}
