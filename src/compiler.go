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

type CompileError struct {
	Message string
}

func (e *CompileError) Error() string {
	return e.Message
}

func (compiler *Compiler) Compile(tree *ProgramNode) ([]*Bytecode, error) {
	program := []*Bytecode{}
	for _, statement := range tree.Statements {
		statementCode, err := compiler.compileStatement(statement)
		if err != nil {
			return nil, err
		}
		program = append(program, statementCode...)
	}
	return program, nil
}

func (compiler *Compiler) compileStatement(tree Statement) ([]*Bytecode, error) {
	switch v := tree.(type) {
	case *ExpressionStatementNode:
		bytecodes, _, err := compiler.compileExpression(v.Expression)
		if err != nil {
			return nil, err
		}
		return bytecodes, nil
	case *LetStatementNode:
		bytecodes, eType, err := compiler.compileExpression(v.Expr)
		if err != nil {
			return nil, err
		}
		compiler.symbolTable.symbols[v.Symbol] = eType
		return append(bytecodes, NewBytecode("STORE_NAME", &VeniceString{v.Symbol})), nil
	default:
		return nil, &CompileError{"unknown statement type"}
	}
}

func (compiler *Compiler) compileExpression(tree Expression) ([]*Bytecode, VeniceType, error) {
	switch v := tree.(type) {
	case *CallNode:
		return compiler.compileCallNode(v)
	case *InfixNode:
		return compiler.compileInfixNode(v)
	case *IntegerNode:
		return []*Bytecode{NewBytecode("PUSH_CONST", &VeniceInteger{v.Value})}, VENICE_TYPE_INTEGER, nil
	case *StringNode:
		return []*Bytecode{NewBytecode("PUSH_CONST", &VeniceString{v.Value})}, VENICE_TYPE_STRING, nil
	case *SymbolNode:
		symbolType, ok := compiler.symbolTable.Get(v.Value)
		if !ok {
			return nil, nil, &CompileError{fmt.Sprintf("undefined symbol: %s", v.Value)}
		}
		return []*Bytecode{NewBytecode("PUSH_NAME", &VeniceString{v.Value})}, symbolType, nil
	default:
		return nil, nil, &CompileError{"unknown expression type"}
	}
}

func (compiler *Compiler) compileCallNode(tree *CallNode) ([]*Bytecode, VeniceType, error) {
	if v, ok := tree.Function.(*SymbolNode); ok {
		if v.Value == "print" {
			if len(tree.Args) != 1 {
				return nil, nil, &CompileError{"`print` takes exactly 1 argument"}
			}

			bytecodes, argType, err := compiler.compileExpression(tree.Args[0])
			if err != nil {
				return nil, nil, err
			}

			if argType != VENICE_TYPE_INTEGER && argType != VENICE_TYPE_STRING {
				return nil, nil, &CompileError{"`print`'s argument must be an integer or string"}
			}

			bytecodes = append(bytecodes, NewBytecode("CALL_BUILTIN", &VeniceString{"print"}))
			return bytecodes, nil, nil
		}
	}

	return nil, nil, &CompileError{"function calls not implemented yet"}
}

func (compiler *Compiler) compileInfixNode(tree *InfixNode) ([]*Bytecode, VeniceType, error) {
	leftBytecodes, leftType, err := compiler.compileExpression(tree.Left)
	if err != nil {
		return nil, nil, err
	}

	leftAtomicType, ok := leftType.(*VeniceAtomicType)
	if !ok || leftAtomicType != VENICE_TYPE_INTEGER {
		return nil, nil, &CompileError{fmt.Sprintf("operand of %s must be an integer", tree.Operator)}
	}

	rightBytecodes, rightType, err := compiler.compileExpression(tree.Right)
	if !ok {
		return nil, nil, err
	}

	rightAtomicType, ok := rightType.(*VeniceAtomicType)
	if !ok || rightAtomicType != VENICE_TYPE_INTEGER {
		return nil, nil, &CompileError{fmt.Sprintf("operand of %s must be an integer", tree.Operator)}
	}

	bytecodes := append(leftBytecodes, rightBytecodes...)
	switch tree.Operator {
	case "+":
		return append(bytecodes, NewBytecode("BINARY_ADD")), VENICE_TYPE_INTEGER, nil
	case "-":
		return append(bytecodes, NewBytecode("BINARY_SUB")), VENICE_TYPE_INTEGER, nil
	case "*":
		return append(bytecodes, NewBytecode("BINARY_MUL")), VENICE_TYPE_INTEGER, nil
	case "/":
		return append(bytecodes, NewBytecode("BINARY_DIV")), VENICE_TYPE_INTEGER, nil
	default:
		return nil, nil, &CompileError{fmt.Sprintf("unknown oeprator: %s", tree.Operator)}
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
