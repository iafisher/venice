package main

type ExpressionNode interface {
	expressionNode()
}

type StatementNode interface {
	statementNode()
}

type TypeNode interface {
	typeNode()
}

type FunctionDeclarationNode struct {
	Name       string
	Params     []*FunctionParamNode
	ReturnType TypeNode
	Body       []StatementNode
}

func (n *FunctionDeclarationNode) statementNode() {}

type FunctionParamNode struct {
	Name      string
	ParamType TypeNode
}

type SimpleTypeNode struct {
	Symbol string
}

func (n *SimpleTypeNode) typeNode() {}

type LetStatementNode struct {
	Symbol string
	Expr   ExpressionNode
}

func (n *LetStatementNode) statementNode() {}

type AssignStatementNode struct {
	Symbol string
	Expr   ExpressionNode
}

func (n *AssignStatementNode) statementNode() {}

type ReturnStatementNode struct {
	Expr ExpressionNode
}

func (n *ReturnStatementNode) statementNode() {}

type IfStatementNode struct {
	Condition   ExpressionNode
	TrueClause  []StatementNode
	FalseClause []StatementNode
}

func (n *IfStatementNode) statementNode() {}

type WhileLoopNode struct {
	Condition ExpressionNode
	Body      []StatementNode
}

func (n *WhileLoopNode) statementNode() {}

type BreakStatementNode struct{}

func (n *BreakStatementNode) statementNode() {}

type ContinueStatementNode struct{}

func (n *ContinueStatementNode) statementNode() {}

type CallNode struct {
	Function ExpressionNode
	Args     []ExpressionNode
}

func (n *CallNode) expressionNode() {}

type IndexNode struct {
	Expr  ExpressionNode
	Index ExpressionNode
}

func (n *IndexNode) expressionNode() {}

type InfixNode struct {
	Operator string
	Left     ExpressionNode
	Right    ExpressionNode
}

func (n *InfixNode) expressionNode() {}

type ListNode struct {
	Values []ExpressionNode
}

func (n *ListNode) expressionNode() {}

type MapNode struct {
	Pairs []*MapPairNode
}

func (n *MapNode) expressionNode() {}

type MapPairNode struct {
	Key   ExpressionNode
	Value ExpressionNode
}

func (n *MapPairNode) expressionNode() {}

type IntegerNode struct {
	Value int
}

func (n *IntegerNode) expressionNode() {}

type SymbolNode struct {
	Value string
}

func (n *SymbolNode) expressionNode() {}

type StringNode struct {
	Value string
}

func (n *StringNode) expressionNode() {}

type BooleanNode struct {
	Value bool
}

func (n *BooleanNode) expressionNode() {}

type ProgramNode struct {
	Statements []StatementNode
}

type ExpressionStatementNode struct {
	Expr ExpressionNode
}

func (n *ExpressionStatementNode) statementNode() {}
