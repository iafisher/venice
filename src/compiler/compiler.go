/**
 * The Venice compiler.
 *
 * The compiler compiles a Venice program (represented as an abstract syntax tree, the
 * output of src/compiler/parser.go) into bytecode instructions. It also checks the static
 * types of the programs and reports any errors.
 */
package compiler

import (
	"fmt"
	"github.com/iafisher/venice/src/common/bytecode"
	"github.com/iafisher/venice/src/common/lex"
)

type Compiler struct {
	symbolTable     *SymbolTable
	typeSymbolTable *SymbolTable
	functionInfo    *FunctionInfo
	compiledProgram *bytecode.CompiledProgram
	nestedLoopCount int
}

func NewCompiler() *Compiler {
	return &Compiler{
		symbolTable:     NewBuiltinSymbolTable(),
		typeSymbolTable: NewBuiltinTypeSymbolTable(),
		compiledProgram: bytecode.NewCompiledProgram(),
		functionInfo:    nil,
		nestedLoopCount: 0,
	}
}

func (compiler *Compiler) Compile(file *File) (*bytecode.CompiledProgram, error) {
	compiler.compiledProgram = bytecode.NewCompiledProgram()
	compiledProgram, _, err := compiler.compileModule("", file)
	return compiledProgram, err
}

func (compiler *Compiler) compileModule(
	moduleName string, file *File,
) (*bytecode.CompiledProgram, VeniceType, error) {
	for _, statementAny := range file.Statements {
		switch statement := statementAny.(type) {
		case *ClassDeclarationNode:
			err := compiler.compileClassDeclaration(statement)
			if err != nil {
				return nil, nil, err
			}
		case *EnumDeclarationNode:
			err := compiler.compileEnumDeclaration(statement)
			if err != nil {
				return nil, nil, err
			}
		case *FunctionDeclarationNode:
			err := compiler.compileFunctionDeclaration(statement)
			if err != nil {
				return nil, nil, err
			}
		case *ImportStatementNode:
			importedFile, err := ParseFile(statement.Path)
			if err != nil {
				return nil, nil, err
			}

			subCompiler := NewCompiler()
			importedProgram, moduleType, err := subCompiler.compileModule(
				statement.Name, importedFile,
			)
			if err != nil {
				return nil, nil, err
			}

			for functionName, functionCode := range importedProgram.Code {
				// TODO(2021-08-24): How should I handle this?
				if functionName == "main" {
					continue
				}

				qualifiedName := fmt.Sprintf("%s::%s", statement.Name, functionName)
				compiler.compiledProgram.Code[qualifiedName] = functionCode
			}

			compiler.symbolTable.Put(statement.Name, moduleType)
		default:
			code, err := compiler.compileStatement(statementAny)
			if err != nil {
				return nil, nil, err
			}
			compiler.compiledProgram.Code["main"] = append(compiler.compiledProgram.Code["main"], code...)
		}
	}
	return compiler.compiledProgram, moduleTypeFromSymbolTable(moduleName, compiler.symbolTable), nil
}

func (compiler *Compiler) GetType(expr ExpressionNode) (VeniceType, error) {
	_, exprType, err := compiler.compileExpression(expr)
	return exprType, err
}

/**
 * Compile statements
 */

func (compiler *Compiler) compileStatement(
	treeAny StatementNode,
) ([]bytecode.Bytecode, error) {
	switch node := treeAny.(type) {
	case *AssignStatementNode:
		return compiler.compileAssignStatement(node)
	case *BreakStatementNode:
		return compiler.compileBreakStatement(node)
	case *ContinueStatementNode:
		return compiler.compileContinueStatement(node)
	case *ExpressionStatementNode:
		code, _, err := compiler.compileExpression(node.Expr)
		return code, err
	case *ForLoopNode:
		return compiler.compileForLoop(node)
	case *IfStatementNode:
		return compiler.compileIfStatement(node)
	case *LetStatementNode:
		return compiler.compileLetStatement(node)
	case *MatchStatementNode:
		return compiler.compileMatchStatement(node)
	case *ReturnStatementNode:
		return compiler.compileReturnStatement(node)
	case *WhileLoopNode:
		return compiler.compileWhileLoop(node)
	default:
		return nil, compiler.customError(treeAny, "unknown statement type: %T", treeAny)
	}
}

func (compiler *Compiler) compileAssignStatement(
	node *AssignStatementNode,
) ([]bytecode.Bytecode, error) {
	switch destination := node.Destination.(type) {
	case *FieldAccessNode:
		return compiler.compileAssignStatementToField(node, destination)
	case *IndexNode:
		return compiler.compileAssignStatementToIndex(node, destination)
	case *SymbolNode:
		return compiler.compileAssignStatementToSymbol(node, destination)
	default:
		return nil, compiler.customError(destination, "cannot assign to non-symbol")
	}
}

func (compiler *Compiler) compileBreakStatement(
	node *BreakStatementNode,
) ([]bytecode.Bytecode, error) {
	if compiler.nestedLoopCount == 0 {
		return nil, compiler.customError(node, "break statement outside of loop")
	}

	// Return a placeholder bytecode instruction that the compiler will later convert
	// to a REL_JUMP instruction.
	return []bytecode.Bytecode{&bytecode.Placeholder{"break"}}, nil
}

func (compiler *Compiler) compileContinueStatement(
	node *ContinueStatementNode,
) ([]bytecode.Bytecode, error) {
	if compiler.nestedLoopCount == 0 {
		return nil, compiler.customError(node, "continue statement outside of loop")
	}

	// Return a placeholder bytecode instruction that the compiler will later convert
	// to a REL_JUMP instruction.
	return []bytecode.Bytecode{&bytecode.Placeholder{"continue"}}, nil
}

func (compiler *Compiler) compileAssignStatementToField(
	node *AssignStatementNode,
	destination *FieldAccessNode,
) ([]bytecode.Bytecode, error) {
	code, eType, err := compiler.compileExpression(node.Expr)
	if err != nil {
		return nil, err
	}

	destinationCode, destinationTypeAny, err := compiler.compileExpression(destination.Expr)
	if err != nil {
		return nil, err
	}

	classType, ok := destinationTypeAny.(*VeniceClassType)
	if !ok {
		return nil, compiler.customError(
			node, "cannot assign to field on type %s", destinationTypeAny.String(),
		)
	}

	for i, field := range classType.Fields {
		if field.Name == destination.Name {
			if !field.Public {
				return nil, compiler.customError(node, "cannot assign to non-public field")
			}

			if !compiler.checkType(field.FieldType, eType) {
				return nil, compiler.customError(
					node.Expr,
					"expected type %s, got %s",
					field.FieldType.String(),
					eType.String(),
				)
			}

			code = append(code, destinationCode...)
			code = append(code, &bytecode.StoreField{i})
			return code, nil
		}
	}

	return nil, compiler.customError(
		node, "field `%s` does not exist on %s", destination.Name, classType.String(),
	)
}

