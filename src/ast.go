package main

import (
	"fmt"
	"strconv"
	"strings"
)

/**
 * AST interface declarations
 */

type Node interface {
	fmt.Stringer
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

type ProgramNode struct {
	Statements []StatementNode
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

type StringNode struct {
	Value    string
	Location *Location
}

type SymbolNode struct {
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

func (n *ProgramNode) getLocation() *Location {
	return &Location{Line: 1, Column: 1}
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

/**
 * String() implementations
 */

func (n *AssignStatementNode) String() string {
	return fmt.Sprintf("(assign %s %s)", n.Symbol, n.Expr.String())
}

func (n *BooleanNode) String() string {
	if n.Value {
		return "true"
	} else {
		return "false"
	}
}

func (n *BreakStatementNode) String() string {
	return "(break)"
}

func (n *CallNode) String() string {
	var sb strings.Builder
	sb.WriteString("(call ")
	sb.WriteString(n.Function.String())
	for _, arg := range n.Args {
		sb.WriteByte(' ')
		sb.WriteString(arg.String())
	}
	sb.WriteByte(')')
	return sb.String()
}

func (n *CharacterNode) String() string {
	// TODO(2021-08-13): Does this handle backslash escapes correctly?
	return fmt.Sprintf("'%c'", n.Value)
}

func (n *ClassDeclarationNode) String() string {
	var sb strings.Builder
	sb.WriteString("(class-declaration ")
	sb.WriteString(n.Name)
	if n.GenericTypeParameter != "" {
		sb.WriteByte(' ')
		sb.WriteString(n.GenericTypeParameter)
	}
	sb.WriteString(" (")
	for i, field := range n.Fields {
		if i != 0 {
			sb.WriteByte(' ')
		}
		sb.WriteString("(class-field ")
		if field.Public {
			sb.WriteString("public ")
		} else {
			sb.WriteString("private ")
		}
		sb.WriteString(field.Name)
		sb.WriteByte(' ')
		sb.WriteString(field.FieldType.String())
		sb.WriteByte(')')
	}
	sb.WriteString("))")
	return sb.String()
}

func (n *ContinueStatementNode) String() string {
	return "(continue)"
}

func (n *EnumDeclarationNode) String() string {
	var sb strings.Builder
	sb.WriteString("(enum-declaration ")
	sb.WriteString(n.Name)
	if n.GenericTypeParameter != "" {
		sb.WriteByte(' ')
		sb.WriteString(n.GenericTypeParameter)
	}
	sb.WriteString(" (")
	for i, enumCase := range n.Cases {
		if i != 0 {
			sb.WriteByte(' ')
		}
		sb.WriteString("(enum-case ")
		sb.WriteString(enumCase.Label)
		for _, caseType := range enumCase.Types {
			sb.WriteByte(' ')
			sb.WriteString(caseType.String())
		}
		sb.WriteByte(')')
	}
	sb.WriteString("))")
	return sb.String()
}

func (n *EnumSymbolNode) String() string {
	return fmt.Sprintf("(enum-case %s %s)", n.Enum, n.Case)
}

func (n *ExpressionStatementNode) String() string {
	return fmt.Sprintf("(expression-statement %s)", n.Expr.String())
}

func (n *FieldAccessNode) String() string {
	return fmt.Sprintf("(field-access %s %s)", n.Expr.String(), n.Name)
}

func (n *ForLoopNode) String() string {
	var sb strings.Builder
	sb.WriteString("(for (")
	for _, variable := range n.Variables {
		sb.WriteByte(' ')
		sb.WriteString(variable)
	}
	sb.WriteString(") ")
	sb.WriteString(n.Iterable.String())
	sb.WriteByte(' ')
	writeBlock(&sb, n.Body)
	sb.WriteByte(')')
	return sb.String()
}

func (n *FunctionDeclarationNode) String() string {
	var sb strings.Builder
	sb.WriteString("(function-declaration ")
	sb.WriteString(n.Name)
	sb.WriteString(" (")
	for _, param := range n.Params {
		sb.WriteString(" (function-param ")
		sb.WriteString(param.Name)
		sb.WriteByte(' ')
		sb.WriteString(param.ParamType.String())
		sb.WriteByte(')')
	}
	sb.WriteByte(')')
	sb.WriteByte(')')
	return sb.String()
}

func (n *IfStatementNode) String() string {
	var sb strings.Builder
	sb.WriteString("(if ")
	sb.WriteString(n.Condition.String())
	sb.WriteByte(' ')
	writeBlock(&sb, n.TrueClause)
	writeBlock(&sb, n.FalseClause)
	sb.WriteByte(')')
	return sb.String()
}

func (n *IndexNode) String() string {
	return fmt.Sprintf("(index %s %s)", n.Expr.String(), n.Index.String())
}

func (n *InfixNode) String() string {
	return fmt.Sprintf("(infix %s %s %s)", n.Operator, n.Left.String(), n.Right.String())
}

func (n *IntegerNode) String() string {
	return fmt.Sprint(n.Value)
}

func (n *LetStatementNode) String() string {
	return fmt.Sprintf("(let %s %s)", n.Symbol, n.Expr.String())
}

func (n *ListNode) String() string {
	var sb strings.Builder
	sb.WriteString("(list")
	for _, value := range n.Values {
		sb.WriteByte(' ')
		sb.WriteString(value.String())
	}
	sb.WriteByte(')')
	return sb.String()
}

func (n *MapNode) String() string {
	var sb strings.Builder
	sb.WriteString("(map")
	for _, pair := range n.Pairs {
		sb.WriteString(" (")
		sb.WriteString(pair.Key.String())
		sb.WriteByte(' ')
		sb.WriteString(pair.Value.String())
		sb.WriteByte(')')
	}
	sb.WriteByte(')')
	return sb.String()
}

func (n *ProgramNode) String() string {
	var sb strings.Builder
	writeBlock(&sb, n.Statements)
	return sb.String()
}

func (n *ReturnStatementNode) String() string {
	if n.Expr != nil {
		return fmt.Sprintf("(return %s)", n.Expr.String())
	} else {
		return "(return)"
	}
}

func (n *SimpleTypeNode) String() string {
	return n.Symbol
}

func (n *StringNode) String() string {
	return strconv.Quote(n.Value)
}

func (n *SymbolNode) String() string {
	return n.Value
}

func (n *TupleFieldAccessNode) String() string {
	return fmt.Sprintf("(tuple-field-access %s %d)", n.Expr.String(), n.Index)
}

func (n *TupleNode) String() string {
	var sb strings.Builder
	sb.WriteString("(tuple")
	for _, value := range n.Values {
		sb.WriteByte(' ')
		sb.WriteString(value.String())
	}
	sb.WriteByte(')')
	return sb.String()
}

func (n *WhileLoopNode) String() string {
	var sb strings.Builder
	sb.WriteString("(while ")
	sb.WriteString(n.Condition.String())
	sb.WriteByte(' ')
	writeBlock(&sb, n.Body)
	sb.WriteByte(')')
	return sb.String()
}

func writeBlock(sb *strings.Builder, block []StatementNode) {
	sb.WriteString("(block")
	for _, statement := range block {
		sb.WriteByte(' ')
		sb.WriteString(statement.String())
	}
	sb.WriteByte(')')
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
func (n *ProgramNode) statementNode()             {}
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
