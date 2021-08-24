package ast

import (
	"fmt"
	"github.com/iafisher/venice/src/lexer"
	"strconv"
	"strings"
)

type File struct {
	Statements []StatementNode
	Imports    []string
}

/**
 * AST interface declarations
 */

type Node interface {
	fmt.Stringer
	GetLocation() *lexer.Location
}

type ExpressionNode interface {
	Node
	expressionNode()
}

type PatternNode interface {
	Node
	patternNode()
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
	Destination ExpressionNode
	Expr        ExpressionNode
	Location    *lexer.Location
}

type BreakStatementNode struct {
	Location *lexer.Location
}

type ClassDeclarationNode struct {
	Name                 string
	GenericTypeParameter string
	Fields               []*ClassFieldNode
	Methods              []*ClassMethodNode
	Location             *lexer.Location
}

// Helper struct - does not implement Node
type ClassFieldNode struct {
	Name      string
	Public    bool
	FieldType TypeNode
}

type ClassMethodNode struct {
	Name       string
	Public     bool
	Params     []*FunctionParamNode
	ReturnType TypeNode
	Body       []StatementNode
	Location   *lexer.Location
}

type ContinueStatementNode struct {
	Location *lexer.Location
}

type EnumDeclarationNode struct {
	Name                 string
	GenericTypeParameter string
	Cases                []*EnumCaseNode
	Location             *lexer.Location
}

// Helper struct - does not implement Node
type EnumCaseNode struct {
	Label    string
	Types    []TypeNode
	Location *lexer.Location
}

type ExpressionStatementNode struct {
	Expr     ExpressionNode
	Location *lexer.Location
}

type ForLoopNode struct {
	Variables []string
	Iterable  ExpressionNode
	Body      []StatementNode
	Location  *lexer.Location
}

type FunctionDeclarationNode struct {
	Name       string
	Params     []*FunctionParamNode
	ReturnType TypeNode
	Body       []StatementNode
	Location   *lexer.Location
}

// Helper struct - does not implement Node
type FunctionParamNode struct {
	Name      string
	ParamType TypeNode
	Location  *lexer.Location
}

type IfStatementNode struct {
	Clauses    []*IfClauseNode
	ElseClause []StatementNode
	Location   *lexer.Location
}

// Helper struct - does not implement Node
type IfClauseNode struct {
	Condition ExpressionNode
	Body      []StatementNode
}

type ImportStatementNode struct {
	Path     string
	Name     string
	Location *lexer.Location
}

type LetStatementNode struct {
	Symbol   string
	Type     TypeNode
	IsVar    bool
	Expr     ExpressionNode
	Location *lexer.Location
}

type MatchStatementNode struct {
	Expr     ExpressionNode
	Clauses  []*MatchClause
	Default  []StatementNode
	Location *lexer.Location
}

// Helper struct - does not implement Node
type MatchClause struct {
	Pattern PatternNode
	Body    []StatementNode
}

type ReturnStatementNode struct {
	Expr     ExpressionNode
	Location *lexer.Location
}

type WhileLoopNode struct {
	Condition ExpressionNode
	Body      []StatementNode
	Location  *lexer.Location
}

/**
 * Pattern nodes
 */

type CompoundPatternNode struct {
	Label    string
	Patterns []PatternNode
	Elided   bool
	Location *lexer.Location
}

/**
 * Expression nodes
 */

type BooleanNode struct {
	Value    bool
	Location *lexer.Location
}

type CallNode struct {
	Function ExpressionNode
	Args     []ExpressionNode
	Location *lexer.Location
}

type CharacterNode struct {
	Value    byte
	Location *lexer.Location
}

type FieldAccessNode struct {
	Expr     ExpressionNode
	Name     string
	Location *lexer.Location
}

type IndexNode struct {
	Expr     ExpressionNode
	Index    ExpressionNode
	Location *lexer.Location
}

type InfixNode struct {
	Operator string
	Left     ExpressionNode
	Right    ExpressionNode
	Location *lexer.Location
}