func (compiler *Compiler) compileAssignStatementToIndex(
	node *AssignStatementNode,
	destination *IndexNode,
) ([]bytecode.Bytecode, error) {
	code, eType, err := compiler.compileExpression(node.Expr)
	if err != nil {
		return nil, err
	}

	destinationCode, destinationTypeAny, err := compiler.compileExpression(destination.Expr)
	if err != nil {
		return nil, err
	}

	switch destinationType := destinationTypeAny.(type) {
	case *VeniceListType:
		if !compiler.checkType(destinationType.ItemType, eType) {
			return nil, compiler.customError(
				node.Expr,
				"expected type %s, got %s",
				destinationType.ItemType.String(),
				eType.String(),
			)
		}

		indexCode, indexType, err := compiler.compileExpression(destination.Index)
		if err != nil {
			return nil, err
		}

		if !compiler.checkType(VENICE_TYPE_INTEGER, indexType) {
			return nil, compiler.customError(
				node, "list index must be of type integer, not %s", indexType.String(),
			)
		}

		code = append(code, indexCode...)
		code = append(code, destinationCode...)
		code = append(code, &bytecode.StoreIndex{})
		return code, nil
	case *VeniceMapType:
		if !compiler.checkType(destinationType.ValueType, eType) {
			return nil, compiler.customError(
				node.Expr,
				"expected type %s, got %s",
				destinationType.ValueType.String(),
				eType.String(),
			)
		}

		indexCode, indexType, err := compiler.compileExpression(destination.Index)
		if err != nil {
			return nil, err
		}

		if !compiler.checkType(destinationType.KeyType, indexType) {
			return nil, compiler.customError(
				node, "map index must be of type integer, not %s", indexType.String(),
			)
		}

		code = append(code, indexCode...)
		code = append(code, destinationCode...)
		code = append(code, &bytecode.StoreMapIndex{})
		return code, nil
	default:
		return nil, compiler.customError(
			node,
			"cannot assign to index on type %s",
			destinationTypeAny.String(),
		)
	}
}

func (compiler *Compiler) compileAssignStatementToSymbol(
	node *AssignStatementNode,
	destination *SymbolNode,
) ([]bytecode.Bytecode, error) {
	code, eType, err := compiler.compileExpression(node.Expr)
	if err != nil {
		return nil, err
	}

	if binding, ok := compiler.symbolTable.GetBinding(destination.Value); ok {
		if !binding.IsVar {
			return nil, compiler.customError(
				destination, "cannot assign to const symbol `%q`", destination.Value,
			)
		}
	}

	expectedType, ok := compiler.symbolTable.Get(destination.Value)
	if !ok {
		return nil, compiler.customError(
			node, "cannot assign to undeclared symbol `%s`", destination.Value,
		)
	} else if !compiler.checkType(expectedType, eType) {
		return nil, compiler.customError(
			node,
			"cannot assign %s to symbol of type %s",
			eType.String(),
			expectedType.String(),
		)
	}

	code = append(code, &bytecode.StoreName{destination.Value})
	return code, nil
}

func (compiler *Compiler) compileClassDeclaration(node *ClassDeclarationNode) error {
	fields := make([]*VeniceClassField, 0, len(node.Fields))
	paramTypes := make([]VeniceType, 0, len(node.Fields))
	for _, field := range node.Fields {
		paramType, err := compiler.resolveType(field.FieldType)
		if err != nil {
			return err
		}
		paramTypes = append(paramTypes, paramType)
		fields = append(
			fields,
			&VeniceClassField{
				field.Name, field.Public, paramType,
			},
		)
	}

	classType := &VeniceClassType{
		Name:              node.Name,
		GenericParameters: []string{},
		Fields:            fields,
	}
	compiler.typeSymbolTable.Put(node.Name, classType)
	return nil
}

func (compiler *Compiler) compileEnumDeclaration(node *EnumDeclarationNode) error {
	// Put in a dummy entry for the enum in the type symbol table so that recursive enum
	// types will type-check properly.
	enumType := &VeniceEnumType{
		Name:              node.Name,
		GenericParameters: nil,
		Cases:             nil,
	}
	compiler.typeSymbolTable.Put(node.Name, enumType)

	if node.GenericTypeParameter != "" {
		compiler.typeSymbolTable = compiler.typeSymbolTable.SpawnChild()
		compiler.typeSymbolTable.Put(
			node.GenericTypeParameter,
			&VeniceSymbolType{node.GenericTypeParameter},
		)
	}

	caseTypes := make([]*VeniceCaseType, 0, len(node.Cases))
	for _, caseNode := range node.Cases {
		types := make([]VeniceType, 0, len(caseNode.Types))
		for _, typeNode := range caseNode.Types {
			veniceType, err := compiler.resolveType(typeNode)
			if err != nil {
				return err
			}
			types = append(types, veniceType)
		}
		caseTypes = append(caseTypes, &VeniceCaseType{caseNode.Label, types})

		functionName := fmt.Sprintf("%s__%s", node.Name, caseNode.Label)
		compiler.compiledProgram.Code[functionName] = []bytecode.Bytecode{
			&bytecode.PushEnum{caseNode.Label, len(caseNode.Types)},
		}
	}

	if node.GenericTypeParameter != "" {
		compiler.typeSymbolTable = compiler.typeSymbolTable.Parent
		enumType.GenericParameters = []string{node.GenericTypeParameter}
	}
	enumType.Cases = caseTypes

	return nil
}

func (compiler *Compiler) compileForLoop(
	node *ForLoopNode,
) ([]bytecode.Bytecode, error) {
	iterableCode, iterableTypeAny, err := compiler.compileExpression(node.Iterable)
	if err != nil {
		return nil, err
	}

	loopSymbolTable := compiler.symbolTable.SpawnChild()
	switch iterableType := iterableTypeAny.(type) {
	case *VeniceListType:
		if len(node.Variables) != 1 {
			return nil, compiler.customError(node, "too many for loop variables")
		}

		loopSymbolTable.Put(node.Variables[0], iterableType.ItemType)
	case *VeniceMapType:
		if len(node.Variables) != 2 {
			return nil, compiler.customError(node, "too many for loop variables")
		}

		loopSymbolTable.Put(node.Variables[0], iterableType.KeyType)
		loopSymbolTable.Put(node.Variables[1], iterableType.ValueType)
	case *VeniceStringType:
		if len(node.Variables) != 1 {
			return nil, compiler.customError(node, "too many for loop variables")
		}

		loopSymbolTable.Put(node.Variables[0], VENICE_TYPE_STRING)
	default:
		return nil, compiler.customError(node.Iterable, "for loop must be of list, map, or string")
	}

	compiler.symbolTable = loopSymbolTable
	compiler.nestedLoopCount += 1
	bodyCode, err := compiler.compileBlock(node.Body)
	compiler.nestedLoopCount -= 1
	compiler.symbolTable = compiler.symbolTable.Parent

	if err != nil {
		return nil, err
	}

	jump := len(node.Variables) + len(bodyCode) + 1
	if len(node.Variables) > 1 {
		// Account for the extra UNPACK_TUPLE instruction.
		jump++
	}

	code := iterableCode
	code = append(code, &bytecode.GetIter{})
	code = append(code, &bytecode.ForIter{jump + 1})

	if len(node.Variables) > 1 {
		code = append(code, &bytecode.UnpackTuple{})
	}
	for _, variable := range node.Variables {
		code = append(code, &bytecode.StoreName{variable})
	}
	code = append(code, bodyCode...)
	code = append(code, &bytecode.RelJump{-jump})
	return code, nil
}

