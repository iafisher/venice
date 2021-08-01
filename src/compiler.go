package main

import "fmt"

type VeniceValue interface {
	veniceValue()
	Serialize() string
}

type VeniceInteger struct {
	Value int64
}

func (v *VeniceInteger) veniceValue() {}
func (v *VeniceInteger) Serialize() string {
	return fmt.Sprintf("%d", v.Value)
}

type VeniceString struct {
	Value string
}

func (v *VeniceString) veniceValue() {}
func (v *VeniceString) Serialize() string {
	return fmt.Sprintf("%q", v.Value)
}

type Bytecode struct {
	Name string
	Args []VeniceValue
}

func NewBytecode(name string, args ...VeniceValue) *Bytecode {
	return &Bytecode{name, args}
}

type VeniceType interface {
	veniceType()
}

type VeniceAtomicType struct {
	Type string
}

func (t *VeniceAtomicType) veniceType() {}

type VeniceFunctionType struct {
	ArgTypes   []VeniceType
	ReturnType VeniceType
}

func (t *VeniceFunctionType) veniceType() {}

type SymbolTable struct {
	parent  *SymbolTable
	symbols map[string]VeniceType
}

func NewBuiltinSymbolTable() *SymbolTable {
	symbols := map[string]VeniceType{}
	return &SymbolTable{nil, symbols}
}

type Compiler struct {
	symbolTable *SymbolTable
}

func NewCompiler() *Compiler {
	return &Compiler{NewBuiltinSymbolTable()}
}

func (compiler *Compiler) Compile(tree *ProgramNode) ([]*Bytecode, bool) {
	program := []*Bytecode{}
	for _, statement := range tree.Statements {
		statementCode, ok := compiler.compileStatement(statement)
		if !ok {
			return nil, false
		}
		program = append(program, statementCode...)
	}
	return program, true
}

func (compiler *Compiler) compileStatement(tree Statement) ([]*Bytecode, bool) {
	switch v := tree.(type) {
	case *ExpressionStatementNode:
		bytecodes, _, ok := compiler.compileExpression(v.Expression)
		if !ok {
			return nil, false
		}
		return bytecodes, true
	case *LetStatementNode:
		bytecodes, eType, ok := compiler.compileExpression(v.Expr)
		if !ok {
			return nil, false
		}
		compiler.symbolTable.symbols[v.Symbol] = eType
		return append(bytecodes, NewBytecode("STORE_NAME", &VeniceString{v.Symbol})), true
	default:
		return nil, false
	}
}

func (compiler *Compiler) compileExpression(tree Expression) ([]*Bytecode, VeniceType, bool) {
	switch v := tree.(type) {
	case *CallNode:
		return compiler.compileCallNode(v)
	case *InfixNode:
		return compiler.compileInfixNode(v)
	case *IntegerNode:
		return []*Bytecode{NewBytecode("PUSH_CONST", &VeniceInteger{v.Value})}, VENICE_TYPE_INTEGER, true
	case *SymbolNode:
		symbolType, ok := compiler.symbolTable.Get(v.Value)
		if !ok {
			return nil, nil, false
		}
		return []*Bytecode{NewBytecode("PUSH_NAME", &VeniceString{v.Value})}, symbolType, true
	default:
		return nil, nil, false
	}
}

func (compiler *Compiler) compileCallNode(tree *CallNode) ([]*Bytecode, VeniceType, bool) {
	if v, ok := tree.Function.(*SymbolNode); ok {
		if v.Value == "print" {
			if len(tree.Args) != 1 {
				return nil, nil, false
			}

			bytecodes, argType, ok := compiler.compileExpression(tree.Args[0])
			if !ok {
				return nil, nil, false
			}

			if argType != VENICE_TYPE_INTEGER && argType != VENICE_TYPE_STRING {
				return nil, nil, false
			}

			bytecodes = append(bytecodes, NewBytecode("CALL_BUILTIN", &VeniceString{"print"}))
			return bytecodes, nil, true
		}
	}

	return nil, nil, false
}

func (compiler *Compiler) compileInfixNode(tree *InfixNode) ([]*Bytecode, VeniceType, bool) {
	leftBytecodes, leftType, ok := compiler.compileExpression(tree.Left)
	if !ok {
		return nil, nil, false
	}

	leftAtomicType, ok := leftType.(*VeniceAtomicType)
	if !ok {
		return nil, nil, false
	}

	if leftAtomicType != VENICE_TYPE_INTEGER {
		return nil, nil, false
	}

	rightBytecodes, rightType, ok := compiler.compileExpression(tree.Right)
	if !ok {
		return nil, nil, false
	}

	rightAtomicType, ok := rightType.(*VeniceAtomicType)
	if !ok {
		return nil, nil, false
	}

	if rightAtomicType != VENICE_TYPE_INTEGER {
		return nil, nil, false
	}

	bytecodes := append(leftBytecodes, rightBytecodes...)
	switch tree.Operator {
	case "+":
		return append(bytecodes, NewBytecode("BINARY_ADD")), VENICE_TYPE_INTEGER, true
	case "-":
		return append(bytecodes, NewBytecode("BINARY_SUB")), VENICE_TYPE_INTEGER, true
	case "*":
		return append(bytecodes, NewBytecode("BINARY_MUL")), VENICE_TYPE_INTEGER, true
	case "/":
		return append(bytecodes, NewBytecode("BINARY_DIV")), VENICE_TYPE_INTEGER, true
	default:
		return nil, nil, false
	}
}

func (symtab *SymbolTable) Get(symbol string) (VeniceType, bool) {
	value, ok := symtab.symbols[symbol]
	if !ok {
		if symtab.parent != nil {
			return symtab.parent.Get(symbol)
		} else {
			return nil, false
		}
	}
	return value, true
}

func (symtab *SymbolTable) Put(symbol string, value VeniceType) {
	symtab.symbols[symbol] = value
}

const (
	VENICE_TYPE_INTEGER_LABEL = "int"
	VENICE_TYPE_STRING_LABEL  = "string"
)

var (
	VENICE_TYPE_INTEGER = &VeniceAtomicType{VENICE_TYPE_INTEGER_LABEL}
	VENICE_TYPE_STRING  = &VeniceAtomicType{VENICE_TYPE_STRING_LABEL}
)
