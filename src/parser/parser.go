/**
 * Parse a Venice program into an abstract syntax tree (defined in src/ast.go).
 *
 * Each `matchXYZ` function expects that `parser.currentToken` is set to the first token
 * of the node to be matched, and after it returns `parser.currentToken` is set to one
 * past the last token of the node.
 */
package parser

import (
	"fmt"
	"github.com/iafisher/venice/src/ast"
	lexer_mod "github.com/iafisher/venice/src/lexer"
	"io/ioutil"
	"strconv"
)

type Parser struct {
	lexer        *lexer_mod.Lexer
	currentToken *lexer_mod.Token
	// Number of currently nested brackets (parentheses, curly braces, square brackets).
	brackets int
}

func NewParser() *Parser {
	return &Parser{lexer: nil, currentToken: nil, brackets: 0}
}

func (p *Parser) ParseString(input string) (*ast.File, error) {
	return p.parseGeneric("", input)
}

func (p *Parser) ParseFile(filePath string) (*ast.File, error) {
	fileContentsBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return p.parseGeneric(filePath, string(fileContentsBytes))
}

func (p *Parser) parseGeneric(filePath string, input string) (*ast.File, error) {
	p.lexer = lexer_mod.NewLexer(filePath, input)
	p.nextTokenSkipNewlines()
	statements := []ast.StatementNode{}

	for {
		statement, err := p.matchStatement()
		if err != nil {
			return nil, err
		}
		statements = append(statements, statement)

		if p.currentToken.Type == lexer_mod.TOKEN_EOF {
			break
		}
	}

	if len(statements) == 0 {
		return nil, p.customError("empty program")
	}

	return &ast.File{Statements: statements, Imports: []string{}}, nil
}

/**
 * Match statements
 */

func (p *Parser) matchStatement() (ast.StatementNode, error) {
	location := p.currentToken.Location
	var tree ast.StatementNode
	var err error
	switch p.currentToken.Type {
	case lexer_mod.TOKEN_BREAK:
		tree = &ast.BreakStatementNode{location}
		p.nextToken()
	case lexer_mod.TOKEN_CLASS:
		return p.matchClassDeclaration()
	case lexer_mod.TOKEN_CONTINUE:
		tree = &ast.ContinueStatementNode{location}
		p.nextToken()
	case lexer_mod.TOKEN_ENUM:
		return p.matchEnumDeclaration()
	case lexer_mod.TOKEN_FN:
		return p.matchFunctionDeclaration()
	case lexer_mod.TOKEN_FOR:
		return p.matchForLoop()
	case lexer_mod.TOKEN_IF:
		return p.matchIfStatement()
	case lexer_mod.TOKEN_LET:
		tree, err = p.matchLetStatement()
	case lexer_mod.TOKEN_RETURN:
		tree, err = p.matchReturnStatement()
	case lexer_mod.TOKEN_WHILE:
		return p.matchWhileLoop()
	default:
		expr, err := p.matchExpression(PRECEDENCE_LOWEST)
		if err != nil {
			return nil, err
		}

		if p.currentToken.Type == lexer_mod.TOKEN_ASSIGN {
			if symbol, ok := expr.(*ast.SymbolNode); ok {
				p.nextToken()
				assignExpr, err := p.matchExpression(PRECEDENCE_LOWEST)
				if err != nil {
					return nil, err
				}

				tree = &ast.AssignStatementNode{symbol.Value, assignExpr, location}
			} else {
				return nil, p.customError("cannot assign to non-symbol")
			}
		} else {
			tree = &ast.ExpressionStatementNode{expr, expr.GetLocation()}
		}
	}

	if err != nil {
		return nil, err
	}

	if p.currentToken.Type != lexer_mod.TOKEN_NEWLINE && p.currentToken.Type != lexer_mod.TOKEN_SEMICOLON && p.currentToken.Type != lexer_mod.TOKEN_EOF && p.currentToken.Type != lexer_mod.TOKEN_RIGHT_CURLY {
		return nil, p.customError("statement must be followed by newline or semicolon (got %s)", p.currentToken.Type)
	}

	if p.currentToken.Type == lexer_mod.TOKEN_NEWLINE || p.currentToken.Type == lexer_mod.TOKEN_SEMICOLON {
		p.nextTokenSkipNewlines()
	}

	return tree, nil
}