func (compiler *Compiler) compileFunctionDeclaration(node *FunctionDeclarationNode) error {
	params := make([]string, 0, len(node.Params))
	paramTypes := make([]VeniceType, 0, len(node.Params))
	bodySymbolTable := compiler.symbolTable.SpawnChild()
	for _, param := range node.Params {
		paramType, err := compiler.resolveType(param.ParamType)
		if err != nil {
			return err
		}

		params = append(params, param.Name)
		paramTypes = append(paramTypes, paramType)

		bodySymbolTable.Put(param.Name, paramType)
	}

	declaredReturnType, err := compiler.resolveType(node.ReturnType)
	if err != nil {
		return err
	}

	// Put the function's entry in the symbol table before compiling the body so that
	// recursive functions can call themselves.
	compiler.symbolTable.Put(
		node.Name,
		&VeniceFunctionType{
			Name:       node.Name,
			ParamTypes: paramTypes,
			ReturnType: declaredReturnType,
			IsBuiltin:  false,
		},
	)

	compiler.symbolTable = bodySymbolTable
	compiler.functionInfo = &FunctionInfo{
		declaredReturnType:  declaredReturnType,
		seenReturnStatement: false,
	}
	bodyCode, err := compiler.compileBlock(node.Body)
	if err != nil {
		return err
	}

	if declaredReturnType != nil && !compiler.functionInfo.seenReturnStatement {
		return compiler.customError(node, "non-void function has no return statement")
	}

	compiler.functionInfo = nil
	compiler.symbolTable = bodySymbolTable.Parent

	paramLoadCode := make([]bytecode.Bytecode, 0, len(node.Params))
	for i := len(node.Params) - 1; i >= 0; i-- {
		param := node.Params[i]
		paramLoadCode = append(paramLoadCode, &bytecode.StoreName{param.Name})
	}
	bodyCode = append(paramLoadCode, bodyCode...)
	compiler.compiledProgram.Code[node.Name] = bodyCode
	return nil
}

func (compiler *Compiler) compileIfStatement(
	node *IfStatementNode,
) ([]bytecode.Bytecode, error) {
	if len(node.Clauses) == 0 {
		return nil, compiler.customError(node, "`if` statement with no clauses")
	}

	code := []bytecode.Bytecode{}
	for _, clause := range node.Clauses {
		conditionCode, conditionType, err := compiler.compileExpression(clause.Condition)
		if err != nil {
			return nil, err
		}

		if !compiler.checkType(VENICE_TYPE_BOOLEAN, conditionType) {
			return nil, compiler.customError(
				clause.Condition, "condition of `if` statement must be a boolean",
			)
		}

		bodyCode, err := compiler.compileBlock(clause.Body)
		if err != nil {
			return nil, err
		}

		code = append(code, conditionCode...)
		code = append(code, &bytecode.RelJumpIfFalse{len(bodyCode) + 2})
		code = append(code, bodyCode...)
		// To be replaced with REL_JUMP instructions once the total length of the code is
		// known.
		code = append(code, &bytecode.Placeholder{"if"})
	}

	if node.ElseClause != nil {
		elseClauseCode, err := compiler.compileBlock(node.ElseClause)
		if err != nil {
			return nil, err
		}

		code = append(code, elseClauseCode...)
	}

	// Fill in the relative jump values now that we have generated all the code.
	for i, bcode := range code {
		if placeholder, ok := bcode.(*bytecode.Placeholder); ok && placeholder.Name == "if" {
			code[i] = &bytecode.RelJump{len(code) - i}
		}
	}

	return code, nil
}

func (compiler *Compiler) compileLetStatement(
	node *LetStatementNode,
) ([]bytecode.Bytecode, error) {
	if _, ok := compiler.symbolTable.Get(node.Symbol); ok {
		return nil, compiler.customError(node, "re-declaration of symbol `%q`", node.Symbol)
	}

	var declaredType VeniceType
	if node.Type != nil {
		var err error
		declaredType, err = compiler.resolveType(node.Type)
		if err != nil {
			return nil, err
		}
	}

	code, eType, err := compiler.compileExpressionWithTypeHint(node.Expr, declaredType)
	if err != nil {
		return nil, err
	}

	if declaredType != nil && !compiler.checkType(declaredType, eType) {
		return nil, compiler.customError(
			node.Expr, "expected %s, got %s", declaredType.String(), eType.String(),
		)
	}

	if node.IsVar {
		compiler.symbolTable.PutVar(node.Symbol, eType)
	} else {
		compiler.symbolTable.Put(node.Symbol, eType)
	}
	return append(code, &bytecode.StoreName{node.Symbol}), nil
}

func (compiler *Compiler) compileMatchStatement(
	node *MatchStatementNode,
) ([]bytecode.Bytecode, error) {
	exprCode, exprType, err := compiler.compileExpression(node.Expr)
	if err != nil {
		return nil, err
	}

	enumType, ok := exprType.(*VeniceEnumType)
	if !ok {
		return nil, compiler.customError(
			node.Expr,
			"cannot match a non-enum type (%s)",
			exprType.String(),
		)
	}

	// TODO(2021-08-22): Ensure matches are exhaustive.
	// TODO(2021-08-22): This code leaves the value on the top of the stack.
	code := exprCode
	for _, clause := range node.Clauses {
		compiler.symbolTable = compiler.symbolTable.SpawnChild()
		patternCode, err := compiler.compilePattern(clause.Pattern, enumType)
		if err != nil {
			return nil, err
		}

		bodyCode, err := compiler.compileBlock(clause.Body)
		if err != nil {
			return nil, err
		}
		compiler.symbolTable = compiler.symbolTable.Parent

		code = append(code, patternCode...)
		code = append(code, &bytecode.RelJumpIfFalseOrPop{len(bodyCode) + 2})
		code = append(code, bodyCode...)
		// To be replaced with REL_JUMP instructions once the total length of the code is
		// known.
		code = append(code, &bytecode.Placeholder{"match"})
	}

	// TODO(2021-08-22): Default clause.

	for i, bcode := range code {
		if placeholder, ok := bcode.(*bytecode.Placeholder); ok && placeholder.Name == "match" {
			code[i] = &bytecode.RelJump{len(code) - i}
		}
	}

	return code, nil
}

func (compiler *Compiler) compilePattern(
	patternAny PatternNode, exprType VeniceType,
) ([]bytecode.Bytecode, error) {
	switch pattern := patternAny.(type) {
	case *CompoundPatternNode:
		enumType, ok := exprType.(*VeniceEnumType)
		if !ok {
			// TODO(2021-08-22): Better error message.
			return nil, compiler.customError(patternAny, "does not match enum type")
		}

		index := -1
		for i, caseType := range enumType.Cases {
			if caseType.Label == pattern.Label {
				index = i
				break
			}
		}

		if index == -1 {
			return nil, compiler.customError(
				patternAny, "enum case `%s` not found", pattern.Label,
			)
		}

		caseType := enumType.Cases[index]

		if len(pattern.Patterns) > len(caseType.Types) {
			return nil, compiler.customError(
				patternAny, "too many patterns for %s", enumType.Name,
			)
		}

		if !pattern.Elided && len(pattern.Patterns) < len(caseType.Types) {
			return nil, compiler.customError(
				patternAny, "too few patterns for %s", enumType.Name,
			)
		}

		code := []bytecode.Bytecode{
			&bytecode.CheckLabel{pattern.Label},
			&bytecode.Placeholder{"pattern"},
		}
		for i, subPattern := range pattern.Patterns {
			code = append(code, &bytecode.PushEnumIndex{i})
			subPatternCode, err := compiler.compilePattern(subPattern, caseType.Types[i])
			if err != nil {
				return nil, err
			}

			code = append(code, subPatternCode...)
		}

		for i, bcode := range code {
			if placeholder, ok := bcode.(*bytecode.Placeholder); ok &&
				placeholder.Name == "pattern" {
				code[i] = &bytecode.RelJumpIfFalseOrPop{len(code) - i}
			}
		}

		return code, nil
	case *SymbolNode:
		compiler.symbolTable.Put(pattern.Value, exprType)
		code := []bytecode.Bytecode{
			&bytecode.StoreName{pattern.Value},
			&bytecode.PushConstBool{true},
		}
		return code, nil
	default:
		return nil, compiler.customError(patternAny, "unknown pattern type %T", patternAny)
	}
}

