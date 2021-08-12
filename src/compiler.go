package main

import "fmt"

type Compiler struct {
	symbolTable     *SymbolTable
	typeSymbolTable *SymbolTable
	functionInfo    *FunctionInfo
	nestedLoopCount int
}

type FunctionInfo struct {
	declaredReturnType  VeniceType
	seenReturnStatement bool
}

func NewCompiler() *Compiler {
	return &Compiler{
		symbolTable:     NewBuiltinSymbolTable(),
		typeSymbolTable: NewBuiltinTypeSymbolTable(),
		functionInfo:    nil,
		nestedLoopCount: 0,
	}
}

type SymbolTable struct {
	parent  *SymbolTable
	symbols map[string]VeniceType
}

func NewBuiltinSymbolTable() *SymbolTable {
	symbols := map[string]VeniceType{}
	return &SymbolTable{nil, symbols}
}

func NewBuiltinTypeSymbolTable() *SymbolTable {
	symbols := map[string]VeniceType{
		"bool":   VENICE_TYPE_BOOLEAN,
		"int":    VENICE_TYPE_INTEGER,
		"string": VENICE_TYPE_STRING,
		"Optional": &VeniceGenericType{
			[]string{"T"},
			&VeniceEnumType{
				[]*VeniceCaseType{
					&VeniceCaseType{
						"Some",
						[]VeniceType{&VeniceGenericParameterType{"T"}},
					},
					&VeniceCaseType{"None", nil},
				},
			},
		},
	}
	return &SymbolTable{nil, symbols}
}

func (compiler *Compiler) Compile(tree *ProgramNode) (CompiledProgram, error) {
	compiledProgram := NewCompiledProgram()
	for _, statementInterface := range tree.Statements {
		switch statement := statementInterface.(type) {
		case *ClassDeclarationNode:
			code, err := compiler.compileClassDeclaration(statement)
			if err != nil {
				return nil, err
			}
			compiledProgram[statement.Name] = code
		case *EnumDeclarationNode:
			err := compiler.compileEnumDeclaration(statement)
			if err != nil {
				return nil, err
			}
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
			return nil, compiler.customError(treeInterface, fmt.Sprintf("cannot assign to undeclared symbol %q", tree.Symbol))
		} else if !areTypesCompatible(expectedType, eType) {
			return nil, compiler.customError(treeInterface, fmt.Sprintf("wrong expression type in assignment to %q", tree.Symbol))
		}

		code = append(code, NewBytecode("STORE_NAME", &VeniceString{tree.Symbol}))
		return code, nil
	case *BreakStatementNode:
		if compiler.nestedLoopCount == 0 {
			return nil, compiler.customError(treeInterface, "break statement outside of loop")
		}

		// BREAK_LOOP is a temporary bytecode instruction that the compiler will later
		// convert to a REL_JUMP instruction.
		return []*Bytecode{NewBytecode("BREAK_LOOP")}, nil
	case *ContinueStatementNode:
		if compiler.nestedLoopCount == 0 {
			return nil, compiler.customError(treeInterface, "break statement outside of loop")
		}

		// CONTINUE_LOOP is a temporary bytecode instruction that the compiler will later
		// convert to a REL_JUMP instruction.
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
		if _, ok := compiler.symbolTable.Get(tree.Symbol); ok {
			return nil, compiler.customError(treeInterface, fmt.Sprintf("re-declaration of symbol %q", tree.Symbol))
		}

		code, eType, err := compiler.compileExpression(tree.Expr)
		if err != nil {
			return nil, err
		}
		compiler.symbolTable.Put(tree.Symbol, eType)
		return append(code, NewBytecode("STORE_NAME", &VeniceString{tree.Symbol})), nil
	case *ReturnStatementNode:
		if compiler.functionInfo == nil {
			return nil, compiler.customError(treeInterface, "return statement outside of function definition")
		}

		if tree.Expr != nil {
			code, exprType, err := compiler.compileExpression(tree.Expr)
			if err != nil {
				return nil, err
			}

			if !areTypesCompatible(compiler.functionInfo.declaredReturnType, exprType) {
				return nil, compiler.customError(treeInterface, "conflicting function return types")
			}

			compiler.functionInfo.seenReturnStatement = true
			code = append(code, NewBytecode("RETURN"))
			return code, err
		} else {
			if !areTypesCompatible(compiler.functionInfo.declaredReturnType, nil) {
				return nil, compiler.customError(treeInterface, "function cannot return void")
			}

			return []*Bytecode{NewBytecode("RETURN")}, nil
		}
	case *WhileLoopNode:
		return compiler.compileWhileLoop(tree)
	default:
		return nil, compiler.customError(treeInterface, fmt.Sprintf("unknown statement type: %T", treeInterface))
	}
}

