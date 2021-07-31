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
	expr, ok := p.matchExpression()
	if !ok {
		return nil, false
	}
	if p.currentToken.Type != TOKEN_EOF {
		return nil, false
	}
	return expr, true
}

func (p *Parser) matchExpression() (Expression, bool) {
	switch p.currentToken.Type {
	case TOKEN_INT:
		token := p.currentToken
		p.nextToken()
		value, err := strconv.ParseInt(token.Value, 10, 64)
		if err != nil {
			return nil, false
		}
		return &IntegerNode{value}, true
	default:
		return nil, false
	}
}

func (p *Parser) nextToken() *Token {
	p.currentToken = p.lexer.NextToken()
	return p.currentToken
}
