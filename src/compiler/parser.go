/**
 * Parse a Venice program into an abstract syntax tree (defined in src/compiler/ast.go).
 *
 * Each `matchXYZ` function expects that `parser.currentToken` is set to the first token
 * of the node to be matched, and after it returns `parser.currentToken` is set to one
 * past the last token of the node.
 */
package compiler

import (
	"fmt"
	"github.com/iafisher/venice/src/common/lex"
	"io/ioutil"
	pathlib "path"
	"strconv"
	"strings"
)

type Parser struct {
	lexer        *lex.Lexer
	currentToken *lex.Token
	// Number of currently nested brackets (parentheses, curly braces, square brackets).
	brackets int
}

func NewParser() *Parser {
	return &Parser{lexer: nil, currentToken: nil, brackets: 0}
}

func (p *Parser) ParseString(input string) (*File, error) {
	return p.parseGeneric("", input)
}

func (p *Parser) ParseFile(filePath string) (*File, error) {
	fileContentsBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return p.parseGeneric(filePath, string(fileContentsBytes))
}

func (p *Parser) parseGeneric(filePath string, input string) (*File, error) {
	p.lexer = lex.NewLexer(filePath, input)
	p.nextTokenSkipNewlines()
	statements := []StatementNode{}

	for {
		statement, err := p.matchStatement()
		if err != nil {
			return nil, err
		}
		statements = append(statements, statement)

		if p.currentToken.Type == lex.TOKEN_EOF {
			break
		}
	}

	if len(statements) == 0 {
		return nil, p.customError("empty program")
	}

	return &File{Statements: statements, Imports: []string{}}, nil
}

/**
 * Match statements
 */

func (p *Parser) matchStatement() (StatementNode, error) {
	location := p.currentToken.Location
	var tree StatementNode
	var err error
	switch p.currentToken.Type {
	case lex.TOKEN_BREAK:
		tree = &BreakStatementNode{location}
		p.nextToken()
	case lex.TOKEN_CLASS:
		return p.matchClassDeclaration()
	case lex.TOKEN_CONTINUE:
		tree = &ContinueStatementNode{location}
		p.nextToken()
	case lex.TOKEN_ENUM:
		return p.matchEnumDeclaration()
	case lex.TOKEN_FUNC:
		return p.matchFunctionDeclaration()
	case lex.TOKEN_FOR:
		return p.matchForLoop()
	case lex.TOKEN_IF:
		return p.matchIfStatement()
	case lex.TOKEN_IMPORT:
		tree, err = p.matchImportStatement()
	case lex.TOKEN_LET:
		tree, err = p.matchLetStatement(false)
	case lex.TOKEN_MATCH:
		tree, err = p.matchMatchStatement()
	case lex.TOKEN_RETURN:
		tree, err = p.matchReturnStatement()
	case lex.TOKEN_VAR:
		tree, err = p.matchLetStatement(true)
	case lex.TOKEN_WHILE:
		return p.matchWhileLoop()
	default:
		expr, err := p.matchExpression(PRECEDENCE_LOWEST)
		if err != nil {
			return nil, err
		}

		if p.currentToken.Type == lex.TOKEN_ASSIGN {
			p.nextToken()
			assignExpr, err := p.matchExpression(PRECEDENCE_LOWEST)
			if err != nil {
				return nil, err
			}

			tree = &AssignStatementNode{
				Destination: expr,
				Expr:        assignExpr,
			}
		} else if operator, ok := compoundAssignOperators[p.currentToken.Type]; ok {
			p.nextToken()
			assignExpr, err := p.matchExpression(PRECEDENCE_LOWEST)
			if err != nil {
				return nil, err
			}

			tree = &AssignStatementNode{
				Destination: expr,
				Expr: &InfixNode{
					Operator: operator,
					Left:     expr,
					Right:    assignExpr,
				},
			}
		} else {
			tree = &ExpressionStatementNode{expr}
		}
	}

	if err != nil {
		return nil, err
	}

	if p.currentToken.Type != lex.TOKEN_NEWLINE &&
		p.currentToken.Type != lex.TOKEN_SEMICOLON &&
		p.currentToken.Type != lex.TOKEN_EOF &&
		p.currentToken.Type != lex.TOKEN_RIGHT_CURLY {
		return nil, p.customError(
			"statement must be followed by newline or semicolon (got %s)",
			p.currentToken.Type,
		)
	}

	if p.currentToken.Type == lex.TOKEN_NEWLINE ||
		p.currentToken.Type == lex.TOKEN_SEMICOLON {
		p.nextTokenSkipNewlines()
	}

	return tree, nil
}