type IntegerNode struct {
	Value    int
	Location *lexer.Location
}

type ListNode struct {
	Values   []ExpressionNode
	Location *lexer.Location
}

type MapNode struct {
	Pairs    []*MapPairNode
	Location *lexer.Location
}

// Helper struct - does not implement Node
type MapPairNode struct {
	Key      ExpressionNode
	Value    ExpressionNode
	Location *lexer.Location
}

type QualifiedSymbolNode struct {
	Enum     string
	Case     string
	Location *lexer.Location
}

type StringNode struct {
	Value    string
	Location *lexer.Location
}

type SymbolNode struct {
	Value    string
	Location *lexer.Location
}

type TernaryIfNode struct {
	Condition   ExpressionNode
	TrueClause  ExpressionNode
	FalseClause ExpressionNode
	Location    *lexer.Location
}

type TupleNode struct {
	Values   []ExpressionNode
	Location *lexer.Location
}

type TupleFieldAccessNode struct {
	Expr     ExpressionNode
	Index    int
	Location *lexer.Location
}

type UnaryNode struct {
	Operator string
	Expr     ExpressionNode
	Location *lexer.Location
}

/**
 * Type nodes
 */

type ParameterizedTypeNode struct {
	Symbol    string
	TypeNodes []TypeNode
	Location  *lexer.Location
}

/**
 * GetLocation() implementations
 */

func (n *AssignStatementNode) GetLocation() *lexer.Location {
	return n.Location
}

func (n *BooleanNode) GetLocation() *lexer.Location {
	return n.Location
}

func (n *BreakStatementNode) GetLocation() *lexer.Location {
	return n.Location
}

func (n *CallNode) GetLocation() *lexer.Location {
	return n.Location
}

func (n *CharacterNode) GetLocation() *lexer.Location {
	return n.Location
}

func (n *ClassDeclarationNode) GetLocation() *lexer.Location {
	return n.Location
}

func (n *ClassMethodNode) GetLocation() *lexer.Location {
	return n.Location
}

func (n *CompoundPatternNode) GetLocation() *lexer.Location {
	return n.Location
}

func (n *ContinueStatementNode) GetLocation() *lexer.Location {
	return n.Location
}

func (n *EnumDeclarationNode) GetLocation() *lexer.Location {
	return n.Location
}

func (n *ExpressionStatementNode) GetLocation() *lexer.Location {
	return n.Location
}

func (n *FieldAccessNode) GetLocation() *lexer.Location {
	return n.Location
}

func (n *ForLoopNode) GetLocation() *lexer.Location {
	return n.Location
}

func (n *FunctionDeclarationNode) GetLocation() *lexer.Location {
	return n.Location
}

func (n *IfStatementNode) GetLocation() *lexer.Location {
	return n.Location
}

func (n *ImportStatementNode) GetLocation() *lexer.Location {
	return n.Location
}

func (n *IndexNode) GetLocation() *lexer.Location {
	return n.Location
}

func (n *InfixNode) GetLocation() *lexer.Location {
	return n.Location
}

func (n *IntegerNode) GetLocation() *lexer.Location {
	return n.Location
}

func (n *LetStatementNode) GetLocation() *lexer.Location {
	return n.Location
}

func (n *ListNode) GetLocation() *lexer.Location {
	return n.Location
}

func (n *MapNode) GetLocation() *lexer.Location {
	return n.Location
}

func (n *MatchStatementNode) GetLocation() *lexer.Location {
	return n.Location
}

func (n *ParameterizedTypeNode) GetLocation() *lexer.Location {
	return n.Location
}

func (n *QualifiedSymbolNode) GetLocation() *lexer.Location {
	return n.Location
}

func (n *ReturnStatementNode) GetLocation() *lexer.Location {
	return n.Location
}

func (n *StringNode) GetLocation() *lexer.Location {
	return n.Location
}

func (n *SymbolNode) GetLocation() *lexer.Location {
	return n.Location
}

func (n *TernaryIfNode) GetLocation() *lexer.Location {
	return n.Location
}