func (compiler *Compiler) compileEnumDeclaration(tree *EnumDeclarationNode) error {
	if tree.GenericTypeParameter != "" {
		subTypeSymbolTable := &SymbolTable{
			compiler.typeSymbolTable,
			map[string]VeniceType{
				tree.GenericTypeParameter: &VeniceGenericParameterType{tree.GenericTypeParameter},
			},
		}
		compiler.typeSymbolTable = subTypeSymbolTable
	}

	caseTypes := []*VeniceCaseType{}
	for _, caseNode := range tree.Cases {
		types := []VeniceType{}
		for _, typeNode := range caseNode.Types {
			veniceType, err := compiler.resolveType(typeNode)
			if err != nil {
				return err
			}
			types = append(types, veniceType)
		}
		caseTypes = append(caseTypes, &VeniceCaseType{caseNode.Label, types})
	}

	if tree.GenericTypeParameter != "" {
		compiler.typeSymbolTable = compiler.typeSymbolTable.parent
		compiler.typeSymbolTable.Put(
			tree.Name,
			&VeniceGenericType{
				[]string{tree.GenericTypeParameter},
				&VeniceEnumType{caseTypes},
			},
		)
	} else {
		compiler.typeSymbolTable.Put(tree.Name, &VeniceEnumType{caseTypes})
	}

	return nil
}

func (compiler *Compiler) compileClassDeclaration(tree *ClassDeclarationNode) ([]*Bytecode, error) {
	if tree.GenericTypeParameter != "" {
		subTypeSymbolTable := &SymbolTable{
			compiler.typeSymbolTable,
			map[string]VeniceType{
				tree.GenericTypeParameter: &VeniceGenericParameterType{tree.GenericTypeParameter},
			},
		}
		compiler.typeSymbolTable = subTypeSymbolTable
	}

	fields := []*VeniceClassField{}
	paramTypes := []VeniceType{}
	for _, field := range tree.Fields {
		paramType, err := compiler.resolveType(field.FieldType)
		if err != nil {
			return nil, err
		}
		paramTypes = append(paramTypes, paramType)
		fields = append(fields, &VeniceClassField{field.Name, field.Public, paramType})
	}

	var classType VeniceType
	if tree.GenericTypeParameter == "" {
		classType = &VeniceClassType{fields}
	} else {
		classType = &VeniceGenericType{[]string{tree.GenericTypeParameter}, &VeniceClassType{fields}}
		compiler.typeSymbolTable = compiler.typeSymbolTable.parent
	}
	compiler.typeSymbolTable.Put(tree.Name, classType)

	constructorType := &VeniceFunctionType{paramTypes, classType}
	compiler.symbolTable.Put(tree.Name, constructorType)

	constructorBytecode := []*Bytecode{NewBytecode("BUILD_CLASS", &VeniceInteger{len(fields)})}
	return constructorBytecode, nil
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
	compiler.functionInfo = &FunctionInfo{
		declaredReturnType:  declaredReturnType,
		seenReturnStatement: false,
	}
	bodyCode, err := compiler.compileBlock(tree.Body)

	if declaredReturnType != nil && !compiler.functionInfo.seenReturnStatement {
		return nil, compiler.customError(tree, "non-void function has no return statement")
	}

	compiler.functionInfo = nil
	compiler.symbolTable = bodySymbolTable.parent

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
		resolvedType, ok := compiler.typeSymbolTable.Get(typeNode.Symbol)
		if !ok {
			return nil, compiler.customError(typeNodeInterface, fmt.Sprintf("unknown type: %s", typeNode.Symbol))
		}
		return resolvedType, nil
	default:
		return nil, compiler.customError(typeNodeInterface, fmt.Sprintf("unknown type node: %T", typeNodeInterface))
	}
}