func (p *Parser) matchClassDeclaration() (*ClassDeclarationNode, error) {
	location := p.currentToken.Location
	p.nextToken()
	if p.currentToken.Type != lex.TOKEN_SYMBOL {
		return nil, p.unexpectedToken("class name")
	}

	name := p.currentToken.Value

	p.nextToken()
	noConstructor := false
	if p.currentToken.Type == lex.TOKEN_NO {
		p.nextToken()
		if p.currentToken.Type != lex.TOKEN_CONSTRUCTOR {
			return nil, p.unexpectedToken("keyword `constructor`")
		}
		noConstructor = true
		p.nextToken()
	}

	if p.currentToken.Type != lex.TOKEN_LEFT_CURLY {
		return nil, p.unexpectedToken("left curly brace")
	}
	p.nextTokenSkipNewlines()

	fieldNodes := []*ClassFieldNode{}
	for {
		if p.currentToken.Type == lex.TOKEN_RIGHT_CURLY {
			p.nextTokenSkipNewlines()
			break
		}

		var public bool
		if p.currentToken.Type == lex.TOKEN_PUBLIC {
			public = true
		} else if p.currentToken.Type == lex.TOKEN_PRIVATE {
			public = false
		} else {
			return nil, p.unexpectedToken("`public` or `private`")
		}

		p.nextToken()
		if p.currentToken.Type == lex.TOKEN_SYMBOL {
			name := p.currentToken.Value

			p.nextToken()
			if p.currentToken.Type != lex.TOKEN_COLON {
				return nil, p.unexpectedToken("colon")
			}

			p.nextToken()
			fieldType, err := p.matchTypeNode()
			if err != nil {
				return nil, err
			}

			if p.currentToken.Type == lex.TOKEN_NEWLINE {
				p.nextTokenSkipNewlines()
			}

			fieldNodes = append(fieldNodes, &ClassFieldNode{name, public, fieldType})
		} else {
			return nil, p.unexpectedToken("symbol")
		}
	}

	return &ClassDeclarationNode{
		Name:          name,
		NoConstructor: noConstructor,
		Fields:        fieldNodes,
		Location:      location,
	}, nil
}

func (p *Parser) matchEnumDeclaration() (*EnumDeclarationNode, error) {
	location := p.currentToken.Location
	p.nextToken()
	if p.currentToken.Type != lex.TOKEN_SYMBOL {
		return nil, p.unexpectedToken("symbol")
	}

	name := p.currentToken.Value

	var genericTypeParameter string
	p.nextToken()
	if p.currentToken.Type == lex.TOKEN_LESS_THAN {
		p.nextToken()
		if p.currentToken.Type != lex.TOKEN_SYMBOL {
			return nil, p.unexpectedToken("type parameter")
		}
		genericTypeParameter = p.currentToken.Value

		p.nextToken()
		if p.currentToken.Type != lex.TOKEN_GREATER_THAN {
			return nil, p.unexpectedToken("right angle bracket")
		}
		p.nextToken()
	}

	if p.currentToken.Type != lex.TOKEN_LEFT_CURLY {
		return nil, p.unexpectedToken("left curly brace")
	}
	p.nextTokenSkipNewlines()

	cases := []*EnumCaseNode{}
	for {
		if p.currentToken.Type == lex.TOKEN_RIGHT_CURLY {
			p.nextTokenSkipNewlines()
			break
		}

		location := p.currentToken.Location
		if p.currentToken.Type != lex.TOKEN_SYMBOL {
			return nil, p.unexpectedToken("symbol or right curly brace")
		}

		label := p.currentToken.Value
		p.nextTokenSkipNewlines()

		if p.currentToken.Type == lex.TOKEN_LEFT_PAREN {
			p.nextTokenSkipNewlines()
			types := []TypeNode{}
			for {
				if p.currentToken.Type == lex.TOKEN_RIGHT_PAREN {
					p.nextTokenSkipNewlines()
					break
				}

				typeNode, err := p.matchTypeNode()
				if err != nil {
					return nil, err
				}

				types = append(types, typeNode)

				if p.currentToken.Type == lex.TOKEN_COMMA {
					p.nextTokenSkipNewlines()
				} else if p.currentToken.Type == lex.TOKEN_RIGHT_PAREN {
					p.nextTokenSkipNewlines()
					break
				} else {
					return nil, p.unexpectedToken("comma or right parenthesis")
				}
			}

			if len(types) == 0 {
				return nil, p.customError(
					"enum case with parentheses must include at least one type",
				)
			}

			cases = append(cases, &EnumCaseNode{label, types, location})
		} else {
			cases = append(cases, &EnumCaseNode{label, []TypeNode{}, location})
		}

		if p.currentToken.Type == lex.TOKEN_COMMA {
			p.nextTokenSkipNewlines()
		} else if p.currentToken.Type == lex.TOKEN_RIGHT_CURLY {
			p.nextTokenSkipNewlines()
			break
		} else {
			return nil, p.unexpectedToken("comma or right curly brace")
		}
	}

	return &EnumDeclarationNode{name, genericTypeParameter, cases, location}, nil
}

