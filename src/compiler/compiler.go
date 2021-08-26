/**
 * The Venice compiler.
 *
 * The compiler compiles a Venice program (represented as an abstract syntax tree, the
 * output of src/parser.go) into bytecode instructions. It also checks the static types
 * of the programs and reports any errors.
 */
package compiler

import (
	"fmt"
	"github.com/iafisher/venice/src/ast"
	"github.com/iafisher/venice/src/bytecode"
	"github.com/iafisher/venice/src/lexer"
	"github.com/iafisher/venice/src/parser"
	"github.com/iafisher/venice/src/vtype"
)

type Compiler struct {
	symbolTable     *SymbolTable
	typeSymbolTable *SymbolTable
	functionInfo    *FunctionInfo
	nestedLoopCount int
}

func NewCompiler() *Compiler {
	return &Compiler{
		symbolTable:     NewBuiltinSymbolTable(),
		typeSymbolTable: NewBuiltinTypeSymbolTable(),
		functionInfo:    nil,
		nestedLoopCount: 0,
	}
}

func (compiler *Compiler) Compile(file *ast.File) (*bytecode.CompiledProgram, error) {
	compiledProgram, _, err := compiler.compileModule("", file)
	return compiledProgram, err
}

func (compiler *Compiler) compileModule(
	moduleName string, file *ast.File,
) (*bytecode.CompiledProgram, vtype.VeniceType, error) {
	compiledProgram := bytecode.NewCompiledProgram()
	for _, statementAny := range file.Statements {
		switch statement := statementAny.(type) {
		case *ast.ClassDeclarationNode:
			err := compiler.compileClassDeclaration(compiledProgram, statement)
			if err != nil {
				return nil, nil, err
			}
		case *ast.EnumDeclarationNode:
			err := compiler.compileEnumDeclaration(statement)
			if err != nil {
				return nil, nil, err
			}
		case *ast.FunctionDeclarationNode:
			code, err := compiler.compileFunctionDeclaration(statement)
			if err != nil {
				return nil, nil, err
			}
			compiledProgram.Code[statement.Name] = code
		case *ast.ImportStatementNode:
			importedFile, err := parser.NewParser().ParseFile(statement.Path)
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
				compiledProgram.Code[qualifiedName] = functionCode
			}

			compiler.symbolTable.Put(statement.Name, moduleType)
		default:
			code, err := compiler.compileStatement(statementAny)
			if err != nil {
				return nil, nil, err
			}
			compiledProgram.Code["main"] = append(compiledProgram.Code["main"], code...)
		}
	}
	return compiledProgram, moduleTypeFromSymbolTable(moduleName, compiler.symbolTable), nil
}

func (compiler *Compiler) GetType(expr ast.ExpressionNode) (vtype.VeniceType, error) {
	_, exprType, err := compiler.compileExpression(expr)
	return exprType, err
}

/**
 * Compile statements
 */

func (compiler *Compiler) compileStatement(
	treeAny ast.StatementNode,
) ([]bytecode.Bytecode, error) {
	switch node := treeAny.(type) {
	case *ast.AssignStatementNode:
		return compiler.compileAssignStatement(node)
	case *ast.BreakStatementNode:
		return compiler.compileBreakStatement(node)
	case *ast.ContinueStatementNode:
		return compiler.compileContinueStatement(node)
	case *ast.ExpressionStatementNode:
		code, _, err := compiler.compileExpression(node.Expr)
		return code, err
	case *ast.ForLoopNode:
		return compiler.compileForLoop(node)
	case *ast.IfStatementNode:
		return compiler.compileIfStatement(node)
	case *ast.LetStatementNode:
		return compiler.compileLetStatement(node)
	case *ast.MatchStatementNode:
		return compiler.compileMatchStatement(node)
	case *ast.ReturnStatementNode:
		return compiler.compileReturnStatement(node)
	case *ast.WhileLoopNode:
		return compiler.compileWhileLoop(node)
	default:
		return nil, compiler.customError(treeAny, "unknown statement type: %T", treeAny)
	}
}

func (compiler *Compiler) compileAssignStatement(
	node *ast.AssignStatementNode,
) ([]bytecode.Bytecode, error) {
	switch destination := node.Destination.(type) {
	case *ast.FieldAccessNode:
		return compiler.compileAssignStatementToField(node, destination)
	case *ast.IndexNode:
		return compiler.compileAssignStatementToIndex(node, destination)
	case *ast.SymbolNode:
		return compiler.compileAssignStatementToSymbol(node, destination)
	default:
		return nil, compiler.customError(destination, "cannot assign to non-symbol")
	}
}

func (compiler *Compiler) compileBreakStatement(
	node *ast.BreakStatementNode,
) ([]bytecode.Bytecode, error) {
	if compiler.nestedLoopCount == 0 {
		return nil, compiler.customError(node, "break statement outside of loop")
	}

	// Return a placeholder bytecode instruction that the compiler will later convert
	// to a REL_JUMP instruction.
	return []bytecode.Bytecode{&bytecode.Placeholder{"break"}}, nil
}