func (compiler *Compiler) compileWhileLoop(tree *WhileLoopNode) ([]*Bytecode, error) {
	conditionCode, conditionType, err := compiler.compileExpression(tree.Condition)
	if err != nil {
		return nil, err
	}

	if !areTypesCompatible(VENICE_TYPE_BOOLEAN, conditionType) {
		return nil, compiler.customError(tree.Condition, "condition of while loop must be a boolean")
	}

	compiler.nestedLoopCount += 1
	bodyCode, err := compiler.compileBlock(tree.Body)
	compiler.nestedLoopCount -= 1
	if err != nil {
		return nil, err
	}

	code := conditionCode
	jumpForward := len(bodyCode) + 2
	code = append(code, NewBytecode("REL_JUMP_IF_FALSE", &VeniceInteger{jumpForward}))
	code = append(code, bodyCode...)
	jumpBack := -(len(conditionCode) + len(bodyCode) + 1)
	code = append(code, NewBytecode("REL_JUMP", &VeniceInteger{jumpBack}))

	for i, bytecode := range code {
		if bytecode.Name == "BREAK_LOOP" {
			bytecode.Name = "REL_JUMP"
			bytecode.Args = append(bytecode.Args, &VeniceInteger{len(code) - i})
		} else if bytecode.Name == "CONTINUE_LOOP" {
			bytecode.Name = "REL_JUMP"
			bytecode.Args = append(bytecode.Args, &VeniceInteger{-i})
		}
	}

	return code, nil
}

func (compiler *Compiler) compileIfStatement(tree *IfStatementNode) ([]*Bytecode, error) {
	conditionCode, conditionType, err := compiler.compileExpression(tree.Condition)
	if err != nil {
		return nil, err
	}

	if !areTypesCompatible(VENICE_TYPE_BOOLEAN, conditionType) {
		return nil, compiler.customError(tree.Condition, "condition of if statement must be a boolean")
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
	case *CharacterNode:
		return []*Bytecode{NewBytecode("PUSH_CONST", &VeniceCharacter{tree.Value})}, VENICE_TYPE_CHARACTER, nil
	case *EnumSymbolNode:
		return compiler.compileEnumSymbolNode(tree)
	case *FieldAccessNode:
		return compiler.compileFieldAccessNode(tree)
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
				return nil, nil, compiler.customError(value, "list elements must all be of same type")
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
				return nil, nil, compiler.customError(pair.Key, "map keys must all be of the same type")
			}

			code = append(code, keyCode...)

			valueCode, thisValueType, err := compiler.compileExpression(pair.Value)
			if err != nil {
				return nil, nil, err
			}

			if valueType == nil {
				valueType = thisValueType
			} else if !areTypesCompatible(valueType, thisValueType) {
				return nil, nil, compiler.customError(pair.Value, "map values must all be of the same type")
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
			return nil, nil, compiler.customError(treeInterface, fmt.Sprintf("undefined symbol: %s", tree.Value))
		}
		return []*Bytecode{NewBytecode("PUSH_NAME", &VeniceString{tree.Value})}, symbolType, nil
	default:
		return nil, nil, compiler.customError(treeInterface, fmt.Sprintf("unknown expression type: %T", treeInterface))
	}
}

func (compiler *Compiler) compileEnumSymbolNode(tree *EnumSymbolNode) ([]*Bytecode, VeniceType, error) {
	enumTypeInterface, err := compiler.resolveType(&SimpleTypeNode{tree.Enum, nil})
	if err != nil {
		return nil, nil, err
	}

	isEnum := false
	enumType, ok := enumTypeInterface.(*VeniceEnumType)
	if ok {
		isEnum = true
	} else {
		genericType, ok := enumTypeInterface.(*VeniceGenericType)
		if ok {
			enumType, ok = genericType.GenericType.(*VeniceEnumType)
			isEnum = ok
		} else {
			isEnum = false
		}
	}

	if !isEnum {
		return nil, nil, compiler.customError(tree, "cannot use double colon after non-enum type")
	}

	for _, enumCase := range enumType.Cases {
		if enumCase.Label == tree.Case {
			if len(enumCase.Types) != 0 {
				return nil, nil, compiler.customError(tree, fmt.Sprintf("%s takes %d argument(s)", enumCase.Label, len(enumCase.Types)))
			}

			return []*Bytecode{NewBytecode("PUSH_ENUM", &VeniceString{tree.Case}, &VeniceInteger{0})}, enumType, nil
		}
	}

	return nil, nil, compiler.customError(tree, fmt.Sprintf("enum %s does not have case %s", tree.Enum, tree.Case))
}