func (p *Parser) matchClassDeclaration() (*ast.ClassDeclarationNode, error) {
	location := p.currentToken.Location
	p.nextToken()
	if p.currentToken.Type != lexer_mod.TOKEN_SYMBOL {
		return nil, p.unexpectedToken("class name")
	}

	name := p.currentToken.Value

	var genericTypeParameter string
	p.nextToken()
	if p.currentToken.Type == lexer_mod.TOKEN_LESS_THAN {
		p.nextToken()
		if p.currentToken.Type != lexer_mod.TOKEN_SYMBOL {
			return nil, p.unexpectedToken("type parameter")
		}
		genericTypeParameter = p.currentToken.Value

		p.nextToken()
		if p.currentToken.Type != lexer_mod.TOKEN_GREATER_THAN {
			return nil, p.unexpectedToken("right angle bracket")
		}
		p.nextToken()
	}

	if p.currentToken.Type != lexer_mod.TOKEN_LEFT_CURLY {
		return nil, p.unexpectedToken("left curly brace")
	}
	p.nextTokenSkipNewlines()

	fieldNodes := []*ast.ClassFieldNode{}
	for {
		if p.currentToken.Type == lexer_mod.TOKEN_RIGHT_CURLY {
			p.nextTokenSkipNewlines()
			break
		}

		var public bool
		if p.currentToken.Type == lexer_mod.TOKEN_PUBLIC {
			public = true
		} else if p.currentToken.Type == lexer_mod.TOKEN_PRIVATE {
			public = false
		} else {
			return nil, p.unexpectedToken("field access identifier (public or private)")
		}

		p.nextToken()
		if p.currentToken.Type != lexer_mod.TOKEN_SYMBOL {
			return nil, p.unexpectedToken("symbol")
		}

		name := p.currentToken.Value

		p.nextToken()
		if p.currentToken.Type != lexer_mod.TOKEN_COLON {
			return nil, p.unexpectedToken("colon")
		}

		p.nextToken()
		fieldType, err := p.matchTypeNode()
		if err != nil {
			return nil, err
		}

		if p.currentToken.Type == lexer_mod.TOKEN_NEWLINE {
			p.nextTokenSkipNewlines()
		}

		fieldNodes = append(fieldNodes, &ast.ClassFieldNode{name, public, fieldType})
	}

	return &ast.ClassDeclarationNode{name, genericTypeParameter, fieldNodes, location}, nil
}

