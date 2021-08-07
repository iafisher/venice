package main

import (
	"fmt"
	"strconv"
)

type Parser struct {
	lexer        *Lexer
	currentToken *Token
	// Number of currently nested brackets (parentheses, curly braces, square brackets).
	brackets int
}

func NewParser(l *Lexer) *Parser {
	parser := &Parser{l, nil, 0}
	parser.nextTokenSkipNewlines()
	return parser
}

func (p *Parser) Parse() (*ProgramNode, error) {
	statements := []StatementNode{}

	for {
		statement, err := p.matchStatement()
		if err != nil {
			return nil, err
		}
		statements = append(statements, statement)

		if p.currentToken.Type == TOKEN_EOF {
			break
		}
	}

	if len(statements) == 0 {
		return nil, p.customError("empty program")
	}

	return &ProgramNode{statements}, nil
}

func (p *Parser) matchStatement() (StatementNode, error) {
	var tree StatementNode
	var err error
	switch p.currentToken.Type {
	case TOKEN_BREAK:
		tree = &BreakStatementNode{}
	case TOKEN_CONTINUE:
		tree = &ContinueStatementNode{}
	case TOKEN_FN:
		return p.matchFunctionDeclaration()
	case TOKEN_IF:
		return p.matchIfStatement()
	case TOKEN_LET:
		tree, err = p.matchLetStatement()
	case TOKEN_RETURN:
		tree, err = p.matchReturnStatement()
	case TOKEN_WHILE:
		return p.matchWhileLoop()
	default:
		expr, err := p.matchExpression(PRECEDENCE_LOWEST)
		if err != nil {
			return nil, err
		}

		if p.currentToken.Type == TOKEN_ASSIGN {
			if symbol, ok := expr.(*SymbolNode); ok {
				p.nextToken()
				assignExpr, err := p.matchExpression(PRECEDENCE_LOWEST)
				if err != nil {
					return nil, err
				}

				tree = &AssignStatementNode{symbol.Value, assignExpr}
			} else {
				return nil, p.customError("cannot assign to non-symbol")
			}
		} else {
			tree = &ExpressionStatementNode{expr}
		}
	}

	if err != nil {
		return nil, err
	}

	if p.currentToken.Type != TOKEN_NEWLINE && p.currentToken.Type != TOKEN_SEMICOLON && p.currentToken.Type != TOKEN_EOF && p.currentToken.Type != TOKEN_RIGHT_CURLY {
		return nil, p.customError("statement must be followed by newline or semicolon")
	}

	if p.currentToken.Type == TOKEN_NEWLINE || p.currentToken.Type == TOKEN_SEMICOLON {
		p.nextTokenSkipNewlines()
	}

	return tree, nil
}

func (p *Parser) matchReturnStatement() (*ReturnStatementNode, error) {
	p.nextToken()
	if p.currentToken.Type == TOKEN_NEWLINE || p.currentToken.Type == TOKEN_SEMICOLON {
		return &ReturnStatementNode{nil}, nil
	}

	expr, err := p.matchExpression(PRECEDENCE_LOWEST)
	if err != nil {
		return nil, err
	}
	return &ReturnStatementNode{expr}, nil
}

