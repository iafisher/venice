package main

import (
	"fmt"
	"strconv"
)

type Expression interface {
	expressionNode()
}

type Statement interface {
	statementNode()
}

type LetStatementNode struct {
	Symbol string
	Expr   Expression
}

func (n *LetStatementNode) statementNode() {}

type IfStatementNode struct {
	Condition   Expression
	TrueClause  []Statement
	FalseClause []Statement
}

func (n *IfStatementNode) statementNode() {}

type CallNode struct {
	Function Expression
	Args     []Expression
}

func (n *CallNode) expressionNode() {}

type IndexNode struct {
	Expr  Expression
	Index Expression
}

func (n *IndexNode) expressionNode() {}

type InfixNode struct {
	Operator string
	Left     Expression
	Right    Expression
}

func (n *InfixNode) expressionNode() {}

type ListNode struct {
	Values []Expression
}

func (n *ListNode) expressionNode() {}

type MapNode struct {
	Pairs []*MapPairNode
}

func (n *MapNode) expressionNode() {}

type MapPairNode struct {
	Key   Expression
	Value Expression
}

func (n *MapPairNode) expressionNode() {}

type IntegerNode struct {
	Value int
}

func (n *IntegerNode) expressionNode() {}

type SymbolNode struct {
	Value string
}

func (n *SymbolNode) expressionNode() {}

type StringNode struct {
	Value string
}

func (n *StringNode) expressionNode() {}

type BooleanNode struct {
	Value bool
}

func (n *BooleanNode) expressionNode() {}

type ProgramNode struct {
	Statements []Statement
}

type ExpressionStatementNode struct {
	Expression Expression
}

func (n *ExpressionStatementNode) statementNode() {}

type Parser struct {
	lexer        *Lexer
	currentToken *Token
}

func NewParser(l *Lexer) *Parser {
	parser := &Parser{l, nil}
	parser.nextToken()
	return parser
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

func (p *Parser) Parse() (*ProgramNode, error) {
	statements := []Statement{}

	for {
		statement, err := p.matchStatement()
		if err != nil {
			return nil, err
		}
		statements = append(statements, statement)
		p.lexer.skipWhitespace()

		if p.currentToken.Type == TOKEN_EOF {
			break
		}
	}

	if len(statements) == 0 {
		return nil, p.customError("empty program")
	}

	return &ProgramNode{statements}, nil
}

func (p *Parser) matchStatement() (Statement, error) {
	switch p.currentToken.Type {
	case TOKEN_IF:
		return p.matchIfStatement()
	case TOKEN_LET:
		return p.matchLetStatement()
	default:
		expr, err := p.matchExpression(PRECEDENCE_LOWEST)
		if err != nil {
			return nil, err
		}
		return &ExpressionStatementNode{expr}, nil
	}
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

func (p *Parser) matchBlock() ([]Statement, error) {
	if p.currentToken.Type != TOKEN_LEFT_CURLY {
		return nil, p.unexpectedToken("left curly brace")
	}

	p.nextToken()
	statements := []Statement{}
	for {
		if p.currentToken.Type == TOKEN_RIGHT_CURLY {
			p.nextToken()
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

func (p *Parser) matchExpression(precedence int) (Expression, error) {
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

func (p *Parser) matchPrefix() (Expression, error) {
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
		p.nextToken()
		pairs, err := p.matchMapPairs()
		if err != nil {
			return nil, err
		}
		return &MapNode{pairs}, nil
	case TOKEN_LEFT_PAREN:
		p.nextToken()
		expr, err := p.matchExpression(PRECEDENCE_LOWEST)
		if err != nil {
			return nil, err
		}
		if p.currentToken.Type != TOKEN_RIGHT_PAREN {
			return nil, p.unexpectedToken("right parenthesis")
		}
		p.nextToken()
		return expr, nil
	case TOKEN_LEFT_SQUARE:
		p.nextToken()
		values, err := p.matchArglist(TOKEN_RIGHT_SQUARE)
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

func (p *Parser) matchArglist(terminator string) ([]Expression, error) {
	arglist := []Expression{}
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

func (p *Parser) matchInfix(left Expression, precedence int) (Expression, error) {
	operator := p.currentToken.Value
	p.nextToken()
	right, err := p.matchExpression(precedence)
	if err != nil {
		return nil, err
	}
	return &InfixNode{operator, left, right}, nil
}

func (p *Parser) nextToken() *Token {
	p.currentToken = p.lexer.NextToken()
	return p.currentToken
}

func (p *Parser) unexpectedToken(expected string) *ParseError {
	return p.customError(fmt.Sprintf("expected %s, got %s", expected, p.currentToken.Type))
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
	TOKEN_ASTERISK:    PRECEDENCE_MUL_DIV,
	TOKEN_EQ:          PRECEDENCE_CMP,
	TOKEN_LEFT_PAREN:  PRECEDENCE_CALL_INDEX,
	TOKEN_LEFT_SQUARE: PRECEDENCE_CALL_INDEX,
	TOKEN_MINUS:       PRECEDENCE_ADD_SUB,
	TOKEN_PLUS:        PRECEDENCE_ADD_SUB,
	TOKEN_SLASH:       PRECEDENCE_MUL_DIV,
}