func (compiler *Compiler) compileReturnStatement(
	node *ReturnStatementNode,
) ([]bytecode.Bytecode, error) {
	if compiler.functionInfo == nil {
		return nil, compiler.customError(
			node, "return statement outside of function definition",
		)
	}

	if node.Expr != nil {
		code, exprType, err := compiler.compileExpression(node.Expr)
		if err != nil {
			return nil, err
		}

		if !compiler.checkType(compiler.functionInfo.declaredReturnType, exprType) {
			return nil, compiler.customError(
				node,
				"conflicting function return types: got %s, expected %s",
				exprType.String(),
				compiler.functionInfo.declaredReturnType.String(),
			)
		}

		compiler.functionInfo.seenReturnStatement = true
		code = append(code, &bytecode.Return{})
		return code, err
	} else {
		if compiler.functionInfo.declaredReturnType != nil {
			return nil, compiler.customError(node, "function cannot return void")
		}

		return []bytecode.Bytecode{&bytecode.Return{}}, nil
	}
}

func (compiler *Compiler) compileWhileLoop(
	node *WhileLoopNode,
) ([]bytecode.Bytecode, error) {
	conditionCode, conditionType, err := compiler.compileExpression(node.Condition)
	if err != nil {
		return nil, err
	}

	if !compiler.checkType(VENICE_TYPE_BOOLEAN, conditionType) {
		return nil, compiler.customError(
			node.Condition, "condition of `while` loop must be a boolean",
		)
	}

	compiler.nestedLoopCount += 1
	bodyCode, err := compiler.compileBlock(node.Body)
	compiler.nestedLoopCount -= 1
	if err != nil {
		return nil, err
	}

	code := conditionCode
	jumpForward := len(bodyCode) + 2
	code = append(code, &bytecode.RelJumpIfFalse{jumpForward})
	code = append(code, bodyCode...)
	jumpBack := -(len(conditionCode) + len(bodyCode) + 1)
	code = append(code, &bytecode.RelJump{jumpBack})

	for i, bcode := range code {
		if placeholder, ok := bcode.(*bytecode.Placeholder); ok {
			if placeholder.Name == "break" {
				code[i] = &bytecode.RelJump{len(code) - i}
			} else if placeholder.Name == "continue" {
				code[i] = &bytecode.RelJump{-i}
			}
		}
	}

	return code, nil
}

func (compiler *Compiler) compileBlock(
	block []StatementNode,
) ([]bytecode.Bytecode, error) {
	compiler.symbolTable = compiler.symbolTable.SpawnChild()
	defer func() { compiler.symbolTable = compiler.symbolTable.Parent }()

	code := []bytecode.Bytecode{}
	for _, statement := range block {
		statementCode, err := compiler.compileStatement(statement)
		if err != nil {
			return nil, err
		}

		code = append(code, statementCode...)
	}
	return code, nil
}

/**
 * Compile expressions
 */

func (compiler *Compiler) compileExpression(
	nodeAny ExpressionNode,
) ([]bytecode.Bytecode, VeniceType, error) {
	return compiler.compileExpressionWithTypeHint(nodeAny, nil)
}

func (compiler *Compiler) compileExpressionWithTypeHint(
	nodeAny ExpressionNode, typeHint VeniceType,
) ([]bytecode.Bytecode, VeniceType, error) {
	switch node := nodeAny.(type) {
	case *BooleanNode:
		return []bytecode.Bytecode{
			&bytecode.PushConstBool{node.Value},
		}, VENICE_TYPE_BOOLEAN, nil
	case *CallNode:
		return compiler.compileCallNode(node)
	case *ConstructorNode:
		return compiler.compileConstructorNode(node)
	case *FieldAccessNode:
		return compiler.compileFieldAccessNode(node)
	case *IndexNode:
		return compiler.compileIndexNode(node)
	case *InfixNode:
		return compiler.compileInfixNode(node)
	case *IntegerNode:
		return []bytecode.Bytecode{
			&bytecode.PushConstInt{node.Value},
		}, VENICE_TYPE_INTEGER, nil
	case *ListNode:
		return compiler.compileListNode(node, typeHint)
	case *MapNode:
		return compiler.compileMapNode(node, typeHint)
	case *QualifiedSymbolNode:
		return compiler.compileQualifiedSymbolNode(node)
	case *RealNumberNode:
		return []bytecode.Bytecode{
			&bytecode.PushConstRealNumber{node.Value},
		}, VENICE_TYPE_REAL_NUMBER, nil
	case *StringNode:
		return []bytecode.Bytecode{
			&bytecode.PushConstStr{node.Value},
		}, VENICE_TYPE_STRING, nil
	case *SymbolNode:
		symbolType, ok := compiler.symbolTable.Get(node.Value)
		if !ok {
			return nil, nil, compiler.customError(
				nodeAny, "undefined symbol `%s`", node.Value,
			)
		}
		// TODO(2021-08-26): Do we need to handle function types separately?
		if functionType, isFunctionType := symbolType.(*VeniceFunctionType); isFunctionType {
			return []bytecode.Bytecode{
					&bytecode.PushConstFunction{
						Name:      functionType.Name,
						IsBuiltin: functionType.IsBuiltin,
					},
				},
				symbolType,
				nil
		} else {
			return []bytecode.Bytecode{&bytecode.PushName{node.Value}}, symbolType, nil
		}
		return []bytecode.Bytecode{&bytecode.PushName{node.Value}}, symbolType, nil
	case *TernaryIfNode:
		return compiler.compileTernaryIfNode(node)
	case *TupleFieldAccessNode:
		return compiler.compileTupleFieldAccessNode(node)
	case *TupleNode:
		return compiler.compileTupleNode(node)
	case *UnaryNode:
		return compiler.compileUnaryNode(node)
	default:
		return nil, nil, compiler.customError(
			nodeAny, "unknown expression type: %T", nodeAny,
		)
	}
}

func (compiler *Compiler) compileCallNode(
	node *CallNode,
) ([]bytecode.Bytecode, VeniceType, error) {
	functionCode, functionTypeAny, err := compiler.compileExpression(node.Function)
	if err != nil {
		return nil, nil, err
	}

	functionType, ok := functionTypeAny.(*VeniceFunctionType)
	if !ok {
		return nil, nil, compiler.customError(
			node.Function,
			"type %s cannot be used as a function",
			functionTypeAny.String(),
		)
	}

	_, isClassMethod := node.Function.(*FieldAccessNode)
	compiler.typeSymbolTable = compiler.typeSymbolTable.SpawnChild()
	code, returnType, err := compiler.compileFunctionArguments(
		node,
		node.Args,
		functionType,
		isClassMethod,
	)
	compiler.typeSymbolTable = compiler.typeSymbolTable.Parent
	if err != nil {
		return nil, nil, err
	}

	code = append(code, functionCode...)
	code = append(code, &bytecode.CallFunction{len(functionType.ParamTypes)})
	return code, returnType, nil
}

