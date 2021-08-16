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

type FunctionInfo struct {
	declaredReturnType  vtype.VeniceType
	seenReturnStatement bool
}

type SymbolTable struct {
	Parent  *SymbolTable
	Symbols map[string]vtype.VeniceType
}

func NewBuiltinSymbolTable() *SymbolTable {
	symbols := map[string]vtype.VeniceType{}
	return &SymbolTable{nil, symbols}
}

func NewBuiltinTypeSymbolTable() *SymbolTable {
	symbols := map[string]vtype.VeniceType{
		"bool":   vtype.VENICE_TYPE_BOOLEAN,
		"int":    vtype.VENICE_TYPE_INTEGER,
		"string": vtype.VENICE_TYPE_STRING,
		"Optional": &vtype.VeniceGenericType{
			[]string{"T"},
			&vtype.VeniceEnumType{
				[]*vtype.VeniceCaseType{
					&vtype.VeniceCaseType{
						"Some",
						[]vtype.VeniceType{&vtype.VeniceGenericParameterType{"T"}},
					},
					&vtype.VeniceCaseType{"None", nil},
				},
			},
		},
	}
	return &SymbolTable{nil, symbols}
}

func (compiler *Compiler) Compile(file *ast.File) (*bytecode.CompiledProgram, error) {
	compiledProgram := bytecode.NewCompiledProgram()
	for _, statementAny := range file.Statements {
		switch statement := statementAny.(type) {
		case *ast.ClassDeclarationNode:
			code, err := compiler.compileClassDeclaration(statement)
			if err != nil {
				return nil, err
			}
			compiledProgram.Code[statement.Name] = code
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

/**
 * Compile statements
 */

func (compiler *Compiler) compileStatement(treeAny ast.StatementNode) ([]bytecode.Bytecode, error) {
	switch node := treeAny.(type) {
	case *ast.AssignStatementNode:
		code, eType, err := compiler.compileExpression(node.Expr)
		if err != nil {
			return nil, err
		}

		expectedType, ok := compiler.SymbolTable.Get(node.Symbol)
		if !ok {
			return nil, compiler.customError(treeAny, fmt.Sprintf("cannot assign to undeclared symbol %q", node.Symbol))
		} else if !areTypesCompatible(expectedType, eType) {
			return nil, compiler.customError(treeAny, fmt.Sprintf("wrong expression type in assignment to %q", node.Symbol))
		}

		code = append(code, &bytecode.StoreName{node.Symbol})
		return code, nil
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
			return nil, compiler.customError(treeAny, fmt.Sprintf("re-declaration of symbol %q", node.Symbol))
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

			if !areTypesCompatible(compiler.functionInfo.declaredReturnType, exprType) {
				return nil, compiler.customError(treeAny, "conflicting function return types")
			}

			compiler.functionInfo.seenReturnStatement = true
			code = append(code, &bytecode.Return{})
			return code, err
		} else {
			if !areTypesCompatible(compiler.functionInfo.declaredReturnType, nil) {
				return nil, compiler.customError(treeAny, "function cannot return void")
			}

			return []bytecode.Bytecode{&bytecode.Return{}}, nil
		}
	case *ast.WhileLoopNode:
		return compiler.compileWhileLoop(node)
	default:
		return nil, compiler.customError(treeAny, fmt.Sprintf("unknown statement type: %T", treeAny))
	}
}

func (compiler *Compiler) compileClassDeclaration(node *ast.ClassDeclarationNode) ([]bytecode.Bytecode, error) {
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
			return nil, err
		}
		paramTypes = append(paramTypes, paramType)
		fields = append(fields, &vtype.VeniceClassField{field.Name, field.Public, paramType})
	}

	var classType vtype.VeniceType
	if node.GenericTypeParameter == "" {
		classType = &vtype.VeniceClassType{fields}
	} else {
		classType = &vtype.VeniceGenericType{[]string{node.GenericTypeParameter}, &vtype.VeniceClassType{fields}}
		compiler.TypeSymbolTable = compiler.TypeSymbolTable.Parent
	}
	compiler.TypeSymbolTable.Put(node.Name, classType)

	constructorType := &vtype.VeniceFunctionType{paramTypes, classType}
	compiler.SymbolTable.Put(node.Name, constructorType)

	constructorBytecode := []bytecode.Bytecode{&bytecode.BuildClass{len(fields)}}
	return constructorBytecode, nil
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
				&vtype.VeniceEnumType{caseTypes},
			},
		)
	} else {
		compiler.TypeSymbolTable.Put(node.Name, &vtype.VeniceEnumType{caseTypes})
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
	compiler.SymbolTable.Put(node.Name, &vtype.VeniceFunctionType{paramTypes, declaredReturnType})

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
	for _, param := range node.Params {
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

		if !areTypesCompatible(vtype.VENICE_TYPE_BOOLEAN, conditionType) {
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

	if !areTypesCompatible(vtype.VENICE_TYPE_BOOLEAN, conditionType) {
		return nil, compiler.customError(node.Condition, "condition of while loop must be a boolean")
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
			return nil, nil, compiler.customError(nodeAny, fmt.Sprintf("undefined symbol: %s", node.Value))
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
		return nil, nil, compiler.customError(nodeAny, fmt.Sprintf("unknown expression type: %T", nodeAny))
	}
}

func (compiler *Compiler) compileCallNode(node *ast.CallNode) ([]bytecode.Bytecode, vtype.VeniceType, error) {
	if v, ok := node.Function.(*ast.SymbolNode); ok {
		if v.Value == "print" {
			if len(node.Args) != 1 {
				return nil, nil, compiler.customError(node, "`print` takes exactly 1 argument")
			}

			code, _, err := compiler.compileExpression(node.Args[0])
			if err != nil {
				return nil, nil, err
			}

			code = append(code, &bytecode.CallBuiltin{"print"})
			return code, nil, nil
		} else if v.Value == "length" {
			if len(node.Args) != 1 {
				return nil, nil, compiler.customError(node, "`length` takes exactly 1 argument")
			}

			code, argTypeAny, err := compiler.compileExpression(node.Args[0])
			if err != nil {
				return nil, nil, err
			}

			switch argType := argTypeAny.(type) {
			case *vtype.VeniceAtomicType:
				if argType != vtype.VENICE_TYPE_STRING {
					return nil, nil, compiler.customError(node.Args[0], "argument of `length` must be string, list, or map")
				}
			case *vtype.VeniceListType:
			case *vtype.VeniceMapType:
			default:
				return nil, nil, compiler.customError(node.Args[0], "argument of `length` must be string, list, or map")
			}

			code = append(code, &bytecode.CallBuiltin{"length"})
			return code, vtype.VENICE_TYPE_INTEGER, nil
		} else {
			valueAny, ok := compiler.SymbolTable.Get(v.Value)
			if !ok {
				return nil, nil, compiler.customError(node, fmt.Sprintf("undefined symbol: %s", v.Value))
			}

			if f, ok := valueAny.(*vtype.VeniceFunctionType); ok {
				if len(f.ParamTypes) != len(node.Args) {
					return nil, nil, compiler.customError(node, fmt.Sprintf("wrong number of arguments: expected %d, got %d", len(f.ParamTypes), len(node.Args)))
				}

				code := []bytecode.Bytecode{}
				genericParameters := []string{}
				concreteTypes := []vtype.VeniceType{}
				for i := 0; i < len(f.ParamTypes); i++ {
					argCode, argType, err := compiler.compileExpression(node.Args[i])
					if err != nil {
						return nil, nil, err
					}

					if genericParamType, ok := f.ParamTypes[i].(*vtype.VeniceGenericParameterType); ok {
						genericParameters = append(genericParameters, genericParamType.Label)
						concreteTypes = append(concreteTypes, argType)
					} else {
						if !areTypesCompatible(f.ParamTypes[i], argType) {
							return nil, nil, compiler.customError(node.Args[i], "wrong function parameter type")
						}
					}

					code = append(code, argCode...)
				}

				code = append(code, &bytecode.CallFunction{v.Value, len(f.ParamTypes)})

				if len(genericParameters) > 0 {
					return code, f.ReturnType.SubstituteGenerics(genericParameters, concreteTypes), nil
				} else {
					return code, f.ReturnType, nil
				}
			} else {
				return nil, nil, compiler.customError(node, "not a function")
			}
		}
	}

	if v, ok := node.Function.(*ast.EnumSymbolNode); ok {
		enumTypeAny, err := compiler.resolveType(&ast.SimpleTypeNode{v.Enum, nil})
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
			if enumCase.Label == v.Case {
				if len(enumCase.Types) != len(node.Args) {
					return nil, nil, compiler.customError(node, fmt.Sprintf("%s takes %d argument(s)", enumCase.Label, len(enumCase.Types)))
				}

				code := []bytecode.Bytecode{}
				genericParameters := []string{}
				concreteTypes := []vtype.VeniceType{}
				for i := 0; i < len(node.Args); i++ {
					argCode, argType, err := compiler.compileExpression(node.Args[i])
					if err != nil {
						return nil, nil, err
					}

					if genericParamType, ok := enumCase.Types[i].(*vtype.VeniceGenericParameterType); ok {
						genericParameters = append(genericParameters, genericParamType.Label)
						concreteTypes = append(concreteTypes, argType)
					} else {
						if !areTypesCompatible(enumCase.Types[i], argType) {
							return nil, nil, compiler.customError(node.Args[i], "wrong enum parameter type")
						}
					}

					code = append(code, argCode...)
				}
				code = append(code, &bytecode.PushEnum{v.Case, len(node.Args)})

				if len(genericParameters) > 0 {
					return code, enumType.SubstituteGenerics(genericParameters, concreteTypes), nil
				} else {
					return code, enumType, nil
				}
			}
		}

		return nil, nil, compiler.customError(node, fmt.Sprintf("enum %s does not have case %s", v.Enum, v.Case))
	}

	return nil, nil, compiler.customError(node, "function calls for non-symbols not implemented yet")
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
				return nil, nil, compiler.customError(node, fmt.Sprintf("%s takes %d argument(s)", enumCase.Label, len(enumCase.Types)))
			}

			return []bytecode.Bytecode{&bytecode.PushEnum{node.Case, 0}}, enumType, nil
		}
	}

	return nil, nil, compiler.customError(node, fmt.Sprintf("enum %s does not have case %s", node.Enum, node.Case))
}

