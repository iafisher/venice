package main

import (
	"fmt"
	"strings"
)

type VeniceValue interface {
	veniceValue()
	Serialize() string
	SerializePrintable() string
	Equals(v VeniceValue) bool
}

type VeniceList struct {
	Values []VeniceValue
}

func (v *VeniceList) veniceValue() {}

func (v *VeniceList) Serialize() string {
	var sb strings.Builder
	sb.WriteByte('[')
	for i, value := range v.Values {
		sb.WriteString(value.Serialize())
		if i != len(v.Values)-1 {
			sb.WriteString(", ")
		}
	}
	sb.WriteByte(']')
	return sb.String()
}

func (v *VeniceList) SerializePrintable() string {
	var sb strings.Builder
	sb.WriteByte('[')
	for i, value := range v.Values {
		sb.WriteString(value.SerializePrintable())
		if i != len(v.Values)-1 {
			sb.WriteString(", ")
		}
	}
	sb.WriteByte(']')
	return sb.String()
}

func (v *VeniceList) Equals(otherUntyped VeniceValue) bool {
	switch other := otherUntyped.(type) {
	case *VeniceList:
		if len(v.Values) != len(other.Values) {
			return false
		}

		for i := 0; i < len(v.Values); i++ {
			if !v.Values[i].Equals(other.Values[i]) {
				return false
			}
		}

		return true
	default:
		return false
	}
}

type VeniceMap struct {
	Pairs []*VeniceMapPair
}

func (v *VeniceMap) veniceValue() {}

func (v *VeniceMap) Serialize() string {
	var sb strings.Builder
	sb.WriteByte('{')
	for i, pair := range v.Pairs {
		sb.WriteString(pair.Key.Serialize())
		sb.WriteString(": ")
		sb.WriteString(pair.Value.Serialize())
		if i != len(v.Pairs)-1 {
			sb.WriteString(", ")
		}
	}
	sb.WriteByte('}')
	return sb.String()
}

func (v *VeniceMap) SerializePrintable() string {
	var sb strings.Builder
	sb.WriteByte('{')
	for i, pair := range v.Pairs {
		sb.WriteString(pair.Key.SerializePrintable())
		sb.WriteString(": ")
		sb.WriteString(pair.Value.SerializePrintable())
		if i != len(v.Pairs)-1 {
			sb.WriteString(", ")
		}
	}
	sb.WriteByte('}')
	return sb.String()
}

func (v *VeniceMap) Equals(otherUntyped VeniceValue) bool {
	// TODO(2021-08-31): Implement.
	return false
}

type VeniceMapPair struct {
	Key   VeniceValue
	Value VeniceValue
}

type VeniceInteger struct {
	Value int
}

func (v *VeniceInteger) veniceValue() {}

func (v *VeniceInteger) Serialize() string {
	return fmt.Sprintf("%d", v.Value)
}

func (v *VeniceInteger) SerializePrintable() string {
	return v.Serialize()
}

func (v *VeniceInteger) Equals(otherUntyped VeniceValue) bool {
	switch other := otherUntyped.(type) {
	case *VeniceInteger:
		return v.Value == other.Value
	default:
		return false
	}
}

type VeniceString struct {
	Value string
}

func (v *VeniceString) veniceValue() {}

func (v *VeniceString) Serialize() string {
	return fmt.Sprintf("%q", v.Value)
}

func (v *VeniceString) SerializePrintable() string {
	return v.Value
}

func (v *VeniceString) Equals(otherUntyped VeniceValue) bool {
	switch other := otherUntyped.(type) {
	case *VeniceString:
		return v.Value == other.Value
	default:
		return false
	}
}

type VeniceBoolean struct {
	Value bool
}

func (v *VeniceBoolean) veniceValue() {}

func (v *VeniceBoolean) Serialize() string {
	if v.Value {
		return "true"
	} else {
		return "false"
	}
}

func (v *VeniceBoolean) SerializePrintable() string {
	return v.Serialize()
}