func (p *Parser) matchForLoop() (*ForLoopNode, error) {
	location := p.currentToken.Location
	p.nextToken()

	variables := []string{}
	for {
		if p.currentToken.Type != lex.TOKEN_SYMBOL {
			return nil, p.unexpectedToken("symbol")
		}
		variables = append(variables, p.currentToken.Value)

		p.nextToken()
		if p.currentToken.Type == lex.TOKEN_COMMA {
			p.nextToken()
		} else if p.currentToken.Type == lex.TOKEN_IN {
			break
		} else {
			return nil, p.unexpectedToken("comma or keyword 'in'")
		}
	}

	p.nextToken()
	iterable, err := p.matchExpression(PRECEDENCE_LOWEST)
	if err != nil {
		return nil, err
	}

	body, err := p.matchBlock()
	if err != nil {
		return nil, err
	}

	return &ForLoopNode{variables, iterable, body, location}, nil
}

func (p *Parser) matchFunctionDeclaration() (*FunctionDeclarationNode, error) {
	location := p.currentToken.Location
	p.nextToken()
	if p.currentToken.Type != lex.TOKEN_SYMBOL {
		return nil, p.unexpectedToken("function name")
	}

	name := p.currentToken.Value
	p.nextToken()

	if p.currentToken.Type != lex.TOKEN_LEFT_PAREN {
		return nil, p.unexpectedToken("left parenthesis")
	}

	p.nextToken()
	params, err := p.matchFunctionParams()
	if err != nil {
		return nil, err
	}

	if p.currentToken.Type != lex.TOKEN_RIGHT_PAREN {
		return nil, p.unexpectedToken("right parenthesis")
	}

	p.nextToken()
	var returnType TypeNode
	if p.currentToken.Type == lex.TOKEN_ARROW {
		p.nextToken()
		returnType, err = p.matchTypeNode()
		if err != nil {
			return nil, err
		}
	}

	body, err := p.matchBlock()
	if err != nil {
		return nil, err
	}

	return &FunctionDeclarationNode{name, params, returnType, body, location}, nil
}

func (p *Parser) matchFunctionParams() ([]*FunctionParamNode, error) {
	params := []*FunctionParamNode{}
	for p.currentToken.Type != lex.TOKEN_RIGHT_PAREN {
		if p.currentToken.Type != lex.TOKEN_SYMBOL {
			return nil, p.unexpectedToken("function parameter name")
		}
		name := p.currentToken.Value
		location := p.currentToken.Location

		p.nextToken()
		if p.currentToken.Type != lex.TOKEN_COLON {
			return nil, p.unexpectedToken("colon")
		}

		p.nextToken()
		paramType, err := p.matchTypeNode()
		if err != nil {
			return nil, err
		}

		params = append(params, &FunctionParamNode{name, paramType, location})

		if p.currentToken.Type == lex.TOKEN_RIGHT_PAREN {
			break
		} else if p.currentToken.Type == lex.TOKEN_COMMA {
			p.nextToken()
		} else {
			return nil, p.unexpectedToken("comma or right parenthesis")
		}
	}
	return params, nil
}

func (p *Parser) matchIfStatement() (*IfStatementNode, error) {
	location := p.currentToken.Location
	p.nextToken()
	condition, err := p.matchExpression(PRECEDENCE_LOWEST)
	if err != nil {
		return nil, err
	}

	trueClauseStatements, err := p.matchBlock()
	if err != nil {
		return nil, err
	}

	var elseClause []StatementNode
	clauses := []*IfClauseNode{&IfClauseNode{condition, trueClauseStatements}}
	for p.currentToken.Type == lex.TOKEN_ELSE {
		p.nextToken()

		if p.currentToken.Type == lex.TOKEN_LEFT_CURLY {
			elseClause, err = p.matchBlock()
			if err != nil {
				return nil, err
			}
			break
		} else if p.currentToken.Type == lex.TOKEN_IF {
			p.nextToken()
			condition, err = p.matchExpression(PRECEDENCE_LOWEST)
			if err != nil {
				return nil, err
			}

			body, err := p.matchBlock()
			if err != nil {
				return nil, err
			}

			clauses = append(clauses, &IfClauseNode{condition, body})
		} else {
			return nil, p.unexpectedToken("`else if` or `else`")
		}
	}

	return &IfStatementNode{clauses, elseClause, location}, nil
}

