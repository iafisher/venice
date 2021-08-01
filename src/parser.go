package main

import "strconv"

type Expression interface {
	expressionNode()
}

type Statement interface {
	statementNode()
}

type Declaration interface {
	declarationNode()
}

type InfixNode struct {
	Operator string
	Left     Expression
	Right    Expression
}

func (n *InfixNode) expressionNode() {}

type IntegerNode struct {
	Value int64
}

func (n *IntegerNode) expressionNode() {}

type SymbolNode struct {
	Value string
}

func (n *SymbolNode) expressionNode() {}

type ProgramNode struct {
	Declarations []Declaration
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

func (p *Parser) Parse() *ProgramNode {
	return nil
}

func (p *Parser) ParseExpression() (Expression, bool) {
	expr, ok := p.matchExpression(PRECEDENCE_LOWEST)
	if !ok {
		return nil, false
	}
	p.lexer.skipWhitespace()
	if p.currentToken.Type != TOKEN_EOF {
		return nil, false
	}
	return expr, true
}

func (p *Parser) matchExpression(precedence int) (Expression, bool) {
	expr, ok := p.matchPrefix()
	if !ok {
		return nil, false
	}

	for {
		if infixPrecedence, ok := precedenceMap[p.currentToken.Type]; ok {
			if precedence < infixPrecedence {
				expr, ok = p.matchInfix(expr, infixPrecedence)
				if !ok {
					return nil, false
				}
			} else {
				break
			}
		} else {
			break
		}
	}

	return expr, true
}

func (p *Parser) matchPrefix() (Expression, bool) {
	switch p.currentToken.Type {
	case TOKEN_INT:
		token := p.currentToken
		p.nextToken()
		value, err := strconv.ParseInt(token.Value, 10, 64)
		if err != nil {
			return nil, false
		}
		return &IntegerNode{value}, true
	case TOKEN_SYMBOL:
		token := p.currentToken
		p.nextToken()
		return &SymbolNode{token.Value}, true
	case TOKEN_LEFT_PAREN:
		p.nextToken()
		expr, ok := p.matchExpression(PRECEDENCE_LOWEST)
		if !ok {
			return nil, false
		}
		if p.currentToken.Type != TOKEN_RIGHT_PAREN {
			return nil, false
		}
		p.nextToken()
		return expr, true
	default:
		return nil, false
	}
}

func (p *Parser) matchInfix(left Expression, precedence int) (Expression, bool) {
	operator := p.currentToken.Value
	p.nextToken()
	right, ok := p.matchExpression(precedence)
	if !ok {
		return nil, false
	}
	return &InfixNode{operator, left, right}, true
}

func (p *Parser) nextToken() *Token {
	p.currentToken = p.lexer.NextToken()
	return p.currentToken
}

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
	TOKEN_ASTERISK: PRECEDENCE_MUL_DIV,
	TOKEN_EQ:       PRECEDENCE_CMP,
	TOKEN_MINUS:    PRECEDENCE_ADD_SUB,
	TOKEN_PLUS:     PRECEDENCE_ADD_SUB,
	TOKEN_SLASH:    PRECEDENCE_MUL_DIV,
}