func (compiler *Compiler) compileContinueStatement(
	node *ast.ContinueStatementNode,
) ([]bytecode.Bytecode, error) {
	if compiler.nestedLoopCount == 0 {
		return nil, compiler.customError(node, "continue statement outside of loop")
	}

	// Return a placeholder bytecode instruction that the compiler will later convert
	// to a REL_JUMP instruction.
	return []bytecode.Bytecode{&bytecode.Placeholder{"continue"}}, nil
}

func (compiler *Compiler) compileAssignStatementToField(
	node *ast.AssignStatementNode,
	destination *ast.FieldAccessNode,
) ([]bytecode.Bytecode, error) {
	code, eType, err := compiler.compileExpression(node.Expr)
	if err != nil {
		return nil, err
	}

	destinationCode, destinationTypeAny, err := compiler.compileExpression(destination.Expr)
	if err != nil {
		return nil, err
	}

	classType, ok := destinationTypeAny.(*vtype.VeniceClassType)
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
	node *ast.AssignStatementNode,
	destination *ast.IndexNode,
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
	case *vtype.VeniceListType:
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

		if !compiler.checkType(vtype.VENICE_TYPE_INTEGER, indexType) {
			return nil, compiler.customError(
				node, "list index must be of type integer, not %s", indexType.String(),
			)
		}

		code = append(code, indexCode...)
		code = append(code, destinationCode...)
		code = append(code, &bytecode.StoreIndex{})
		return code, nil
	case *vtype.VeniceMapType:
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
	node *ast.AssignStatementNode,
	destination *ast.SymbolNode,
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
			node, "wrong expression type in assignment to `%s`", destination.Value,
		)
	}

	code = append(code, &bytecode.StoreName{destination.Value})
	return code, nil
}

func (compiler *Compiler) compileClassDeclaration(
	compiledProgram *bytecode.CompiledProgram, node *ast.ClassDeclarationNode,
) error {
	if node.GenericTypeParameter != "" {
		compiler.typeSymbolTable = compiler.typeSymbolTable.SpawnChild()
		compiler.typeSymbolTable.Put(
			node.GenericTypeParameter,
			&vtype.VeniceSymbolType{node.GenericTypeParameter},
		)
	}

	fields := []*vtype.VeniceClassField{}
	paramTypes := []vtype.VeniceType{}
	for _, field := range node.Fields {
		paramType, err := compiler.resolveType(field.FieldType)
		if err != nil {
			return err
		}
		paramTypes = append(paramTypes, paramType)
		fields = append(
			fields,
			&vtype.VeniceClassField{
				field.Name, field.Public, paramType,
			},
		)
	}

	methods := []*vtype.VeniceFunctionType{}
	for _, method := range node.Methods {
		methodParams := []string{}
		methodParamTypes := []vtype.VeniceType{}
		for _, param := range method.Params {
			paramType, err := compiler.resolveType(param.ParamType)
			if err != nil {
				return err
			}

			methodParams = append(methodParams, param.Name)
			methodParamTypes = append(methodParamTypes, paramType)
		}

		declaredReturnType, err := compiler.resolveType(method.ReturnType)
		if err != nil {
			return err
		}

		methodType := &vtype.VeniceFunctionType{
			Name:       method.Name,
			Public:     method.Public,
			ParamTypes: paramTypes,
			ReturnType: declaredReturnType,
			IsBuiltin:  false,
		}
		methods = append(methods, methodType)
	}

	var classType vtype.VeniceType
	if node.GenericTypeParameter == "" {
		classType = &vtype.VeniceClassType{
			Name: node.Name, Fields: fields, Methods: methods,
		}
	} else {
		classType = &vtype.VeniceClassType{
			Name:              node.Name,
			GenericParameters: []string{node.GenericTypeParameter},
			Fields:            fields,
			Methods:           methods,
		}
		compiler.typeSymbolTable = compiler.typeSymbolTable.Parent
	}
	compiler.typeSymbolTable.Put(node.Name, classType)
	constructorType := &vtype.VeniceFunctionType{
		Name:       node.Name,
		ParamTypes: paramTypes,
		ReturnType: classType,
		IsBuiltin:  false,
	}
	compiler.symbolTable.Put(node.Name, constructorType)

	for i := 0; i < len(methods); i++ {
		method := node.Methods[i]
		methodType := methods[i]
		bodySymbolTable := compiler.symbolTable.SpawnChild()
		for j := 0; j < len(method.Params); j++ {
			bodySymbolTable.Put(method.Params[j].Name, methodType.ParamTypes[j])
		}

		declaredReturnType := methodType.ReturnType

		// Put `self` in the symbol table.
		bodySymbolTable.Put("self", classType)

		compiler.symbolTable = bodySymbolTable
		compiler.functionInfo = &FunctionInfo{
			declaredReturnType:  declaredReturnType,
			seenReturnStatement: false,
		}
		bodyCode, err := compiler.compileBlock(method.Body)
		if err != nil {
			return err
		}

		if declaredReturnType != nil && !compiler.functionInfo.seenReturnStatement {
			return compiler.customError(node, "non-void function has no return statement")
		}

		compiler.functionInfo = nil
		compiler.symbolTable = bodySymbolTable.Parent

		paramLoadCode := []bytecode.Bytecode{
			&bytecode.StoreName{"self"},
		}
		for i := len(method.Params) - 1; i >= 0; i-- {
			param := method.Params[i]
			paramLoadCode = append(paramLoadCode, &bytecode.StoreName{param.Name})
		}
		bodyCode = append(paramLoadCode, bodyCode...)

		compiledProgram.Code[fmt.Sprintf("%s__%s", node.Name, method.Name)] = bodyCode
	}

	constructorBytecode := []bytecode.Bytecode{&bytecode.BuildClass{node.Name, len(fields)}}
	compiledProgram.Code[node.Name] = constructorBytecode
	return nil
}