func (p *Parser) matchImportStatement() (*ImportStatementNode, error) {
	location := p.currentToken.Location
	p.nextToken()
	if p.currentToken.Type != lex.TOKEN_STRING {
		return nil, p.unexpectedToken("string")
	}
	path := p.currentToken.Value

	if !strings.HasPrefix(path, "./") {
		path = pathlib.Join("/usr/lib/venice0.1", path) + ".vn"
	} else if p.currentToken.Location.FilePath != "" {
		path = pathlib.Join(pathlib.Dir(p.currentToken.Location.FilePath), path)
	}

	p.nextToken()
	if p.currentToken.Type != lex.TOKEN_AS {
		return nil, p.unexpectedToken("keyword `as`")
	}

	p.nextToken()
	if p.currentToken.Type != lex.TOKEN_SYMBOL {
		return nil, p.unexpectedToken("string")
	}
	name := p.currentToken.Value

	p.nextToken()
	return &ImportStatementNode{Path: path, Name: name, Location: location}, nil
}

func (p *Parser) matchLetStatement(isVar bool) (*LetStatementNode, error) {
	location := p.currentToken.Location
	p.nextToken()
	if p.currentToken.Type != lex.TOKEN_SYMBOL {
		return nil, p.unexpectedToken("symbol")
	}
	symbol := p.currentToken.Value

	p.nextToken()
	var typeNode TypeNode
	if p.currentToken.Type == lex.TOKEN_COLON {
		var err error
		p.nextToken()
		typeNode, err = p.matchTypeNode()
		if err != nil {
			return nil, err
		}
	}

	if p.currentToken.Type != lex.TOKEN_ASSIGN {
		return nil, p.unexpectedToken("equals sign")
	}

	p.nextToken()
	expr, err := p.matchExpression(PRECEDENCE_LOWEST)
	if err != nil {
		return nil, err
	}

	return &LetStatementNode{
		Symbol:   symbol,
		Type:     typeNode,
		IsVar:    isVar,
		Expr:     expr,
		Location: location,
	}, nil
}

func (p *Parser) matchMatchStatement() (*MatchStatementNode, error) {
	location := p.currentToken.Location

	p.nextToken()
	expr, err := p.matchExpression(PRECEDENCE_LOWEST)
	if err != nil {
		return nil, err
	}

	if p.currentToken.Type != lex.TOKEN_LEFT_CURLY {
		return nil, p.unexpectedToken("left curly brace")
	}

	p.nextTokenSkipNewlines()
	clauses := []*MatchClause{}
	for {
		if p.currentToken.Type == lex.TOKEN_CASE {
			p.nextToken()
			pattern, err := p.matchPatternNode()
			if err != nil {
				return nil, err
			}

			body, err := p.matchBlock()
			if err != nil {
				return nil, err
			}

			clauses = append(clauses, &MatchClause{Pattern: pattern, Body: body})
		} else if p.currentToken.Type == lex.TOKEN_DEFAULT ||
			p.currentToken.Type == lex.TOKEN_RIGHT_CURLY {
			break
		} else {
			return nil, p.unexpectedToken(
				"`case` keyword, `default` keyword, or right curly brace",
			)
		}
	}

	defaultNode := []StatementNode{}
	if p.currentToken.Type == lex.TOKEN_DEFAULT {
		p.nextToken()
		defaultNode, err = p.matchBlock()
		if err != nil {
			return nil, err
		}
	}

	if p.currentToken.Type != lex.TOKEN_RIGHT_CURLY {
		return nil, p.unexpectedToken("right curly brace")
	}

	p.nextToken()
	return &MatchStatementNode{
		Expr:     expr,
		Clauses:  clauses,
		Default:  defaultNode,
		Location: location,
	}, nil
}

func (p *Parser) matchReturnStatement() (*ReturnStatementNode, error) {
	location := p.currentToken.Location
	p.nextToken()
	if p.currentToken.Type == lex.TOKEN_NEWLINE ||
		p.currentToken.Type == lex.TOKEN_SEMICOLON {
		return &ReturnStatementNode{nil, location}, nil
	}

	expr, err := p.matchExpression(PRECEDENCE_LOWEST)
	if err != nil {
		return nil, err
	}
	return &ReturnStatementNode{expr, location}, nil
}

func (p *Parser) matchWhileLoop() (*WhileLoopNode, error) {
	location := p.currentToken.Location
	p.nextToken()
	condition, err := p.matchExpression(PRECEDENCE_LOWEST)
	if err != nil {
		return nil, err
	}

	body, err := p.matchBlock()
	if err != nil {
		return nil, err
	}

	return &WhileLoopNode{condition, body, location}, nil
}

/**
 * Match expressions
 */

