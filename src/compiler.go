package main

import "fmt"

type Compiler struct {
	symbolTable           *SymbolTable
	typeSymbolTable       map[string]VeniceType
	inFunctionDeclaration bool
	functionReturnType    VeniceType
}

func NewCompiler() *Compiler {
	return &Compiler{NewBuiltinSymbolTable(), NewBuiltinTypeSymbolTable(), false, nil}
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

func (compiler *Compiler) Compile(tree *ProgramNode) (CompiledProgram, error) {
	compiledProgram := NewCompiledProgram()
	for _, statementInterface := range tree.Statements {
		switch statement := statementInterface.(type) {
		case *FunctionDeclarationNode:
			code, err := compiler.compileFunctionDeclaration(statement)
			if err != nil {
				return nil, err
			}
			compiledProgram[statement.Name] = code
		default:
			code, err := compiler.compileStatement(statementInterface)
			if err != nil {
				return nil, err
			}
			compiledProgram["main"] = append(compiledProgram["main"], code...)
		}
	}
	return compiledProgram, nil
}

func (compiler *Compiler) compileStatement(treeInterface StatementNode) ([]*Bytecode, error) {
	switch tree := treeInterface.(type) {
	case *AssignStatementNode:
		code, eType, err := compiler.compileExpression(tree.Expr)
		if err != nil {
			return nil, err
		}

		expectedType, ok := compiler.symbolTable.Get(tree.Symbol)
		if !ok {
			return nil, &CompileError{fmt.Sprintf("cannot assign to undeclared symbol %q", tree.Symbol)}
		} else if !areTypesCompatible(expectedType, eType) {
			return nil, &CompileError{fmt.Sprintf("wrong expression type in assignment to %q", tree.Symbol)}
		}

		code = append(code, NewBytecode("STORE_NAME", &VeniceString{tree.Symbol}))
		return code, nil
	case *BreakStatementNode:
		return []*Bytecode{NewBytecode("BREAK_LOOP")}, nil
	case *ContinueStatementNode:
		return []*Bytecode{NewBytecode("CONTINUE_LOOP")}, nil
	case *ExpressionStatementNode:
		code, _, err := compiler.compileExpression(tree.Expr)
		if err != nil {
			return nil, err
		}
		return code, nil
	case *IfStatementNode:
		return compiler.compileIfStatement(tree)
	case *LetStatementNode:
		code, eType, err := compiler.compileExpression(tree.Expr)
		if err != nil {
			return nil, err
		}
		compiler.symbolTable.Put(tree.Symbol, eType)
		return append(code, NewBytecode("STORE_NAME", &VeniceString{tree.Symbol})), nil
	case *ReturnStatementNode:
		if !compiler.inFunctionDeclaration {
			return nil, &CompileError{"return statement outside of function definition"}
		}

		if tree.Expr != nil {
			code, exprType, err := compiler.compileExpression(tree.Expr)
			if err != nil {
				return nil, err
			}

			if !areTypesCompatible(compiler.functionReturnType, exprType) {
				return nil, &CompileError{"conflicting function return types"}
			}

			code = append(code, NewBytecode("RETURN"))
			return code, err
		} else {
			if !areTypesCompatible(compiler.functionReturnType, nil) {
				return nil, &CompileError{"function cannot return void"}
			}

			return []*Bytecode{NewBytecode("RETURN")}, nil
		}
	case *WhileLoopNode:
		return compiler.compileWhileLoop(tree)
	default:
		return nil, &CompileError{fmt.Sprintf("unknown statement type: %T", treeInterface)}
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

	declaredReturnType, err := compiler.resolveType(tree.ReturnType)
	if err != nil {
		return nil, err
	}

	// Put the function's entry in the symbol table before compiling the body so that
	// recursive functions can call themselves.
	compiler.symbolTable.Put(tree.Name, &VeniceFunctionType{paramTypes, declaredReturnType})

	compiler.symbolTable = bodySymbolTable
	compiler.inFunctionDeclaration = true
	compiler.functionReturnType = declaredReturnType
	bodyCode, err := compiler.compileBlock(tree.Body)
	compiler.inFunctionDeclaration = false
	compiler.functionReturnType = nil
	compiler.symbolTable = bodySymbolTable.parent

	// TODO(2021-08-07): Detect non-void functions which lack a return statement.

	paramLoadCode := []*Bytecode{}
	for _, param := range tree.Params {
		paramLoadCode = append(paramLoadCode, NewBytecode("STORE_NAME", &VeniceString{param.Name}))
	}
	bodyCode = append(paramLoadCode, bodyCode...)

	if err != nil {
		return nil, err
	}

	return bodyCode, nil
}

func (compiler *Compiler) resolveType(typeNodeInterface TypeNode) (VeniceType, error) {
	if typeNodeInterface == nil {
		return nil, nil
	}

	switch typeNode := typeNodeInterface.(type) {
	case *SimpleTypeNode:
		resolvedType, ok := compiler.typeSymbolTable[typeNode.Symbol]
		if !ok {
			return nil, &CompileError{fmt.Sprintf("unknown type: %s", typeNode.Symbol)}
		}
		return resolvedType, nil
	default:
		return nil, &CompileError{fmt.Sprintf("unknown type node: %T", typeNodeInterface)}
	}
}

func (compiler *Compiler) compileWhileLoop(tree *WhileLoopNode) ([]*Bytecode, error) {
	conditionCode, conditionType, err := compiler.compileExpression(tree.Condition)
	if err != nil {
		return nil, err
	}

	if !areTypesCompatible(VENICE_TYPE_BOOLEAN, conditionType) {
		return nil, &CompileError{"condition of while loop must be a boolean"}
	}

	bodyCode, err := compiler.compileBlock(tree.Body)
	if err != nil {
		return nil, err
	}

	code := conditionCode
	jumpForward := len(bodyCode) + 2
	code = append(code, NewBytecode("REL_JUMP_IF_FALSE", &VeniceInteger{jumpForward}))
	code = append(code, bodyCode...)
	jumpBack := -(len(conditionCode) + len(bodyCode) + 1)
	code = append(code, NewBytecode("REL_JUMP", &VeniceInteger{jumpBack}))
	return code, nil
}

func (compiler *Compiler) compileIfStatement(tree *IfStatementNode) ([]*Bytecode, error) {
	conditionCode, conditionType, err := compiler.compileExpression(tree.Condition)
	if err != nil {
		return nil, err
	}

	if !areTypesCompatible(VENICE_TYPE_BOOLEAN, conditionType) {
		return nil, &CompileError{"condition of if statement must be a boolean"}
	}

	trueClauseCode, err := compiler.compileBlock(tree.TrueClause)
	if err != nil {
		return nil, err
	}

	code := conditionCode
	code = append(code, NewBytecode("REL_JUMP_IF_FALSE", &VeniceInteger{len(trueClauseCode) + 1}))
	code = append(code, trueClauseCode...)

	if tree.FalseClause != nil {
		falseClauseCode, err := compiler.compileBlock(tree.FalseClause)
		if err != nil {
			return nil, err
		}

		code = append(code, NewBytecode("REL_JUMP", &VeniceInteger{len(falseClauseCode) + 1}))
		code = append(code, falseClauseCode...)
	}

	return code, nil
}

func (compiler *Compiler) compileBlock(block []StatementNode) ([]*Bytecode, error) {
	code := []*Bytecode{}
	for _, statement := range block {
		statementCode, err := compiler.compileStatement(statement)
		if err != nil {
			return nil, err
		}

		code = append(code, statementCode...)
	}
	return code, nil
}

func (compiler *Compiler) compileExpression(treeInterface ExpressionNode) ([]*Bytecode, VeniceType, error) {
	switch tree := treeInterface.(type) {
	case *BooleanNode:
		return []*Bytecode{NewBytecode("PUSH_CONST", &VeniceBoolean{tree.Value})}, VENICE_TYPE_BOOLEAN, nil
	case *CallNode:
		return compiler.compileCallNode(tree)
	case *IndexNode:
		return compiler.compileIndexNode(tree)
	case *InfixNode:
		return compiler.compileInfixNode(tree)
	case *IntegerNode:
		return []*Bytecode{NewBytecode("PUSH_CONST", &VeniceInteger{tree.Value})}, VENICE_TYPE_INTEGER, nil
	case *ListNode:
		code := []*Bytecode{}
		var itemType VeniceType
		for i := len(tree.Values) - 1; i >= 0; i-- {
			value := tree.Values[i]
			valueCode, valueType, err := compiler.compileExpression(value)
			if err != nil {
				return nil, nil, err
			}

			if itemType == nil {
				itemType = valueType
			} else if !areTypesCompatible(itemType, valueType) {
				return nil, nil, &CompileError{"list elements must all be of same type"}
			}

			code = append(code, valueCode...)
		}
		code = append(code, NewBytecode("BUILD_LIST", &VeniceInteger{len(tree.Values)}))
		return code, &VeniceListType{itemType}, nil
	case *MapNode:
		code := []*Bytecode{}
		var keyType VeniceType
		var valueType VeniceType
		for i := len(tree.Pairs) - 1; i >= 0; i-- {
			pair := tree.Pairs[i]
			keyCode, thisKeyType, err := compiler.compileExpression(pair.Key)
			if err != nil {
				return nil, nil, err
			}

			if keyType == nil {
				keyType = thisKeyType
			} else if !areTypesCompatible(keyType, thisKeyType) {
				return nil, nil, &CompileError{"map keys must all be of the same type"}
			}

			code = append(code, keyCode...)

			valueCode, thisValueType, err := compiler.compileExpression(pair.Value)
			if err != nil {
				return nil, nil, err
			}

			if valueType == nil {
				valueType = thisValueType
			} else if !areTypesCompatible(valueType, thisValueType) {
				return nil, nil, &CompileError{"map values must all be of the same type"}
			}

			code = append(code, valueCode...)
		}
		code = append(code, NewBytecode("BUILD_MAP", &VeniceInteger{len(tree.Pairs)}))
		return code, &VeniceMapType{keyType, valueType}, nil
	case *StringNode:
		return []*Bytecode{NewBytecode("PUSH_CONST", &VeniceString{tree.Value})}, VENICE_TYPE_STRING, nil
	case *SymbolNode:
		symbolType, ok := compiler.symbolTable.Get(tree.Value)
		if !ok {
			return nil, nil, &CompileError{fmt.Sprintf("undefined symbol: %s", tree.Value)}
		}
		return []*Bytecode{NewBytecode("PUSH_NAME", &VeniceString{tree.Value})}, symbolType, nil
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

			code, argType, err := compiler.compileExpression(tree.Args[0])
			if err != nil {
				return nil, nil, err
			}

			if argType != VENICE_TYPE_INTEGER && argType != VENICE_TYPE_STRING {
				return nil, nil, &CompileError{"`print`'s argument must be an integer or string"}
			}

			code = append(code, NewBytecode("CALL_BUILTIN", &VeniceString{"print"}))
			return code, nil, nil
		} else {
			valueInterface, ok := compiler.symbolTable.Get(v.Value)
			if !ok {
				return nil, nil, &CompileError{fmt.Sprintf("undefined symbol: %s", v.Value)}
			}

			if f, ok := valueInterface.(*VeniceFunctionType); ok {
				if len(f.ParamTypes) != len(tree.Args) {
					return nil, nil, &CompileError{fmt.Sprintf("wrong number of arguments: expected %d, got %d", len(f.ParamTypes), len(tree.Args))}
				}

				code := []*Bytecode{}
				for i := 0; i < len(f.ParamTypes); i++ {
					argCode, argType, err := compiler.compileExpression(tree.Args[i])
					if err != nil {
						return nil, nil, err
					}

					if !areTypesCompatible(f.ParamTypes[i], argType) {
						return nil, nil, &CompileError{"wrong function parameter type"}
					}

					code = append(code, argCode...)
				}

				code = append(code, NewBytecode("CALL_FUNCTION", &VeniceString{v.Value}, &VeniceInteger{len(f.ParamTypes)}))
				return code, f.ReturnType, nil
			} else {
				return nil, nil, &CompileError{"not a function"}
			}
		}
	}

	return nil, nil, &CompileError{"function calls not implemented yet"}
}

func (compiler *Compiler) compileIndexNode(tree *IndexNode) ([]*Bytecode, VeniceType, error) {
	exprCode, exprTypeInterface, err := compiler.compileExpression(tree.Expr)
	if err != nil {
		return nil, nil, err
	}

	indexCode, indexType, err := compiler.compileExpression(tree.Index)
	if err != nil {
		return nil, nil, err
	}

	code := append(exprCode, indexCode...)

	switch exprType := exprTypeInterface.(type) {
	case *VeniceListType:
		if !areTypesCompatible(VENICE_TYPE_INTEGER, indexType) {
			return nil, nil, &CompileError{"list index must be integer"}
		}

		code = append(code, NewBytecode("BINARY_LIST_INDEX"))
		return code, exprType.ItemType, nil
	case *VeniceMapType:
		if !areTypesCompatible(exprType.KeyType, indexType) {
			return nil, nil, &CompileError{"wrong map key type in index expression"}
		}

		code = append(code, NewBytecode("BINARY_MAP_INDEX"))
		return code, exprType.KeyType, nil
	default:
		return nil, nil, &CompileError{"only maps and lists can be indexed"}
	}
}

func (compiler *Compiler) compileInfixNode(tree *InfixNode) ([]*Bytecode, VeniceType, error) {
	leftCode, leftType, err := compiler.compileExpression(tree.Left)
	if err != nil {
		return nil, nil, err
	}

	if !checkInfixLeftType(tree.Operator, leftType) {
		return nil, nil, &CompileError{fmt.Sprintf("invalid type for left operand of %s", tree.Operator)}
	}

	rightCode, rightType, err := compiler.compileExpression(tree.Right)
	if err != nil {
		return nil, nil, err
	}

	if !checkInfixRightType(tree.Operator, leftType, rightType) {
		return nil, nil, &CompileError{fmt.Sprintf("invalid type for right operand of %s", tree.Operator)}
	}

	code := append(leftCode, rightCode...)
	switch tree.Operator {
	case "+":
		return append(code, NewBytecode("BINARY_ADD")), VENICE_TYPE_INTEGER, nil
	case "/":
		return append(code, NewBytecode("BINARY_DIV")), VENICE_TYPE_INTEGER, nil
	case "==":
		return append(code, NewBytecode("BINARY_EQ")), VENICE_TYPE_BOOLEAN, nil
	case ">":
		return append(code, NewBytecode("BINARY_GT")), VENICE_TYPE_BOOLEAN, nil
	case ">=":
		return append(code, NewBytecode("BINARY_GT_EQ")), VENICE_TYPE_BOOLEAN, nil
	case "<":
		return append(code, NewBytecode("BINARY_LT")), VENICE_TYPE_BOOLEAN, nil
	case "<=":
		return append(code, NewBytecode("BINARY_LT_EQ")), VENICE_TYPE_BOOLEAN, nil
	case "*":
		return append(code, NewBytecode("BINARY_MUL")), VENICE_TYPE_INTEGER, nil
	case "!=":
		return append(code, NewBytecode("BINARY_NOT_EQ")), VENICE_TYPE_BOOLEAN, nil
	case "-":
		return append(code, NewBytecode("BINARY_SUB")), VENICE_TYPE_INTEGER, nil
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

func areTypesCompatible(expectedTypeInterface VeniceType, actualTypeInterface VeniceType) bool {
	if expectedTypeInterface == nil && actualTypeInterface == nil {
		return true
	}

	switch expectedType := expectedTypeInterface.(type) {
	case *VeniceAtomicType:
		switch actualType := actualTypeInterface.(type) {
		case *VeniceAtomicType:
			return expectedType.Type == actualType.Type
		default:
			return false
		}
	case *VeniceListType:
		switch actualType := actualTypeInterface.(type) {
		case *VeniceListType:
			return areTypesCompatible(expectedType.ItemType, actualType.ItemType)
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