func (compiler *Compiler) compileQualifiedSymbolNode(
	node *QualifiedSymbolNode,
) ([]bytecode.Bytecode, VeniceType, error) {
	enumTypeAny, err := compiler.resolveType(&SymbolNode{node.Enum, nil})
	if err != nil {
		otherType, ok := compiler.symbolTable.Get(node.Enum)
		if ok {
			moduleType, ok := otherType.(*VeniceModuleType)
			if !ok {
				return nil, nil, compiler.customError(node, "%s is not a module", node.Enum)
			}

			memberType, ok := moduleType.Types[node.Case]
			if !ok {
				return nil, nil, compiler.customError(node, "module %s has no member %s", node.Enum, node.Case)
			}

			memberFunctionType, ok := memberType.(*VeniceFunctionType)
			if !ok {
				return nil, nil, compiler.customError(
					node, "cannot only access functions from modules, not %s", memberType.String(),
				)
			}

			functionName := fmt.Sprintf("%s::%s", node.Enum, node.Case)
			code := []bytecode.Bytecode{
				&bytecode.PushConstFunction{Name: functionName, IsBuiltin: memberFunctionType.IsBuiltin},
			}
			return code, memberFunctionType, nil
		} else {
			return nil, nil, err
		}
	}

	enumType, ok := enumTypeAny.(*VeniceEnumType)
	if !ok {
		return nil, nil, compiler.customError(
			node, "cannot use double colon after non-enum type",
		)
	}

	for _, enumCase := range enumType.Cases {
		if enumCase.Label == node.Case {
			if len(enumCase.Types) > 0 {
				functionName := fmt.Sprintf("%s__%s", enumType.Name, enumCase.Label)
				code := []bytecode.Bytecode{&bytecode.PushConstFunction{functionName, false}}
				return code, enumCase.AsFunctionType(enumType), nil
			} else {
				return []bytecode.Bytecode{&bytecode.PushEnum{node.Case, 0}}, enumType, nil
			}
		}
	}

	return nil, nil, compiler.customError(
		node, "enum %s does not have case %s", node.Enum, node.Case,
	)
}

func (compiler *Compiler) compileConstructorNode(
	node *ConstructorNode,
) ([]bytecode.Bytecode, VeniceType, error) {
	classTypeAny, ok := compiler.typeSymbolTable.Get(node.Name)
	if !ok {
		return nil, nil, compiler.customError(node, "no such type `%s`", node.Name)
	}

	classType, ok := classTypeAny.(*VeniceClassType)
	if !ok {
		return nil, nil, compiler.customError(node, "`%s` is not a class type", node.Name)
	}

	if len(classType.Fields) != len(node.Fields) {
		return nil, nil, compiler.customError(node, "too many fields in class constructor")
	}

	code := []bytecode.Bytecode{}
	for i := len(node.Fields) - 1; i >= 0; i-- {
		field := node.Fields[i]
		exprCode, exprType, err := compiler.compileExpression(field.Value)
		if err != nil {
			return nil, nil, err
		}

		found := false
		for _, fieldType := range classType.Fields {
			if fieldType.Name == field.Name {
				if !compiler.checkType(fieldType.FieldType, exprType) {
					return nil, nil, compiler.customError(
						node,
						"expected type %s, got %s for class field `%s`",
						fieldType.FieldType.String(),
						exprType.String(),
						fieldType.Name,
					)
				}

				found = true
				break
			}
		}

		if !found {
			return nil, nil, compiler.customError(
				node,
				"no such field `%s` on class `%s`",
				field.Name,
				node.Name,
			)
		}

		code = append(code, exprCode...)
	}

	code = append(code, &bytecode.BuildClass{node.Name, len(node.Fields)})
	return code, classType, nil
}

func (compiler *Compiler) compileFieldAccessNode(
	node *FieldAccessNode,
) ([]bytecode.Bytecode, VeniceType, error) {
	code, typeAny, err := compiler.compileExpression(node.Expr)
	if err != nil {
		return nil, nil, err
	}

	switch concreteType := typeAny.(type) {
	case *VeniceClassType:
		for i, field := range concreteType.Fields {
			if field.Name == node.Name {
				// TODO(2021-08-09): Allow this when inside the class itself.
				if !field.Public {
					return nil, nil, compiler.customError(node, "use of private field")
				}

				code = append(code, &bytecode.PushField{i})
				return code, field.FieldType, nil
			}
		}

		for _, methodType := range concreteType.Methods {
			// TODO(2021-08-17): Allow this when inside the class itself.
			if methodType.Name == node.Name {
				if !methodType.Public {
					return nil, nil, compiler.customError(node, "use of private method")
				}
				code = append(code, &bytecode.LookupMethod{methodType.Name})
				return code, methodType, nil
			}
		}

		return nil, nil, compiler.customError(
			node, "no such field or method `%s`", node.Name,
		)
	case *VeniceListType:
		methodType, ok := listBuiltins[node.Name]
		if !ok {
			return nil, nil, compiler.customError(
				node, "no such field or method `%s` on list type", node.Name,
			)
		}

		genericParametersMap := map[string]VeniceType{"T": concreteType.ItemType}
		methodType, err = compiler.substituteGenerics(genericParametersMap, methodType)
		if err != nil {
			return nil, nil, err
		}

		code = append(code, &bytecode.LookupMethod{node.Name})
		return code, methodType, nil
	case *VeniceMapType:
		methodType, ok := mapBuiltins[node.Name]
		if !ok {
			return nil, nil, compiler.customError(
				node, "no such field or method `%s` on map type", node.Name,
			)
		}

		genericParametersMap := map[string]VeniceType{
			"K": concreteType.KeyType,
			"V": concreteType.ValueType,
		}
		methodType, err = compiler.substituteGenerics(genericParametersMap, methodType)
		if err != nil {
			return nil, nil, err
		}

		code = append(code, &bytecode.LookupMethod{node.Name})
		return code, methodType, nil
	case *VeniceStringType:
		methodType, ok := stringBuiltins[node.Name]
		if !ok {
			return nil, nil, compiler.customError(
				node, "no such field or method `%s` on string type", node.Name,
			)
		}
		code = append(code, &bytecode.LookupMethod{node.Name})
		return code, methodType, nil
	default:
		return nil, nil, compiler.customError(
			node, "no such field or method `%s` on type %s", node.Name, typeAny.String(),
		)
	}
}

func (compiler *Compiler) compileFunctionArguments(
	node Node,
	args []ExpressionNode,
	functionType *VeniceFunctionType,
	isClassMethod bool,
) ([]bytecode.Bytecode, VeniceType, error) {
	actualNumberOfArgs := len(args)
	if isClassMethod {
		actualNumberOfArgs++
	}

	if len(functionType.ParamTypes) != actualNumberOfArgs {
		return nil, nil, compiler.customError(
			node,
			"wrong number of arguments: expected %d, got %d",
			len(functionType.ParamTypes),
			actualNumberOfArgs,
		)
	}

	genericParameterMap := map[string]VeniceType{}
	for _, genericParameter := range functionType.GenericParameters {
		genericParameterMap[genericParameter] = nil
	}

	code := []bytecode.Bytecode{}
	for i := len(args) - 1; i >= 0; i-- {
		argCode, argType, err := compiler.compileExpression(args[i])
		if err != nil {
			return nil, nil, err
		}

		var paramType VeniceType
		if isClassMethod {
			paramType = functionType.ParamTypes[i+1]
		} else {
			paramType = functionType.ParamTypes[i]
		}

		ok := compiler.checkTypeAndSubstituteGenerics(
			paramType,
			argType,
		)
		if !ok {
			return nil, nil, compiler.customError(
				node,
				"wrong function parameter type",
			)
		}

		code = append(code, argCode...)
	}

	genericParametersMap := map[string]VeniceType{}
	for key, value := range compiler.typeSymbolTable.Symbols {
		genericParametersMap[key] = value.Type
	}

	returnType, err := compiler.substituteGenerics(
		genericParametersMap,
		functionType.ReturnType,
	)
	if err != nil {
		return nil, nil, err
	}

	return code, returnType, nil
}

