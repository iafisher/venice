package parser

import (
	"github.com/iafisher/venice/src/ast"
	lexer_mod "github.com/iafisher/venice/src/lexer"
	"testing"
)

func TestParseAssignStatements(t *testing.T) {
	checkParseStatement(t, "x = 21 * 2", "(assign x (infix * 21 2))")
}

func TestParseBreakStatements(t *testing.T) {
	checkParseStatement(t, "break", "(break)")
}

func TestParseClassDeclarationStatements(t *testing.T) {
	checkParseStatement(t, "class Point {\n  public x: int\n  public y: int\n}", "(class-declaration Point ((class-field public x int) (class-field public y int)))")
}

func TestParseContinueStatements(t *testing.T) {
	checkParseStatement(t, "continue", "(continue)")
}

func TestParseEnumDeclarationStatements(t *testing.T) {
	checkParseStatement(t, "enum Operator { Plus, Minus }", "(enum-declaration Operator ((enum-case Plus) (enum-case Minus)))")
	checkParseStatement(t, "enum Optional { Some(int), None }", "(enum-declaration Optional ((enum-case Some int) (enum-case None)))")
}

func TestParseLetStatements(t *testing.T) {
	checkParseStatement(t, "let x = 10", "(let x 10)")
}

func TestParseInfixExpressions(t *testing.T) {
	checkParseExpression(t, "1 + 2 * 3", "(infix + 1 (infix * 2 3))")
	checkParseExpression(t, "1 * 2 + 3", "(infix + (infix * 1 2) 3)")
	checkParseExpression(t, "1 * (2 + 3)", "(infix * 1 (infix + 2 3))")
	checkParseExpression(t, "1 + 1 in x", "(infix in (infix + 1 1) x)")
	// checkParseExpression(t, "1 + 1 not in x", "(unary not (infix in (infix + 1 1) x))")
}

func TestParseSimpleExpressions(t *testing.T) {
	checkParseExpression(t, "123", "123")
	checkParseExpression(t, "abc", "abc")
}

func TestParseUnaryOperators(t *testing.T) {
	checkParseExpression(t, "-(123)", "(unary - 123)")
	checkParseExpression(t, "- 123 + 2", "(infix + (unary - 123) 2)")
	checkParseExpression(t, "not true", "(unary not true)")
}

func checkParseExpression(t *testing.T, input string, expectedOutput string) {
	tree, err := NewParser(lexer_mod.NewLexer(input)).Parse()
	if err != nil {
		t.Fatalf("Parse error: %s\n\nInput: %q", err, input)
	}

	if len(tree.Statements) != 1 {
		t.Fatalf("Expected exactly 1 statement, got %d", len(tree.Statements))
	}

	expressionStatement, ok := tree.Statements[0].(*ast.ExpressionStatementNode)
	if !ok {
		t.Fatalf("Expected expression, got %s", tree.Statements[0].String())
	}

	actualOutput := expressionStatement.Expr.String()
	if actualOutput != expectedOutput {
		t.Fatalf("Expected %q, got %q", expectedOutput, actualOutput)
	}
}

func checkParseStatement(t *testing.T, input string, expectedOutput string) {
	tree, err := NewParser(lexer_mod.NewLexer(input)).Parse()
	if err != nil {
		t.Fatalf("Parse error: %s\n\nInput: %q", err, input)
	}

	if len(tree.Statements) != 1 {
		t.Fatalf("Expected exactly 1 statement, got %d", len(tree.Statements))
	}

	actualOutput := tree.Statements[0].String()
	if actualOutput != expectedOutput {
		t.Fatalf("Expected %q, got %q", expectedOutput, actualOutput)
	}
}

func checkParseStatements(t *testing.T, input string, expectedOutput string) {
	tree, err := NewParser(lexer_mod.NewLexer(input)).Parse()
	if err != nil {
		t.Fatalf("Parse error: %s\n\nInput: %q", err, input)
	}

	actualOutput := tree.String()
	if actualOutput != expectedOutput {
		t.Fatalf("Expected %q, got %q", expectedOutput, actualOutput)
	}
}