func (p *Parser) matchEnumDeclaration() (*ast.EnumDeclarationNode, error) {
	location := p.currentToken.Location
	p.nextToken()
	if p.currentToken.Type != lexer_mod.TOKEN_SYMBOL {
		return nil, p.unexpectedToken("symbol")
	}

	name := p.currentToken.Value

	var genericTypeParameter string
	p.nextToken()
	if p.currentToken.Type == lexer_mod.TOKEN_LESS_THAN {
		p.nextToken()
		if p.currentToken.Type != lexer_mod.TOKEN_SYMBOL {
			return nil, p.unexpectedToken("type parameter")
		}
		genericTypeParameter = p.currentToken.Value

		p.nextToken()
		if p.currentToken.Type != lexer_mod.TOKEN_GREATER_THAN {
			return nil, p.unexpectedToken("right angle bracket")
		}
		p.nextToken()
	}

	if p.currentToken.Type != lexer_mod.TOKEN_LEFT_CURLY {
		return nil, p.unexpectedToken("left curly brace")
	}
	p.nextTokenSkipNewlines()

	cases := []*ast.EnumCaseNode{}
	for {
		if p.currentToken.Type == lexer_mod.TOKEN_RIGHT_CURLY {
			p.nextTokenSkipNewlines()
			break
		}

		location := p.currentToken.Location
		if p.currentToken.Type != lexer_mod.TOKEN_SYMBOL {
			return nil, p.unexpectedToken("symbol or right curly brace")
		}

		label := p.currentToken.Value
		p.nextTokenSkipNewlines()

		if p.currentToken.Type == lexer_mod.TOKEN_LEFT_PAREN {
			p.nextTokenSkipNewlines()
			types := []ast.TypeNode{}
			for {
				if p.currentToken.Type == lexer_mod.TOKEN_RIGHT_PAREN {
					p.nextTokenSkipNewlines()
					break
				}

				typeNode, err := p.matchTypeNode()
				if err != nil {
					return nil, err
				}

				types = append(types, typeNode)

				if p.currentToken.Type == lexer_mod.TOKEN_COMMA {
					p.nextTokenSkipNewlines()
				} else if p.currentToken.Type == lexer_mod.TOKEN_RIGHT_PAREN {
					p.nextTokenSkipNewlines()
					break
				} else {
					return nil, p.unexpectedToken("comma or right parenthesis")
				}
			}

			if len(types) == 0 {
				return nil, p.customError("enum case with parentheses must include at least one type")
			}

			cases = append(cases, &ast.EnumCaseNode{label, types, location})
		} else {
			cases = append(cases, &ast.EnumCaseNode{label, []ast.TypeNode{}, location})
		}

		if p.currentToken.Type == lexer_mod.TOKEN_COMMA {
			p.nextTokenSkipNewlines()
		} else if p.currentToken.Type == lexer_mod.TOKEN_RIGHT_CURLY {
			p.nextTokenSkipNewlines()
			break
		} else {
			return nil, p.unexpectedToken("comma or right curly brace")
		}
	}

	return &ast.EnumDeclarationNode{name, genericTypeParameter, cases, location}, nil
}