func (compiler *Compiler) compileEnumDeclaration(node *ast.EnumDeclarationNode) error {
	// Put in a dummy entry for the enum in the type symbol table so that recursive enum
	// types will type-check properly.
	enumType := &vtype.VeniceEnumType{
		Name:              node.Name,
		GenericParameters: nil,
		Cases:             nil,
	}
	compiler.typeSymbolTable.Put(node.Name, enumType)

	if node.GenericTypeParameter != "" {
		compiler.typeSymbolTable = compiler.typeSymbolTable.SpawnChild()
		compiler.typeSymbolTable.Put(
			node.GenericTypeParameter,
			&vtype.VeniceSymbolType{node.GenericTypeParameter},
		)
	}

	caseTypes := []*vtype.VeniceCaseType{}
	for _, caseNode := range node.Cases {
		types := []vtype.VeniceType{}
		for _, typeNode := range caseNode.Types {
			veniceType, err := compiler.resolveType(typeNode)
			if err != nil {
				return err
			}
			types = append(types, veniceType)
		}
		caseTypes = append(caseTypes, &vtype.VeniceCaseType{caseNode.Label, types})
	}

	if node.GenericTypeParameter != "" {
		compiler.typeSymbolTable = compiler.typeSymbolTable.Parent
		enumType.GenericParameters = []string{node.GenericTypeParameter}
	}
	enumType.Cases = caseTypes

	return nil
}

func (compiler *Compiler) compileForLoop(
	node *ast.ForLoopNode,
) ([]bytecode.Bytecode, error) {
	iterableCode, iterableTypeAny, err := compiler.compileExpression(node.Iterable)
	if err != nil {
		return nil, err
	}

	loopSymbolTable := compiler.symbolTable.SpawnChild()
	switch iterableType := iterableTypeAny.(type) {
	case *vtype.VeniceListType:
		if len(node.Variables) != 1 {
			return nil, compiler.customError(node, "too many for loop variables")
		}

		loopSymbolTable.Put(node.Variables[0], iterableType.ItemType)
	case *vtype.VeniceMapType:
		if len(node.Variables) != 2 {
			return nil, compiler.customError(node, "too many for loop variables")
		}

		loopSymbolTable.Put(node.Variables[0], iterableType.KeyType)
		loopSymbolTable.Put(node.Variables[1], iterableType.ValueType)
	default:
		return nil, compiler.customError(node.Iterable, "for loop must be of list")
	}

	compiler.symbolTable = loopSymbolTable
	compiler.nestedLoopCount += 1
	bodyCode, err := compiler.compileBlock(node.Body)
	compiler.nestedLoopCount -= 1
	compiler.symbolTable = compiler.symbolTable.Parent

	if err != nil {
		return nil, err
	}

	code := iterableCode
	code = append(code, &bytecode.GetIter{})
	code = append(code, &bytecode.ForIter{len(bodyCode) + len(node.Variables) + 2})
	for i := len(node.Variables) - 1; i >= 0; i-- {
		code = append(code, &bytecode.StoreName{node.Variables[i]})
	}
	code = append(code, bodyCode...)
	code = append(code, &bytecode.RelJump{-(len(bodyCode) + len(node.Variables) + 1)})
	return code, nil
}

