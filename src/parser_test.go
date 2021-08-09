package main

import (
	"fmt"
	"reflect"
	"testing"
)

func TestParseExpressions(t *testing.T) {
	var testCases = []struct {
		input    string
		expected ExpressionNode
	}{
		{"123", &IntegerNode{123, &Location{1, 1}}},
		{"abc", &SymbolNode{"abc", &Location{1, 1}}},
		{
			"1 + 2",
			&InfixNode{
				"+",
				&IntegerNode{1, &Location{1, 1}},
				&IntegerNode{2, &Location{1, 5}},
				&Location{1, 1},
			},
		},
		{
			"1 + 2 * 3",
			&InfixNode{
				"+",
				&IntegerNode{1, &Location{1, 1}},
				&InfixNode{
					"*",
					&IntegerNode{2, &Location{1, 5}},
					&IntegerNode{3, &Location{1, 9}},
					&Location{1, 5},
				},
				&Location{1, 1},
			},
		},
		{
			"1 * 2 + 3",
			&InfixNode{
				"+",
				&InfixNode{
					"*",
					&IntegerNode{1, &Location{1, 1}},
					&IntegerNode{2, &Location{1, 5}},
					&Location{1, 1},
				},
				&IntegerNode{3, &Location{1, 9}},
				&Location{1, 1},
			},
		},
		{
			"1 * (2 + 3)",
			&InfixNode{
				"*",
				&IntegerNode{1, &Location{1, 1}},
				&InfixNode{
					"+",
					&IntegerNode{2, &Location{1, 6}},
					&IntegerNode{3, &Location{1, 10}},
					&Location{1, 6},
				},
				&Location{1, 1},
			},
		},
	}

	for i, testCase := range testCases {
		testName := fmt.Sprintf("%d", i)
		t.Run(testName, func(t *testing.T) {
			program, err := NewParser(NewLexer(testCase.input)).Parse()
			if err != nil {
				t.Fatalf("Parse error: %s\n\nInput: %q", err, testCase.input)
			}

			if len(program.Statements) != 1 {
				t.Fatalf("Expected exactly 1 statement, got %d", len(program.Statements))
			}

			expressionStatement, ok := program.Statements[0].(*ExpressionStatementNode)
			if !ok {
				t.Fatalf("Expected expression, got %+v for %q", program, testCase.input)
			}

			answer := expressionStatement.Expr
			if !reflect.DeepEqual(testCase.expected, answer) {
				t.Fatalf("expected %+[1]v (%[1]T), got %+[2]v (%[2]T) for %[3]q", testCase.expected, answer, testCase.input)
			}
		})
	}
}

func TestParseStatements(t *testing.T) {
	var testCases = []struct {
		input    string
		expected StatementNode
	}{
		{"let x = 10", &LetStatementNode{"x", &IntegerNode{10, &Location{1, 9}}, &Location{1, 1}}},
	}

	for i, testCase := range testCases {
		testName := fmt.Sprintf("%d", i)
		t.Run(testName, func(t *testing.T) {
			program, err := NewParser(NewLexer(testCase.input)).Parse()
			if err != nil {
				t.Fatalf("Parse error: %s\n\nInput: %q", err, testCase.input)
			}

			if len(program.Statements) != 1 {
				t.Fatalf("Expected exactly 1 statement, got %d", len(program.Statements))
			}

			answer := program.Statements[0]
			if !reflect.DeepEqual(testCase.expected, answer) {
				t.Fatalf("expected %+[1]v (%[1]T), got %+[2]v (%[2]T) for %[3]q", testCase.expected, answer, testCase.input)
			}
		})
	}
}