func (compiler *Compiler) compileFieldAccessNode(node *ast.FieldAccessNode) ([]bytecode.Bytecode, vtype.VeniceType, error) {
	code, typeAny, err := compiler.compileExpression(node.Expr)
	if err != nil {
		return nil, nil, err
	}

	classType, ok := typeAny.(*vtype.VeniceClassType)
	if !ok {
		return nil, nil, compiler.customError(node, "left-hand side of dot must be a class object")
	}

	for i, field := range classType.Fields {
		if field.Name == node.Name {
			// TODO(2021-08-09): Allow this when inside the class itself.
			if !field.Public {
				return nil, nil, compiler.customError(node, "use of private field")
			}

			code = append(code, &bytecode.PushField{i})
			return code, field.FieldType, nil
		}
	}

	return nil, nil, compiler.customError(node, fmt.Sprintf("no such field: %s", node.Name))
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
			return nil, nil, compiler.customError(node, fmt.Sprintf("%s cannot be indexed", exprType.String()))
		}
		if !areTypesCompatible(vtype.VENICE_TYPE_INTEGER, indexType) {
			return nil, nil, compiler.customError(node.Expr, "string index must be integer")
		}

		code = append(code, &bytecode.BinaryStringIndex{})
		return code, vtype.VENICE_TYPE_CHARACTER, nil
	case *vtype.VeniceListType:
		if !areTypesCompatible(vtype.VENICE_TYPE_INTEGER, indexType) {
			return nil, nil, compiler.customError(node.Expr, "list index must be integer")
		}

		code = append(code, &bytecode.BinaryListIndex{})
		return code, exprType.ItemType, nil
	case *vtype.VeniceMapType:
		if !areTypesCompatible(exprType.KeyType, indexType) {
			return nil, nil, compiler.customError(node.Expr, "wrong map key type in index expression")
		}

		code = append(code, &bytecode.BinaryMapIndex{})
		return code, exprType.KeyType, nil
	default:
		return nil, nil, compiler.customError(node, fmt.Sprintf("%s cannot be indexed", exprTypeAny))
	}
}