func (p *Parser) matchExpression(precedence int) (ExpressionNode, error) {
	expr, err := p.matchPrefix()
	if err != nil {
		return nil, err
	}

	for {
		if infixPrecedence, ok := precedenceMap[p.currentToken.Type]; ok {
			if precedence < infixPrecedence {
				if p.currentToken.Type == lex.TOKEN_LEFT_PAREN {
					arglist, err := p.matchArglist(lex.TOKEN_RIGHT_PAREN)
					if err != nil {
						return nil, err
					}

					expr = &CallNode{expr, arglist}
				} else if p.currentToken.Type == lex.TOKEN_LEFT_SQUARE {
					p.nextToken()
					indexExpr, err := p.matchExpression(PRECEDENCE_LOWEST)
					if err != nil {
						return nil, err
					}

					if p.currentToken.Type != lex.TOKEN_RIGHT_SQUARE {
						return nil, p.unexpectedToken("right square bracket")
					}
					p.nextToken()

					expr = &IndexNode{expr, indexExpr}
				} else if p.currentToken.Type == lex.TOKEN_DOT {
					p.nextToken()
					if p.currentToken.Type == lex.TOKEN_SYMBOL {
						expr = &FieldAccessNode{expr, p.currentToken.Value}
					} else if p.currentToken.Type == lex.TOKEN_INT {
						index, err := strconv.ParseInt(p.currentToken.Value, 10, 0)
						if err != nil {
							return nil, p.customError("could not convert integer token")
						}
						expr = &TupleFieldAccessNode{expr, int(index)}
					} else {
						return nil, p.customError("right-hand side of dot must be a symbol")
					}

					p.nextToken()
				} else if p.currentToken.Type == lex.TOKEN_IF {
					p.nextToken()
					conditionExpr, err := p.matchExpression(PRECEDENCE_LOWEST)
					if err != nil {
						return nil, err
					}

					if p.currentToken.Type != lex.TOKEN_ELSE {
						return nil, p.customError("`else`")
					}

					p.nextToken()
					elseExpr, err := p.matchExpression(PRECEDENCE_LOWEST)
					if err != nil {
						return nil, err
					}

					expr = &TernaryIfNode{
						Condition:   conditionExpr,
						TrueClause:  expr,
						FalseClause: elseExpr,
					}
				} else if p.currentToken.Type == lex.TOKEN_NOT {
					location := p.currentToken.Location
					p.nextToken()
					if p.currentToken.Type != lex.TOKEN_IN {
						return nil, p.customError(
							"expected `in` after `not` in infix position",
						)
					}

					p.nextToken()
					right, err := p.matchExpression(infixPrecedence)
					if err != nil {
						return nil, err
					}

					expr = &UnaryNode{
						Operator: "not",
						Expr: &InfixNode{
							Operator: "in",
							Left:     expr,
							Right:    right,
						},
						Location: location,
					}
				} else {
					expr, err = p.matchInfix(expr, infixPrecedence)
					if err != nil {
						return nil, err
					}
				}
			} else {
				break
			}
		} else {
			break
		}
	}

	return expr, nil
}

func (p *Parser) matchInfix(left ExpressionNode, precedence int) (ExpressionNode, error) {
	operator := p.currentToken.Value
	p.nextToken()
	right, err := p.matchExpression(precedence)
	if err != nil {
		return nil, err
	}

	// If the infix expression is a double comparison like `0 <= x < 100`, then refactor
	// it into `0 <= x and x < 100`.
	leftInfix, ok := left.(*InfixNode)
	if ok {
		if comparisonOperators[operator] && comparisonOperators[leftInfix.Operator] {
			return &InfixNode{
				Operator: "and",
				Left:     left,
				Right: &InfixNode{
					Operator: operator,
					Left:     leftInfix.Right,
					Right:    right,
				},
			}, nil
		}
	}

	return &InfixNode{operator, left, right}, nil
}

