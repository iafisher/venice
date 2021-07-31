package main

import "testing"

func TestParseInteger(t *testing.T) {
	expression := parseExpressionHelper(t, "123")

	integerNode, ok := expression.(*IntegerNode)
	if !ok {
		t.Fatalf("Wrong AST type: expected *IntegerNode, got %T", expression)
	}

	if integerNode.Value != 123 {
		t.Fatalf("Expected 123, got %d", integerNode.Value)
	}
}

func parseExpressionHelper(t *testing.T, input string) Expression {
	expr, ok := NewParser(NewLexer(input)).ParseExpression()
	if !ok {
		t.Fatalf("Failed to parse")
	}
	return expr
}