func (compiler *Compiler) compileInfixNode(node *ast.InfixNode) ([]bytecode.Bytecode, vtype.VeniceType, error) {
	// TODO(2021-08-07): Boolean operators should short-circuit.
	opBytecode, ok := opsToBytecodes[node.Operator]
	if !ok {
		return nil, nil, compiler.customError(node, fmt.Sprintf("unknown operator: %s", node.Operator))
	}

	leftCode, leftType, err := compiler.compileExpression(node.Left)
	if err != nil {
		return nil, nil, err
	}

	if !checkInfixLeftType(node.Operator, leftType) {
		return nil, nil, compiler.customError(node.Left, fmt.Sprintf("invalid type for left operand of %s", node.Operator))
	}

	rightCode, rightType, err := compiler.compileExpression(node.Right)
	if err != nil {
		return nil, nil, err
	}

	resultType, ok := checkInfixRightType(node.Operator, leftType, rightType)
	if !ok {
		return nil, nil, compiler.customError(node.Right, fmt.Sprintf("invalid type for right operand of %s", node.Operator))
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
		} else if !areTypesCompatible(itemType, valueType) {
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
	code = append(code, &bytecode.BuildMap{len(node.Pairs)})
	return code, &vtype.VeniceMapType{keyType, valueType}, nil
}

func (compiler *Compiler) compileTernaryIfNode(node *ast.TernaryIfNode) ([]bytecode.Bytecode, vtype.VeniceType, error) {
	conditionCode, conditionType, err := compiler.compileExpression(node.Condition)
	if err != nil {
		return nil, nil, err
	}

	if !areTypesCompatible(vtype.VENICE_TYPE_BOOLEAN, conditionType) {
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

	if !areTypesCompatible(trueClauseType, falseClauseType) {
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
			return nil, nil, compiler.customError(node, fmt.Sprintf("argument to unary minus must be integer, not %s", exprType.String()))
		}

		code = append(code, &bytecode.UnaryMinus{})
		return code, vtype.VENICE_TYPE_INTEGER, nil
	case "not":
		code, exprType, err := compiler.compileExpression(node.Expr)
		if err != nil {
			return nil, nil, err
		}

		if exprType != vtype.VENICE_TYPE_BOOLEAN {
			return nil, nil, compiler.customError(node, fmt.Sprintf("argument to `not` must be boolean, not %s", exprType.String()))
		}

		code = append(code, &bytecode.UnaryNot{})
		return code, vtype.VENICE_TYPE_BOOLEAN, nil
	default:
		return nil, nil, compiler.customError(node, fmt.Sprintf("unknown unary operator: %s", node.Operator))
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
		return areTypesCompatible(vtype.VENICE_TYPE_BOOLEAN, leftType)
	case "++":
		if _, ok := leftType.(*vtype.VeniceListType); ok {
			return true
		} else {
			return areTypesCompatible(vtype.VENICE_TYPE_STRING, leftType)
		}
	default:
		return areTypesCompatible(vtype.VENICE_TYPE_INTEGER, leftType)
	}
}

func checkInfixRightType(operator string, leftType vtype.VeniceType, rightType vtype.VeniceType) (vtype.VeniceType, bool) {
	switch operator {
	case "==", "!=":
		return vtype.VENICE_TYPE_BOOLEAN, areTypesCompatible(leftType, rightType)
	case "and", "or":
		return vtype.VENICE_TYPE_BOOLEAN, areTypesCompatible(vtype.VENICE_TYPE_BOOLEAN, rightType)
	case ">", ">=", "<", "<=":
		return vtype.VENICE_TYPE_BOOLEAN, areTypesCompatible(vtype.VENICE_TYPE_INTEGER, rightType)
	case "++":
		if areTypesCompatible(vtype.VENICE_TYPE_STRING, leftType) {
			return vtype.VENICE_TYPE_STRING, areTypesCompatible(vtype.VENICE_TYPE_STRING, rightType)
		} else {
			return leftType, areTypesCompatible(leftType, rightType)
		}
	case "in":
		switch rightConcreteType := rightType.(type) {
		case *vtype.VeniceAtomicType:
			if rightConcreteType == vtype.VENICE_TYPE_STRING {
				return vtype.VENICE_TYPE_BOOLEAN, areTypesCompatible(vtype.VENICE_TYPE_CHARACTER, leftType) || areTypesCompatible(vtype.VENICE_TYPE_STRING, leftType)

			} else {
				return nil, false
			}
		case *vtype.VeniceListType:
			return vtype.VENICE_TYPE_BOOLEAN, areTypesCompatible(rightConcreteType.ItemType, leftType)
		case *vtype.VeniceMapType:
			return vtype.VENICE_TYPE_BOOLEAN, areTypesCompatible(rightConcreteType.KeyType, leftType)
		default:
			return nil, false
		}
	default:
		return vtype.VENICE_TYPE_INTEGER, areTypesCompatible(vtype.VENICE_TYPE_INTEGER, rightType)
	}
}

func areTypesCompatible(expectedTypeAny vtype.VeniceType, actualTypeAny vtype.VeniceType) bool {
	if expectedTypeAny == nil && actualTypeAny == nil {
		return true
	}

	switch expectedType := expectedTypeAny.(type) {
	case *vtype.VeniceAtomicType:
		actualType, ok := actualTypeAny.(*vtype.VeniceAtomicType)
		if !ok {
			return false
		}
		return expectedType.Type == actualType.Type
	case *vtype.VeniceEnumType:
		actualType, ok := actualTypeAny.(*vtype.VeniceEnumType)
		if !ok {
			return false
		}
		return expectedType == actualType
	case *vtype.VeniceListType:
		actualType, ok := actualTypeAny.(*vtype.VeniceListType)
		if !ok {
			return false
		}
		return areTypesCompatible(expectedType.ItemType, actualType.ItemType)
	case *vtype.VeniceMapType:
		actualType, ok := actualTypeAny.(*vtype.VeniceMapType)
		if !ok {
			return false
		}

		return areTypesCompatible(expectedType.KeyType, actualType.KeyType) && areTypesCompatible(expectedType.ValueType, actualType.ValueType)
	default:
		return false
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
			return nil, compiler.customError(typeNodeAny, fmt.Sprintf("unknown type: %s", typeNode.Symbol))
		}
		return resolvedType, nil
	default:
		return nil, compiler.customError(typeNodeAny, fmt.Sprintf("unknown type node: %T", typeNodeAny))
	}
}

/**
 * Symbol table methods
 */

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

func (compiler *Compiler) customError(node ast.Node, message string) *CompileError {
	return &CompileError{message, node.GetLocation()}
}
