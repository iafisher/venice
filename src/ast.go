package main

/**
 * AST interface declarations
 */

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

// Root node - does not implement Node
type ProgramNode struct {
	Statements []StatementNode
}

/**
 * Statement nodes
 */

type AssignStatementNode struct {
	Symbol   string
	Expr     ExpressionNode
	Location *Location
}

type BreakStatementNode struct {
	Location *Location
}

type ClassDeclarationNode struct {
	Name                 string
	GenericTypeParameter string
	Fields               []*ClassFieldNode
	Location             *Location
}

// Helper struct - does not implement Node
type ClassFieldNode struct {
	Name      string
	Public    bool
	FieldType TypeNode
}

type ContinueStatementNode struct {
	Location *Location
}

type EnumDeclarationNode struct {
	Name                 string
	GenericTypeParameter string
	Cases                []*EnumCaseNode
	Location             *Location
}

// Helper struct - does not implement Node
type EnumCaseNode struct {
	Label    string
	Types    []TypeNode
	Location *Location
}

type ExpressionStatementNode struct {
	Expr     ExpressionNode
	Location *Location
}

type ForLoopNode struct {
	Variables []string
	Iterable  ExpressionNode
	Body      []StatementNode
	Location  *Location
}

type FunctionDeclarationNode struct {
	Name       string
	Params     []*FunctionParamNode
	ReturnType TypeNode
	Body       []StatementNode
	Location   *Location
}

// Helper struct - does not implement Node
type FunctionParamNode struct {
	Name      string
	ParamType TypeNode
	Location  *Location
}

type IfStatementNode struct {
	Condition   ExpressionNode
	TrueClause  []StatementNode
	FalseClause []StatementNode
	Location    *Location
}

type LetStatementNode struct {
	Symbol   string
	Expr     ExpressionNode
	Location *Location
}

type ReturnStatementNode struct {
	Expr     ExpressionNode
	Location *Location
}

type WhileLoopNode struct {
	Condition ExpressionNode
	Body      []StatementNode
	Location  *Location
}

/**
 * Expression nodes
 */

type BooleanNode struct {
	Value    bool
	Location *Location
}

type CallNode struct {
	Function ExpressionNode
	Args     []ExpressionNode
	Location *Location
}

type CharacterNode struct {
	Value    byte
	Location *Location
}

type EnumSymbolNode struct {
	Enum     string
	Case     string
	Location *Location
}

type FieldAccessNode struct {
	Expr     ExpressionNode
	Name     string
	Location *Location
}

type IndexNode struct {
	Expr     ExpressionNode
	Index    ExpressionNode
	Location *Location
}

type InfixNode struct {
	Operator string
	Left     ExpressionNode
	Right    ExpressionNode
	Location *Location
}

type IntegerNode struct {
	Value    int
	Location *Location
}

type ListNode struct {
	Values   []ExpressionNode
	Location *Location
}

type StringNode struct {
	Value    string
	Location *Location
}

type TupleNode struct {
	Values   []ExpressionNode
	Location *Location
}

type TupleFieldAccessNode struct {
	Expr     ExpressionNode
	Index    int
	Location *Location
}

type MapNode struct {
	Pairs    []*MapPairNode
	Location *Location
}

// Helper struct - does not implement Node
type MapPairNode struct {
	Key      ExpressionNode
	Value    ExpressionNode
	Location *Location
}

type SymbolNode struct {
	Value    string
	Location *Location
}

/**
 * Type nodes
 */

type SimpleTypeNode struct {
	Symbol   string
	Location *Location
}

/**
 * getLocation() implementations
 */

func (n *AssignStatementNode) getLocation() *Location {
	return n.Location
}

func (n *BooleanNode) getLocation() *Location {
	return n.Location
}

func (n *BreakStatementNode) getLocation() *Location {
	return n.Location
}

func (n *CallNode) getLocation() *Location {
	return n.Location
}

func (n *CharacterNode) getLocation() *Location {
	return n.Location
}

func (n *ClassDeclarationNode) getLocation() *Location {
	return n.Location
}

func (n *ContinueStatementNode) getLocation() *Location {
	return n.Location
}

func (n *EnumDeclarationNode) getLocation() *Location {
	return n.Location
}

func (n *EnumSymbolNode) getLocation() *Location {
	return n.Location
}

func (n *ExpressionStatementNode) getLocation() *Location {
	return n.Location
}

func (n *FieldAccessNode) getLocation() *Location {
	return n.Location
}

func (n *ForLoopNode) getLocation() *Location {
	return n.Location
}

func (n *FunctionDeclarationNode) getLocation() *Location {
	return n.Location
}

func (n *IfStatementNode) getLocation() *Location {
	return n.Location
}

func (n *IndexNode) getLocation() *Location {
	return n.Location
}

func (n *InfixNode) getLocation() *Location {
	return n.Location
}

func (n *IntegerNode) getLocation() *Location {
	return n.Location
}

func (n *LetStatementNode) getLocation() *Location {
	return n.Location
}

func (n *ListNode) getLocation() *Location {
	return n.Location
}

func (n *MapNode) getLocation() *Location {
	return n.Location
}

func (n *ReturnStatementNode) getLocation() *Location {
	return n.Location
}

func (n *SimpleTypeNode) getLocation() *Location {
	return n.Location
}

func (n *StringNode) getLocation() *Location {
	return n.Location
}

func (n *SymbolNode) getLocation() *Location {
	return n.Location
}

func (n *TupleFieldAccessNode) getLocation() *Location {
	return n.Location
}

func (n *TupleNode) getLocation() *Location {
	return n.Location
}

func (n *WhileLoopNode) getLocation() *Location {
	return n.Location
}

func (n *AssignStatementNode) statementNode()     {}
func (n *BreakStatementNode) statementNode()      {}
func (n *ClassDeclarationNode) statementNode()    {}
func (n *ContinueStatementNode) statementNode()   {}
func (n *EnumDeclarationNode) statementNode()     {}
func (n *ExpressionStatementNode) statementNode() {}
func (n *ForLoopNode) statementNode()             {}
func (n *FunctionDeclarationNode) statementNode() {}
func (n *IfStatementNode) statementNode()         {}
func (n *LetStatementNode) statementNode()        {}
func (n *ReturnStatementNode) statementNode()     {}
func (n *WhileLoopNode) statementNode()           {}

func (n *BooleanNode) expressionNode()          {}
func (n *CallNode) expressionNode()             {}
func (n *CharacterNode) expressionNode()        {}
func (n *EnumSymbolNode) expressionNode()       {}
func (n *FieldAccessNode) expressionNode()      {}
func (n *IndexNode) expressionNode()            {}
func (n *InfixNode) expressionNode()            {}
func (n *IntegerNode) expressionNode()          {}
func (n *ListNode) expressionNode()             {}
func (n *MapNode) expressionNode()              {}
func (n *StringNode) expressionNode()           {}
func (n *SymbolNode) expressionNode()           {}
func (n *TupleFieldAccessNode) expressionNode() {}
func (n *TupleNode) expressionNode()            {}

func (n *SimpleTypeNode) typeNode() {}