func (compiler *Compiler) compileFieldAccessNode(tree *FieldAccessNode) ([]*Bytecode, VeniceType, error) {
	code, typeInterface, err := compiler.compileExpression(tree.Expr)
	if err != nil {
		return nil, nil, err
	}

	if classType, ok := typeInterface.(*VeniceClassType); ok {
		for i, field := range classType.Fields {
			if field.Name == tree.Name {
				// TODO(2021-08-09): Allow this when inside the class itself.
				if !field.Public {
					return nil, nil, compiler.customError(tree, "use of private field")
				}

				code = append(code, NewBytecode("PUSH_FIELD", &VeniceInteger{i}))
				return code, field.FieldType, nil
			}
		}

		return nil, nil, compiler.customError(tree, fmt.Sprintf("no such field: %s", tree.Name))
	} else {
		return nil, nil, compiler.customError(tree, "left-hand side of dot must be a class object")
	}
}

func (compiler *Compiler) compileCallNode(tree *CallNode) ([]*Bytecode, VeniceType, error) {
	if v, ok := tree.Function.(*SymbolNode); ok {
		if v.Value == "print" {
			if len(tree.Args) != 1 {
				return nil, nil, compiler.customError(tree, "`print` takes exactly 1 argument")
			}

			code, _, err := compiler.compileExpression(tree.Args[0])
			if err != nil {
				return nil, nil, err
			}

			code = append(code, NewBytecode("CALL_BUILTIN", &VeniceString{"print"}))
			return code, nil, nil
		} else {
			valueInterface, ok := compiler.symbolTable.Get(v.Value)
			if !ok {
				return nil, nil, compiler.customError(tree, fmt.Sprintf("undefined symbol: %s", v.Value))
			}

			if f, ok := valueInterface.(*VeniceFunctionType); ok {
				if len(f.ParamTypes) != len(tree.Args) {
					return nil, nil, compiler.customError(tree, fmt.Sprintf("wrong number of arguments: expected %d, got %d", len(f.ParamTypes), len(tree.Args)))
				}

				code := []*Bytecode{}
				genericParameters := []string{}
				concreteTypes := []VeniceType{}
				for i := 0; i < len(f.ParamTypes); i++ {
					argCode, argType, err := compiler.compileExpression(tree.Args[i])
					if err != nil {
						return nil, nil, err
					}

					if genericParamType, ok := f.ParamTypes[i].(*VeniceGenericParameterType); ok {
						genericParameters = append(genericParameters, genericParamType.Label)
						concreteTypes = append(concreteTypes, argType)
					} else {
						if !areTypesCompatible(f.ParamTypes[i], argType) {
							return nil, nil, compiler.customError(tree.Args[i], "wrong function parameter type")
						}
					}

					code = append(code, argCode...)
				}

				code = append(code, NewBytecode("CALL_FUNCTION", &VeniceString{v.Value}, &VeniceInteger{len(f.ParamTypes)}))

				if len(genericParameters) > 0 {
					return code, f.ReturnType.SubstituteGenerics(genericParameters, concreteTypes), nil
				} else {
					return code, f.ReturnType, nil
				}
			} else {
				return nil, nil, compiler.customError(tree, "not a function")
			}
		}
	}

	if v, ok := tree.Function.(*EnumSymbolNode); ok {
		enumTypeInterface, err := compiler.resolveType(&SimpleTypeNode{v.Enum, nil})
		if err != nil {
			return nil, nil, err
		}

		isEnum := false
		enumType, ok := enumTypeInterface.(*VeniceEnumType)
		if ok {
			isEnum = true
		} else {
			genericType, ok := enumTypeInterface.(*VeniceGenericType)
			if ok {
				enumType, ok = genericType.GenericType.(*VeniceEnumType)
				isEnum = ok
			} else {
				isEnum = false
			}
		}

		if !isEnum {
			return nil, nil, compiler.customError(tree, "cannot use double colon after non-enum type")
		}

		for _, enumCase := range enumType.Cases {
			if enumCase.Label == v.Case {
				if len(enumCase.Types) != len(tree.Args) {
					return nil, nil, compiler.customError(tree, fmt.Sprintf("%s takes %d argument(s)", enumCase.Label, len(enumCase.Types)))
				}

				code := []*Bytecode{}
				genericParameters := []string{}
				concreteTypes := []VeniceType{}
				for i := 0; i < len(tree.Args); i++ {
					argCode, argType, err := compiler.compileExpression(tree.Args[i])
					if err != nil {
						return nil, nil, err
					}

					if genericParamType, ok := enumCase.Types[i].(*VeniceGenericParameterType); ok {
						genericParameters = append(genericParameters, genericParamType.Label)
						concreteTypes = append(concreteTypes, argType)
					} else {
						if !areTypesCompatible(enumCase.Types[i], argType) {
							return nil, nil, compiler.customError(tree.Args[i], "wrong enum parameter type")
						}
					}

					code = append(code, argCode...)
				}
				code = append(code, NewBytecode("PUSH_ENUM", &VeniceString{v.Case}, &VeniceInteger{len(tree.Args)}))

				if len(genericParameters) > 0 {
					return code, enumType.SubstituteGenerics(genericParameters, concreteTypes), nil
				} else {
					return code, enumType, nil
				}
			}
		}

		return nil, nil, compiler.customError(tree, fmt.Sprintf("enum %s does not have case %s", v.Enum, v.Case))
	}

	return nil, nil, compiler.customError(tree, "function calls for non-symbols not implemented yet")
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
			return nil, nil, compiler.customError(tree.Expr, "list index must be integer")
		}

		code = append(code, NewBytecode("BINARY_LIST_INDEX"))
		return code, exprType.ItemType, nil
	case *VeniceMapType:
		if !areTypesCompatible(exprType.KeyType, indexType) {
			return nil, nil, compiler.customError(tree.Expr, "wrong map key type in index expression")
		}

		code = append(code, NewBytecode("BINARY_MAP_INDEX"))
		return code, exprType.KeyType, nil
	default:
		return nil, nil, compiler.customError(tree, "only maps and lists can be indexed")
	}
}