func (compiler *Compiler) compileFunctionDeclaration(
	node *ast.FunctionDeclarationNode,
) ([]bytecode.Bytecode, error) {
	params := []string{}
	paramTypes := []vtype.VeniceType{}
	bodySymbolTable := compiler.symbolTable.SpawnChild()
	for _, param := range node.Params {
		paramType, err := compiler.resolveType(param.ParamType)
		if err != nil {
			return nil, err
		}

		params = append(params, param.Name)
		paramTypes = append(paramTypes, paramType)

		bodySymbolTable.Put(param.Name, paramType)
	}

	declaredReturnType, err := compiler.resolveType(node.ReturnType)
	if err != nil {
		return nil, err
	}

	// Put the function's entry in the symbol table before compiling the body so that
	// recursive functions can call themselves.
	compiler.symbolTable.Put(
		node.Name,
		&vtype.VeniceFunctionType{
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
		return nil, err
	}

	if declaredReturnType != nil && !compiler.functionInfo.seenReturnStatement {
		return nil, compiler.customError(node, "non-void function has no return statement")
	}

	compiler.functionInfo = nil
	compiler.symbolTable = bodySymbolTable.Parent

	paramLoadCode := []bytecode.Bytecode{}
	for i := len(node.Params) - 1; i >= 0; i-- {
		param := node.Params[i]
		paramLoadCode = append(paramLoadCode, &bytecode.StoreName{param.Name})
	}
	bodyCode = append(paramLoadCode, bodyCode...)

	return bodyCode, nil
}

func (compiler *Compiler) compileIfStatement(
	node *ast.IfStatementNode,
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

		if !compiler.checkType(vtype.VENICE_TYPE_BOOLEAN, conditionType) {
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
	node *ast.LetStatementNode,
) ([]bytecode.Bytecode, error) {
	if _, ok := compiler.symbolTable.Get(node.Symbol); ok {
		return nil, compiler.customError(node, "re-declaration of symbol `%q`", node.Symbol)
	}

	code, eType, err := compiler.compileExpression(node.Expr)
	if err != nil {
		return nil, err
	}

	if node.Type != nil {
		declaredType, err := compiler.resolveType(node.Type)
		if err != nil {
			return nil, err
		}

		if !compiler.checkType(declaredType, eType) {
			return nil, compiler.customError(
				node.Expr, "expected %s, got %s", declaredType.String(), eType.String(),
			)
		}
	}

	if node.IsVar {
		compiler.symbolTable.PutVar(node.Symbol, eType)
	} else {
		compiler.symbolTable.Put(node.Symbol, eType)
	}
	return append(code, &bytecode.StoreName{node.Symbol}), nil
}

func (compiler *Compiler) compileMatchStatement(
	node *ast.MatchStatementNode,
) ([]bytecode.Bytecode, error) {
	exprCode, exprType, err := compiler.compileExpression(node.Expr)
	if err != nil {
		return nil, err
	}

	enumType, ok := exprType.(*vtype.VeniceEnumType)
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
	patternAny ast.PatternNode, exprType vtype.VeniceType,
) ([]bytecode.Bytecode, error) {
	switch pattern := patternAny.(type) {
	case *ast.CompoundPatternNode:
		enumType, ok := exprType.(*vtype.VeniceEnumType)
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
	case *ast.SymbolNode:
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
	node *ast.ReturnStatementNode,
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
	node *ast.WhileLoopNode,
) ([]bytecode.Bytecode, error) {
	conditionCode, conditionType, err := compiler.compileExpression(node.Condition)
	if err != nil {
		return nil, err
	}

	if !compiler.checkType(vtype.VENICE_TYPE_BOOLEAN, conditionType) {
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
	block []ast.StatementNode,
) ([]bytecode.Bytecode, error) {
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
	nodeAny ast.ExpressionNode,
) ([]bytecode.Bytecode, vtype.VeniceType, error) {
	switch node := nodeAny.(type) {
	case *ast.BooleanNode:
		return []bytecode.Bytecode{
			&bytecode.PushConstBool{node.Value},
		}, vtype.VENICE_TYPE_BOOLEAN, nil
	case *ast.CallNode:
		return compiler.compileCallNode(node)
	case *ast.CharacterNode:
		return []bytecode.Bytecode{
			&bytecode.PushConstChar{node.Value},
		}, vtype.VENICE_TYPE_CHARACTER, nil
	case *ast.FieldAccessNode:
		return compiler.compileFieldAccessNode(node)
	case *ast.IndexNode:
		return compiler.compileIndexNode(node)
	case *ast.InfixNode:
		return compiler.compileInfixNode(node)
	case *ast.IntegerNode:
		return []bytecode.Bytecode{
			&bytecode.PushConstInt{node.Value},
		}, vtype.VENICE_TYPE_INTEGER, nil
	case *ast.ListNode:
		return compiler.compileListNode(node)
	case *ast.MapNode:
		return compiler.compileMapNode(node)
	case *ast.QualifiedSymbolNode:
		return compiler.compileQualifiedSymbolNode(node)
	case *ast.StringNode:
		return []bytecode.Bytecode{
			&bytecode.PushConstStr{node.Value},
		}, vtype.VENICE_TYPE_STRING, nil
	case *ast.SymbolNode:
		symbolType, ok := compiler.symbolTable.Get(node.Value)
		if !ok {
			return nil, nil, compiler.customError(
				nodeAny, "undefined symbol: %s", node.Value,
			)
		}
		// TODO(2021-08-26): Do we need to handle function types separately?
		if functionType, isFunctionType := symbolType.(*vtype.VeniceFunctionType); isFunctionType {
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
	case *ast.TernaryIfNode:
		return compiler.compileTernaryIfNode(node)
	case *ast.TupleFieldAccessNode:
		return compiler.compileTupleFieldAccessNode(node)
	case *ast.TupleNode:
		return compiler.compileTupleNode(node)
	case *ast.UnaryNode:
		return compiler.compileUnaryNode(node)
	default:
		return nil, nil, compiler.customError(
			nodeAny, "unknown expression type: %T", nodeAny,
		)
	}
}

func (compiler *Compiler) compileCallNode(
	node *ast.CallNode,
) ([]bytecode.Bytecode, vtype.VeniceType, error) {
	if nodeAsEnumSymbol, ok := node.Function.(*ast.QualifiedSymbolNode); ok {
		return compiler.compileEnumCallNode(nodeAsEnumSymbol, node)
	}

	functionCode, functionTypeAny, err := compiler.compileExpression(node.Function)
	if err != nil {
		return nil, nil, err
	}

	functionType, ok := functionTypeAny.(*vtype.VeniceFunctionType)
	if !ok {
		return nil, nil, compiler.customError(
			node.Function,
			"type %s cannot be used as a function",
			functionTypeAny.String(),
		)
	}

	_, isClassMethod := node.Function.(*ast.FieldAccessNode)
	code, returnType, err := compiler.compileFunctionArguments(
		node,
		node.Args,
		functionType,
		isClassMethod,
	)
	if err != nil {
		return nil, nil, err
	}

	code = append(code, functionCode...)
	code = append(code, &bytecode.CallFunction{len(functionType.ParamTypes)})
	return code, returnType, nil
}

func (compiler *Compiler) compileEnumCallNode(
	enumSymbolNode *ast.QualifiedSymbolNode, callNode *ast.CallNode,
) ([]bytecode.Bytecode, vtype.VeniceType, error) {
	// TODO(2021-08-25): Clean this method up - enum calls and module calls should be separate.
	enumTypeAny, err := compiler.resolveType(&ast.SymbolNode{enumSymbolNode.Enum, nil})
	if err != nil {
		moduleTypeAny, ok1 := compiler.symbolTable.Get(enumSymbolNode.Enum)
		moduleType, ok2 := moduleTypeAny.(*vtype.VeniceModuleType)
		if !ok1 || !ok2 {
			// TODO(2021-08-25): Better error message.
			return nil, nil, compiler.customError(callNode, "invalid use of qualified symbol")
		}

		functionTypeAny, ok := moduleType.Types[enumSymbolNode.Case]
		if !ok {
			return nil, nil, compiler.customError(
				callNode,
				"module `%s` has no member `%s`",
				enumSymbolNode.Enum,
				enumSymbolNode.Case,
			)
		}

		functionType, ok := functionTypeAny.(*vtype.VeniceFunctionType)
		if !ok {
			return nil, nil, compiler.customError(
				callNode.Function,
				"type %s cannot be used as a function",
				functionTypeAny.String(),
			)
		}

		code, returnType, err := compiler.compileFunctionArguments(
			callNode,
			callNode.Args,
			functionType,
			false,
		)
		if err != nil {
			return nil, nil, err
		}

		code = append(
			code,
			&bytecode.PushConstFunction{
				Name:      fmt.Sprintf("%s::%s", enumSymbolNode.Enum, functionType.Name),
				IsBuiltin: functionType.IsBuiltin,
			},
		)
		code = append(code, &bytecode.CallFunction{len(functionType.ParamTypes)})
		return code, returnType, nil
	}

	enumType, ok := enumTypeAny.(*vtype.VeniceEnumType)
	if !ok {
		return nil, nil, compiler.customError(callNode, "cannot use double colon after non-enum type")
	}

	for _, enumCase := range enumType.Cases {
		if enumCase.Label == enumSymbolNode.Case {
			code, returnType, err := compiler.compileFunctionArguments(
				callNode, callNode.Args, enumCase.AsFunctionType(enumType), false,
			)
			if err != nil {
				return nil, nil, err
			}

			code = append(code, &bytecode.PushEnum{enumSymbolNode.Case, len(callNode.Args)})
			return code, returnType, nil
		}
	}

	return nil, nil, compiler.customError(
		callNode,
		"enum %s does not have case %s",
		enumSymbolNode.Enum,
		enumSymbolNode.Case,
	)
}

func (compiler *Compiler) compileQualifiedSymbolNode(
	node *ast.QualifiedSymbolNode,
) ([]bytecode.Bytecode, vtype.VeniceType, error) {
	enumTypeAny, err := compiler.resolveType(&ast.SymbolNode{node.Enum, nil})
	if err != nil {
		return nil, nil, err
	}

	enumType, ok := enumTypeAny.(*vtype.VeniceEnumType)
	if !ok {
		return nil, nil, compiler.customError(
			node, "cannot use double colon after non-enum type",
		)
	}

	for _, enumCase := range enumType.Cases {
		if enumCase.Label == node.Case {
			if len(enumCase.Types) != 0 {
				return nil, nil, compiler.customError(
					node,
					"%s takes %d argument(s)",
					enumCase.Label,
					len(enumCase.Types),
				)
			}

			if len(enumCase.Types) > 0 {
				functionName := fmt.Sprintf("%s__%s", enumType.Name, enumCase.Label)
				return []bytecode.Bytecode{&bytecode.PushConstFunction{functionName, false}},
					enumCase.AsFunctionType(enumType),
					nil
			} else {
				return []bytecode.Bytecode{&bytecode.PushEnum{node.Case, 0}}, enumType, nil
			}
		}
	}

	return nil, nil, compiler.customError(
		node, "enum %s does not have case %s", node.Enum, node.Case,
	)
}

func (compiler *Compiler) compileFieldAccessNode(
	node *ast.FieldAccessNode,
) ([]bytecode.Bytecode, vtype.VeniceType, error) {
	code, typeAny, err := compiler.compileExpression(node.Expr)
	if err != nil {
		return nil, nil, err
	}

	switch concreteType := typeAny.(type) {
	case *vtype.VeniceClassType:
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
	case *vtype.VeniceListType:
		methodType, ok := listBuiltins[node.Name]
		if !ok {
			return nil, nil, compiler.customError(
				node, "no such field or method `%s` on list type", node.Name,
			)
		}
		code = append(code, &bytecode.LookupMethod{node.Name})
		return code, methodType, nil
	case *vtype.VeniceMapType:
		methodType, ok := mapBuiltins[node.Name]
		if !ok {
			return nil, nil, compiler.customError(
				node, "no such field or method `%s` on map type", node.Name,
			)
		}
		code = append(code, &bytecode.LookupMethod{node.Name})
		return code, methodType, nil
	case *vtype.VeniceStringType:
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
	node ast.Node,
	args []ast.ExpressionNode,
	functionType *vtype.VeniceFunctionType,
	isClassMethod bool,
) ([]bytecode.Bytecode, vtype.VeniceType, error) {
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

	genericParameterMap := map[string]vtype.VeniceType{}
	for _, genericParameter := range functionType.GenericParameters {
		genericParameterMap[genericParameter] = nil
	}

	code := []bytecode.Bytecode{}
	for i := len(args) - 1; i >= 0; i-- {
		argCode, argType, err := compiler.compileExpression(args[i])
		if err != nil {
			return nil, nil, err
		}

		var paramType vtype.VeniceType
		if isClassMethod {
			paramType = functionType.ParamTypes[i+1]
		} else {
			paramType = functionType.ParamTypes[i]
		}

		err = compiler.checkFunctionArgType(
			args[i], paramType, argType, genericParameterMap,
		)
		if err != nil {
			return nil, nil, err
		}

		code = append(code, argCode...)
	}

	for key, value := range genericParameterMap {
		if value == nil {
			return nil, nil, compiler.customError(
				node, "unsubstituted generic parameter `%s`", key,
			)
		}
	}

	if len(genericParameterMap) > 0 && functionType.ReturnType != nil {
		return code, functionType.ReturnType.SubstituteGenerics(genericParameterMap), nil
	} else {
		return code, functionType.ReturnType, nil
	}
}

func (compiler *Compiler) checkFunctionArgType(
	node ast.Node,
	paramType vtype.VeniceType,
	argType vtype.VeniceType,
	genericParameterMap map[string]vtype.VeniceType,
) error {
	err := paramType.MatchGenerics(genericParameterMap, argType)
	if err != nil {
		return err
	}

	if !compiler.checkType(paramType, argType) {
		return compiler.customError(node, "wrong function parameter type")
	}

	return nil
}

func (compiler *Compiler) compileIndexNode(
	node *ast.IndexNode,
) ([]bytecode.Bytecode, vtype.VeniceType, error) {
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
	case *vtype.VeniceListType:
		if !compiler.checkType(vtype.VENICE_TYPE_INTEGER, indexType) {
			return nil, nil, compiler.customError(node.Expr, "list index must be integer")
		}

		code = append(code, &bytecode.BinaryListIndex{})
		return code, exprType.ItemType, nil
	case *vtype.VeniceMapType:
		if !compiler.checkType(exprType.KeyType, indexType) {
			return nil, nil, compiler.customError(
				node.Expr, "wrong map key type in index expression",
			)
		}

		code = append(code, &bytecode.BinaryMapIndex{})
		return code, vtype.VeniceOptionalTypeOf(exprType.ValueType), nil
	case *vtype.VeniceStringType:
		if !compiler.checkType(vtype.VENICE_TYPE_INTEGER, indexType) {
			return nil, nil, compiler.customError(node.Expr, "string index must be integer")
		}

		code = append(code, &bytecode.BinaryStringIndex{})
		return code, vtype.VENICE_TYPE_CHARACTER, nil
	default:
		return nil, nil, compiler.customError(node, "%s cannot be indexed", exprTypeAny)
	}
}

func (compiler *Compiler) compileInfixNode(
	node *ast.InfixNode,
) ([]bytecode.Bytecode, vtype.VeniceType, error) {
	// TODO(2021-08-07): Boolean operators should short-circuit.
	opBytecode, ok := opsToBytecodes[node.Operator]
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
		code = append(code, opBytecode)
	}

	return code, resultType, nil
}

func (compiler *Compiler) compileListNode(
	node *ast.ListNode,
) ([]bytecode.Bytecode, vtype.VeniceType, error) {
	code := []bytecode.Bytecode{}
	var itemType vtype.VeniceType
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
	code = append(code, &bytecode.BuildList{len(node.Values)})
	return code, &vtype.VeniceListType{itemType}, nil
}

func (compiler *Compiler) compileMapNode(
	node *ast.MapNode,
) ([]bytecode.Bytecode, vtype.VeniceType, error) {
	code := []bytecode.Bytecode{}
	var keyType vtype.VeniceType
	var valueType vtype.VeniceType
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
	code = append(code, &bytecode.BuildMap{len(node.Pairs)})
	return code, &vtype.VeniceMapType{keyType, valueType}, nil
}

func (compiler *Compiler) compileTernaryIfNode(
	node *ast.TernaryIfNode,
) ([]bytecode.Bytecode, vtype.VeniceType, error) {
	conditionCode, conditionType, err := compiler.compileExpression(node.Condition)
	if err != nil {
		return nil, nil, err
	}

	if !compiler.checkType(vtype.VENICE_TYPE_BOOLEAN, conditionType) {
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
	node *ast.TupleFieldAccessNode,
) ([]bytecode.Bytecode, vtype.VeniceType, error) {
	code, typeAny, err := compiler.compileExpression(node.Expr)
	if err != nil {
		return nil, nil, err
	}

	tupleType, ok := typeAny.(*vtype.VeniceTupleType)
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
	node *ast.TupleNode,
) ([]bytecode.Bytecode, vtype.VeniceType, error) {
	code := []bytecode.Bytecode{}
	itemTypes := []vtype.VeniceType{}
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
	return code, &vtype.VeniceTupleType{itemTypes}, nil
}

func (compiler *Compiler) compileUnaryNode(
	node *ast.UnaryNode,
) ([]bytecode.Bytecode, vtype.VeniceType, error) {
	switch node.Operator {
	case "-":
		code, exprType, err := compiler.compileExpression(node.Expr)
		if err != nil {
			return nil, nil, err
		}

		if exprType != vtype.VENICE_TYPE_INTEGER {
			return nil, nil, compiler.customError(
				node,
				"argument to unary minus must be integer, not %s",
				exprType.String(),
			)
		}

		code = append(code, &bytecode.UnaryMinus{})
		return code, vtype.VENICE_TYPE_INTEGER, nil
	case "not":
		code, exprType, err := compiler.compileExpression(node.Expr)
		if err != nil {
			return nil, nil, err
		}

		if exprType != vtype.VENICE_TYPE_BOOLEAN {
			return nil, nil, compiler.customError(
				node,
				"argument to `not` must be boolean, not %s",
				exprType.String(),
			)
		}

		code = append(code, &bytecode.UnaryNot{})
		return code, vtype.VENICE_TYPE_BOOLEAN, nil
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
	"/":   &bytecode.BinaryDiv{},
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
}

func (compiler *Compiler) checkInfixLeftType(operator string, leftType vtype.VeniceType) bool {
	switch operator {
	case "==", "in":
		return true
	case "and", "or":
		return compiler.checkType(vtype.VENICE_TYPE_BOOLEAN, leftType)
	case "++":
		if _, ok := leftType.(*vtype.VeniceListType); ok {
			return true
		} else {
			return compiler.checkType(vtype.VENICE_TYPE_STRING, leftType)
		}
	default:
		return compiler.checkType(vtype.VENICE_TYPE_INTEGER, leftType)
	}
}

func (compiler *Compiler) checkInfixRightType(
	operator string, leftType vtype.VeniceType, rightType vtype.VeniceType,
) (vtype.VeniceType, bool) {
	switch operator {
	case "==", "!=":
		return vtype.VENICE_TYPE_BOOLEAN, compiler.checkType(leftType, rightType)
	case "and", "or":
		return vtype.VENICE_TYPE_BOOLEAN,
			compiler.checkType(vtype.VENICE_TYPE_BOOLEAN, rightType)
	case ">", ">=", "<", "<=":
		return vtype.VENICE_TYPE_BOOLEAN,
			compiler.checkType(vtype.VENICE_TYPE_INTEGER, rightType)
	case "++":
		if compiler.checkType(vtype.VENICE_TYPE_STRING, leftType) {
			return vtype.VENICE_TYPE_STRING,
				compiler.checkType(vtype.VENICE_TYPE_STRING, rightType)
		} else {
			return leftType, compiler.checkType(leftType, rightType)
		}
	case "in":
		switch rightConcreteType := rightType.(type) {
		case *vtype.VeniceStringType:
			return vtype.VENICE_TYPE_BOOLEAN,
				compiler.checkType(vtype.VENICE_TYPE_CHARACTER, leftType) ||
					compiler.checkType(vtype.VENICE_TYPE_STRING, leftType)
		case *vtype.VeniceListType:
			return vtype.VENICE_TYPE_BOOLEAN,
				compiler.checkType(rightConcreteType.ItemType, leftType)
		case *vtype.VeniceMapType:
			return vtype.VENICE_TYPE_BOOLEAN,
				compiler.checkType(rightConcreteType.KeyType, leftType)
		default:
			return nil, false
		}
	default:
		return vtype.VENICE_TYPE_INTEGER,
			compiler.checkType(vtype.VENICE_TYPE_INTEGER, rightType)
	}
}

func (compiler *Compiler) checkType(
	expectedTypeAny vtype.VeniceType, actualTypeAny vtype.VeniceType,
) bool {
	switch expectedType := expectedTypeAny.(type) {
	case *vtype.VeniceAnyType:
		return true
	case *vtype.VeniceEnumType:
		actualType, ok := actualTypeAny.(*vtype.VeniceEnumType)
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
				// TODO(2021-08-25): Fix this dirty hack.
				if _, isSymbolType := actualType.Cases[i].Types[j].(*vtype.VeniceSymbolType); isSymbolType {
					continue
				}

				if !compiler.checkType(expectedType.Cases[i].Types[j], actualType.Cases[i].Types[j]) {
					return false
				}
			}
		}

		return true
	case *vtype.VeniceListType:
		actualType, ok := actualTypeAny.(*vtype.VeniceListType)
		return ok && compiler.checkType(expectedType.ItemType, actualType.ItemType)
	case *vtype.VeniceMapType:
		actualType, ok := actualTypeAny.(*vtype.VeniceMapType)
		return ok &&
			compiler.checkType(expectedType.KeyType, actualType.KeyType) &&
			compiler.checkType(expectedType.ValueType, actualType.ValueType)
	case *vtype.VeniceSymbolType:
		return true
		// TODO(2021-08-21)
		// symbolType, ok := compiler.typeSymbolTable.Get(expectedType.Label)
		// if !ok {
		// 	// TODO(2021-08-21): Return an error?
		// 	return false
		// }
		// return compiler.checkType(symbolType, actualTypeAny)
	case *vtype.VeniceTupleType:
		actualType, ok := actualTypeAny.(*vtype.VeniceTupleType)
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
	case *vtype.VeniceUnionType:
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

func (compiler *Compiler) resolveType(typeNodeAny ast.TypeNode) (vtype.VeniceType, error) {
	if typeNodeAny == nil {
		return nil, nil
	}

	switch typeNode := typeNodeAny.(type) {
	case *ast.ParameterizedTypeNode:
		switch typeNode.Symbol {
		case "map":
			if len(typeNode.TypeNodes) != 2 {
				return nil, compiler.customError(
					typeNodeAny, "map type requires 2 type parameters",
				)
			}

			keyType, err := compiler.resolveType(typeNode.TypeNodes[0])
			if err != nil {
				return nil, err
			}

			valueType, err := compiler.resolveType(typeNode.TypeNodes[1])
			if err != nil {
				return nil, err
			}

			return &vtype.VeniceMapType{KeyType: keyType, ValueType: valueType}, nil
		case "list":
			if len(typeNode.TypeNodes) != 1 {
				return nil, compiler.customError(
					typeNodeAny, "list type requires 1 type parameter",
				)
			}

			itemType, err := compiler.resolveType(typeNode.TypeNodes[0])
			if err != nil {
				return nil, err
			}

			return &vtype.VeniceListType{itemType}, nil
		case "tuple":
			types := []vtype.VeniceType{}
			for _, subTypeNode := range typeNode.TypeNodes {
				subType, err := compiler.resolveType(subTypeNode)
				if err != nil {
					return nil, err
				}
				types = append(types, subType)
			}
			return &vtype.VeniceTupleType{types}, nil
		default:
			resolvedType, ok := compiler.typeSymbolTable.Get(typeNode.Symbol)
			if !ok {
				return nil, compiler.customError(
					typeNodeAny, "unknown type `%s`", typeNode.Symbol,
				)
			}
			// TODO(2021-08-25): This is hard-coded to work only for Optional.
			subType, err := compiler.resolveType(typeNode.TypeNodes[0])
			if err != nil {
				return nil, err
			}
			genericsMap := map[string]vtype.VeniceType{"T": subType}
			return resolvedType.SubstituteGenerics(genericsMap), nil
		}
	case *ast.SymbolNode:
		resolvedType, ok := compiler.typeSymbolTable.Get(typeNode.Value)
		if !ok {
			return nil, compiler.customError(
				typeNodeAny, "unknown type `%s`", typeNode.Value,
			)
		}
		return resolvedType, nil
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
	declaredReturnType  vtype.VeniceType
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

func moduleTypeFromSymbolTable(name string, symbolTable *SymbolTable) *vtype.VeniceModuleType {
	moduleTypes := map[string]vtype.VeniceType{}
	for key, value := range symbolTable.Symbols {
		moduleTypes[key] = value.Type
	}
	return &vtype.VeniceModuleType{Name: name, Types: moduleTypes}
}

type CompileError struct {
	Message  string
	Location *lexer.Location
}

func (e *CompileError) Error() string {
	if e.Location != nil {
		return fmt.Sprintf("%s at %s", e.Message, e.Location)
	} else {
		return e.Message
	}
}

func (compiler *Compiler) customError(
	node ast.Node, message string, args ...interface{},
) *CompileError {
	return &CompileError{fmt.Sprintf(message, args...), node.GetLocation()}
}
