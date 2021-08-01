package main

type VeniceValue interface {
	veniceValue()
}

type VeniceInteger struct {
	Value int64
}

func (v *VeniceInteger) veniceValue() {}

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

type SymbolTable struct {
	parent  *SymbolTable
	symbols map[string]VeniceType
}

type Compiler struct {
	symbolTable *SymbolTable
}

func NewCompiler() *Compiler {
	return &Compiler{&SymbolTable{nil, make(map[string]VeniceType)}}
}

func (compiler *Compiler) Compile(tree *ProgramNode) ([]*Bytecode, bool) {
	program := []*Bytecode{}
	for _, declaration := range tree.Declarations {
		declarationCode, ok := compiler.compileDeclaration(declaration)
		if !ok {
			return nil, false
		}
		program = append(program, declarationCode...)
	}
	return program, true
}

func (compiler *Compiler) CompileExpression(tree Expression) ([]*Bytecode, VeniceType, bool) {
	switch v := tree.(type) {
	case *IntegerNode:
		return []*Bytecode{NewBytecode("PUSH_CONST", &VeniceInteger{v.Value})}, VENICE_TYPE_INTEGER, true
	case *InfixNode:
		leftBytecodes, leftType, ok := compiler.CompileExpression(v.Left)
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

		rightBytecodes, rightType, ok := compiler.CompileExpression(v.Right)
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
		switch v.Operator {
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
	default:
		return nil, nil, false
	}
}

func (compiler *Compiler) compileDeclaration(declaration Declaration) ([]*Bytecode, bool) {
	return nil, false
}

const (
	VENICE_TYPE_INTEGER_LABEL = "int"
	VENICE_TYPE_STRING_LABEL  = "string"
)

var (
	VENICE_TYPE_INTEGER = &VeniceAtomicType{VENICE_TYPE_INTEGER_LABEL}
	VENICE_TYPE_STRING  = &VeniceAtomicType{VENICE_TYPE_STRING_LABEL}
)