func (p *Parser) matchPrefix() (ExpressionNode, error) {
	location := p.currentToken.Location
	switch p.currentToken.Type {
	case lex.TOKEN_CHARACTER:
		value := p.currentToken.Value
		p.nextToken()
		return &CharacterNode{value[0], location}, nil
	case lex.TOKEN_FALSE:
		p.nextToken()
		return &BooleanNode{false, location}, nil
	case lex.TOKEN_INT:
		token := p.currentToken
		p.nextToken()
		value, err := strconv.ParseInt(token.Value, 0, 0)
		if err != nil {
			return nil, p.customError("invalid integer literal")
		}
		return &IntegerNode{int(value), location}, nil
	case lex.TOKEN_LEFT_CURLY:
		p.brackets++
		p.nextToken()
		pairs, err := p.matchMapPairs()
		if err != nil {
			return nil, err
		}
		return &MapNode{pairs, location}, nil
	case lex.TOKEN_LEFT_PAREN:
		values, err := p.matchArglist(lex.TOKEN_RIGHT_PAREN)
		if err != nil {
			return nil, err
		}

		if len(values) == 1 {
			return values[0], nil
		} else {
			return &TupleNode{values, location}, nil
		}
	case lex.TOKEN_LEFT_SQUARE:
		values, err := p.matchArglist(lex.TOKEN_RIGHT_SQUARE)
		if err != nil {
			return nil, err
		}
		return &ListNode{values, location}, nil
	case lex.TOKEN_MINUS, lex.TOKEN_NOT:
		operator := p.currentToken.Value
		p.nextToken()
		expr, err := p.matchExpression(PRECEDENCE_PREFIX)
		if err != nil {
			return nil, err
		}
		return &UnaryNode{operator, expr, location}, nil
	case lex.TOKEN_NEW:
		return p.matchConstructor()
	case lex.TOKEN_REAL_NUMBER:
		token := p.currentToken
		p.nextToken()
		value, err := strconv.ParseFloat(token.Value, 64)
		if err != nil {
			return nil, p.customError("invalid real number literal")
		}
		return &RealNumberNode{value, location}, nil
	case lex.TOKEN_SELF:
		p.nextToken()
		return &SymbolNode{"self", location}, nil
	case lex.TOKEN_STRING:
		value := p.currentToken.Value
		p.nextToken()
		return &StringNode{value, location}, nil
	case lex.TOKEN_SYMBOL:
		value := p.currentToken.Value
		p.nextToken()

		if p.currentToken.Type == lex.TOKEN_DOUBLE_COLON {
			p.nextToken()
			if p.currentToken.Type != lex.TOKEN_SYMBOL {
				return nil, p.unexpectedToken("symbol")
			}
			secondValue := p.currentToken.Value
			p.nextToken()
			return &QualifiedSymbolNode{value, secondValue, location}, nil
		} else {
			return &SymbolNode{value, location}, nil
		}
	case lex.TOKEN_TRUE:
		p.nextToken()
		return &BooleanNode{true, location}, nil
	default:
		return nil, p.unexpectedToken("start of expression")
	}
}

func (p *Parser) matchConstructor() (*ConstructorNode, error) {
	location := p.currentToken.Location
	p.nextToken()

	if p.currentToken.Type != lex.TOKEN_SYMBOL {
		return nil, p.unexpectedToken("class name")
	}
	name := p.currentToken.Value

	p.nextToken()
	if p.currentToken.Type != lex.TOKEN_LEFT_CURLY {
		return nil, p.unexpectedToken("left curly brace")
	}

	p.brackets++

	p.nextToken()
	fields := []*ConstructorFieldNode{}
	for {
		if p.currentToken.Type == lex.TOKEN_RIGHT_CURLY {
			break
		}

		if p.currentToken.Type != lex.TOKEN_SYMBOL {
			p.brackets--
			return nil, p.unexpectedToken("field name")
		}
		fieldName := p.currentToken.Value

		p.nextToken()
		if p.currentToken.Type != lex.TOKEN_COLON {
			p.brackets--
			return nil, p.unexpectedToken("colon")
		}

		p.nextToken()
		expr, err := p.matchExpression(PRECEDENCE_LOWEST)
		if err != nil {
			p.brackets--
			return nil, err
		}

		fields = append(
			fields,
			&ConstructorFieldNode{Name: fieldName, Value: expr},
		)

		if p.currentToken.Type == lex.TOKEN_RIGHT_CURLY {
			break
		} else if p.currentToken.Type != lex.TOKEN_COMMA {
			p.brackets--
			return nil, p.unexpectedToken("comma")
		}

		p.nextToken()
	}

	p.brackets--
	p.nextToken()
	return &ConstructorNode{
		Name:     name,
		Fields:   fields,
		Location: location,
	}, nil
}

func (p *Parser) matchMapPairs() ([]*MapPairNode, error) {
	pairs := []*MapPairNode{}
	for {
		location := p.currentToken.Location
		if p.currentToken.Type == lex.TOKEN_RIGHT_CURLY {
			p.nextToken()
			break
		}

		key, err := p.matchExpression(PRECEDENCE_LOWEST)
		if err != nil {
			return nil, err
		}

		if p.currentToken.Type != lex.TOKEN_COLON {
			return nil, p.unexpectedToken("colon")
		}

		p.nextToken()
		value, err := p.matchExpression(PRECEDENCE_LOWEST)
		if err != nil {
			return nil, err
		}

		pairs = append(pairs, &MapPairNode{key, value, location})

		if p.currentToken.Type == lex.TOKEN_COMMA {
			p.nextToken()
		} else if p.currentToken.Type == lex.TOKEN_RIGHT_CURLY {
			p.brackets--
			p.nextToken()
			break
		} else {
			return nil, p.unexpectedToken("comma or right curly brace")
		}
	}
	return pairs, nil
}

/**
 * Match types
 */