func (v *VeniceBoolean) Equals(otherUntyped VeniceValue) bool {
	switch other := otherUntyped.(type) {
	case *VeniceBoolean:
		return v.Value == other.Value
	default:
		return false
	}
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

type VeniceMapType struct {
	KeyType   VeniceType
	ValueType VeniceType
}

func (t *VeniceMapType) veniceType() {}

type VeniceListType struct {
	ItemType VeniceType
}

func (t *VeniceListType) veniceType() {}

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
	case *IfStatementNode:
		return compiler.compileIfStatement(v)
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

func (compiler *Compiler) compileIfStatement(tree *IfStatementNode) ([]*Bytecode, error) {
	conditionBytecodes, conditionType, err := compiler.compileExpression(tree.Condition)
	if err != nil {
		return nil, err
	}

	if conditionType != VENICE_TYPE_BOOLEAN {
		return nil, &CompileError{"condition of if statement must be a boolean"}
	}

	trueClauseBytecodes, err := compiler.compileBlock(tree.TrueClause)
	if err != nil {
		return nil, err
	}

	bytecodes := conditionBytecodes
	bytecodes = append(bytecodes, NewBytecode("REL_JUMP_IF_FALSE", &VeniceInteger{len(trueClauseBytecodes) + 1}))
	bytecodes = append(bytecodes, trueClauseBytecodes...)

	if tree.FalseClause != nil {
		falseClauseBytecodes, err := compiler.compileBlock(tree.FalseClause)
		if err != nil {
			return nil, err
		}

		bytecodes = append(bytecodes, NewBytecode("REL_JUMP", &VeniceInteger{len(falseClauseBytecodes) + 1}))
		bytecodes = append(bytecodes, falseClauseBytecodes...)
	}

	return bytecodes, nil
}

func (compiler *Compiler) compileBlock(block []Statement) ([]*Bytecode, error) {
	bytecodes := []*Bytecode{}
	for _, statement := range block {
		statementBytecodes, err := compiler.compileStatement(statement)
		if err != nil {
			return nil, err
		}
		bytecodes = append(bytecodes, statementBytecodes...)
	}
	return bytecodes, nil
}

func (compiler *Compiler) compileExpression(tree Expression) ([]*Bytecode, VeniceType, error) {
	switch v := tree.(type) {
	case *BooleanNode:
		return []*Bytecode{NewBytecode("PUSH_CONST", &VeniceBoolean{v.Value})}, VENICE_TYPE_BOOLEAN, nil
	case *CallNode:
		return compiler.compileCallNode(v)
	case *IndexNode:
		return compiler.compileIndexNode(v)
	case *InfixNode:
		return compiler.compileInfixNode(v)
	case *IntegerNode:
		return []*Bytecode{NewBytecode("PUSH_CONST", &VeniceInteger{v.Value})}, VENICE_TYPE_INTEGER, nil
	case *ListNode:
		bytecodes := []*Bytecode{}
		var itemType VeniceType
		for i := len(v.Values) - 1; i >= 0; i-- {
			value := v.Values[i]
			valueBytecodes, valueType, err := compiler.compileExpression(value)
			if err != nil {
				return nil, nil, err
			}

			if itemType == nil {
				itemType = valueType
			} else if !areTypesCompatible(itemType, valueType) {
				return nil, nil, &CompileError{"list elements must all be of same type"}
			}

			bytecodes = append(bytecodes, valueBytecodes...)
		}
		bytecodes = append(bytecodes, NewBytecode("BUILD_LIST", &VeniceInteger{len(v.Values)}))
		return bytecodes, &VeniceListType{itemType}, nil
	case *MapNode:
		bytecodes := []*Bytecode{}
		var keyType VeniceType
		var valueType VeniceType
		for i := len(v.Pairs) - 1; i >= 0; i-- {
			pair := v.Pairs[i]
			keyBytecodes, thisKeyType, err := compiler.compileExpression(pair.Key)
			if err != nil {
				return nil, nil, err
			}

			if keyType == nil {
				keyType = thisKeyType
			} else if !areTypesCompatible(keyType, thisKeyType) {
				return nil, nil, &CompileError{"map keys must all be of the same type"}
			}

			bytecodes = append(bytecodes, keyBytecodes...)

			valueBytecodes, thisValueType, err := compiler.compileExpression(pair.Value)
			if err != nil {
				return nil, nil, err
			}

			if valueType == nil {
				valueType = thisValueType
			} else if !areTypesCompatible(valueType, thisValueType) {
				return nil, nil, &CompileError{"map values must all be of the same type"}
			}

			bytecodes = append(bytecodes, valueBytecodes...)
		}
		bytecodes = append(bytecodes, NewBytecode("BUILD_MAP", &VeniceInteger{len(v.Pairs)}))
		return bytecodes, &VeniceMapType{keyType, valueType}, nil
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

func (compiler *Compiler) compileIndexNode(tree *IndexNode) ([]*Bytecode, VeniceType, error) {
	exprBytecodes, exprType, err := compiler.compileExpression(tree.Expr)
	if err != nil {
		return nil, nil, err
	}

	indexBytecodes, indexType, err := compiler.compileExpression(tree.Index)
	if err != nil {
		return nil, nil, err
	}

	bytecodes := append(exprBytecodes, indexBytecodes...)

	switch exprConcreteType := exprType.(type) {
	case *VeniceListType:
		if !areTypesCompatible(VENICE_TYPE_INTEGER, indexType) {
			return nil, nil, &CompileError{"list index must be integer"}
		}

		bytecodes = append(bytecodes, NewBytecode("BINARY_LIST_INDEX"))
		return bytecodes, exprConcreteType.ItemType, nil
	case *VeniceMapType:
		if !areTypesCompatible(exprConcreteType.KeyType, indexType) {
			return nil, nil, &CompileError{"wrong map key type in index expression"}
		}

		bytecodes = append(bytecodes, NewBytecode("BINARY_MAP_INDEX"))
		return bytecodes, exprConcreteType.KeyType, nil
	default:
		return nil, nil, &CompileError{"only maps and lists can be indexed"}
	}
}

func (compiler *Compiler) compileInfixNode(tree *InfixNode) ([]*Bytecode, VeniceType, error) {
	leftBytecodes, leftType, err := compiler.compileExpression(tree.Left)
	if err != nil {
		return nil, nil, err
	}

	if !checkInfixLeftType(tree.Operator, leftType) {
		return nil, nil, &CompileError{fmt.Sprintf("invalid type for left operand of %s", tree.Operator)}
	}

	rightBytecodes, rightType, err := compiler.compileExpression(tree.Right)
	if err != nil {
		return nil, nil, err
	}

	if !checkInfixRightType(tree.Operator, leftType, rightType) {
		return nil, nil, &CompileError{fmt.Sprintf("invalid type for right operand of %s", tree.Operator)}
	}

	bytecodes := append(leftBytecodes, rightBytecodes...)
	switch tree.Operator {
	case "+":
		return append(bytecodes, NewBytecode("BINARY_ADD")), VENICE_TYPE_INTEGER, nil
	case "/":
		return append(bytecodes, NewBytecode("BINARY_DIV")), VENICE_TYPE_INTEGER, nil
	case "==":
		return append(bytecodes, NewBytecode("BINARY_EQ")), VENICE_TYPE_BOOLEAN, nil
	case "*":
		return append(bytecodes, NewBytecode("BINARY_MUL")), VENICE_TYPE_INTEGER, nil
	case "-":
		return append(bytecodes, NewBytecode("BINARY_SUB")), VENICE_TYPE_INTEGER, nil
	default:
		return nil, nil, &CompileError{fmt.Sprintf("unknown oeprator: %s", tree.Operator)}
	}
}

func checkInfixLeftType(operator string, leftType VeniceType) bool {
	switch operator {
	case "==":
		return true
	default:
		return areTypesCompatible(VENICE_TYPE_INTEGER, leftType)
	}
}

func checkInfixRightType(operator string, leftType VeniceType, rightType VeniceType) bool {
	switch operator {
	case "==":
		return areTypesCompatible(leftType, rightType)
	default:
		return areTypesCompatible(VENICE_TYPE_INTEGER, leftType)
	}
}

func areTypesCompatible(expectedType VeniceType, actualType VeniceType) bool {
	switch v1 := expectedType.(type) {
	case *VeniceAtomicType:
		switch v2 := actualType.(type) {
		case *VeniceAtomicType:
			return v1.Type == v2.Type
		default:
			return false
		}
	case *VeniceListType:
		switch v2 := actualType.(type) {
		case *VeniceListType:
			return areTypesCompatible(v1.ItemType, v2.ItemType)
		default:
			return false
		}
	default:
		return false
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
	VENICE_TYPE_BOOLEAN_LABEL = "bool"
	VENICE_TYPE_INTEGER_LABEL = "int"
	VENICE_TYPE_STRING_LABEL  = "string"
)

var (
	VENICE_TYPE_BOOLEAN = &VeniceAtomicType{VENICE_TYPE_BOOLEAN_LABEL}
	VENICE_TYPE_INTEGER = &VeniceAtomicType{VENICE_TYPE_INTEGER_LABEL}
	VENICE_TYPE_STRING  = &VeniceAtomicType{VENICE_TYPE_STRING_LABEL}
)