func (p *Parser) matchFunctionDeclaration() (*FunctionDeclarationNode, error) {
	p.nextToken()
	if p.currentToken.Type != TOKEN_SYMBOL {
		return nil, p.unexpectedToken("function name")
	}

	name := p.currentToken.Value
	p.nextToken()

	if p.currentToken.Type != TOKEN_LEFT_PAREN {
		return nil, p.unexpectedToken("left parenthesis")
	}

	p.nextToken()
	params, err := p.matchFunctionParams()
	if err != nil {
		return nil, err
	}

	if p.currentToken.Type != TOKEN_RIGHT_PAREN {
		return nil, p.unexpectedToken("right parenthesis")
	}

	p.nextToken()
	var returnType TypeNode
	if p.currentToken.Type == TOKEN_ARROW {
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

	return &FunctionDeclarationNode{name, params, returnType, body}, nil
}

func (p *Parser) matchFunctionParams() ([]*FunctionParamNode, error) {
	params := []*FunctionParamNode{}
	for p.currentToken.Type != TOKEN_RIGHT_PAREN {
		if p.currentToken.Type != TOKEN_SYMBOL {
			return nil, p.unexpectedToken("function parameter name")
		}
		name := p.currentToken.Value

		p.nextToken()
		if p.currentToken.Type != TOKEN_COLON {
			return nil, p.unexpectedToken("colon")
		}

		p.nextToken()
		paramType, err := p.matchTypeNode()
		if err != nil {
			return nil, err
		}

		params = append(params, &FunctionParamNode{name, paramType})

		if p.currentToken.Type == TOKEN_RIGHT_PAREN {
			break
		} else if p.currentToken.Type == TOKEN_COMMA {
			p.nextToken()
		} else {
			return nil, p.unexpectedToken("comma or right parenthesis")
		}
	}
	return params, nil
}

func (p *Parser) matchTypeNode() (TypeNode, error) {
	if p.currentToken.Type != TOKEN_SYMBOL {
		return nil, p.unexpectedToken("type name")
	}

	name := p.currentToken.Value
	p.nextToken()
	return &SimpleTypeNode{name}, nil
}

func (p *Parser) matchWhileLoop() (*WhileLoopNode, error) {
	p.nextToken()
	condition, err := p.matchExpression(PRECEDENCE_LOWEST)
	if err != nil {
		return nil, err
	}

	body, err := p.matchBlock()
	if err != nil {
		return nil, err
	}

	return &WhileLoopNode{condition, body}, nil
}

func (p *Parser) matchIfStatement() (*IfStatementNode, error) {
	p.nextToken()
	condition, err := p.matchExpression(PRECEDENCE_LOWEST)
	if err != nil {
		return nil, err
	}

	trueClauseStatements, err := p.matchBlock()
	if err != nil {
		return nil, err
	}

	if p.currentToken.Type == TOKEN_ELSE {
		p.nextToken()
		falseClauseStatements, err := p.matchBlock()
		if err != nil {
			return nil, err
		}
		return &IfStatementNode{condition, trueClauseStatements, falseClauseStatements}, nil
	}

	return &IfStatementNode{condition, trueClauseStatements, nil}, nil
}

func (p *Parser) matchLetStatement() (*LetStatementNode, error) {
	p.nextToken()
	if p.currentToken.Type != TOKEN_SYMBOL {
		return nil, p.unexpectedToken("symbol")
	}
	symbol := p.currentToken.Value

	p.nextToken()
	if p.currentToken.Type != TOKEN_ASSIGN {
		return nil, p.unexpectedToken("equals sign")
	}

	p.nextToken()
	expr, err := p.matchExpression(PRECEDENCE_LOWEST)
	if err != nil {
		return nil, err
	}

	return &LetStatementNode{symbol, expr}, nil
}

func (p *Parser) matchBlock() ([]StatementNode, error) {
	if p.currentToken.Type != TOKEN_LEFT_CURLY {
		return nil, p.unexpectedToken("left curly brace")
	}

	p.nextTokenSkipNewlines()
	statements := []StatementNode{}
	for {
		if p.currentToken.Type == TOKEN_RIGHT_CURLY {
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

func (p *Parser) matchExpression(precedence int) (ExpressionNode, error) {
	expr, err := p.matchPrefix()
	if err != nil {
		return nil, err
	}

	for {
		if infixPrecedence, ok := precedenceMap[p.currentToken.Type]; ok {
			if precedence < infixPrecedence {
				if p.currentToken.Type == TOKEN_LEFT_PAREN {
					p.nextToken()
					arglist, err := p.matchArglist(TOKEN_RIGHT_PAREN)
					if err != nil {
						return nil, err
					}

					expr = &CallNode{expr, arglist}
				} else if p.currentToken.Type == TOKEN_LEFT_SQUARE {
					p.nextToken()
					indexExpr, err := p.matchExpression(PRECEDENCE_LOWEST)
					if err != nil {
						return nil, err
					}

					if p.currentToken.Type != TOKEN_RIGHT_SQUARE {
						return nil, p.unexpectedToken("right square bracket")
					}
					p.nextToken()

					expr = &IndexNode{expr, indexExpr}
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

func (p *Parser) matchPrefix() (ExpressionNode, error) {
	switch p.currentToken.Type {
	case TOKEN_FALSE:
		p.nextToken()
		return &BooleanNode{false}, nil
	case TOKEN_INT:
		token := p.currentToken
		p.nextToken()
		value, err := strconv.ParseInt(token.Value, 10, 0)
		if err != nil {
			return nil, p.customError("could not convert integer token")
		}
		return &IntegerNode{int(value)}, nil
	case TOKEN_LEFT_CURLY:
		p.brackets++
		p.nextToken()
		pairs, err := p.matchMapPairs()
		p.brackets--
		if err != nil {
			return nil, err
		}
		return &MapNode{pairs}, nil
	case TOKEN_LEFT_PAREN:
		p.brackets++
		p.nextToken()
		expr, err := p.matchExpression(PRECEDENCE_LOWEST)
		p.brackets--
		if err != nil {
			return nil, err
		}
		if p.currentToken.Type != TOKEN_RIGHT_PAREN {
			return nil, p.unexpectedToken("right parenthesis")
		}
		p.nextToken()
		return expr, nil
	case TOKEN_LEFT_SQUARE:
		p.brackets++
		p.nextToken()
		values, err := p.matchArglist(TOKEN_RIGHT_SQUARE)
		p.brackets--
		if err != nil {
			return nil, err
		}
		return &ListNode{values}, nil
	case TOKEN_STRING:
		token := p.currentToken
		p.nextToken()
		return &StringNode{token.Value}, nil
	case TOKEN_SYMBOL:
		token := p.currentToken
		p.nextToken()
		return &SymbolNode{token.Value}, nil
	case TOKEN_TRUE:
		p.nextToken()
		return &BooleanNode{true}, nil
	default:
		return nil, p.unexpectedToken("start of expression")
	}
}

func (p *Parser) matchMapPairs() ([]*MapPairNode, error) {
	pairs := []*MapPairNode{}
	for {
		if p.currentToken.Type == TOKEN_RIGHT_CURLY {
			p.nextToken()
			break
		}

		key, err := p.matchExpression(PRECEDENCE_LOWEST)
		if err != nil {
			return nil, err
		}

		if p.currentToken.Type != TOKEN_COLON {
			return nil, p.unexpectedToken("colon")
		}

		p.nextToken()
		value, err := p.matchExpression(PRECEDENCE_LOWEST)
		if err != nil {
			return nil, err
		}

		pairs = append(pairs, &MapPairNode{key, value})

		if p.currentToken.Type == TOKEN_COMMA {
			p.nextToken()
		} else if p.currentToken.Type == TOKEN_RIGHT_CURLY {
			p.nextToken()
			break
		} else {
			return nil, p.unexpectedToken("comma or right curly brace")
		}
	}
	return pairs, nil
}

func (p *Parser) matchArglist(terminator string) ([]ExpressionNode, error) {
	arglist := []ExpressionNode{}
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

		if p.currentToken.Type == TOKEN_COMMA {
			p.nextToken()
		} else if p.currentToken.Type == terminator {
			p.nextToken()
			break
		} else {
			return nil, p.unexpectedToken(fmt.Sprintf("comma or %s", terminator))
		}
	}
	return arglist, nil
}

func (p *Parser) matchInfix(left ExpressionNode, precedence int) (ExpressionNode, error) {
	operator := p.currentToken.Value
	p.nextToken()
	right, err := p.matchExpression(precedence)
	if err != nil {
		return nil, err
	}
	return &InfixNode{operator, left, right}, nil
}

func (p *Parser) nextToken() *Token {
	if p.brackets > 0 {
		p.currentToken = p.lexer.NextTokenSkipNewlines()
	} else {
		p.currentToken = p.lexer.NextToken()
	}
	return p.currentToken
}

func (p *Parser) nextTokenSkipNewlines() *Token {
	p.currentToken = p.lexer.NextTokenSkipNewlines()
	return p.currentToken
}

type ParseError struct {
	Message  string
	Location *Location
}

func (e *ParseError) Error() string {
	if e.Location != nil {
		return fmt.Sprintf("%s at line %d, column %d", e.Message, e.Location.Line, e.Location.Column)
	} else {
		return e.Message
	}
}

func (p *Parser) unexpectedToken(expected string) *ParseError {
	if p.currentToken.Type == TOKEN_EOF {
		// Don't change the start of this error message or multi-line parsing in the REPL will break.
		return p.customError(fmt.Sprintf("premature end of input (expected %s)", expected))
	} else {
		return p.customError(fmt.Sprintf("expected %s, got %s", expected, p.currentToken.Type))
	}
}

func (p *Parser) customError(message string) *ParseError {
	return &ParseError{message, p.currentToken.Loc}
}

// TODO(2021-08-03): Double-check this order.
const (
	_ int = iota
	PRECEDENCE_LOWEST
	PRECEDENCE_OR
	PRECEDENCE_AND
	PRECEDENCE_CMP
	PRECEDENCE_ADD_SUB
	PRECEDENCE_MUL_DIV
	PRECEDENCE_PREFIX
	PRECEDENCE_CALL_INDEX
)

var precedenceMap = map[string]int{
	TOKEN_ASTERISK:               PRECEDENCE_MUL_DIV,
	TOKEN_EQUALS:                 PRECEDENCE_CMP,
	TOKEN_GREATER_THAN:           PRECEDENCE_CMP,
	TOKEN_GREATER_THAN_OR_EQUALS: PRECEDENCE_CMP,
	TOKEN_LEFT_PAREN:             PRECEDENCE_CALL_INDEX,
	TOKEN_LEFT_SQUARE:            PRECEDENCE_CALL_INDEX,
	TOKEN_LESS_THAN:              PRECEDENCE_CMP,
	TOKEN_LESS_THAN_OR_EQUALS:    PRECEDENCE_CMP,
	TOKEN_MINUS:                  PRECEDENCE_ADD_SUB,
	TOKEN_NOT_EQUALS:             PRECEDENCE_CMP,
	TOKEN_PLUS:                   PRECEDENCE_ADD_SUB,
	TOKEN_SLASH:                  PRECEDENCE_MUL_DIV,
}
