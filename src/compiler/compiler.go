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
	"github.com/iafisher/venice/src/vtype"
)

type Compiler struct {
	SymbolTable     *SymbolTable
	TypeSymbolTable *SymbolTable
	functionInfo    *FunctionInfo
	nestedLoopCount int
}

func NewCompiler() *Compiler {
	return &Compiler{
		SymbolTable:     NewBuiltinSymbolTable(),
		TypeSymbolTable: NewBuiltinTypeSymbolTable(),
		functionInfo:    nil,
		nestedLoopCount: 0,
	}
}

func (compiler *Compiler) Compile(file *ast.File) (*bytecode.CompiledProgram, error) {
	compiledProgram := bytecode.NewCompiledProgram()
	for _, statementAny := range file.Statements {
		switch statement := statementAny.(type) {
		case *ast.ClassDeclarationNode:
			err := compiler.compileClassDeclaration(compiledProgram, statement)
			if err != nil {
				return nil, err
			}
		case *ast.EnumDeclarationNode:
			err := compiler.compileEnumDeclaration(statement)
			if err != nil {
				return nil, err
			}
		case *ast.FunctionDeclarationNode:
			code, err := compiler.compileFunctionDeclaration(statement)
			if err != nil {
				return nil, err
			}
			compiledProgram.Code[statement.Name] = code
		default:
			code, err := compiler.compileStatement(statementAny)
			if err != nil {
				return nil, err
			}
			compiledProgram.Code["main"] = append(compiledProgram.Code["main"], code...)
		}
	}
	return compiledProgram, nil
}

func (compiler *Compiler) GetType(expr ast.ExpressionNode) (vtype.VeniceType, error) {
	_, exprType, err := compiler.compileExpression(expr)
	return exprType, err
}

/**
 * Compile statements
 */

func (compiler *Compiler) compileStatement(treeAny ast.StatementNode) ([]bytecode.Bytecode, error) {
	switch node := treeAny.(type) {
	case *ast.AssignStatementNode:
		return compiler.compileAssignStatement(node)
	case *ast.BreakStatementNode:
		if compiler.nestedLoopCount == 0 {
			return nil, compiler.customError(treeAny, "break statement outside of loop")
		}

		// BREAK_LOOP is a temporary bytecode instruction that the compiler will later
		// convert to a REL_JUMP instruction.
		return []bytecode.Bytecode{&bytecode.BreakLoop{}}, nil
	case *ast.ContinueStatementNode:
		if compiler.nestedLoopCount == 0 {
			return nil, compiler.customError(treeAny, "break statement outside of loop")
		}

		// CONTINUE_LOOP is a temporary bytecode instruction that the compiler will later
		// convert to a REL_JUMP instruction.
		return []bytecode.Bytecode{&bytecode.ContinueLoop{}}, nil
	case *ast.ExpressionStatementNode:
		code, _, err := compiler.compileExpression(node.Expr)
		if err != nil {
			return nil, err
		}
		return code, nil
	case *ast.ForLoopNode:
		return compiler.compileForLoop(node)
	case *ast.IfStatementNode:
		return compiler.compileIfStatement(node)
	case *ast.LetStatementNode:
		if _, ok := compiler.SymbolTable.Get(node.Symbol); ok {
			return nil, compiler.customError(treeAny, "re-declaration of symbol `%q`", node.Symbol)
		}

		code, eType, err := compiler.compileExpression(node.Expr)
		if err != nil {
			return nil, err
		}
		compiler.SymbolTable.Put(node.Symbol, eType)
		return append(code, &bytecode.StoreName{node.Symbol}), nil
	case *ast.ReturnStatementNode:
		if compiler.functionInfo == nil {
			return nil, compiler.customError(treeAny, "return statement outside of function definition")
		}

		if node.Expr != nil {
			code, exprType, err := compiler.compileExpression(node.Expr)
			if err != nil {
				return nil, err
			}

			if !compiler.functionInfo.declaredReturnType.Check(exprType) {
				return nil, compiler.customError(treeAny, "conflicting function return types")
			}

			compiler.functionInfo.seenReturnStatement = true
			code = append(code, &bytecode.Return{})
			return code, err
		} else {
			if compiler.functionInfo.declaredReturnType != nil {
				return nil, compiler.customError(treeAny, "function cannot return void")
			}

			return []bytecode.Bytecode{&bytecode.Return{}}, nil
		}
	case *ast.WhileLoopNode:
		return compiler.compileWhileLoop(node)
	default:
		return nil, compiler.customError(treeAny, "unknown statement type: %T", treeAny)
	}
}

