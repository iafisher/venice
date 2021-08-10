package main

type Node interface {
	getLocation() *Location
}

type ExpressionNode interface {
	Node
	expressionNode()
}

type StatementNode interface {
	Node
	statementNode()
}

type TypeNode interface {
	Node
	typeNode()
}

type FunctionDeclarationNode struct {
	Name       string
	Params     []*FunctionParamNode
	ReturnType TypeNode
	Body       []StatementNode
	Location   *Location
}

func (n *FunctionDeclarationNode) statementNode() {}

func (n *FunctionDeclarationNode) getLocation() *Location {
	return n.Location
}

type FunctionParamNode struct {
	Name      string
	ParamType TypeNode
	Location  *Location
}

type ClassDeclarationNode struct {
	Name                 string
	GenericTypeParameter string
	Fields               []*ClassFieldNode
	Location             *Location
}

func (n *ClassDeclarationNode) statementNode() {}

func (n *ClassDeclarationNode) getLocation() *Location {
	return n.Location
}

type EnumDeclarationNode struct {
	Name     string
	Cases    []*EnumCaseNode
	Location *Location
}

func (n *EnumDeclarationNode) statementNode() {}

func (n *EnumDeclarationNode) getLocation() *Location {
	return n.Location
}

type EnumCaseNode struct {
	Label    string
	Types    []TypeNode
	Location *Location
}

func (n *EnumCaseNode) typeNode() {}

func (n *EnumCaseNode) getLocation() *Location {
	return n.Location
}

type ClassFieldNode struct {
	Name      string
	Public    bool
	FieldType TypeNode
}

type SimpleTypeNode struct {
	Symbol   string
	Location *Location
}

func (n *SimpleTypeNode) typeNode() {}

func (n *SimpleTypeNode) getLocation() *Location {
	return n.Location
}

type LetStatementNode struct {
	Symbol   string
	Expr     ExpressionNode
	Location *Location
}

func (n *LetStatementNode) statementNode() {}

func (n *LetStatementNode) getLocation() *Location {
	return n.Location
}

type AssignStatementNode struct {
	Symbol   string
	Expr     ExpressionNode
	Location *Location
}

func (n *AssignStatementNode) statementNode() {}

func (n *AssignStatementNode) getLocation() *Location {
	return n.Location
}

type ReturnStatementNode struct {
	Expr     ExpressionNode
	Location *Location
}

func (n *ReturnStatementNode) statementNode() {}

func (n *ReturnStatementNode) getLocation() *Location {
	return n.Location
}

type IfStatementNode struct {
	Condition   ExpressionNode
	TrueClause  []StatementNode
	FalseClause []StatementNode
	Location    *Location
}

func (n *IfStatementNode) statementNode() {}

func (n *IfStatementNode) getLocation() *Location {
	return n.Location
}

type WhileLoopNode struct {
	Condition ExpressionNode
	Body      []StatementNode
	Location  *Location
}

func (n *WhileLoopNode) statementNode() {}

func (n *WhileLoopNode) getLocation() *Location {
	return n.Location
}

type BreakStatementNode struct {
	Location *Location
}

func (n *BreakStatementNode) statementNode() {}

func (n *BreakStatementNode) getLocation() *Location {
	return n.Location
}

type ContinueStatementNode struct {
	Location *Location
}

func (n *ContinueStatementNode) statementNode() {}

func (n *ContinueStatementNode) getLocation() *Location {
	return n.Location
}

type CallNode struct {
	Function ExpressionNode
	Args     []ExpressionNode
	Location *Location
}

func (n *CallNode) expressionNode() {}

func (n *CallNode) getLocation() *Location {
	return n.Location
}

type IndexNode struct {
	Expr     ExpressionNode
	Index    ExpressionNode
	Location *Location
}

func (n *IndexNode) expressionNode() {}

func (n *IndexNode) getLocation() *Location {
	return n.Location
}

type InfixNode struct {
	Operator string
	Left     ExpressionNode
	Right    ExpressionNode
	Location *Location
}

func (n *InfixNode) expressionNode() {}

func (n *InfixNode) getLocation() *Location {
	return n.Location
}

type ListNode struct {
	Values   []ExpressionNode
	Location *Location
}

func (n *ListNode) expressionNode() {}

func (n *ListNode) getLocation() *Location {
	return n.Location
}

type MapNode struct {
	Pairs    []*MapPairNode
	Location *Location
}

func (n *MapNode) expressionNode() {}

func (n *MapNode) getLocation() *Location {
	return n.Location
}

type MapPairNode struct {
	Key      ExpressionNode
	Value    ExpressionNode
	Location *Location
}

type FieldAccessNode struct {
	Expr     ExpressionNode
	Name     string
	Location *Location
}

func (n *FieldAccessNode) expressionNode() {}

func (n *FieldAccessNode) getLocation() *Location {
	return n.Location
}

type IntegerNode struct {
	Value    int
	Location *Location
}

func (n *IntegerNode) expressionNode() {}

func (n *IntegerNode) getLocation() *Location {
	return n.Location
}

type SymbolNode struct {
	Value    string
	Location *Location
}

func (n *SymbolNode) expressionNode() {}

func (n *SymbolNode) getLocation() *Location {
	return n.Location
}

type EnumSymbolNode struct {
	Enum     string
	Case     string
	Location *Location
}

func (n *EnumSymbolNode) expressionNode() {}

func (n *EnumSymbolNode) getLocation() *Location {
	return n.Location
}

type StringNode struct {
	Value    string
	Location *Location
}

func (n *StringNode) expressionNode() {}

func (n *StringNode) getLocation() *Location {
	return n.Location
}

type BooleanNode struct {
	Value    bool
	Location *Location
}

func (n *BooleanNode) expressionNode() {}

func (n *BooleanNode) getLocation() *Location {
	return n.Location
}

type ProgramNode struct {
	Statements []StatementNode
}

type ExpressionStatementNode struct {
	Expr     ExpressionNode
	Location *Location
}

func (n *ExpressionStatementNode) statementNode() {}

func (n *ExpressionStatementNode) getLocation() *Location {
	return n.Location
}