func (compiler *Compiler) compileIndexNode(
	node *IndexNode,
) ([]bytecode.Bytecode, VeniceType, error) {
	exprCode, exprTypeAny, err := compiler.compileExpression(node.Expr)
	if err != nil {
		return nil, nil, err
	}

	indexCode, indexType, err := compiler.compileExpression(node.Index)
	if err != nil {
		return nil, nil, err
	}

	code := append(exprCode, indexCode...)

	switch exprType := exprTypeAny.(type) {
	case *VeniceListType:
		if !compiler.checkType(VENICE_TYPE_INTEGER, indexType) {
			return nil, nil, compiler.customError(node.Expr, "list index must be integer")
		}

		code = append(code, &bytecode.BinaryListIndex{})
		return code, exprType.ItemType, nil
	case *VeniceMapType:
		if !compiler.checkType(exprType.KeyType, indexType) {
			return nil, nil, compiler.customError(
				node.Expr, "wrong map key type in index expression",
			)
		}

		code = append(code, &bytecode.BinaryMapIndex{})
		return code, VeniceOptionalTypeOf(exprType.ValueType), nil
	case *VeniceStringType:
		if !compiler.checkType(VENICE_TYPE_INTEGER, indexType) {
			return nil, nil, compiler.customError(node.Expr, "string index must be integer")
		}

		code = append(code, &bytecode.BinaryStringIndex{})
		return code, VENICE_TYPE_STRING, nil
	default:
		return nil, nil, compiler.customError(node, "%s cannot be indexed", exprTypeAny)
	}
}

func (compiler *Compiler) compileInfixNode(
	node *InfixNode,
) ([]bytecode.Bytecode, VeniceType, error) {
	// TODO(2021-08-07): Boolean operators should short-circuit.
	_, ok := opsToBytecodes[node.Operator]
	if !ok {
		return nil, nil, compiler.customError(node, "unknown operator `%s`", node.Operator)
	}

	leftCode, leftType, err := compiler.compileExpression(node.Left)
	if err != nil {
		return nil, nil, err
	}

	if !compiler.checkInfixLeftType(node.Operator, leftType) {
		return nil, nil, compiler.customError(
			node.Left, "invalid type for left operand of %s", node.Operator,
		)
	}

	rightCode, rightType, err := compiler.compileExpression(node.Right)
	if err != nil {
		return nil, nil, err
	}

	resultType, ok := compiler.checkInfixRightType(node.Operator, leftType, rightType)
	if !ok {
		return nil, nil, compiler.customError(
			node.Right, "invalid type for right operand of %s", node.Operator,
		)
	}

	var code []bytecode.Bytecode
	if node.Operator == "and" {
		code = leftCode
		code = append(code, &bytecode.RelJumpIfFalseOrPop{len(rightCode) + 1})
		code = append(code, rightCode...)
	} else if node.Operator == "or" {
		code = leftCode
		code = append(code, &bytecode.RelJumpIfTrueOrPop{len(rightCode) + 1})
		code = append(code, rightCode...)
	} else {
		code = append(leftCode, rightCode...)

		var opBytecode bytecode.Bytecode
		if _, ok := resultType.(*VeniceRealNumberType); ok {
			opBytecode = opsToBytecodesReal[node.Operator]
		} else {
			opBytecode = opsToBytecodes[node.Operator]
		}
		code = append(code, opBytecode)
	}

	return code, resultType, nil
}

func (compiler *Compiler) compileListNode(
	node *ListNode, typeHint VeniceType,
) ([]bytecode.Bytecode, VeniceType, error) {
	code := []bytecode.Bytecode{}
	var itemType VeniceType
	for i := len(node.Values) - 1; i >= 0; i-- {
		value := node.Values[i]
		valueCode, valueType, err := compiler.compileExpression(value)
		if err != nil {
			return nil, nil, err
		}

		if itemType == nil {
			itemType = valueType
		} else if !compiler.checkType(itemType, valueType) {
			return nil, nil, compiler.customError(
				value, "list elements must all be of same type",
			)
		}

		code = append(code, valueCode...)
	}

	if len(node.Values) == 0 {
		listTypeHint, ok := typeHint.(*VeniceListType)
		if ok {
			itemType = listTypeHint.ItemType
		} else {
			return nil, nil, compiler.customError(node, "empty list has unknown type")
		}
	}

	code = append(code, &bytecode.BuildList{len(node.Values)})
	return code, &VeniceListType{itemType}, nil
}

func (compiler *Compiler) compileMapNode(
	node *MapNode, typeHint VeniceType,
) ([]bytecode.Bytecode, VeniceType, error) {
	code := []bytecode.Bytecode{}
	var keyType VeniceType
	var valueType VeniceType
	for i := len(node.Pairs) - 1; i >= 0; i-- {
		pair := node.Pairs[i]
		keyCode, thisKeyType, err := compiler.compileExpression(pair.Key)
		if err != nil {
			return nil, nil, err
		}

		if keyType == nil {
			keyType = thisKeyType
		} else if !compiler.checkType(keyType, thisKeyType) {
			return nil, nil, compiler.customError(
				pair.Key, "map keys must all be of the same type",
			)
		}

		code = append(code, keyCode...)

		valueCode, thisValueType, err := compiler.compileExpression(pair.Value)
		if err != nil {
			return nil, nil, err
		}

		if valueType == nil {
			valueType = thisValueType
		} else if !compiler.checkType(valueType, thisValueType) {
			return nil, nil, compiler.customError(
				pair.Value, "map values must all be of the same type",
			)
		}

		code = append(code, valueCode...)
	}

	if len(node.Pairs) == 0 {
		mapTypeHint, ok := typeHint.(*VeniceMapType)
		if ok {
			keyType = mapTypeHint.KeyType
			valueType = mapTypeHint.ValueType
		} else {
			return nil, nil, compiler.customError(node, "empty map has unknown type")
		}
	}

	code = append(code, &bytecode.BuildMap{len(node.Pairs)})
	return code, &VeniceMapType{keyType, valueType}, nil
}

func (compiler *Compiler) compileTernaryIfNode(
	node *TernaryIfNode,
) ([]bytecode.Bytecode, VeniceType, error) {
	conditionCode, conditionType, err := compiler.compileExpression(node.Condition)
	if err != nil {
		return nil, nil, err
	}

	if !compiler.checkType(VENICE_TYPE_BOOLEAN, conditionType) {
		return nil, nil, compiler.customError(
			node, "condition of `if` expression must be a boolean",
		)
	}

	trueClauseCode, trueClauseType, err := compiler.compileExpression(node.TrueClause)
	if err != nil {
		return nil, nil, err
	}

	falseClauseCode, falseClauseType, err := compiler.compileExpression(node.FalseClause)
	if err != nil {
		return nil, nil, err
	}

	if !compiler.checkType(trueClauseType, falseClauseType) {
		return nil, nil, compiler.customError(
			node, "branches of `if` expression are of different types",
		)
	}

	code := conditionCode
	code = append(code, &bytecode.RelJumpIfFalse{len(trueClauseCode) + 2})
	code = append(code, trueClauseCode...)
	code = append(code, &bytecode.RelJump{len(falseClauseCode) + 1})
	code = append(code, falseClauseCode...)
	return code, trueClauseType, nil
}

func (compiler *Compiler) compileTupleFieldAccessNode(
	node *TupleFieldAccessNode,
) ([]bytecode.Bytecode, VeniceType, error) {
	code, typeAny, err := compiler.compileExpression(node.Expr)
	if err != nil {
		return nil, nil, err
	}

	tupleType, ok := typeAny.(*VeniceTupleType)
	if !ok {
		return nil, nil, compiler.customError(
			node, "left-hand side of dot must be a tuple object",
		)
	}

	if node.Index < 0 || node.Index >= len(tupleType.ItemTypes) {
		return nil, nil, compiler.customError(node, "tuple index out of bounds")
	}

	code = append(code, &bytecode.PushTupleField{node.Index})
	return code, tupleType.ItemTypes[node.Index], nil
}