func (p *Parser) matchTypeNode() (TypeNode, error) {
	location := p.currentToken.Location
	if p.currentToken.Type == lex.TOKEN_LEFT_SQUARE {
		p.nextToken()
		itemTypeNode, err := p.matchTypeNode()
		if err != nil {
			return nil, err
		}

		if p.currentToken.Type != lex.TOKEN_RIGHT_SQUARE {
			return nil, p.unexpectedToken("right square bracket")
		}

		p.nextToken()
		return &ListTypeNode{itemTypeNode, location}, nil
	} else if p.currentToken.Type == lex.TOKEN_LEFT_CURLY {
		p.nextToken()
		keyTypeNode, err := p.matchTypeNode()
		if err != nil {
			return nil, err
		}

		if p.currentToken.Type != lex.TOKEN_COMMA {
			return nil, p.unexpectedToken("comma")
		}

		p.nextToken()
		valueTypeNode, err := p.matchTypeNode()
		if err != nil {
			return nil, err
		}

		if p.currentToken.Type != lex.TOKEN_RIGHT_CURLY {
			return nil, p.unexpectedToken("right curly bracket")
		}

		p.nextToken()
		return &MapTypeNode{keyTypeNode, valueTypeNode, location}, nil
	} else if p.currentToken.Type == lex.TOKEN_LEFT_PAREN {
		p.nextToken()
		typeNodes := []TypeNode{}
		for p.currentToken.Type != lex.TOKEN_RIGHT_PAREN {
			typeNode, err := p.matchTypeNode()
			if err != nil {
				return nil, err
			}
			typeNodes = append(typeNodes, typeNode)

			// TODO(2021-08-29): Disallow trailing comma.
			if p.currentToken.Type == lex.TOKEN_COMMA {
				p.nextToken()
			} else if p.currentToken.Type != lex.TOKEN_RIGHT_PAREN {
				return nil, p.unexpectedToken("comma or right parenthesis")
			}
		}
		return &TupleTypeNode{typeNodes, location}, nil
	} else if p.currentToken.Type == lex.TOKEN_SYMBOL {
		name := p.currentToken.Value
		p.nextToken()
		if p.currentToken.Type == lex.TOKEN_LESS_THAN {
			p.nextToken()
			typeNodes := []TypeNode{}

			firstType, err := p.matchTypeNode()
			if err != nil {
				return nil, err
			}
			typeNodes = append(typeNodes, firstType)

			for {
				if p.currentToken.Type == lex.TOKEN_GREATER_THAN {
					p.nextToken()
					break
				} else if p.currentToken.Type != lex.TOKEN_COMMA {
					return nil, p.unexpectedToken("comma or right angle bracket")
				}

				p.nextToken()
				subType, err := p.matchTypeNode()
				if err != nil {
					return nil, err
				}
				typeNodes = append(typeNodes, subType)
			}

			return &ParameterizedTypeNode{
				Symbol:    name,
				TypeNodes: typeNodes,
				Location:  location,
			}, nil
		} else {
			return &SymbolNode{name, location}, nil
		}
	} else {
		return nil, p.unexpectedToken("type name")
	}
}

/**
 * Match patterns
 */

func (p *Parser) matchPatternNode() (PatternNode, error) {
	location := p.currentToken.Location
	if p.currentToken.Type == lex.TOKEN_SYMBOL {
		value := p.currentToken.Value
		p.nextToken()
		if p.currentToken.Type == lex.TOKEN_LEFT_PAREN {
			patterns := []PatternNode{}

			p.nextToken()
			first_pattern, err := p.matchPatternNode()
			if err != nil {
				return nil, err
			}
			patterns = append(patterns, first_pattern)

			for {
				if p.currentToken.Type == lex.TOKEN_COMMA {
					p.nextToken()
				} else if p.currentToken.Type == lex.TOKEN_RIGHT_PAREN {
					p.nextToken()
					break
				} else if p.currentToken.Type == lex.TOKEN_ELLIPSIS {
					p.nextToken()
					if p.currentToken.Type != lex.TOKEN_RIGHT_PAREN {
						return nil, p.unexpectedToken("right parenthesis")
					}
					p.nextToken()
					break
				} else {
					return nil, p.unexpectedToken("comma or right parenthesis")
				}

				pattern, err := p.matchPatternNode()
				if err != nil {
					return nil, err
				}

				patterns = append(patterns, pattern)
			}
			return &CompoundPatternNode{
				Label:    value,
				Patterns: patterns,
				Location: location,
			}, nil
		} else {
			return &SymbolNode{Value: value, Location: location}, nil
		}
	} else {
		return nil, p.unexpectedToken("match pattern")
	}
}

/**
 * Helper functions
 */

func (p *Parser) matchArglist(terminator string) ([]ExpressionNode, error) {
	p.brackets++
	p.nextToken()

	arglist := []ExpressionNode{}
	for {
		if p.currentToken.Type == terminator {
			p.brackets--
			p.nextToken()
			break
		}

		expr, err := p.matchExpression(PRECEDENCE_LOWEST)
		if err != nil {
			return nil, err
		}
		arglist = append(arglist, expr)

		if p.currentToken.Type == lex.TOKEN_COMMA {
			p.nextToken()
		} else if p.currentToken.Type == terminator {
			p.brackets--
			p.nextToken()
			break
		} else {
			return nil, p.unexpectedToken(fmt.Sprintf("comma or %s", terminator))
		}
	}
	return arglist, nil
}