func (p *Parser) matchForLoop() (*ast.ForLoopNode, error) {
	location := p.currentToken.Location
	p.nextToken()

	variables := []string{}
	for {
		if p.currentToken.Type != lexer_mod.TOKEN_SYMBOL {
			return nil, p.unexpectedToken("symbol")
		}
		variables = append(variables, p.currentToken.Value)

		p.nextToken()
		if p.currentToken.Type == lexer_mod.TOKEN_COMMA {
			p.nextToken()
		} else if p.currentToken.Type == lexer_mod.TOKEN_IN {
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

	return &ast.ForLoopNode{variables, iterable, body, location}, nil
}

func (p *Parser) matchFunctionDeclaration() (*ast.FunctionDeclarationNode, error) {
	location := p.currentToken.Location
	p.nextToken()
	if p.currentToken.Type != lexer_mod.TOKEN_SYMBOL {
		return nil, p.unexpectedToken("function name")
	}

	name := p.currentToken.Value
	p.nextToken()

	if p.currentToken.Type != lexer_mod.TOKEN_LEFT_PAREN {
		return nil, p.unexpectedToken("left parenthesis")
	}

	p.nextToken()
	params, err := p.matchFunctionParams()
	if err != nil {
		return nil, err
	}

	if p.currentToken.Type != lexer_mod.TOKEN_RIGHT_PAREN {
		return nil, p.unexpectedToken("right parenthesis")
	}

	p.nextToken()
	var returnType ast.TypeNode
	if p.currentToken.Type == lexer_mod.TOKEN_ARROW {
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

	return &ast.FunctionDeclarationNode{name, params, returnType, body, location}, nil
}

func (p *Parser) matchFunctionParams() ([]*ast.FunctionParamNode, error) {
	params := []*ast.FunctionParamNode{}
	for p.currentToken.Type != lexer_mod.TOKEN_RIGHT_PAREN {
		if p.currentToken.Type != lexer_mod.TOKEN_SYMBOL {
			return nil, p.unexpectedToken("function parameter name")
		}
		name := p.currentToken.Value
		location := p.currentToken.Location

		p.nextToken()
		if p.currentToken.Type != lexer_mod.TOKEN_COLON {
			return nil, p.unexpectedToken("colon")
		}

		p.nextToken()
		paramType, err := p.matchTypeNode()
		if err != nil {
			return nil, err
		}

		params = append(params, &ast.FunctionParamNode{name, paramType, location})

		if p.currentToken.Type == lexer_mod.TOKEN_RIGHT_PAREN {
			break
		} else if p.currentToken.Type == lexer_mod.TOKEN_COMMA {
			p.nextToken()
		} else {
			return nil, p.unexpectedToken("comma or right parenthesis")
		}
	}
	return params, nil
}

func (p *Parser) matchIfStatement() (*ast.IfStatementNode, error) {
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

	var elseClause []ast.StatementNode
	clauses := []*ast.IfClauseNode{&ast.IfClauseNode{condition, trueClauseStatements}}
	for p.currentToken.Type == lexer_mod.TOKEN_ELSE {
		p.nextToken()

		if p.currentToken.Type == lexer_mod.TOKEN_LEFT_CURLY {
			elseClause, err = p.matchBlock()
			if err != nil {
				return nil, err
			}
			break
		} else if p.currentToken.Type == lexer_mod.TOKEN_IF {
			p.nextToken()
			condition, err = p.matchExpression(PRECEDENCE_LOWEST)
			if err != nil {
				return nil, err
			}

			body, err := p.matchBlock()
			if err != nil {
				return nil, err
			}

			clauses = append(clauses, &ast.IfClauseNode{condition, body})
		} else {
			return nil, p.unexpectedToken("`else if` or `else`")
		}
	}

	return &ast.IfStatementNode{clauses, elseClause, location}, nil
}

func (p *Parser) matchLetStatement() (*ast.LetStatementNode, error) {
	location := p.currentToken.Location
	p.nextToken()
	if p.currentToken.Type != lexer_mod.TOKEN_SYMBOL {
		return nil, p.unexpectedToken("symbol")
	}
	symbol := p.currentToken.Value

	p.nextToken()
	if p.currentToken.Type != lexer_mod.TOKEN_ASSIGN {
		return nil, p.unexpectedToken("equals sign")
	}

	p.nextToken()
	expr, err := p.matchExpression(PRECEDENCE_LOWEST)
	if err != nil {
		return nil, err
	}

	return &ast.LetStatementNode{symbol, expr, location}, nil
}

func (p *Parser) matchReturnStatement() (*ast.ReturnStatementNode, error) {
	location := p.currentToken.Location
	p.nextToken()
	if p.currentToken.Type == lexer_mod.TOKEN_NEWLINE || p.currentToken.Type == lexer_mod.TOKEN_SEMICOLON {
		return &ast.ReturnStatementNode{nil, location}, nil
	}

	expr, err := p.matchExpression(PRECEDENCE_LOWEST)
	if err != nil {
		return nil, err
	}
	return &ast.ReturnStatementNode{expr, location}, nil
}

func (p *Parser) matchWhileLoop() (*ast.WhileLoopNode, error) {
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

	return &ast.WhileLoopNode{condition, body, location}, nil
}

/**
 * Match expressions
 */

func (p *Parser) matchExpression(precedence int) (ast.ExpressionNode, error) {
	location := p.currentToken.Location
	expr, err := p.matchPrefix()
	if err != nil {
		return nil, err
	}

	for {
		if infixPrecedence, ok := precedenceMap[p.currentToken.Type]; ok {
			if precedence < infixPrecedence {
				if p.currentToken.Type == lexer_mod.TOKEN_LEFT_PAREN {
					arglist, err := p.matchArglist(lexer_mod.TOKEN_RIGHT_PAREN)
					if err != nil {
						return nil, err
					}

					expr = &ast.CallNode{expr, arglist, location}
				} else if p.currentToken.Type == lexer_mod.TOKEN_LEFT_SQUARE {
					p.nextToken()
					indexExpr, err := p.matchExpression(PRECEDENCE_LOWEST)
					if err != nil {
						return nil, err
					}

					if p.currentToken.Type != lexer_mod.TOKEN_RIGHT_SQUARE {
						return nil, p.unexpectedToken("right square bracket")
					}
					p.nextToken()

					expr = &ast.IndexNode{expr, indexExpr, location}
				} else if p.currentToken.Type == lexer_mod.TOKEN_DOT {
					p.nextToken()
					if p.currentToken.Type == lexer_mod.TOKEN_SYMBOL {
						expr = &ast.FieldAccessNode{expr, p.currentToken.Value, location}
					} else if p.currentToken.Type == lexer_mod.TOKEN_INT {
						index, err := strconv.ParseInt(p.currentToken.Value, 10, 0)
						if err != nil {
							return nil, p.customError("could not convert integer token")
						}
						expr = &ast.TupleFieldAccessNode{expr, int(index), location}
					} else {
						return nil, p.customError("right-hand side of dot must be a symbol")
					}

					p.nextToken()
				} else if p.currentToken.Type == lexer_mod.TOKEN_IF {
					p.nextToken()
					conditionExpr, err := p.matchExpression(PRECEDENCE_LOWEST)
					if err != nil {
						return nil, err
					}

					if p.currentToken.Type != lexer_mod.TOKEN_ELSE {
						return nil, p.customError("`else`")
					}

					p.nextToken()
					elseExpr, err := p.matchExpression(PRECEDENCE_LOWEST)
					if err != nil {
						return nil, err
					}

					expr = &ast.TernaryIfNode{
						Condition:   conditionExpr,
						TrueClause:  expr,
						FalseClause: elseExpr,
						Location:    location,
					}
				} else if p.currentToken.Type == lexer_mod.TOKEN_NOT {
					unaryLocation := p.currentToken.Location
					p.nextToken()
					if p.currentToken.Type != lexer_mod.TOKEN_IN {
						return nil, p.customError("expected `in` after `not` in infix position")
					}

					p.nextToken()
					right, err := p.matchExpression(infixPrecedence)
					if err != nil {
						return nil, err
					}

					expr = &ast.UnaryNode{
						Operator: "not",
						Expr: &ast.InfixNode{
							Operator: "in",
							Left:     expr,
							Right:    right,
							Location: location,
						},
						Location: unaryLocation,
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

func (p *Parser) matchInfix(left ast.ExpressionNode, precedence int) (ast.ExpressionNode, error) {
	operator := p.currentToken.Value
	p.nextToken()
	right, err := p.matchExpression(precedence)
	if err != nil {
		return nil, err
	}

	// If the infix expression is a double comparison like `0 <= x < 100`, then refactor
	// it into `0 <= x and x < 100`.
	leftInfix, ok := left.(*ast.InfixNode)
	if ok {
		if comparisonOperators[operator] && comparisonOperators[leftInfix.Operator] {
			return &ast.InfixNode{
				Operator: "and",
				Left:     left,
				Right: &ast.InfixNode{
					Operator: operator,
					Left:     leftInfix.Right,
					Right:    right,
					Location: right.GetLocation(),
				},
				Location: left.GetLocation(),
			}, nil
		}
	}

	return &ast.InfixNode{operator, left, right, left.GetLocation()}, nil
}

func (p *Parser) matchMapPairs() ([]*ast.MapPairNode, error) {
	pairs := []*ast.MapPairNode{}
	for {
		location := p.currentToken.Location
		if p.currentToken.Type == lexer_mod.TOKEN_RIGHT_CURLY {
			p.nextToken()
			break
		}

		key, err := p.matchExpression(PRECEDENCE_LOWEST)
		if err != nil {
			return nil, err
		}

		if p.currentToken.Type != lexer_mod.TOKEN_COLON {
			return nil, p.unexpectedToken("colon")
		}

		p.nextToken()
		value, err := p.matchExpression(PRECEDENCE_LOWEST)
		if err != nil {
			return nil, err
		}

		pairs = append(pairs, &ast.MapPairNode{key, value, location})

		if p.currentToken.Type == lexer_mod.TOKEN_COMMA {
			p.nextToken()
		} else if p.currentToken.Type == lexer_mod.TOKEN_RIGHT_CURLY {
			p.brackets--
			p.nextToken()
			break
		} else {
			return nil, p.unexpectedToken("comma or right curly brace")
		}
	}
	return pairs, nil
}

func (p *Parser) matchPrefix() (ast.ExpressionNode, error) {
	location := p.currentToken.Location
	switch p.currentToken.Type {
	case lexer_mod.TOKEN_CHARACTER:
		value := p.currentToken.Value
		p.nextToken()
		return &ast.CharacterNode{value[0], location}, nil
	case lexer_mod.TOKEN_FALSE:
		p.nextToken()
		return &ast.BooleanNode{false, location}, nil
	case lexer_mod.TOKEN_INT:
		token := p.currentToken
		p.nextToken()
		value, err := strconv.ParseInt(token.Value, 0, 0)
		if err != nil {
			return nil, p.customError("could not convert integer token")
		}
		return &ast.IntegerNode{int(value), location}, nil
	case lexer_mod.TOKEN_LEFT_CURLY:
		p.brackets++
		p.nextToken()
		pairs, err := p.matchMapPairs()
		if err != nil {
			return nil, err
		}
		return &ast.MapNode{pairs, location}, nil
	case lexer_mod.TOKEN_LEFT_PAREN:
		values, err := p.matchArglist(lexer_mod.TOKEN_RIGHT_PAREN)
		if err != nil {
			return nil, err
		}

		if len(values) == 1 {
			return values[0], nil
		} else {
			return &ast.TupleNode{values, location}, nil
		}
	case lexer_mod.TOKEN_LEFT_SQUARE:
		values, err := p.matchArglist(lexer_mod.TOKEN_RIGHT_SQUARE)
		if err != nil {
			return nil, err
		}
		return &ast.ListNode{values, location}, nil
	case lexer_mod.TOKEN_MINUS, lexer_mod.TOKEN_NOT:
		operator := p.currentToken.Value
		p.nextToken()
		expr, err := p.matchExpression(PRECEDENCE_PREFIX)
		if err != nil {
			return nil, err
		}
		return &ast.UnaryNode{operator, expr, location}, nil
	case lexer_mod.TOKEN_STRING:
		value := p.currentToken.Value
		p.nextToken()
		return &ast.StringNode{value, location}, nil
	case lexer_mod.TOKEN_SYMBOL:
		value := p.currentToken.Value
		p.nextToken()

		if p.currentToken.Type == lexer_mod.TOKEN_DOUBLE_COLON {
			p.nextToken()
			if p.currentToken.Type != lexer_mod.TOKEN_SYMBOL {
				return nil, p.unexpectedToken("symbol")
			}
			secondValue := p.currentToken.Value
			p.nextToken()
			return &ast.EnumSymbolNode{value, secondValue, location}, nil
		} else {
			return &ast.SymbolNode{value, location}, nil
		}
	case lexer_mod.TOKEN_TRUE:
		p.nextToken()
		return &ast.BooleanNode{true, location}, nil
	default:
		return nil, p.unexpectedToken("start of expression")
	}
}

/**
 * Match types
 */

func (p *Parser) matchTypeNode() (ast.TypeNode, error) {
	if p.currentToken.Type != lexer_mod.TOKEN_SYMBOL {
		return nil, p.unexpectedToken("type name")
	}

	name := p.currentToken.Value
	location := p.currentToken.Location
	p.nextToken()
	return &ast.SimpleTypeNode{name, location}, nil
}

/**
 * Helper functions
 */

func (p *Parser) matchArglist(terminator string) ([]ast.ExpressionNode, error) {
	p.brackets++
	p.nextToken()

	arglist := []ast.ExpressionNode{}
	for {
		if p.currentToken.Type == terminator {
			p.nextToken()
			break
		}

		expr, err := p.matchExpression(PRECEDENCE_LOWEST)
		if err != nil {
			return nil, err
		}
		arglist = append(arglist, expr)

		if p.currentToken.Type == lexer_mod.TOKEN_COMMA {
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

func (p *Parser) matchBlock() ([]ast.StatementNode, error) {
	if p.currentToken.Type != lexer_mod.TOKEN_LEFT_CURLY {
		return nil, p.unexpectedToken("left curly brace")
	}

	p.nextTokenSkipNewlines()
	statements := []ast.StatementNode{}
	for {
		if p.currentToken.Type == lexer_mod.TOKEN_RIGHT_CURLY {
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

func (p *Parser) nextToken() *lexer_mod.Token {
	if p.brackets > 0 {
		p.currentToken = p.lexer.NextTokenSkipNewlines()
	} else {
		p.currentToken = p.lexer.NextToken()
	}
	return p.currentToken
}

func (p *Parser) nextTokenSkipNewlines() *lexer_mod.Token {
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
	lexer_mod.TOKEN_AND:                    PRECEDENCE_AND,
	lexer_mod.TOKEN_ASTERISK:               PRECEDENCE_MUL_DIV,
	lexer_mod.TOKEN_DOT:                    PRECEDENCE_DOT,
	lexer_mod.TOKEN_DOUBLE_PLUS:            PRECEDENCE_ADD_SUB,
	lexer_mod.TOKEN_EQUALS:                 PRECEDENCE_CMP,
	lexer_mod.TOKEN_GREATER_THAN:           PRECEDENCE_CMP,
	lexer_mod.TOKEN_GREATER_THAN_OR_EQUALS: PRECEDENCE_CMP,
	lexer_mod.TOKEN_IF:                     PRECEDENCE_TERNARY_IF,
	lexer_mod.TOKEN_IN:                     PRECEDENCE_CMP,
	lexer_mod.TOKEN_LEFT_PAREN:             PRECEDENCE_CALL_INDEX,
	lexer_mod.TOKEN_LEFT_SQUARE:            PRECEDENCE_CALL_INDEX,
	lexer_mod.TOKEN_LESS_THAN:              PRECEDENCE_CMP,
	lexer_mod.TOKEN_LESS_THAN_OR_EQUALS:    PRECEDENCE_CMP,
	lexer_mod.TOKEN_MINUS:                  PRECEDENCE_ADD_SUB,
	lexer_mod.TOKEN_NOT:                    PRECEDENCE_CMP, // `not` is used in the binary operator `not in`.
	lexer_mod.TOKEN_NOT_EQUALS:             PRECEDENCE_CMP,
	lexer_mod.TOKEN_OR:                     PRECEDENCE_OR,
	lexer_mod.TOKEN_PLUS:                   PRECEDENCE_ADD_SUB,
	lexer_mod.TOKEN_SLASH:                  PRECEDENCE_MUL_DIV,
}

var comparisonOperators = map[string]bool{
	"<":  true,
	"<=": true,
	">":  true,
	">=": true,
}

type ParseError struct {
	Message  string
	Location *lexer_mod.Location
}

func (e *ParseError) Error() string {
	if e.Location != nil {
		return fmt.Sprintf("%s at %s", e.Message, e.Location)
	} else {
		return e.Message
	}
}

func (p *Parser) unexpectedToken(expected string) *ParseError {
	if p.currentToken.Type == lexer_mod.TOKEN_EOF {
		// Don't change the start of this error message or multi-line parsing in the REPL will break.
		return p.customError("premature end of input (expected %s)", expected)
	} else if p.currentToken.Type == lexer_mod.TOKEN_ERROR {
		return p.customError("%s", p.currentToken.Value)
	} else {
		return p.customError("expected %s, got %s", expected, p.currentToken.Type)
	}
}

func (p *Parser) customError(message string, args ...interface{}) *ParseError {
	return &ParseError{fmt.Sprintf(message, args...), p.currentToken.Location}
}