func (n *TupleFieldAccessNode) GetLocation() *lexer.Location {
	return n.Location
}

func (n *TupleNode) GetLocation() *lexer.Location {
	return n.Location
}

func (n *UnaryNode) GetLocation() *lexer.Location {
	return n.Location
}

func (n *WhileLoopNode) GetLocation() *lexer.Location {
	return n.Location
}

/**
 * String() implementations
 */

func (n *AssignStatementNode) String() string {
	return fmt.Sprintf("(assign %s %s)", n.Destination.String(), n.Expr.String())
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
	return fmt.Sprintf("%q", n.Value)
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

	for _, method := range n.Methods {
		sb.WriteByte(' ')
		sb.WriteString(method.String())
	}

	sb.WriteString("))")
	return sb.String()
}

func (n *ClassMethodNode) String() string {
	var sb strings.Builder
	sb.WriteString("(class-method ")
	if n.Public {
		sb.WriteString("public ")
	} else {
		sb.WriteString("private ")
	}
	sb.WriteString(n.Name)
	sb.WriteString(" (")
	for i, param := range n.Params {
		if i != 0 {
			sb.WriteByte(' ')
		}
		sb.WriteString("(function-param ")
		sb.WriteString(param.Name)
		sb.WriteByte(' ')
		sb.WriteString(param.ParamType.String())
		sb.WriteByte(')')
	}
	sb.WriteByte(')')
	if n.ReturnType != nil {
		sb.WriteByte(' ')
		sb.WriteString(n.ReturnType.String())
	} else {
		sb.WriteString(" void")
	}
	sb.WriteByte(' ')
	writeBlock(&sb, n.Body)
	sb.WriteByte(')')
	return sb.String()
}

