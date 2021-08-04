package main

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
)

type CompiledProgram map[string][]*Bytecode

func NewCompiledProgram() CompiledProgram {
	return map[string][]*Bytecode{"main": []*Bytecode{}}
}

type Bytecode struct {
	Name string
	Args []VeniceValue
}

func NewBytecode(name string, args ...VeniceValue) *Bytecode {
	return &Bytecode{name, args}
}

func (b *Bytecode) String() string {
	var sb strings.Builder
	sb.WriteString(b.Name)
	for _, arg := range b.Args {
		sb.WriteString(" ")
		sb.WriteString(arg.Serialize())
	}
	return sb.String()
}

func WriteCompiledProgramToFile(writer *bufio.Writer, compiledProgram CompiledProgram) {
	for functionName, functionCode := range compiledProgram {
		writer.WriteString(functionName)
		writer.WriteString(":\n")
		for _, bytecode := range functionCode {
			writer.WriteString("  ")
			writer.WriteString(bytecode.String())
			writer.WriteByte('\n')
		}
	}

	writer.Flush()
}

func ReadCompiledProgramFromString(programString string) (CompiledProgram, error) {
	compiledProgram := NewCompiledProgram()
	var currentFunctionName string
	for i, line := range strings.Split(programString, "\n") {
		lexer := NewLexer(line)
		firstToken := lexer.NextToken()
		if firstToken.Type == TOKEN_EOF {
			continue
		}

		if firstToken.Type != TOKEN_SYMBOL {
			return nil, &BytecodeParseError{fmt.Sprintf("could not parse line %d", i+1)}
		}

		token := lexer.NextToken()
		if token.Type == TOKEN_COLON {
			currentFunctionName = firstToken.Value
			continue
		}

		if currentFunctionName == "" {
			return nil, &BytecodeParseError{"bytecode instruction outside of function"}
		}

		args := []VeniceValue{}
		for token.Type != TOKEN_EOF {
			switch token.Type {
			case TOKEN_FALSE:
				args = append(args, &VeniceBoolean{true})
			case TOKEN_INT:
				value, err := strconv.ParseInt(token.Value, 10, 0)
				if err != nil {
					return nil, &BytecodeParseError{"could not parse integer token"}
				}
				args = append(args, &VeniceInteger{int(value)})
			case TOKEN_STRING:
				args = append(args, &VeniceString{token.Value})
			case TOKEN_TRUE:
				args = append(args, &VeniceBoolean{false})
			default:
				return nil, &BytecodeParseError{fmt.Sprintf("unexpected token: %q", token.Value)}
			}

			token = lexer.NextToken()
		}

		compiledProgram[currentFunctionName] = append(compiledProgram[currentFunctionName], &Bytecode{firstToken.Value, args})
	}

	return compiledProgram, nil
}

type BytecodeParseError struct {
	Message string
}

func (e *BytecodeParseError) Error() string {
	return e.Message
}