func (compiler *Compiler) compileInfixNode(tree *InfixNode) ([]*Bytecode, VeniceType, error) {
	leftCode, leftType, err := compiler.compileExpression(tree.Left)
	if err != nil {
		return nil, nil, err
	}

	if !checkInfixLeftType(tree.Operator, leftType) {
		return nil, nil, compiler.customError(tree.Left, fmt.Sprintf("invalid type for left operand of %s", tree.Operator))
	}

	rightCode, rightType, err := compiler.compileExpression(tree.Right)
	if err != nil {
		return nil, nil, err
	}

	if !checkInfixRightType(tree.Operator, leftType, rightType) {
		return nil, nil, compiler.customError(tree.Right, fmt.Sprintf("invalid type for right operand of %s", tree.Operator))
	}

	code := append(leftCode, rightCode...)
	// TODO(2021-08-07): Boolean operators should short-circuit.
	switch tree.Operator {
	case "+":
		return append(code, NewBytecode("BINARY_ADD")), VENICE_TYPE_INTEGER, nil
	case "and":
		return append(code, NewBytecode("BINARY_AND")), VENICE_TYPE_BOOLEAN, nil
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
	case "or":
		return append(code, NewBytecode("BINARY_OR")), VENICE_TYPE_BOOLEAN, nil
	case "-":
		return append(code, NewBytecode("BINARY_SUB")), VENICE_TYPE_INTEGER, nil
	default:
		return nil, nil, compiler.customError(tree, fmt.Sprintf("unknown operator: %s", tree.Operator))
	}
}

func checkInfixLeftType(operator string, leftType VeniceType) bool {
	switch operator {
	case "==":
		return true
	case "and", "or":
		return areTypesCompatible(VENICE_TYPE_BOOLEAN, leftType)
	default:
		return areTypesCompatible(VENICE_TYPE_INTEGER, leftType)
	}
}

func checkInfixRightType(operator string, leftType VeniceType, rightType VeniceType) bool {
	switch operator {
	case "==":
		return areTypesCompatible(leftType, rightType)
	case "and", "or":
		return areTypesCompatible(VENICE_TYPE_BOOLEAN, rightType)
	default:
		return areTypesCompatible(VENICE_TYPE_INTEGER, rightType)
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
	case *VeniceEnumType:
		switch actualType := actualTypeInterface.(type) {
		case *VeniceEnumType:
			return expectedType == actualType
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
	Message  string
	Location *Location
}

func (e *CompileError) Error() string {
	if e.Location != nil {
		return fmt.Sprintf("%s at line %d, column %d", e.Message, e.Location.Line, e.Location.Column)
	} else {
		return e.Message
	}
}

func (compiler *Compiler) customError(node Node, message string) *CompileError {
	return &CompileError{message, node.getLocation()}
}