func (p *Parser) matchBlock() ([]StatementNode, error) {
	if p.currentToken.Type != lex.TOKEN_LEFT_CURLY {
		return nil, p.unexpectedToken("left curly brace")
	}

	p.nextTokenSkipNewlines()
	statements := []StatementNode{}
	for {
		if p.currentToken.Type == lex.TOKEN_RIGHT_CURLY {
			p.nextTokenSkipNewlines()
			break
		}

		statement, err := p.matchStatement()
		if err != nil {
			return nil, err
		}
		statements = append(statements, statement)
	}
	return statements, nil
}

/**
 * Parser methods
 */

func (p *Parser) nextToken() *lex.Token {
	if p.brackets > 0 {
		p.currentToken = p.lexer.NextTokenSkipNewlines()
	} else {
		p.currentToken = p.lexer.NextToken()
	}
	return p.currentToken
}

func (p *Parser) nextTokenSkipNewlines() *lex.Token {
	p.currentToken = p.lexer.NextTokenSkipNewlines()
	return p.currentToken
}

/**
 * Precedence table
 */

// TODO(2021-08-03): Double-check this order.
const (
	_ int = iota
	PRECEDENCE_LOWEST
	PRECEDENCE_TERNARY_IF
	PRECEDENCE_OR
	PRECEDENCE_AND
	PRECEDENCE_CMP
	PRECEDENCE_ADD_SUB
	PRECEDENCE_MUL_DIV
	PRECEDENCE_PREFIX
	PRECEDENCE_CALL_INDEX
	PRECEDENCE_DOT
)

var precedenceMap = map[string]int{
	lex.TOKEN_AND:                    PRECEDENCE_AND,
	lex.TOKEN_ASTERISK:               PRECEDENCE_MUL_DIV,
	lex.TOKEN_DOT:                    PRECEDENCE_DOT,
	lex.TOKEN_DOUBLE_PLUS:            PRECEDENCE_ADD_SUB,
	lex.TOKEN_EQUALS:                 PRECEDENCE_CMP,
	lex.TOKEN_GREATER_THAN:           PRECEDENCE_CMP,
	lex.TOKEN_GREATER_THAN_OR_EQUALS: PRECEDENCE_CMP,
	lex.TOKEN_IF:                     PRECEDENCE_TERNARY_IF,
	lex.TOKEN_IN:                     PRECEDENCE_CMP,
	lex.TOKEN_LEFT_PAREN:             PRECEDENCE_CALL_INDEX,
	lex.TOKEN_LEFT_SQUARE:            PRECEDENCE_CALL_INDEX,
	lex.TOKEN_LESS_THAN:              PRECEDENCE_CMP,
	lex.TOKEN_LESS_THAN_OR_EQUALS:    PRECEDENCE_CMP,
	lex.TOKEN_MINUS:                  PRECEDENCE_ADD_SUB,
	lex.TOKEN_NOT:                    PRECEDENCE_CMP, // `not` is used in the binary operator `not in`.
	lex.TOKEN_NOT_EQUALS:             PRECEDENCE_CMP,
	lex.TOKEN_OR:                     PRECEDENCE_OR,
	lex.TOKEN_PLUS:                   PRECEDENCE_ADD_SUB,
	lex.TOKEN_SLASH:                  PRECEDENCE_MUL_DIV,
}

var comparisonOperators = map[string]bool{
	"<":  true,
	"<=": true,
	">":  true,
	">=": true,
}

var compoundAssignOperators = map[string]string{
	lex.TOKEN_ASSIGN_ADD: "+",
	lex.TOKEN_ASSIGN_DIV: "/",
	lex.TOKEN_ASSIGN_MUL: "*",
	lex.TOKEN_ASSIGN_SUB: "-",
}

type ParseError struct {
	Message  string
	Location *lex.Location
}

func (e *ParseError) Error() string {
	if e.Location != nil {
		return fmt.Sprintf("%s at %s", e.Message, e.Location)
	} else {
		return e.Message
	}
}

func (p *Parser) unexpectedToken(expected string) *ParseError {
	if p.currentToken.Type == lex.TOKEN_EOF {
		// Don't change the start of this error message or multi-line parsing in the REPL
		// will break.
		return p.customError("premature end of input (expected %s)", expected)
	} else if p.currentToken.Type == lex.TOKEN_ERROR {
		return p.customError("%s", p.currentToken.Value)
	} else {
		return p.customError("expected %s, got %s", expected, p.currentToken.Type)
	}
}

func (p *Parser) customError(message string, args ...interface{}) *ParseError {
	return &ParseError{fmt.Sprintf(message, args...), p.currentToken.Location}
}