func (compiler *Compiler) compileAssignStatement(node *ast.AssignStatementNode) ([]bytecode.Bytecode, error) {
	code, eType, err := compiler.compileExpression(node.Expr)
	if err != nil {
		return nil, err
	}

	switch destination := node.Destination.(type) {
	case *ast.FieldAccessNode:
		destinationCode, destinationTypeAny, err := compiler.compileExpression(destination.Expr)
		if err != nil {
			return nil, err
		}

		classType, ok := destinationTypeAny.(*vtype.VeniceClassType)
		if !ok {
			return nil, compiler.customError(node, "cannot assign to field on type %s", destinationTypeAny.String())
		}

		for i, field := range classType.Fields {
			if field.Name == destination.Name {
				if !field.Public {
					return nil, compiler.customError(node, "cannot assign to non-public field")
				}

				code = append(code, destinationCode...)
				code = append(code, &bytecode.StoreField{i})
				return code, nil
			}
		}

		return nil, compiler.customError(node, "field `%s` does not exist on %s", destination.Name, classType.String())
	case *ast.SymbolNode:
		expectedType, ok := compiler.SymbolTable.Get(destination.Value)
		if !ok {
			return nil, compiler.customError(node, "cannot assign to undeclared symbol `%s`", destination.Value)
		} else if !expectedType.Check(eType) {
			return nil, compiler.customError(node, "wrong expression type in assignment to `%s`", destination.Value)
		}

		code = append(code, &bytecode.StoreName{destination.Value})
		return code, nil
	default:
		return nil, compiler.customError(destination, "cannot assign to non-symbol")
	}
}

func (compiler *Compiler) compileClassDeclaration(compiledProgram *bytecode.CompiledProgram, node *ast.ClassDeclarationNode) error {
	if node.GenericTypeParameter != "" {
		subTypeSymbolTable := &SymbolTable{
			compiler.TypeSymbolTable,
			map[string]vtype.VeniceType{
				node.GenericTypeParameter: &vtype.VeniceGenericParameterType{node.GenericTypeParameter},
			},
		}
		compiler.TypeSymbolTable = subTypeSymbolTable
	}

	fields := []*vtype.VeniceClassField{}
	paramTypes := []vtype.VeniceType{}
	for _, field := range node.Fields {
		paramType, err := compiler.resolveType(field.FieldType)
		if err != nil {
			return err
		}
		paramTypes = append(paramTypes, paramType)
		fields = append(fields, &vtype.VeniceClassField{field.Name, field.Public, paramType})
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
		classType = &vtype.VeniceClassType{Name: node.Name, Fields: fields, Methods: methods}
	} else {
		classType = &vtype.VeniceGenericType{
			[]string{node.GenericTypeParameter},
			&vtype.VeniceClassType{Name: node.Name, Fields: fields, Methods: methods},
		}
		compiler.TypeSymbolTable = compiler.TypeSymbolTable.Parent
	}
	compiler.TypeSymbolTable.Put(node.Name, classType)
	constructorType := &vtype.VeniceFunctionType{
		Name:       node.Name,
		ParamTypes: paramTypes,
		ReturnType: classType,
		IsBuiltin:  false,
	}
	compiler.SymbolTable.Put(node.Name, constructorType)

	for i := 0; i < len(methods); i++ {
		method := node.Methods[i]
		methodType := methods[i]
		bodySymbolTableMap := map[string]vtype.VeniceType{}
		for j := 0; j < len(method.Params); j++ {
			bodySymbolTableMap[method.Params[j].Name] = methodType.ParamTypes[j]
		}
		bodySymbolTable := &SymbolTable{compiler.SymbolTable, bodySymbolTableMap}

		declaredReturnType := methodType.ReturnType

		// Put `self` in the symbol table.
		bodySymbolTable.Put("self", classType)

		compiler.SymbolTable = bodySymbolTable
		compiler.functionInfo = &FunctionInfo{
			declaredReturnType:  declaredReturnType,
			seenReturnStatement: false,
		}
		bodyCode, err := compiler.compileBlock(method.Body)

		if declaredReturnType != nil && !compiler.functionInfo.seenReturnStatement {
			return compiler.customError(node, "non-void function has no return statement")
		}

		compiler.functionInfo = nil
		compiler.SymbolTable = bodySymbolTable.Parent

		paramLoadCode := []bytecode.Bytecode{
			&bytecode.StoreName{"self"},
		}
		for i := len(method.Params) - 1; i >= 0; i-- {
			param := method.Params[i]
			paramLoadCode = append(paramLoadCode, &bytecode.StoreName{param.Name})
		}
		bodyCode = append(paramLoadCode, bodyCode...)

		if err != nil {
			return err
		}

		compiledProgram.Code[fmt.Sprintf("%s__%s", node.Name, method.Name)] = bodyCode
	}

	constructorBytecode := []bytecode.Bytecode{&bytecode.BuildClass{node.Name, len(fields)}}
	compiledProgram.Code[node.Name] = constructorBytecode
	return nil
}