func (n *CompoundPatternNode) String() string {
	var sb strings.Builder
	sb.WriteByte('(')
	sb.WriteString(n.Label)
	for _, pattern := range n.Patterns {
		sb.WriteByte(' ')
		sb.WriteString(pattern.String())
	}

	if n.Elided {
		sb.WriteString(" ...")
	}

	sb.WriteByte(')')
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

func (n *ExpressionStatementNode) String() string {
	return fmt.Sprintf("(expression-statement %s)", n.Expr.String())
}

func (n *FieldAccessNode) String() string {
	return fmt.Sprintf("(field-access %s %s)", n.Expr.String(), n.Name)
}

func (f *File) String() string {
	var sb strings.Builder
	writeBlock(&sb, f.Statements)
	return sb.String()
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
	for i, param := range n.Params {
		if i != 0 {
			sb.WriteByte(' ')
		}
		sb.WriteString("(function-param ")
		sb.WriteString(param.Name)
		sb.WriteByte(' ')
		sb.WriteString(param.ParamType.String())
		sb.WriteByte(')')
	}
	sb.WriteByte(')')
	if n.ReturnType != nil {
		sb.WriteByte(' ')
		sb.WriteString(n.ReturnType.String())
	} else {
		sb.WriteString(" void")
	}
	sb.WriteByte(' ')
	writeBlock(&sb, n.Body)
	sb.WriteByte(')')
	return sb.String()
}

func (n *IfStatementNode) String() string {
	var sb strings.Builder
	sb.WriteString("(if")

	for _, clause := range n.Clauses {
		sb.WriteString(" (")
		sb.WriteString(clause.Condition.String())
		sb.WriteByte(' ')
		writeBlock(&sb, clause.Body)
		sb.WriteByte(')')
	}

	if len(n.ElseClause) > 0 {
		sb.WriteString(" (else ")
		writeBlock(&sb, n.ElseClause)
		sb.WriteByte(')')
	}

	sb.WriteByte(')')
	return sb.String()
}

func (n *ImportStatementNode) String() string {
	return fmt.Sprintf("(import %s %q)", n.Name, n.Path)
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
	var keyword string
	if n.IsVar {
		keyword = "var"
	} else {
		keyword = "let"
	}

	var typeString string
	if n.Type != nil {
		typeString = n.Type.String()
	} else {
		typeString = "nil"
	}

	return fmt.Sprintf("(%s %s %s %s)", keyword, typeString, n.Symbol, n.Expr.String())
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

func (n *MatchClause) String() string {
	var sb strings.Builder
	sb.WriteString("(match-case ")
	sb.WriteString(n.Pattern.String())
	sb.WriteByte(' ')
	writeBlock(&sb, n.Body)
	sb.WriteByte(')')
	return sb.String()
}

func (n *MatchStatementNode) String() string {
	var sb strings.Builder
	sb.WriteString("(match ")
	sb.WriteString(n.Expr.String())
	for _, clause := range n.Clauses {
		sb.WriteByte(' ')
		sb.WriteString(clause.String())
	}

	if len(n.Default) > 0 {
		sb.WriteString(" (match-default ")
		writeBlock(&sb, n.Default)
		sb.WriteByte(')')
	}
	sb.WriteByte(')')
	return sb.String()
}

func (n *ParameterizedTypeNode) String() string {
	var sb strings.Builder
	sb.WriteString("(type ")
	sb.WriteString(n.Symbol)
	for _, subTypeNode := range n.TypeNodes {
		sb.WriteByte(' ')
		sb.WriteString(subTypeNode.String())
	}
	sb.WriteByte(')')
	return sb.String()
}

func (n *QualifiedSymbolNode) String() string {
	return fmt.Sprintf("(enum-case %s %s)", n.Enum, n.Case)
}

func (n *ReturnStatementNode) String() string {
	if n.Expr != nil {
		return fmt.Sprintf("(return %s)", n.Expr.String())
	} else {
		return "(return)"
	}
}

func (n *StringNode) String() string {
	return strconv.Quote(n.Value)
}

func (n *SymbolNode) String() string {
	return n.Value
}

func (n *TernaryIfNode) String() string {
	return fmt.Sprintf("(ternary-if %s %s %s)", n.Condition.String(), n.TrueClause.String(), n.FalseClause.String())
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

func (n *UnaryNode) String() string {
	return fmt.Sprintf("(unary %s %s)", n.Operator, n.Expr.String())
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

func (n *BooleanNode) expressionNode()          {}
func (n *CallNode) expressionNode()             {}
func (n *CharacterNode) expressionNode()        {}
func (n *FieldAccessNode) expressionNode()      {}
func (n *IndexNode) expressionNode()            {}
func (n *InfixNode) expressionNode()            {}
func (n *IntegerNode) expressionNode()          {}
func (n *ListNode) expressionNode()             {}
func (n *MapNode) expressionNode()              {}
func (n *QualifiedSymbolNode) expressionNode()  {}
func (n *StringNode) expressionNode()           {}
func (n *SymbolNode) expressionNode()           {}
func (n *TernaryIfNode) expressionNode()        {}
func (n *TupleFieldAccessNode) expressionNode() {}
func (n *TupleNode) expressionNode()            {}
func (n *UnaryNode) expressionNode()            {}

func (n *CompoundPatternNode) patternNode() {}
func (n *SymbolNode) patternNode()          {}

func (n *AssignStatementNode) statementNode()     {}
func (n *BreakStatementNode) statementNode()      {}
func (n *ClassDeclarationNode) statementNode()    {}
func (n *ClassMethodNode) statementNode()         {}
func (n *ContinueStatementNode) statementNode()   {}
func (n *EnumDeclarationNode) statementNode()     {}
func (n *ExpressionStatementNode) statementNode() {}
func (n *ForLoopNode) statementNode()             {}
func (n *FunctionDeclarationNode) statementNode() {}
func (n *IfStatementNode) statementNode()         {}
func (n *ImportStatementNode) statementNode()     {}
func (n *LetStatementNode) statementNode()        {}
func (n *MatchStatementNode) statementNode()      {}
func (n *ReturnStatementNode) statementNode()     {}
func (n *WhileLoopNode) statementNode()           {}

func (n *ParameterizedTypeNode) typeNode() {}
func (n *SymbolNode) typeNode()            {}