func (compiler *Compiler) compileTupleNode(
	node *TupleNode,
) ([]bytecode.Bytecode, VeniceType, error) {
	code := []bytecode.Bytecode{}
	itemTypes := make([]VeniceType, 0, len(node.Values))
	for i := len(node.Values) - 1; i >= 0; i-- {
		value := node.Values[i]
		valueCode, valueType, err := compiler.compileExpression(value)
		if err != nil {
			return nil, nil, err
		}

		itemTypes = append(itemTypes, valueType)
		code = append(code, valueCode...)
	}

	for i, j := 0, len(itemTypes)-1; i < j; i, j = i+1, j-1 {
		itemTypes[i], itemTypes[j] = itemTypes[j], itemTypes[i]
	}

	code = append(code, &bytecode.BuildTuple{len(node.Values)})
	return code, &VeniceTupleType{itemTypes}, nil
}

func (compiler *Compiler) compileUnaryNode(
	node *UnaryNode,
) ([]bytecode.Bytecode, VeniceType, error) {
	switch node.Operator {
	case "-":
		code, exprType, err := compiler.compileExpression(node.Expr)
		if err != nil {
			return nil, nil, err
		}

		if exprType != VENICE_TYPE_INTEGER {
			return nil, nil, compiler.customError(
				node,
				"argument to unary minus must be integer, not %s",
				exprType.String(),
			)
		}

		code = append(code, &bytecode.UnaryMinus{})
		return code, VENICE_TYPE_INTEGER, nil
	case "not":
		code, exprType, err := compiler.compileExpression(node.Expr)
		if err != nil {
			return nil, nil, err
		}

		if exprType != VENICE_TYPE_BOOLEAN {
			return nil, nil, compiler.customError(
				node,
				"argument to `not` must be boolean, not %s",
				exprType.String(),
			)
		}

		code = append(code, &bytecode.UnaryNot{})
		return code, VENICE_TYPE_BOOLEAN, nil
	default:
		return nil, nil, compiler.customError(
			node, "unknown unary operator `%s`", node.Operator,
		)
	}
}

/**
 * Type utility functions
 */

var opsToBytecodes = map[string]bytecode.Bytecode{
	"+":   &bytecode.BinaryAdd{},
	"and": &bytecode.BinaryAnd{},
	"++":  &bytecode.BinaryConcat{},
	"==":  &bytecode.BinaryEq{},
	">":   &bytecode.BinaryGt{},
	">=":  &bytecode.BinaryGtEq{},
	"in":  &bytecode.BinaryIn{},
	"<":   &bytecode.BinaryLt{},
	"<=":  &bytecode.BinaryLtEq{},
	"*":   &bytecode.BinaryMul{},
	"!=":  &bytecode.BinaryNotEq{},
	"or":  &bytecode.BinaryOr{},
	"-":   &bytecode.BinarySub{},
	"/":   nil,
}

var opsToBytecodesReal = map[string]bytecode.Bytecode{
	"+": &bytecode.BinaryRealAdd{},
	"/": &bytecode.BinaryRealDiv{},
	"*": &bytecode.BinaryRealMul{},
	"-": &bytecode.BinaryRealSub{},
}

func (compiler *Compiler) checkInfixLeftType(operator string, leftType VeniceType) bool {
	switch operator {
	case "==", "in":
		return true
	case "and", "or":
		return compiler.checkType(VENICE_TYPE_BOOLEAN, leftType)
	case "++":
		if _, ok := leftType.(*VeniceListType); ok {
			return true
		} else {
			return compiler.checkType(VENICE_TYPE_STRING, leftType)
		}
	default:
		return compiler.checkType(VENICE_TYPE_INTEGER, leftType) ||
			compiler.checkType(VENICE_TYPE_REAL_NUMBER, leftType)
	}
}

func (compiler *Compiler) checkInfixRightType(
	operator string, leftType VeniceType, rightType VeniceType,
) (VeniceType, bool) {
	switch operator {
	case "==", "!=":
		return VENICE_TYPE_BOOLEAN, compiler.checkType(leftType, rightType)
	case "and", "or":
		return VENICE_TYPE_BOOLEAN,
			compiler.checkType(VENICE_TYPE_BOOLEAN, rightType)
	case ">", ">=", "<", "<=":
		return VENICE_TYPE_BOOLEAN,
			compiler.checkType(VENICE_TYPE_INTEGER, rightType)
	case "++":
		if compiler.checkType(VENICE_TYPE_STRING, leftType) {
			return VENICE_TYPE_STRING,
				compiler.checkType(VENICE_TYPE_STRING, rightType)
		} else {
			return leftType, compiler.checkType(leftType, rightType)
		}
	case "in":
		switch rightConcreteType := rightType.(type) {
		case *VeniceStringType:
			return VENICE_TYPE_BOOLEAN, compiler.checkType(VENICE_TYPE_STRING, leftType)
		case *VeniceListType:
			return VENICE_TYPE_BOOLEAN,
				compiler.checkType(rightConcreteType.ItemType, leftType)
		case *VeniceMapType:
			return VENICE_TYPE_BOOLEAN,
				compiler.checkType(rightConcreteType.KeyType, leftType)
		default:
			return nil, false
		}
	case "/":
		return VENICE_TYPE_REAL_NUMBER,
			compiler.checkType(VENICE_TYPE_INTEGER, rightType) ||
				compiler.checkType(VENICE_TYPE_REAL_NUMBER, rightType)
	default:
		_, leftIsRealNumber := leftType.(*VeniceRealNumberType)
		_, rightIsRealNumber := rightType.(*VeniceRealNumberType)

		var returnType VeniceType
		if leftIsRealNumber || rightIsRealNumber {
			returnType = VENICE_TYPE_REAL_NUMBER
		} else {
			returnType = VENICE_TYPE_INTEGER
		}
		return returnType, compiler.checkType(VENICE_TYPE_INTEGER, rightType) ||
			compiler.checkType(VENICE_TYPE_REAL_NUMBER, rightType)
	}
}

func (compiler *Compiler) checkType(
	expectedTypeAny VeniceType,
	actualTypeAny VeniceType,
) bool {
	return compiler.checkTypeCore(expectedTypeAny, actualTypeAny, false)
}

func (compiler *Compiler) checkTypeAndSubstituteGenerics(
	expectedTypeAny VeniceType,
	actualTypeAny VeniceType,
) bool {
	return compiler.checkTypeCore(expectedTypeAny, actualTypeAny, true)
}

func (compiler *Compiler) checkTypeCore(
	expectedTypeAny VeniceType,
	actualTypeAny VeniceType,
	substituteGenerics bool,
) bool {
	switch expectedType := expectedTypeAny.(type) {
	case *VeniceAnyType:
		return true
	case *VeniceEnumType:
		actualType, ok := actualTypeAny.(*VeniceEnumType)
		if !ok {
			return false
		}

		if expectedType.Name != actualType.Name {
			return false
		}

		if len(expectedType.Cases) != len(actualType.Cases) {
			return false
		}

		for i := 0; i < len(expectedType.Cases); i++ {
			if expectedType.Cases[i].Label != actualType.Cases[i].Label {
				return false
			}

			if len(expectedType.Cases[i].Types) != len(actualType.Cases[i].Types) {
				return false
			}

			for j := 0; j < len(expectedType.Cases[i].Types); j++ {
				if !compiler.checkType(
					expectedType.Cases[i].Types[j],
					actualType.Cases[i].Types[j],
				) {
					return false
				}
			}
		}

		return true
	case *VeniceListType:
		actualType, ok := actualTypeAny.(*VeniceListType)
		return ok && compiler.checkType(expectedType.ItemType, actualType.ItemType)
	case *VeniceMapType:
		actualType, ok := actualTypeAny.(*VeniceMapType)
		return ok &&
			compiler.checkType(expectedType.KeyType, actualType.KeyType) &&
			compiler.checkType(expectedType.ValueType, actualType.ValueType)
	case *VeniceSymbolType:
		if substituteGenerics {
			compiler.typeSymbolTable.Put(expectedType.Label, actualTypeAny)
			return true
		} else {
			return false
		}
	case *VeniceTupleType:
		actualType, ok := actualTypeAny.(*VeniceTupleType)
		if !ok {
			return false
		}

		if len(expectedType.ItemTypes) != len(actualType.ItemTypes) {
			return false
		}

		for i, itemType := range expectedType.ItemTypes {
			if !compiler.checkType(itemType, actualType.ItemTypes[i]) {
				return false
			}
		}

		return true
	case *VeniceUnionType:
		for _, subType := range expectedType.Types {
			if compiler.checkType(subType, actualTypeAny) {
				return true
			}
		}

		return false
	default:
		return expectedTypeAny == actualTypeAny
	}
}