func (compiler *Compiler) compileEnumDeclaration(node *ast.EnumDeclarationNode) error {
	if node.GenericTypeParameter != "" {
		subTypeSymbolTable := &SymbolTable{
			compiler.TypeSymbolTable,
			map[string]vtype.VeniceType{
				node.GenericTypeParameter: &vtype.VeniceGenericParameterType{node.GenericTypeParameter},
			},
		}
		compiler.TypeSymbolTable = subTypeSymbolTable
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
		compiler.TypeSymbolTable = compiler.TypeSymbolTable.Parent
		compiler.TypeSymbolTable.Put(
			node.Name,
			&vtype.VeniceGenericType{
				[]string{node.GenericTypeParameter},
				&vtype.VeniceEnumType{Name: node.Name, Cases: caseTypes},
			},
		)
	} else {
		compiler.TypeSymbolTable.Put(node.Name, &vtype.VeniceEnumType{Name: node.Name, Cases: caseTypes})
	}

	return nil
}

func (compiler *Compiler) compileForLoop(node *ast.ForLoopNode) ([]bytecode.Bytecode, error) {
	iterableCode, iterableTypeAny, err := compiler.compileExpression(node.Iterable)
	if err != nil {
		return nil, err
	}

	var symbolTable *SymbolTable
	switch iterableType := iterableTypeAny.(type) {
	case *vtype.VeniceListType:
		if len(node.Variables) != 1 {
			return nil, compiler.customError(node, "too many for loop variables")
		}

		symbolTable = &SymbolTable{
			Parent:  compiler.SymbolTable,
			Symbols: map[string]vtype.VeniceType{node.Variables[0]: iterableType.ItemType},
		}
	case *vtype.VeniceMapType:
		if len(node.Variables) != 2 {
			return nil, compiler.customError(node, "too many for loop variables")
		}

		symbolTable = &SymbolTable{
			Parent: compiler.SymbolTable,
			Symbols: map[string]vtype.VeniceType{
				node.Variables[0]: iterableType.KeyType,
				node.Variables[1]: iterableType.ValueType,
			},
		}
	default:
		return nil, compiler.customError(node.Iterable, "for loop must be of list")
	}

	compiler.SymbolTable = symbolTable
	compiler.nestedLoopCount += 1
	bodyCode, err := compiler.compileBlock(node.Body)
	compiler.nestedLoopCount -= 1
	compiler.SymbolTable = compiler.SymbolTable.Parent

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

func (compiler *Compiler) compileFunctionDeclaration(node *ast.FunctionDeclarationNode) ([]bytecode.Bytecode, error) {
	params := []string{}
	paramTypes := []vtype.VeniceType{}
	bodySymbolTableMap := map[string]vtype.VeniceType{}
	for _, param := range node.Params {
		paramType, err := compiler.resolveType(param.ParamType)
		if err != nil {
			return nil, err
		}

		params = append(params, param.Name)
		paramTypes = append(paramTypes, paramType)

		bodySymbolTableMap[param.Name] = paramType
	}
	bodySymbolTable := &SymbolTable{compiler.SymbolTable, bodySymbolTableMap}

	declaredReturnType, err := compiler.resolveType(node.ReturnType)
	if err != nil {
		return nil, err
	}

	// Put the function's entry in the symbol table before compiling the body so that
	// recursive functions can call themselves.
	compiler.SymbolTable.Put(
		node.Name,
		&vtype.VeniceFunctionType{
			Name:       node.Name,
			ParamTypes: paramTypes,
			ReturnType: declaredReturnType,
			IsBuiltin:  false,
		},
	)

	compiler.SymbolTable = bodySymbolTable
	compiler.functionInfo = &FunctionInfo{
		declaredReturnType:  declaredReturnType,
		seenReturnStatement: false,
	}
	bodyCode, err := compiler.compileBlock(node.Body)

	if declaredReturnType != nil && !compiler.functionInfo.seenReturnStatement {
		return nil, compiler.customError(node, "non-void function has no return statement")
	}

	compiler.functionInfo = nil
	compiler.SymbolTable = bodySymbolTable.Parent

	paramLoadCode := []bytecode.Bytecode{}
	for i := len(node.Params) - 1; i >= 0; i-- {
		param := node.Params[i]
		paramLoadCode = append(paramLoadCode, &bytecode.StoreName{param.Name})
	}
	bodyCode = append(paramLoadCode, bodyCode...)

	if err != nil {
		return nil, err
	}

	return bodyCode, nil
}

func (compiler *Compiler) compileIfStatement(node *ast.IfStatementNode) ([]bytecode.Bytecode, error) {
	if len(node.Clauses) == 0 {
		return nil, compiler.customError(node, "`if` statement with no clauses")
	}

	code := []bytecode.Bytecode{}
	for _, clause := range node.Clauses {
		conditionCode, conditionType, err := compiler.compileExpression(clause.Condition)
		if err != nil {
			return nil, err
		}

		if !vtype.VENICE_TYPE_BOOLEAN.Check(conditionType) {
			return nil, compiler.customError(clause.Condition, "condition of `if` statement must be a boolean")
		}

		bodyCode, err := compiler.compileBlock(clause.Body)
		if err != nil {
			return nil, err
		}

		code = append(code, conditionCode...)
		code = append(code, &bytecode.RelJumpIfFalse{len(bodyCode) + 2})
		code = append(code, bodyCode...)
		// The relative jump values will be filled in later.
		code = append(code, &bytecode.RelJump{0})
	}

	if node.ElseClause != nil {
		elseClauseCode, err := compiler.compileBlock(node.ElseClause)
		if err != nil {
			return nil, err
		}

		code = append(code, elseClauseCode...)
	}

	// Fill in the relative jump values now that we have generated all the code.
	for i, bcodeAny := range code {
		if bcode, ok := bcodeAny.(*bytecode.RelJump); ok {
			bcode.N = len(code) - i
		}
	}

	return code, nil
}

func (compiler *Compiler) compileWhileLoop(node *ast.WhileLoopNode) ([]bytecode.Bytecode, error) {
	conditionCode, conditionType, err := compiler.compileExpression(node.Condition)
	if err != nil {
		return nil, err
	}

	if !vtype.VENICE_TYPE_BOOLEAN.Check(conditionType) {
		return nil, compiler.customError(node.Condition, "condition of `while` loop must be a boolean")
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

	for i, bytecodeAny := range code {
		switch bytecodeAny.(type) {
		case *bytecode.BreakLoop:
			code[i] = &bytecode.RelJump{len(code) - i}
		case *bytecode.ContinueLoop:
			code[i] = &bytecode.RelJump{-i}
		}
	}

	return code, nil
}

func (compiler *Compiler) compileBlock(block []ast.StatementNode) ([]bytecode.Bytecode, error) {
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

func (compiler *Compiler) compileExpression(nodeAny ast.ExpressionNode) ([]bytecode.Bytecode, vtype.VeniceType, error) {
	switch node := nodeAny.(type) {
	case *ast.BooleanNode:
		return []bytecode.Bytecode{&bytecode.PushConstBool{node.Value}}, vtype.VENICE_TYPE_BOOLEAN, nil
	case *ast.CallNode:
		return compiler.compileCallNode(node)
	case *ast.CharacterNode:
		return []bytecode.Bytecode{&bytecode.PushConstChar{node.Value}}, vtype.VENICE_TYPE_CHARACTER, nil
	case *ast.EnumSymbolNode:
		return compiler.compileEnumSymbolNode(node)
	case *ast.FieldAccessNode:
		return compiler.compileFieldAccessNode(node)
	case *ast.IndexNode:
		return compiler.compileIndexNode(node)
	case *ast.InfixNode:
		return compiler.compileInfixNode(node)
	case *ast.IntegerNode:
		return []bytecode.Bytecode{&bytecode.PushConstInt{node.Value}}, vtype.VENICE_TYPE_INTEGER, nil
	case *ast.ListNode:
		return compiler.compileListNode(node)
	case *ast.MapNode:
		return compiler.compileMapNode(node)
	case *ast.StringNode:
		return []bytecode.Bytecode{&bytecode.PushConstStr{node.Value}}, vtype.VENICE_TYPE_STRING, nil
	case *ast.SymbolNode:
		symbolType, ok := compiler.SymbolTable.Get(node.Value)
		if !ok {
			return nil, nil, compiler.customError(nodeAny, "undefined symbol: %s", node.Value)
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
		return nil, nil, compiler.customError(nodeAny, "unknown expression type: %T", nodeAny)
	}
}

func (compiler *Compiler) compileCallNode(node *ast.CallNode) ([]bytecode.Bytecode, vtype.VeniceType, error) {
	if nodeAsEnumSymbol, ok := node.Function.(*ast.EnumSymbolNode); ok {
		return compiler.compileEnumCallNode(nodeAsEnumSymbol, node)
	}

	functionCode, functionTypeAny, err := compiler.compileExpression(node.Function)
	if err != nil {
		return nil, nil, err
	}

	functionType, ok := functionTypeAny.(*vtype.VeniceFunctionType)
	if !ok {
		return nil, nil, compiler.customError(node.Function, "type %s cannot be used as a function", functionTypeAny.String())
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

	if isClassMethod {
		code = append(code, functionCode...)
		code = append(code, &bytecode.LookupMethod{functionType.Name})
	} else {
		code = append(code, &bytecode.PushConstStr{functionType.Name})
	}

	if functionType.IsBuiltin {
		code = append(code, &bytecode.CallBuiltin{len(functionType.ParamTypes)})
	} else {
		code = append(code, &bytecode.CallFunction{len(functionType.ParamTypes)})
	}

	return code, returnType, nil
}

func (compiler *Compiler) compileEnumCallNode(enumSymbolNode *ast.EnumSymbolNode, callNode *ast.CallNode) ([]bytecode.Bytecode, vtype.VeniceType, error) {
	enumTypeAny, err := compiler.resolveType(&ast.SimpleTypeNode{enumSymbolNode.Enum, nil})
	if err != nil {
		return nil, nil, err
	}

	isEnum := false
	enumType, ok := enumTypeAny.(*vtype.VeniceEnumType)
	if ok {
		isEnum = true
	} else {
		genericType, ok := enumTypeAny.(*vtype.VeniceGenericType)
		if ok {
			enumType, ok = genericType.GenericType.(*vtype.VeniceEnumType)
			isEnum = ok
		} else {
			isEnum = false
		}
	}

	if !isEnum {
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

	return nil, nil, compiler.customError(callNode, "enum %s does not have case %s", enumSymbolNode.Enum, enumSymbolNode.Case)
}

func (compiler *Compiler) compileEnumSymbolNode(node *ast.EnumSymbolNode) ([]bytecode.Bytecode, vtype.VeniceType, error) {
	enumTypeAny, err := compiler.resolveType(&ast.SimpleTypeNode{node.Enum, nil})
	if err != nil {
		return nil, nil, err
	}

	isEnum := false
	enumType, ok := enumTypeAny.(*vtype.VeniceEnumType)
	if ok {
		isEnum = true
	} else {
		genericType, ok := enumTypeAny.(*vtype.VeniceGenericType)
		if ok {
			enumType, ok = genericType.GenericType.(*vtype.VeniceEnumType)
			isEnum = ok
		} else {
			isEnum = false
		}
	}

	if !isEnum {
		return nil, nil, compiler.customError(node, "cannot use double colon after non-enum type")
	}

	for _, enumCase := range enumType.Cases {
		if enumCase.Label == node.Case {
			if len(enumCase.Types) != 0 {
				return nil, nil, compiler.customError(node, "%s takes %d argument(s)", enumCase.Label, len(enumCase.Types))
			}

			return []bytecode.Bytecode{&bytecode.PushEnum{node.Case, 0}}, enumType, nil
		}
	}

	return nil, nil, compiler.customError(node, "enum %s does not have case %s", node.Enum, node.Case)
}

func (compiler *Compiler) compileFieldAccessNode(node *ast.FieldAccessNode) ([]bytecode.Bytecode, vtype.VeniceType, error) {
	code, typeAny, err := compiler.compileExpression(node.Expr)
	if err != nil {
		return nil, nil, err
	}

	switch concreteType := typeAny.(type) {
	case *vtype.VeniceAtomicType:
		if concreteType == vtype.VENICE_TYPE_STRING {
			methodType, ok := stringBuiltins[node.Name]
			if !ok {
				return nil, nil, compiler.customError(node, "no such field or method `%s` on string type", node.Name)
			}
			return code, methodType, nil
		}
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
				return code, methodType, nil
			}
		}

		return nil, nil, compiler.customError(node, "no such field or method `%s`", node.Name)
	case *vtype.VeniceListType:
		methodType, ok := listBuiltins[node.Name]
		if !ok {
			return nil, nil, compiler.customError(node, "no such field or method `%s` on list type", node.Name)
		}
		return code, methodType, nil
	}

	return nil, nil, compiler.customError(node, "no such field or method `%s` on type %s", node.Name, typeAny.String())
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

	code := []bytecode.Bytecode{}
	genericParameterMap := map[string]vtype.VeniceType{}
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

		err = compiler.checkFunctionArgType(args[i], paramType, argType, genericParameterMap)
		if err != nil {
			return nil, nil, err
		}

		code = append(code, argCode...)
	}

	if len(genericParameterMap) > 0 {
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
	if genericParamType, ok := paramType.(*vtype.VeniceGenericParameterType); ok {
		genericParameterMap[genericParamType.Label] = argType
	} else {
		if !paramType.Check(argType) {
			return compiler.customError(node, "wrong function parameter type")
		}
	}

	return nil
}

func (compiler *Compiler) compileIndexNode(node *ast.IndexNode) ([]bytecode.Bytecode, vtype.VeniceType, error) {
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
	case *vtype.VeniceAtomicType:
		if exprType != vtype.VENICE_TYPE_STRING {
			return nil, nil, compiler.customError(node, "%s cannot be indexed", exprType.String())
		}
		if !vtype.VENICE_TYPE_INTEGER.Check(indexType) {
			return nil, nil, compiler.customError(node.Expr, "string index must be integer")
		}

		code = append(code, &bytecode.BinaryStringIndex{})
		return code, vtype.VENICE_TYPE_CHARACTER, nil
	case *vtype.VeniceListType:
		if !vtype.VENICE_TYPE_INTEGER.Check(indexType) {
			return nil, nil, compiler.customError(node.Expr, "list index must be integer")
		}

		code = append(code, &bytecode.BinaryListIndex{})
		return code, exprType.ItemType, nil
	case *vtype.VeniceMapType:
		if !exprType.KeyType.Check(indexType) {
			return nil, nil, compiler.customError(node.Expr, "wrong map key type in index expression")
		}

		code = append(code, &bytecode.BinaryMapIndex{})
		return code, exprType.KeyType, nil
	default:
		return nil, nil, compiler.customError(node, "%s cannot be indexed", exprTypeAny)
	}
}

func (compiler *Compiler) compileInfixNode(node *ast.InfixNode) ([]bytecode.Bytecode, vtype.VeniceType, error) {
	// TODO(2021-08-07): Boolean operators should short-circuit.
	opBytecode, ok := opsToBytecodes[node.Operator]
	if !ok {
		return nil, nil, compiler.customError(node, "unknown operator `%s`", node.Operator)
	}

	leftCode, leftType, err := compiler.compileExpression(node.Left)
	if err != nil {
		return nil, nil, err
	}

	if !checkInfixLeftType(node.Operator, leftType) {
		return nil, nil, compiler.customError(node.Left, "invalid type for left operand of %s", node.Operator)
	}

	rightCode, rightType, err := compiler.compileExpression(node.Right)
	if err != nil {
		return nil, nil, err
	}

	resultType, ok := checkInfixRightType(node.Operator, leftType, rightType)
	if !ok {
		return nil, nil, compiler.customError(node.Right, "invalid type for right operand of %s", node.Operator)
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

func (compiler *Compiler) compileListNode(node *ast.ListNode) ([]bytecode.Bytecode, vtype.VeniceType, error) {
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
		} else if !itemType.Check(valueType) {
			return nil, nil, compiler.customError(value, "list elements must all be of same type")
		}

		code = append(code, valueCode...)
	}
	code = append(code, &bytecode.BuildList{len(node.Values)})
	return code, &vtype.VeniceListType{itemType}, nil
}

func (compiler *Compiler) compileMapNode(node *ast.MapNode) ([]bytecode.Bytecode, vtype.VeniceType, error) {
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
		} else if !keyType.Check(thisKeyType) {
			return nil, nil, compiler.customError(pair.Key, "map keys must all be of the same type")
		}

		code = append(code, keyCode...)

		valueCode, thisValueType, err := compiler.compileExpression(pair.Value)
		if err != nil {
			return nil, nil, err
		}

		if valueType == nil {
			valueType = thisValueType
		} else if !valueType.Check(thisValueType) {
			return nil, nil, compiler.customError(pair.Value, "map values must all be of the same type")
		}

		code = append(code, valueCode...)
	}
	code = append(code, &bytecode.BuildMap{len(node.Pairs)})
	return code, &vtype.VeniceMapType{keyType, valueType}, nil
}

func (compiler *Compiler) compileTernaryIfNode(node *ast.TernaryIfNode) ([]bytecode.Bytecode, vtype.VeniceType, error) {
	conditionCode, conditionType, err := compiler.compileExpression(node.Condition)
	if err != nil {
		return nil, nil, err
	}

	if !vtype.VENICE_TYPE_BOOLEAN.Check(conditionType) {
		return nil, nil, compiler.customError(node, "condition of `if` expression must be a boolean")
	}

	trueClauseCode, trueClauseType, err := compiler.compileExpression(node.TrueClause)
	if err != nil {
		return nil, nil, err
	}

	falseClauseCode, falseClauseType, err := compiler.compileExpression(node.FalseClause)
	if err != nil {
		return nil, nil, err
	}

	if !trueClauseType.Check(falseClauseType) {
		return nil, nil, compiler.customError(node, "branches of `if` expression are of different types")
	}

	code := conditionCode
	code = append(code, &bytecode.RelJumpIfFalse{len(trueClauseCode) + 2})
	code = append(code, trueClauseCode...)
	code = append(code, &bytecode.RelJump{len(falseClauseCode) + 1})
	code = append(code, falseClauseCode...)
	return code, trueClauseType, nil
}

func (compiler *Compiler) compileTupleFieldAccessNode(node *ast.TupleFieldAccessNode) ([]bytecode.Bytecode, vtype.VeniceType, error) {
	code, typeAny, err := compiler.compileExpression(node.Expr)
	if err != nil {
		return nil, nil, err
	}

	tupleType, ok := typeAny.(*vtype.VeniceTupleType)
	if !ok {
		return nil, nil, compiler.customError(node, "left-hand side of dot must be a tuple object")
	}

	if node.Index < 0 || node.Index >= len(tupleType.ItemTypes) {
		return nil, nil, compiler.customError(node, "tuple index out of bounds")
	}

	code = append(code, &bytecode.PushTupleField{node.Index})
	return code, tupleType.ItemTypes[node.Index], nil
}

func (compiler *Compiler) compileTupleNode(node *ast.TupleNode) ([]bytecode.Bytecode, vtype.VeniceType, error) {
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

func (compiler *Compiler) compileUnaryNode(node *ast.UnaryNode) ([]bytecode.Bytecode, vtype.VeniceType, error) {
	switch node.Operator {
	case "-":
		code, exprType, err := compiler.compileExpression(node.Expr)
		if err != nil {
			return nil, nil, err
		}

		if exprType != vtype.VENICE_TYPE_INTEGER {
			return nil, nil, compiler.customError(node, "argument to unary minus must be integer, not %s", exprType.String())
		}

		code = append(code, &bytecode.UnaryMinus{})
		return code, vtype.VENICE_TYPE_INTEGER, nil
	case "not":
		code, exprType, err := compiler.compileExpression(node.Expr)
		if err != nil {
			return nil, nil, err
		}

		if exprType != vtype.VENICE_TYPE_BOOLEAN {
			return nil, nil, compiler.customError(node, "argument to `not` must be boolean, not %s", exprType.String())
		}

		code = append(code, &bytecode.UnaryNot{})
		return code, vtype.VENICE_TYPE_BOOLEAN, nil
	default:
		return nil, nil, compiler.customError(node, "unknown unary operator `%s`", node.Operator)
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

func checkInfixLeftType(operator string, leftType vtype.VeniceType) bool {
	switch operator {
	case "==", "in":
		return true
	case "and", "or":
		return vtype.VENICE_TYPE_BOOLEAN.Check(leftType)
	case "++":
		if _, ok := leftType.(*vtype.VeniceListType); ok {
			return true
		} else {
			return vtype.VENICE_TYPE_STRING.Check(leftType)
		}
	default:
		return vtype.VENICE_TYPE_INTEGER.Check(leftType)
	}
}

func checkInfixRightType(operator string, leftType vtype.VeniceType, rightType vtype.VeniceType) (vtype.VeniceType, bool) {
	switch operator {
	case "==", "!=":
		return vtype.VENICE_TYPE_BOOLEAN, leftType.Check(rightType)
	case "and", "or":
		return vtype.VENICE_TYPE_BOOLEAN, vtype.VENICE_TYPE_BOOLEAN.Check(rightType)
	case ">", ">=", "<", "<=":
		return vtype.VENICE_TYPE_BOOLEAN, vtype.VENICE_TYPE_INTEGER.Check(rightType)
	case "++":
		if vtype.VENICE_TYPE_STRING.Check(leftType) {
			return vtype.VENICE_TYPE_STRING, vtype.VENICE_TYPE_STRING.Check(rightType)
		} else {
			return leftType, leftType.Check(rightType)
		}
	case "in":
		switch rightConcreteType := rightType.(type) {
		case *vtype.VeniceAtomicType:
			if rightConcreteType == vtype.VENICE_TYPE_STRING {
				return vtype.VENICE_TYPE_BOOLEAN, vtype.VENICE_TYPE_CHARACTER.Check(leftType) || vtype.VENICE_TYPE_STRING.Check(leftType)

			} else {
				return nil, false
			}
		case *vtype.VeniceListType:
			return vtype.VENICE_TYPE_BOOLEAN, rightConcreteType.ItemType.Check(leftType)
		case *vtype.VeniceMapType:
			return vtype.VENICE_TYPE_BOOLEAN, rightConcreteType.KeyType.Check(leftType)
		default:
			return nil, false
		}
	default:
		return vtype.VENICE_TYPE_INTEGER, vtype.VENICE_TYPE_INTEGER.Check(rightType)
	}
}

func (compiler *Compiler) resolveType(typeNodeAny ast.TypeNode) (vtype.VeniceType, error) {
	if typeNodeAny == nil {
		return nil, nil
	}

	switch typeNode := typeNodeAny.(type) {
	case *ast.SimpleTypeNode:
		resolvedType, ok := compiler.TypeSymbolTable.Get(typeNode.Symbol)
		if !ok {
			return nil, compiler.customError(typeNodeAny, "unknown type `%s`", typeNode.Symbol)
		}
		return resolvedType, nil
	default:
		return nil, compiler.customError(typeNodeAny, "unknown type node: %T", typeNodeAny)
	}
}

var listBuiltins = map[string]vtype.VeniceType{
	"length": &vtype.VeniceFunctionType{
		Name: "length",
		ParamTypes: []vtype.VeniceType{
			&vtype.VeniceListType{vtype.VENICE_TYPE_ANY},
		},
		ReturnType: vtype.VENICE_TYPE_INTEGER,
		IsBuiltin:  true,
	},
}

var stringBuiltins = map[string]vtype.VeniceType{
	"find": &vtype.VeniceFunctionType{
		Name:       "find",
		ParamTypes: []vtype.VeniceType{vtype.VENICE_TYPE_STRING, vtype.VENICE_TYPE_CHARACTER},
		ReturnType: vtype.VeniceOptionalTypeOf(vtype.VENICE_TYPE_INTEGER),
		IsBuiltin:  true,
	},
	"length": &vtype.VeniceFunctionType{
		Name:       "length",
		ParamTypes: []vtype.VeniceType{vtype.VENICE_TYPE_STRING},
		ReturnType: vtype.VENICE_TYPE_INTEGER,
		IsBuiltin:  true,
	},
	"to_lower": &vtype.VeniceFunctionType{
		Name:       "to_lower",
		ParamTypes: []vtype.VeniceType{vtype.VENICE_TYPE_STRING},
		ReturnType: vtype.VENICE_TYPE_STRING,
		IsBuiltin:  true,
	},
	"to_upper": &vtype.VeniceFunctionType{
		Name:       "to_upper",
		ParamTypes: []vtype.VeniceType{vtype.VENICE_TYPE_STRING},
		ReturnType: vtype.VENICE_TYPE_STRING,
		IsBuiltin:  true,
	},
}

/**
 * Symbol table methods
 */

type FunctionInfo struct {
	declaredReturnType  vtype.VeniceType
	seenReturnStatement bool
}

type SymbolTable struct {
	Parent  *SymbolTable
	Symbols map[string]vtype.VeniceType
}

func NewBuiltinSymbolTable() *SymbolTable {
	symbols := map[string]vtype.VeniceType{
		"length": &vtype.VeniceFunctionType{
			Name: "length",
			ParamTypes: []vtype.VeniceType{
				&vtype.VeniceUnionType{
					[]vtype.VeniceType{
						vtype.VENICE_TYPE_STRING,
						&vtype.VeniceListType{vtype.VENICE_TYPE_ANY},
						&vtype.VeniceMapType{vtype.VENICE_TYPE_ANY, vtype.VENICE_TYPE_ANY},
					},
				},
			},
			ReturnType: vtype.VENICE_TYPE_INTEGER,
			IsBuiltin:  true,
		},
		"print": &vtype.VeniceFunctionType{
			Name:       "print",
			ParamTypes: []vtype.VeniceType{vtype.VENICE_TYPE_ANY},
			ReturnType: nil,
			IsBuiltin:  true,
		},
	}
	return &SymbolTable{nil, symbols}
}

func NewBuiltinTypeSymbolTable() *SymbolTable {
	symbols := map[string]vtype.VeniceType{
		"bool":     vtype.VENICE_TYPE_BOOLEAN,
		"int":      vtype.VENICE_TYPE_INTEGER,
		"string":   vtype.VENICE_TYPE_STRING,
		"Optional": vtype.VENICE_TYPE_OPTIONAL,
	}
	return &SymbolTable{nil, symbols}
}

func (symtab *SymbolTable) Get(symbol string) (vtype.VeniceType, bool) {
	value, ok := symtab.Symbols[symbol]
	if !ok {
		if symtab.Parent != nil {
			return symtab.Parent.Get(symbol)
		} else {
			return nil, false
		}
	}
	return value, true
}

func (symtab *SymbolTable) Put(symbol string, value vtype.VeniceType) {
	symtab.Symbols[symbol] = value
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

func (compiler *Compiler) customError(node ast.Node, message string, args ...interface{}) *CompileError {
	return &CompileError{fmt.Sprintf(message, args...), node.GetLocation()}
}
