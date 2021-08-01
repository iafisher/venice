package main

import (
	"fmt"
	"reflect"
	"testing"
)

func TestParseExpressions(t *testing.T) {
	var testCases = []struct {
		input    string
		expected Expression
	}{
		{"123", &IntegerNode{123}},
		{"abc", &SymbolNode{"abc"}},
		{"1 + 2", &InfixNode{"+", &IntegerNode{1}, &IntegerNode{2}}},
		{"1 + 2 * 3", &InfixNode{"+", &IntegerNode{1}, &InfixNode{"*", &IntegerNode{2}, &IntegerNode{3}}}},
		{"1 * 2 + 3", &InfixNode{"+", &InfixNode{"*", &IntegerNode{1}, &IntegerNode{2}}, &IntegerNode{3}}},
		{"1 * (2 + 3)", &InfixNode{"*", &IntegerNode{1}, &InfixNode{"+", &IntegerNode{2}, &IntegerNode{3}}}},
	}

	for i, testCase := range testCases {
		testName := fmt.Sprintf("%d", i)
		t.Run(testName, func(t *testing.T) {
			program, ok := NewParser(NewLexer(testCase.input)).Parse()
			if !ok {
				t.Fatalf("Failed to parse %q", testCase.input)
			}

			if len(program.Statements) != 1 {
				t.Fatalf("Expected exactly 1 statement, got %d", len(program.Statements))
			}

			expressionStatement, ok := program.Statements[0].(*ExpressionStatementNode)
			if !ok {
				t.Fatalf("Expected expression, got %+v for %q", program, testCase.input)
			}

			answer := expressionStatement.Expression
			if !reflect.DeepEqual(testCase.expected, answer) {
				t.Fatalf("expected %+[1]v (%[1]T), got %+[2]v (%[2]T) for %[3]q", testCase.expected, answer, testCase.input)
			}
		})
	}
}

func TestParseStatements(t *testing.T) {
	var testCases = []struct {
		input    string
		expected Statement
	}{
		{"let x = 10", &LetStatementNode{"x", &IntegerNode{10}}},
	}

	for i, testCase := range testCases {
		testName := fmt.Sprintf("%d", i)
		t.Run(testName, func(t *testing.T) {
			program, ok := NewParser(NewLexer(testCase.input)).Parse()
			if !ok {
				t.Fatalf("Failed to parse %q", testCase.input)
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
