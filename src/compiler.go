package main

import "fmt"

type Compiler struct {
	symbolTable     *SymbolTable
	typeSymbolTable map[string]VeniceType
}

func NewCompiler() *Compiler {
	return &Compiler{NewBuiltinSymbolTable(), NewBuiltinTypeSymbolTable()}
}

type SymbolTable struct {
	parent  *SymbolTable
	symbols map[string]VeniceType
}

func NewBuiltinSymbolTable() *SymbolTable {
	symbols := map[string]VeniceType{}
	return &SymbolTable{nil, symbols}
}

func NewBuiltinTypeSymbolTable() map[string]VeniceType {
	return map[string]VeniceType{
		"int": VENICE_TYPE_INTEGER,
	}
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

func (compiler *Compiler) compileStatement(tree StatementNode) ([]*Bytecode, error) {
	switch v := tree.(type) {
	case *BreakStatementNode:
		return []*Bytecode{NewBytecode("BREAK_LOOP")}, nil
	case *ContinueStatementNode:
		return []*Bytecode{NewBytecode("CONTINUE_LOOP")}, nil
	case *ExpressionStatementNode:
		bytecodes, _, err := compiler.compileExpression(v.Expr)
		if err != nil {
			return nil, err
		}
		return bytecodes, nil
	case *FunctionDeclarationNode:
		return compiler.compileFunctionDeclaration(v)
	case *IfStatementNode:
		return compiler.compileIfStatement(v)
	case *LetStatementNode:
		bytecodes, eType, err := compiler.compileExpression(v.Expr)
		if err != nil {
			return nil, err
		}
		compiler.symbolTable.symbols[v.Symbol] = eType
		return append(bytecodes, NewBytecode("STORE_NAME", &VeniceString{v.Symbol})), nil
	case *WhileLoopNode:
		return compiler.compileWhileLoop(v)
	default:
		return nil, &CompileError{"unknown statement type"}
	}
}

func (compiler *Compiler) compileStatementWithReturn(tree StatementNode) ([]*Bytecode, VeniceType, error) {
	switch v := tree.(type) {
	case *ReturnStatementNode:
		bytecodes, exprType, err := compiler.compileExpression(v.Expr)
		if err != nil {
			return nil, nil, err
		}
		bytecodes = append(bytecodes, NewBytecode("RETURN"))
		return bytecodes, exprType, err
	default:
		bytecodes, err := compiler.compileStatement(tree)
		return bytecodes, nil, err
	}
}

func (compiler *Compiler) compileFunctionDeclaration(tree *FunctionDeclarationNode) ([]*Bytecode, error) {
	params := []string{}
	paramTypes := []VeniceType{}
	bodySymbolTableMap := map[string]VeniceType{}
	for _, param := range tree.Params {
		paramType, err := compiler.resolveType(param.ParamType)
		if err != nil {
			return nil, err
		}

		params = append(params, param.Name)
		paramTypes = append(paramTypes, paramType)

		bodySymbolTableMap[param.Name] = paramType
	}
	bodySymbolTable := &SymbolTable{compiler.symbolTable, bodySymbolTableMap}

	compiler.symbolTable = bodySymbolTable
	bodyBytecodes, returnType, err := compiler.compileBlockWithReturn(tree.Body)
	compiler.symbolTable = bodySymbolTable.parent

	if err != nil {
		return nil, err
	}

	declaredReturnType, err := compiler.resolveType(tree.ReturnType)
	if err != nil {
		return nil, err
	}

	if !areTypesCompatible(declaredReturnType, returnType) {
		return nil, &CompileError{"actual return type does not match declared return type"}
	}

	bytecodes := []*Bytecode{
		// TODO(2021-08-03): This is not serializable.
		NewBytecode("PUSH_CONST", &VeniceFunction{params, bodyBytecodes}),
		NewBytecode("STORE_NAME", &VeniceString{tree.Name}),
	}
	compiler.symbolTable.Put(tree.Name, &VeniceFunctionType{paramTypes, returnType})
	return bytecodes, nil
}

func (compiler *Compiler) resolveType(typeNodeUntyped TypeNode) (VeniceType, error) {
	switch typeNode := typeNodeUntyped.(type) {
	case *SimpleTypeNode:
		resolvedType, ok := compiler.typeSymbolTable[typeNode.Symbol]
		if !ok {
			return nil, &CompileError{fmt.Sprintf("unknown type: %s", typeNode.Symbol)}
		}
		return resolvedType, nil
	default:
		return nil, &CompileError{"unknown type node"}
	}
}

func (compiler *Compiler) compileWhileLoop(tree *WhileLoopNode) ([]*Bytecode, error) {
	conditionBytecodes, conditionType, err := compiler.compileExpression(tree.Condition)
	if err != nil {
		return nil, err
	}

	if !areTypesCompatible(VENICE_TYPE_BOOLEAN, conditionType) {
		return nil, &CompileError{"condition of while loop must be a boolean"}
	}

	bodyBytecodes, err := compiler.compileBlock(tree.Body)
	if err != nil {
		return nil, err
	}

	bytecodes := conditionBytecodes
	jumpForward := len(bodyBytecodes) + 2
	bytecodes = append(bytecodes, NewBytecode("REL_JUMP_IF_FALSE", &VeniceInteger{jumpForward}))
	bytecodes = append(bytecodes, bodyBytecodes...)
	jumpBack := -(len(conditionBytecodes) + len(bodyBytecodes) + 1)
	bytecodes = append(bytecodes, NewBytecode("REL_JUMP", &VeniceInteger{jumpBack}))
	return bytecodes, nil
}

func (compiler *Compiler) compileIfStatement(tree *IfStatementNode) ([]*Bytecode, error) {
	conditionBytecodes, conditionType, err := compiler.compileExpression(tree.Condition)
	if err != nil {
		return nil, err
	}

	if !areTypesCompatible(VENICE_TYPE_BOOLEAN, conditionType) {
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

func (compiler *Compiler) compileBlock(block []StatementNode) ([]*Bytecode, error) {
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

func (compiler *Compiler) compileBlockWithReturn(block []StatementNode) ([]*Bytecode, VeniceType, error) {
	bytecodes := []*Bytecode{}
	var returnType VeniceType
	for _, statement := range block {
		statementBytecodes, thisReturnType, err := compiler.compileStatementWithReturn(statement)
		if err != nil {
			return nil, nil, err
		}

		if thisReturnType != nil {
			if returnType == nil {
				returnType = thisReturnType
			} else if !areTypesCompatible(returnType, thisReturnType) {
				return nil, nil, &CompileError{"multiple incompatible return statements in function"}
			}
		}

		bytecodes = append(bytecodes, statementBytecodes...)
	}
	return bytecodes, returnType, nil
}

func (compiler *Compiler) compileExpression(tree ExpressionNode) ([]*Bytecode, VeniceType, error) {
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
		} else {
			fUntyped, ok := compiler.symbolTable.Get(v.Value)
			if !ok {
				return nil, nil, &CompileError{fmt.Sprintf("undefined symbol: %s", v.Value)}
			}

			if f, ok := fUntyped.(*VeniceFunctionType); ok {
				if len(f.ParamTypes) != len(tree.Args) {
					return nil, nil, &CompileError{fmt.Sprintf("wrong number of arguments: expected %d, got %d", len(f.ParamTypes), len(tree.Args))}
				}

				bytecodes := []*Bytecode{}
				for i := 0; i < len(f.ParamTypes); i++ {
					argBytecodes, argType, err := compiler.compileExpression(tree.Args[i])
					if err != nil {
						return nil, nil, err
					}

					if !areTypesCompatible(f.ParamTypes[i], argType) {
						return nil, nil, &CompileError{"wrong function parameter type"}
					}

					bytecodes = append(bytecodes, argBytecodes...)
				}

				bytecodes = append(bytecodes, NewBytecode("PUSH_NAME", &VeniceString{v.Value}))
				bytecodes = append(bytecodes, NewBytecode("CALL_FUNCTION", &VeniceInteger{len(f.ParamTypes)}))
				return bytecodes, f.ReturnType, nil
			} else {
				return nil, nil, &CompileError{"not a function"}
			}
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

type CompileError struct {
	Message string
}

func (e *CompileError) Error() string {
	return e.Message
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
