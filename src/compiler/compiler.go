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
	"github.com/iafisher/venice/src/lexer"
	"github.com/iafisher/venice/src/vtype"
	"github.com/iafisher/venice/src/vval"
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

func (compiler *Compiler) Compile(tree *ast.ProgramNode) (vval.CompiledProgram, error) {
	compiledProgram := vval.NewCompiledProgram()
	for _, statementInterface := range tree.Statements {
		switch statement := statementInterface.(type) {
		case *ast.ClassDeclarationNode:
			code, err := compiler.compileClassDeclaration(statement)
			if err != nil {
				return nil, err
			}
			compiledProgram[statement.Name] = code
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

/**
 * Compile statements
 */

func (compiler *Compiler) compileStatement(treeInterface ast.StatementNode) ([]*vval.Bytecode, error) {
	switch tree := treeInterface.(type) {
	case *ast.AssignStatementNode:
		code, eType, err := compiler.compileExpression(tree.Expr)
		if err != nil {
			return nil, err
		}

		expectedType, ok := compiler.SymbolTable.Get(tree.Symbol)
		if !ok {
			return nil, compiler.customError(treeInterface, fmt.Sprintf("cannot assign to undeclared symbol %q", tree.Symbol))
		} else if !areTypesCompatible(expectedType, eType) {
			return nil, compiler.customError(treeInterface, fmt.Sprintf("wrong expression type in assignment to %q", tree.Symbol))
		}

		code = append(code, vval.NewBytecode("STORE_NAME", &vval.VeniceString{tree.Symbol}))
		return code, nil
	case *ast.BreakStatementNode:
		if compiler.nestedLoopCount == 0 {
			return nil, compiler.customError(treeInterface, "break statement outside of loop")
		}

		// BREAK_LOOP is a temporary bytecode instruction that the compiler will later
		// convert to a REL_JUMP instruction.
		return []*vval.Bytecode{vval.NewBytecode("BREAK_LOOP")}, nil
	case *ast.ContinueStatementNode:
		if compiler.nestedLoopCount == 0 {
			return nil, compiler.customError(treeInterface, "break statement outside of loop")
		}

		// CONTINUE_LOOP is a temporary bytecode instruction that the compiler will later
		// convert to a REL_JUMP instruction.
		return []*vval.Bytecode{vval.NewBytecode("CONTINUE_LOOP")}, nil
	case *ast.ExpressionStatementNode:
		code, _, err := compiler.compileExpression(tree.Expr)
		if err != nil {
			return nil, err
		}
		return code, nil
	case *ast.ForLoopNode:
		return compiler.compileForLoop(tree)
	case *ast.IfStatementNode:
		return compiler.compileIfStatement(tree)
	case *ast.LetStatementNode:
		if _, ok := compiler.SymbolTable.Get(tree.Symbol); ok {
			return nil, compiler.customError(treeInterface, fmt.Sprintf("re-declaration of symbol %q", tree.Symbol))
		}

		code, eType, err := compiler.compileExpression(tree.Expr)
		if err != nil {
			return nil, err
		}
		compiler.SymbolTable.Put(tree.Symbol, eType)
		return append(code, vval.NewBytecode("STORE_NAME", &vval.VeniceString{tree.Symbol})), nil
	case *ast.ReturnStatementNode:
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
			code = append(code, vval.NewBytecode("RETURN"))
			return code, err
		} else {
			if !areTypesCompatible(compiler.functionInfo.declaredReturnType, nil) {
				return nil, compiler.customError(treeInterface, "function cannot return void")
			}

			return []*vval.Bytecode{vval.NewBytecode("RETURN")}, nil
		}
	case *ast.WhileLoopNode:
		return compiler.compileWhileLoop(tree)
	default:
		return nil, compiler.customError(treeInterface, fmt.Sprintf("unknown statement type: %T", treeInterface))
	}
}

func (compiler *Compiler) compileClassDeclaration(tree *ast.ClassDeclarationNode) ([]*vval.Bytecode, error) {
	if tree.GenericTypeParameter != "" {
		subTypeSymbolTable := &SymbolTable{
			compiler.TypeSymbolTable,
			map[string]vtype.VeniceType{
				tree.GenericTypeParameter: &vtype.VeniceGenericParameterType{tree.GenericTypeParameter},
			},
		}
		compiler.TypeSymbolTable = subTypeSymbolTable
	}

	fields := []*vtype.VeniceClassField{}
	paramTypes := []vtype.VeniceType{}
	for _, field := range tree.Fields {
		paramType, err := compiler.resolveType(field.FieldType)
		if err != nil {
			return nil, err
		}
		paramTypes = append(paramTypes, paramType)
		fields = append(fields, &vtype.VeniceClassField{field.Name, field.Public, paramType})
	}

	var classType vtype.VeniceType
	if tree.GenericTypeParameter == "" {
		classType = &vtype.VeniceClassType{fields}
	} else {
		classType = &vtype.VeniceGenericType{[]string{tree.GenericTypeParameter}, &vtype.VeniceClassType{fields}}
		compiler.TypeSymbolTable = compiler.TypeSymbolTable.Parent
	}
	compiler.TypeSymbolTable.Put(tree.Name, classType)

	constructorType := &vtype.VeniceFunctionType{paramTypes, classType}
	compiler.SymbolTable.Put(tree.Name, constructorType)

	constructorBytecode := []*vval.Bytecode{vval.NewBytecode("BUILD_CLASS", &vval.VeniceInteger{len(fields)})}
	return constructorBytecode, nil
}

func (compiler *Compiler) compileEnumDeclaration(tree *ast.EnumDeclarationNode) error {
	if tree.GenericTypeParameter != "" {
		subTypeSymbolTable := &SymbolTable{
			compiler.TypeSymbolTable,
			map[string]vtype.VeniceType{
				tree.GenericTypeParameter: &vtype.VeniceGenericParameterType{tree.GenericTypeParameter},
			},
		}
		compiler.TypeSymbolTable = subTypeSymbolTable
	}

	caseTypes := []*vtype.VeniceCaseType{}
	for _, caseNode := range tree.Cases {
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

	if tree.GenericTypeParameter != "" {
		compiler.TypeSymbolTable = compiler.TypeSymbolTable.Parent
		compiler.TypeSymbolTable.Put(
			tree.Name,
			&vtype.VeniceGenericType{
				[]string{tree.GenericTypeParameter},
				&vtype.VeniceEnumType{caseTypes},
			},
		)
	} else {
		compiler.TypeSymbolTable.Put(tree.Name, &vtype.VeniceEnumType{caseTypes})
	}

	return nil
}

func (compiler *Compiler) compileForLoop(tree *ast.ForLoopNode) ([]*vval.Bytecode, error) {
	iterableCode, iterableTypeAny, err := compiler.compileExpression(tree.Iterable)
	if err != nil {
		return nil, err
	}

	var symbolTable *SymbolTable
	switch iterableType := iterableTypeAny.(type) {
	case *vtype.VeniceListType:
		if len(tree.Variables) != 1 {
			return nil, compiler.customError(tree, "too many for loop variables")
		}

		symbolTable = &SymbolTable{
			Parent:  compiler.SymbolTable,
			Symbols: map[string]vtype.VeniceType{tree.Variables[0]: iterableType.ItemType},
		}
	case *vtype.VeniceMapType:
		if len(tree.Variables) != 2 {
			return nil, compiler.customError(tree, "too many for loop variables")
		}

		symbolTable = &SymbolTable{
			Parent: compiler.SymbolTable,
			Symbols: map[string]vtype.VeniceType{
				tree.Variables[0]: iterableType.KeyType,
				tree.Variables[1]: iterableType.ValueType,
			},
		}
	default:
		return nil, compiler.customError(tree.Iterable, "for loop must be of list")
	}

	compiler.SymbolTable = symbolTable
	compiler.nestedLoopCount += 1
	bodyCode, err := compiler.compileBlock(tree.Body)
	compiler.nestedLoopCount -= 1
	compiler.SymbolTable = compiler.SymbolTable.Parent

	if err != nil {
		return nil, err
	}

	code := iterableCode
	code = append(code, vval.NewBytecode("GET_ITER"))
	code = append(code, vval.NewBytecode("FOR_ITER", &vval.VeniceInteger{len(bodyCode) + len(tree.Variables) + 2}))
	for i := len(tree.Variables) - 1; i >= 0; i-- {
		code = append(code, vval.NewBytecode("STORE_NAME", &vval.VeniceString{tree.Variables[i]}))
	}
	code = append(code, bodyCode...)
	code = append(code, vval.NewBytecode("REL_JUMP", &vval.VeniceInteger{-(len(bodyCode) + len(tree.Variables) + 1)}))
	return code, nil
}

func (compiler *Compiler) compileFunctionDeclaration(tree *ast.FunctionDeclarationNode) ([]*vval.Bytecode, error) {
	params := []string{}
	paramTypes := []vtype.VeniceType{}
	bodySymbolTableMap := map[string]vtype.VeniceType{}
	for _, param := range tree.Params {
		paramType, err := compiler.resolveType(param.ParamType)
		if err != nil {
			return nil, err
		}

		params = append(params, param.Name)
		paramTypes = append(paramTypes, paramType)

		bodySymbolTableMap[param.Name] = paramType
	}
	bodySymbolTable := &SymbolTable{compiler.SymbolTable, bodySymbolTableMap}

	declaredReturnType, err := compiler.resolveType(tree.ReturnType)
	if err != nil {
		return nil, err
	}

	// Put the function's entry in the symbol table before compiling the body so that
	// recursive functions can call themselves.
	compiler.SymbolTable.Put(tree.Name, &vtype.VeniceFunctionType{paramTypes, declaredReturnType})

	compiler.SymbolTable = bodySymbolTable
	compiler.functionInfo = &FunctionInfo{
		declaredReturnType:  declaredReturnType,
		seenReturnStatement: false,
	}
	bodyCode, err := compiler.compileBlock(tree.Body)

	if declaredReturnType != nil && !compiler.functionInfo.seenReturnStatement {
		return nil, compiler.customError(tree, "non-void function has no return statement")
	}

	compiler.functionInfo = nil
	compiler.SymbolTable = bodySymbolTable.Parent

	paramLoadCode := []*vval.Bytecode{}
	for _, param := range tree.Params {
		paramLoadCode = append(paramLoadCode, vval.NewBytecode("STORE_NAME", &vval.VeniceString{param.Name}))
	}
	bodyCode = append(paramLoadCode, bodyCode...)

	if err != nil {
		return nil, err
	}

	return bodyCode, nil
}

func (compiler *Compiler) compileIfStatement(tree *ast.IfStatementNode) ([]*vval.Bytecode, error) {
	conditionCode, conditionType, err := compiler.compileExpression(tree.Condition)
	if err != nil {
		return nil, err
	}

	if !areTypesCompatible(vtype.VENICE_TYPE_BOOLEAN, conditionType) {
		return nil, compiler.customError(tree.Condition, "condition of if statement must be a boolean")
	}

	trueClauseCode, err := compiler.compileBlock(tree.TrueClause)
	if err != nil {
		return nil, err
	}

	code := conditionCode
	code = append(code, vval.NewBytecode("REL_JUMP_IF_FALSE", &vval.VeniceInteger{len(trueClauseCode) + 1}))
	code = append(code, trueClauseCode...)

	if tree.FalseClause != nil {
		falseClauseCode, err := compiler.compileBlock(tree.FalseClause)
		if err != nil {
			return nil, err
		}

		code = append(code, vval.NewBytecode("REL_JUMP", &vval.VeniceInteger{len(falseClauseCode) + 1}))
		code = append(code, falseClauseCode...)
	}

	return code, nil
}

func (compiler *Compiler) compileWhileLoop(tree *ast.WhileLoopNode) ([]*vval.Bytecode, error) {
	conditionCode, conditionType, err := compiler.compileExpression(tree.Condition)
	if err != nil {
		return nil, err
	}

	if !areTypesCompatible(vtype.VENICE_TYPE_BOOLEAN, conditionType) {
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
	code = append(code, vval.NewBytecode("REL_JUMP_IF_FALSE", &vval.VeniceInteger{jumpForward}))
	code = append(code, bodyCode...)
	jumpBack := -(len(conditionCode) + len(bodyCode) + 1)
	code = append(code, vval.NewBytecode("REL_JUMP", &vval.VeniceInteger{jumpBack}))

	for i, bytecode := range code {
		if bytecode.Name == "BREAK_LOOP" {
			bytecode.Name = "REL_JUMP"
			bytecode.Args = append(bytecode.Args, &vval.VeniceInteger{len(code) - i})
		} else if bytecode.Name == "CONTINUE_LOOP" {
			bytecode.Name = "REL_JUMP"
			bytecode.Args = append(bytecode.Args, &vval.VeniceInteger{-i})
		}
	}

	return code, nil
}

func (compiler *Compiler) compileBlock(block []ast.StatementNode) ([]*vval.Bytecode, error) {
	code := []*vval.Bytecode{}
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

func (compiler *Compiler) compileExpression(treeInterface ast.ExpressionNode) ([]*vval.Bytecode, vtype.VeniceType, error) {
	switch tree := treeInterface.(type) {
	case *ast.BooleanNode:
		return []*vval.Bytecode{vval.NewBytecode("PUSH_CONST", &vval.VeniceBoolean{tree.Value})}, vtype.VENICE_TYPE_BOOLEAN, nil
	case *ast.CallNode:
		return compiler.compileCallNode(tree)
	case *ast.CharacterNode:
		return []*vval.Bytecode{vval.NewBytecode("PUSH_CONST", &vval.VeniceCharacter{tree.Value})}, vtype.VENICE_TYPE_CHARACTER, nil
	case *ast.EnumSymbolNode:
		return compiler.compileEnumSymbolNode(tree)
	case *ast.FieldAccessNode:
		return compiler.compileFieldAccessNode(tree)
	case *ast.IndexNode:
		return compiler.compileIndexNode(tree)
	case *ast.InfixNode:
		return compiler.compileInfixNode(tree)
	case *ast.IntegerNode:
		return []*vval.Bytecode{vval.NewBytecode("PUSH_CONST", &vval.VeniceInteger{tree.Value})}, vtype.VENICE_TYPE_INTEGER, nil
	case *ast.ListNode:
		code := []*vval.Bytecode{}
		var itemType vtype.VeniceType
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
		code = append(code, vval.NewBytecode("BUILD_LIST", &vval.VeniceInteger{len(tree.Values)}))
		return code, &vtype.VeniceListType{itemType}, nil
	case *ast.MapNode:
		code := []*vval.Bytecode{}
		var keyType vtype.VeniceType
		var valueType vtype.VeniceType
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
		code = append(code, vval.NewBytecode("BUILD_MAP", &vval.VeniceInteger{len(tree.Pairs)}))
		return code, &vtype.VeniceMapType{keyType, valueType}, nil
	case *ast.StringNode:
		return []*vval.Bytecode{vval.NewBytecode("PUSH_CONST", &vval.VeniceString{tree.Value})}, vtype.VENICE_TYPE_STRING, nil
	case *ast.SymbolNode:
		symbolType, ok := compiler.SymbolTable.Get(tree.Value)
		if !ok {
			return nil, nil, compiler.customError(treeInterface, fmt.Sprintf("undefined symbol: %s", tree.Value))
		}
		return []*vval.Bytecode{vval.NewBytecode("PUSH_NAME", &vval.VeniceString{tree.Value})}, symbolType, nil
	case *ast.TupleFieldAccessNode:
		return compiler.compileTupleFieldAccessNode(tree)
	case *ast.TupleNode:
		code := []*vval.Bytecode{}
		itemTypes := []vtype.VeniceType{}
		for i := len(tree.Values) - 1; i >= 0; i-- {
			value := tree.Values[i]
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

		code = append(code, vval.NewBytecode("BUILD_TUPLE", &vval.VeniceInteger{len(tree.Values)}))
		return code, &vtype.VeniceTupleType{itemTypes}, nil
	case *ast.UnaryNode:
		switch tree.Operator {
		case "-":
			code, exprType, err := compiler.compileExpression(tree.Expr)
			if err != nil {
				return nil, nil, err
			}

			if exprType != vtype.VENICE_TYPE_INTEGER {
				return nil, nil, compiler.customError(tree, fmt.Sprintf("argument to unary minus must be integer, not %s", exprType.String()))
			}

			code = append(code, vval.NewBytecode("UNARY_MINUS"))
			return code, vtype.VENICE_TYPE_INTEGER, nil
		case "not":
			code, exprType, err := compiler.compileExpression(tree.Expr)
			if err != nil {
				return nil, nil, err
			}

			if exprType != vtype.VENICE_TYPE_BOOLEAN {
				return nil, nil, compiler.customError(tree, fmt.Sprintf("argument to `not` must be boolean, not %s", exprType.String()))
			}

			code = append(code, vval.NewBytecode("UNARY_NOT"))
			return code, vtype.VENICE_TYPE_BOOLEAN, nil
		default:
			return nil, nil, compiler.customError(tree, fmt.Sprintf("unknown unary operator: %s", tree.Operator))
		}
	default:
		return nil, nil, compiler.customError(treeInterface, fmt.Sprintf("unknown expression type: %T", treeInterface))
	}
}

func (compiler *Compiler) compileCallNode(tree *ast.CallNode) ([]*vval.Bytecode, vtype.VeniceType, error) {
	if v, ok := tree.Function.(*ast.SymbolNode); ok {
		if v.Value == "print" {
			if len(tree.Args) != 1 {
				return nil, nil, compiler.customError(tree, "`print` takes exactly 1 argument")
			}

			code, _, err := compiler.compileExpression(tree.Args[0])
			if err != nil {
				return nil, nil, err
			}

			code = append(code, vval.NewBytecode("CALL_BUILTIN", &vval.VeniceString{"print"}))
			return code, nil, nil
		} else {
			valueInterface, ok := compiler.SymbolTable.Get(v.Value)
			if !ok {
				return nil, nil, compiler.customError(tree, fmt.Sprintf("undefined symbol: %s", v.Value))
			}

			if f, ok := valueInterface.(*vtype.VeniceFunctionType); ok {
				if len(f.ParamTypes) != len(tree.Args) {
					return nil, nil, compiler.customError(tree, fmt.Sprintf("wrong number of arguments: expected %d, got %d", len(f.ParamTypes), len(tree.Args)))
				}

				code := []*vval.Bytecode{}
				genericParameters := []string{}
				concreteTypes := []vtype.VeniceType{}
				for i := 0; i < len(f.ParamTypes); i++ {
					argCode, argType, err := compiler.compileExpression(tree.Args[i])
					if err != nil {
						return nil, nil, err
					}

					if genericParamType, ok := f.ParamTypes[i].(*vtype.VeniceGenericParameterType); ok {
						genericParameters = append(genericParameters, genericParamType.Label)
						concreteTypes = append(concreteTypes, argType)
					} else {
						if !areTypesCompatible(f.ParamTypes[i], argType) {
							return nil, nil, compiler.customError(tree.Args[i], "wrong function parameter type")
						}
					}

					code = append(code, argCode...)
				}

				code = append(code, vval.NewBytecode("CALL_FUNCTION", &vval.VeniceString{v.Value}, &vval.VeniceInteger{len(f.ParamTypes)}))

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

	if v, ok := tree.Function.(*ast.EnumSymbolNode); ok {
		enumTypeInterface, err := compiler.resolveType(&ast.SimpleTypeNode{v.Enum, nil})
		if err != nil {
			return nil, nil, err
		}

		isEnum := false
		enumType, ok := enumTypeInterface.(*vtype.VeniceEnumType)
		if ok {
			isEnum = true
		} else {
			genericType, ok := enumTypeInterface.(*vtype.VeniceGenericType)
			if ok {
				enumType, ok = genericType.GenericType.(*vtype.VeniceEnumType)
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

				code := []*vval.Bytecode{}
				genericParameters := []string{}
				concreteTypes := []vtype.VeniceType{}
				for i := 0; i < len(tree.Args); i++ {
					argCode, argType, err := compiler.compileExpression(tree.Args[i])
					if err != nil {
						return nil, nil, err
					}

					if genericParamType, ok := enumCase.Types[i].(*vtype.VeniceGenericParameterType); ok {
						genericParameters = append(genericParameters, genericParamType.Label)
						concreteTypes = append(concreteTypes, argType)
					} else {
						if !areTypesCompatible(enumCase.Types[i], argType) {
							return nil, nil, compiler.customError(tree.Args[i], "wrong enum parameter type")
						}
					}

					code = append(code, argCode...)
				}
				code = append(code, vval.NewBytecode("PUSH_ENUM", &vval.VeniceString{v.Case}, &vval.VeniceInteger{len(tree.Args)}))

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

func (compiler *Compiler) compileEnumSymbolNode(tree *ast.EnumSymbolNode) ([]*vval.Bytecode, vtype.VeniceType, error) {
	enumTypeInterface, err := compiler.resolveType(&ast.SimpleTypeNode{tree.Enum, nil})
	if err != nil {
		return nil, nil, err
	}

	isEnum := false
	enumType, ok := enumTypeInterface.(*vtype.VeniceEnumType)
	if ok {
		isEnum = true
	} else {
		genericType, ok := enumTypeInterface.(*vtype.VeniceGenericType)
		if ok {
			enumType, ok = genericType.GenericType.(*vtype.VeniceEnumType)
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

			return []*vval.Bytecode{vval.NewBytecode("PUSH_ENUM", &vval.VeniceString{tree.Case}, &vval.VeniceInteger{0})}, enumType, nil
		}
	}

	return nil, nil, compiler.customError(tree, fmt.Sprintf("enum %s does not have case %s", tree.Enum, tree.Case))
}

func (compiler *Compiler) compileFieldAccessNode(tree *ast.FieldAccessNode) ([]*vval.Bytecode, vtype.VeniceType, error) {
	code, typeInterface, err := compiler.compileExpression(tree.Expr)
	if err != nil {
		return nil, nil, err
	}

	classType, ok := typeInterface.(*vtype.VeniceClassType)
	if !ok {
		return nil, nil, compiler.customError(tree, "left-hand side of dot must be a class object")
	}

	for i, field := range classType.Fields {
		if field.Name == tree.Name {
			// TODO(2021-08-09): Allow this when inside the class itself.
			if !field.Public {
				return nil, nil, compiler.customError(tree, "use of private field")
			}

			code = append(code, vval.NewBytecode("PUSH_FIELD", &vval.VeniceInteger{i}))
			return code, field.FieldType, nil
		}
	}

	return nil, nil, compiler.customError(tree, fmt.Sprintf("no such field: %s", tree.Name))
}

func (compiler *Compiler) compileIndexNode(tree *ast.IndexNode) ([]*vval.Bytecode, vtype.VeniceType, error) {
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
	case *vtype.VeniceAtomicType:
		if exprType != vtype.VENICE_TYPE_STRING {
			return nil, nil, compiler.customError(tree, fmt.Sprintf("%s cannot be indexed", exprType.String()))
		}
		if !areTypesCompatible(vtype.VENICE_TYPE_INTEGER, indexType) {
			return nil, nil, compiler.customError(tree.Expr, "string index must be integer")
		}

		code = append(code, vval.NewBytecode("BINARY_STRING_INDEX"))
		return code, vtype.VENICE_TYPE_CHARACTER, nil
	case *vtype.VeniceListType:
		if !areTypesCompatible(vtype.VENICE_TYPE_INTEGER, indexType) {
			return nil, nil, compiler.customError(tree.Expr, "list index must be integer")
		}

		code = append(code, vval.NewBytecode("BINARY_LIST_INDEX"))
		return code, exprType.ItemType, nil
	case *vtype.VeniceMapType:
		if !areTypesCompatible(exprType.KeyType, indexType) {
			return nil, nil, compiler.customError(tree.Expr, "wrong map key type in index expression")
		}

		code = append(code, vval.NewBytecode("BINARY_MAP_INDEX"))
		return code, exprType.KeyType, nil
	default:
		return nil, nil, compiler.customError(tree, fmt.Sprintf("%s cannot be indexed", exprTypeInterface))
	}
}

func (compiler *Compiler) compileInfixNode(tree *ast.InfixNode) ([]*vval.Bytecode, vtype.VeniceType, error) {
	// TODO(2021-08-07): Boolean operators should short-circuit.
	bytecodeName, ok := opsToBytecodeNames[tree.Operator]
	if !ok {
		return nil, nil, compiler.customError(tree, fmt.Sprintf("unknown operator: %s", tree.Operator))
	}

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

	resultType, ok := checkInfixRightType(tree.Operator, leftType, rightType)
	if !ok {
		return nil, nil, compiler.customError(tree.Right, fmt.Sprintf("invalid type for right operand of %s", tree.Operator))
	}

	code := append(leftCode, rightCode...)
	code = append(code, vval.NewBytecode(bytecodeName))
	return code, resultType, nil
}

func (compiler *Compiler) compileTupleFieldAccessNode(tree *ast.TupleFieldAccessNode) ([]*vval.Bytecode, vtype.VeniceType, error) {
	code, typeInterface, err := compiler.compileExpression(tree.Expr)
	if err != nil {
		return nil, nil, err
	}

	tupleType, ok := typeInterface.(*vtype.VeniceTupleType)
	if !ok {
		return nil, nil, compiler.customError(tree, "left-hand side of dot must be a tuple object")
	}

	if tree.Index < 0 || tree.Index >= len(tupleType.ItemTypes) {
		return nil, nil, compiler.customError(tree, "tuple index out of bounds")
	}

	code = append(code, vval.NewBytecode("PUSH_TUPLE_FIELD", &vval.VeniceInteger{tree.Index}))
	return code, tupleType.ItemTypes[tree.Index], nil
}

/**
 * Type utility functions
 */

var opsToBytecodeNames = map[string]string{
	"+":   "BINARY_ADD",
	"and": "BINARY_AND",
	"++":  "BINARY_CONCAT",
	"/":   "BINARY_DIV",
	"==":  "BINARY_EQ",
	">":   "BINARY_GT",
	">=":  "BINARY_GT_EQ",
	"in":  "BINARY_IN",
	"<":   "BINARY_LT",
	"<=":  "BINARY_LT_EQ",
	"*":   "BINARY_MUL",
	"!=":  "BINARY_NOT_EQ",
	"or":  "BINARY_OR",
	"-":   "BINARY_SUB",
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

func areTypesCompatible(expectedTypeInterface vtype.VeniceType, actualTypeInterface vtype.VeniceType) bool {
	if expectedTypeInterface == nil && actualTypeInterface == nil {
		return true
	}

	switch expectedType := expectedTypeInterface.(type) {
	case *vtype.VeniceAtomicType:
		switch actualType := actualTypeInterface.(type) {
		case *vtype.VeniceAtomicType:
			return expectedType.Type == actualType.Type
		default:
			return false
		}
	case *vtype.VeniceEnumType:
		switch actualType := actualTypeInterface.(type) {
		case *vtype.VeniceEnumType:
			return expectedType == actualType
		default:
			return false
		}
	case *vtype.VeniceListType:
		switch actualType := actualTypeInterface.(type) {
		case *vtype.VeniceListType:
			return areTypesCompatible(expectedType.ItemType, actualType.ItemType)
		default:
			return false
		}
	default:
		return false
	}
}

func (compiler *Compiler) resolveType(typeNodeInterface ast.TypeNode) (vtype.VeniceType, error) {
	if typeNodeInterface == nil {
		return nil, nil
	}

	switch typeNode := typeNodeInterface.(type) {
	case *ast.SimpleTypeNode:
		resolvedType, ok := compiler.TypeSymbolTable.Get(typeNode.Symbol)
		if !ok {
			return nil, compiler.customError(typeNodeInterface, fmt.Sprintf("unknown type: %s", typeNode.Symbol))
		}
		return resolvedType, nil
	default:
		return nil, compiler.customError(typeNodeInterface, fmt.Sprintf("unknown type node: %T", typeNodeInterface))
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
		return fmt.Sprintf("%s at line %d, column %d of %s", e.Message, e.Location.Line, e.Location.Column, e.Location.FilePath)
	} else {
		return e.Message
	}
}

func (compiler *Compiler) customError(node ast.Node, message string) *CompileError {
	return &CompileError{message, node.GetLocation()}
}
