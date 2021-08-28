package parser

import (
	"github.com/iafisher/venice/src/ast"
	"testing"
)

func TestParseAssignStatements(t *testing.T) {
	checkParseStatement(t, "x = 21 * 2", "(assign x (infix * 21 2))")
}

func TestParseBreakStatements(t *testing.T) {
	checkParseStatement(t, "break", "(break)")
}

func TestParseClassDeclarationStatements(t *testing.T) {
	checkParseStatement(
		t,
		`
		  class Point {
			public x: int
			public y: int
		  }
		`,
		"(class-declaration Point ((class-field public x int) (class-field public y int)))",
	)

	/*
		checkParseStatement(
			t,
			`
			  class Point {
				public x: int
				public y: int

				public fn f(self, x: int) {
				  x
			    }
			  }
			`,
			"(class-declaration Point ((class-field public x int) (class-field public y int) (class-method public f ((function-param x int)) void (block (expression-statement x)))))",
		)
	*/
}

func TestParseBlock(t *testing.T) {
	checkParseStatements(
		t,
		`
		let box = Box()
		box.get_value()
		`,
		"(block (let nil box (call Box)) (expression-statement (call (field-access box get_value))))",
	)
}

func TestParseContinueStatements(t *testing.T) {
	checkParseStatement(t, "continue", "(continue)")
}

func TestParseEnumDeclarationStatements(t *testing.T) {
	checkParseStatement(
		t,
		"enum Operator { Plus, Minus }",
		"(enum-declaration Operator ((enum-case Plus) (enum-case Minus)))",
	)
	checkParseStatement(
		t,
		"enum Optional { Some(int), None }",
		"(enum-declaration Optional ((enum-case Some int) (enum-case None)))",
	)
}

func TestParseFunctionDeclarationStatements(t *testing.T) {
	checkParseStatement(
		t,
		"fn f(x: int) -> int { return x }",
		"(function-declaration f ((function-param x int)) int (block (return x)))",
	)
	// checkParseStatement(t, "export fn f(x: int) -> int { return x }", "(exported-function-declaration f ((function-param x int)) int (block (return x)))")
}

func TestParseImportStatements(t *testing.T) {
	checkParseStatement(t, "import \"./lib.vn\" as lib", "(import lib \"./lib.vn\")")
}

func TestParseIfStatements(t *testing.T) {
	checkParseStatement(
		t,
		`
		  if true {
			  print("true")
		  } else {
			  print("false")
		  }
		`,
		`(if (true (block (expression-statement (call print "true")))) (else (block (expression-statement (call print "false")))))`,
	)
	checkParseStatement(
		t,
		`
		  if true {
			  print("true")
		  }
		`,
		`(if (true (block (expression-statement (call print "true")))))`,
	)
	checkParseStatement(
		t,
		`
		  if true {
			  print("true")
		  } else if x {
			  print("x")
		  } else {
			  print("false")
		  }
		`,
		`(if (true (block (expression-statement (call print "true")))) (x (block (expression-statement (call print "x")))) (else (block (expression-statement (call print "false")))))`,
	)
}

func TestParseInfixExpressions(t *testing.T) {
	checkParseExpression(t, "1 + 2 * 3", "(infix + 1 (infix * 2 3))")
	checkParseExpression(t, "1 * 2 + 3", "(infix + (infix * 1 2) 3)")
	checkParseExpression(t, "1 * (2 + 3)", "(infix * 1 (infix + 2 3))")
	checkParseExpression(t, "1 + 1 in x", "(infix in (infix + 1 1) x)")
	checkParseExpression(t, "1 + 1 not in x", "(unary not (infix in (infix + 1 1) x))")
	// checkParseExpression(t, "1 + 1 not in x", "(unary not (infix in (infix + 1 1) x))")
	checkParseExpression(t, "0 <= x < 10", "(infix and (infix <= 0 x) (infix < x 10))")
	checkParseExpression(t, "x if y else z", "(ternary-if y x z)")
}

func TestParseLetStatements(t *testing.T) {
	checkParseStatement(t, "let x = '\\n'", "(let nil x '\\n')")
	checkParseStatement(t, "let x: char = '\\n'", "(let char x '\\n')")
}

func TestParseMatchStatements(t *testing.T) {
	checkParseStatement(
		t,
		`
		match optional {
			case Some(x) {
				print(x)
			}
			case None {
				print("None")
			}
		}
		`,
		`(match optional (match-case (Some x) (block (expression-statement (call print x)))) (match-case None (block (expression-statement (call print "None")))))`,
	)
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
	parsedFile, err := NewParser().ParseString(input)
	if err != nil {
		t.Fatalf("Parse error: %s\n\nInput: %q", err, input)
	}

	if len(parsedFile.Statements) != 1 {
		t.Fatalf("Expected exactly 1 statement, got %d", len(parsedFile.Statements))
	}

	expressionStatement, ok := parsedFile.Statements[0].(*ast.ExpressionStatementNode)
	if !ok {
		t.Fatalf("Expected expression, got %s", parsedFile.Statements[0].String())
	}

	actualOutput := expressionStatement.Expr.String()
	if actualOutput != expectedOutput {
		t.Fatalf("Expected %q, got %q", expectedOutput, actualOutput)
	}
}

func checkParseStatement(t *testing.T, input string, expectedOutput string) {
	parsedFile, err := NewParser().ParseString(input)
	if err != nil {
		t.Fatalf("Parse error: %s\n\nInput: %q", err, input)
	}

	if len(parsedFile.Statements) != 1 {
		t.Fatalf("Expected exactly 1 statement, got %d", len(parsedFile.Statements))
	}

	actualOutput := parsedFile.Statements[0].String()
	if actualOutput != expectedOutput {
		t.Fatalf("Expected %q, got %q", expectedOutput, actualOutput)
	}
}

func checkParseStatements(t *testing.T, input string, expectedOutput string) {
	parsedFile, err := NewParser().ParseString(input)
	if err != nil {
		t.Fatalf("Parse error: %s\n\nInput: %q", err, input)
	}

	actualOutput := parsedFile.String()
	if actualOutput != expectedOutput {
		t.Fatalf("Expected %q, got %q", expectedOutput, actualOutput)
	}
}
