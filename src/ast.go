package main

type Expression interface {
	expressionNode()
}

type Statement interface {
	statementNode()
}

type TypeNode interface {
	typeNode()
}

type FunctionDeclarationNode struct {
	Name       string
	Params     []*FunctionParamNode
	ReturnType TypeNode
	Body       []Statement
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
	Expr   Expression
}

func (n *LetStatementNode) statementNode() {}

type ReturnStatementNode struct {
	Expr Expression
}

func (n *ReturnStatementNode) statementNode() {}

type IfStatementNode struct {
	Condition   Expression
	TrueClause  []Statement
	FalseClause []Statement
}

func (n *IfStatementNode) statementNode() {}

type WhileLoopNode struct {
	Condition Expression
	Body      []Statement
}

func (n *WhileLoopNode) statementNode() {}

type BreakStatementNode struct{}

func (n *BreakStatementNode) statementNode() {}

type ContinueStatementNode struct{}

func (n *ContinueStatementNode) statementNode() {}

type CallNode struct {
	Function Expression
	Args     []Expression
}

func (n *CallNode) expressionNode() {}

type IndexNode struct {
	Expr  Expression
	Index Expression
}

func (n *IndexNode) expressionNode() {}

type InfixNode struct {
	Operator string
	Left     Expression
	Right    Expression
}

func (n *InfixNode) expressionNode() {}

type ListNode struct {
	Values []Expression
}

func (n *ListNode) expressionNode() {}

type MapNode struct {
	Pairs []*MapPairNode
}

func (n *MapNode) expressionNode() {}

type MapPairNode struct {
	Key   Expression
	Value Expression
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
	Statements []Statement
}

type ExpressionStatementNode struct {
	Expression Expression
}

func (n *ExpressionStatementNode) statementNode() {}