func (compiler *Compiler) substituteGenerics(
	genericParametersMap map[string]VeniceType,
	vtypeAny VeniceType,
) (VeniceType, error) {
	switch vtype := vtypeAny.(type) {
	case *VeniceEnumType:
		cases := make([]*VeniceCaseType, 0, len(vtype.Cases))
		for _, enumCase := range vtype.Cases {
			caseTypes := make([]VeniceType, 0, len(enumCase.Types))
			for _, caseType := range enumCase.Types {
				newCaseType, err := compiler.substituteGenerics(
					genericParametersMap,
					caseType,
				)
				if err != nil {
					return nil, err
				}
				caseTypes = append(caseTypes, newCaseType)
			}
			cases = append(cases, &VeniceCaseType{enumCase.Label, caseTypes})
		}
		return &VeniceEnumType{Name: vtype.Name, Cases: cases}, nil
	case *VeniceFunctionType:
		newParamTypes := make([]VeniceType, 0, len(vtype.ParamTypes))
		for _, paramType := range vtype.ParamTypes {
			newParamType, err := compiler.substituteGenerics(
				genericParametersMap,
				paramType,
			)
			if err != nil {
				return nil, err
			}
			newParamTypes = append(newParamTypes, newParamType)
		}

		newReturnType, err := compiler.substituteGenerics(
			genericParametersMap,
			vtype.ReturnType,
		)
		if err != nil {
			return nil, err
		}

		return &VeniceFunctionType{
			Name:       vtype.Name,
			Public:     vtype.Public,
			ParamTypes: newParamTypes,
			ReturnType: newReturnType,
			IsBuiltin:  vtype.IsBuiltin,
		}, nil
	case *VeniceListType:
		itemType, err := compiler.substituteGenerics(genericParametersMap, vtype.ItemType)
		if err != nil {
			return nil, err
		}
		return &VeniceListType{itemType}, nil
	case *VeniceMapType:
		keyType, err := compiler.substituteGenerics(genericParametersMap, vtype.KeyType)
		if err != nil {
			return nil, err
		}

		valueType, err := compiler.substituteGenerics(genericParametersMap, vtype.ValueType)
		if err != nil {
			return nil, err
		}

		return &VeniceMapType{KeyType: keyType, ValueType: valueType}, nil
	case *VeniceSymbolType:
		concreteType, ok := genericParametersMap[vtype.Label]
		if ok {
			return concreteType, nil
		} else {
			return vtype, nil
		}
	case *VeniceTupleType:
		newItemTypes := make([]VeniceType, 0, len(vtype.ItemTypes))
		for _, itemType := range vtype.ItemTypes {
			newItemType, err := compiler.substituteGenerics(
				genericParametersMap,
				itemType,
			)
			if err != nil {
				return nil, err
			}
			newItemTypes = append(newItemTypes, newItemType)
		}
		return &VeniceTupleType{newItemTypes}, nil
	default:
		return vtypeAny, nil
	}
}

func (compiler *Compiler) resolveType(typeNodeAny TypeNode) (VeniceType, error) {
	if typeNodeAny == nil {
		return nil, nil
	}

	switch typeNode := typeNodeAny.(type) {
	case *ListTypeNode:
		itemType, err := compiler.resolveType(typeNode.ItemTypeNode)
		if err != nil {
			return nil, err
		}

		return &VeniceListType{itemType}, nil
	case *MapTypeNode:
		keyType, err := compiler.resolveType(typeNode.KeyTypeNode)
		if err != nil {
			return nil, err
		}

		valueType, err := compiler.resolveType(typeNode.ValueTypeNode)
		if err != nil {
			return nil, err
		}

		return &VeniceMapType{KeyType: keyType, ValueType: valueType}, nil
	case *ParameterizedTypeNode:
		resolvedType, ok := compiler.typeSymbolTable.Get(typeNode.Symbol)
		if !ok {
			return nil, compiler.customError(
				typeNodeAny, "unknown type `%s`", typeNode.Symbol,
			)
		}

		genericParameters := resolvedType.GetGenericParameters()
		if len(typeNode.TypeNodes) != len(genericParameters) {
			return nil, compiler.customError(
				typeNode,
				"expected %d generic parameter(s), got %d",
				len(genericParameters),
				len(typeNode.TypeNodes),
			)
		}

		genericParametersMap := map[string]VeniceType{}
		for i := 0; i < len(typeNode.TypeNodes); i++ {
			subTypeNode := typeNode.TypeNodes[i]
			genericParameter := genericParameters[i]

			subType, err := compiler.resolveType(subTypeNode)
			if err != nil {
				return nil, err
			}

			genericParametersMap[genericParameter] = subType
		}

		resolvedType, err := compiler.substituteGenerics(
			genericParametersMap,
			resolvedType,
		)
		if err != nil {
			return nil, err
		}

		return resolvedType, nil
	case *SymbolNode:
		resolvedType, ok := compiler.typeSymbolTable.Get(typeNode.Value)
		if !ok {
			return nil, compiler.customError(
				typeNodeAny, "unknown type `%s`", typeNode.Value,
			)
		}
		return resolvedType, nil
	case *TupleTypeNode:
		types := make([]VeniceType, 0, len(typeNode.TypeNodes))
		for _, subTypeNode := range typeNode.TypeNodes {
			subType, err := compiler.resolveType(subTypeNode)
			if err != nil {
				return nil, err
			}
			types = append(types, subType)
		}
		return &VeniceTupleType{types}, nil
	default:
		return nil, compiler.customError(
			typeNodeAny, "unknown type node: %T", typeNodeAny,
		)
	}
}

/**
 * Miscellaneous types and methods
 */

type FunctionInfo struct {
	declaredReturnType  VeniceType
	seenReturnStatement bool
}

func (compiler *Compiler) PrintSymbolTable() {
	for key, value := range compiler.symbolTable.Symbols {
		fmt.Printf("%s: %s\n", key, value.Type.String())
	}
}

func (compiler *Compiler) PrintTypeSymbolTable() {
	for key, value := range compiler.typeSymbolTable.Symbols {
		fmt.Printf("%s: %s\n", key, value.Type.String())
	}
}

func moduleTypeFromSymbolTable(name string, symbolTable *SymbolTable) *VeniceModuleType {
	moduleTypes := map[string]VeniceType{}
	for key, value := range symbolTable.Symbols {
		moduleTypes[key] = value.Type
	}
	return &VeniceModuleType{Name: name, Types: moduleTypes}
}

type CompileError struct {
	Message  string
	Location *lex.Location
}

func (e *CompileError) Error() string {
	if e.Location != nil {
		return fmt.Sprintf("%s at %s", e.Message, e.Location)
	} else {
		return e.Message
	}
}

func (compiler *Compiler) customError(
	node Node, message string, args ...interface{},
) *CompileError {
	return &CompileError{fmt.Sprintf(message, args...), node.GetLocation()}
}
