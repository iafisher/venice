package main

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
)

type Bytecode struct {
	Name string
	Args []VeniceValue
}

func NewBytecode(name string, args ...VeniceValue) *Bytecode {
	return &Bytecode{name, args}
}

func WriteBytecodeListToFile(writer *bufio.Writer, bytecodeList []*Bytecode) {
	for _, bytecode := range bytecodeList {
		writer.WriteString(bytecode.Name)
		for _, arg := range bytecode.Args {
			writer.WriteString(" ")
			writer.WriteString(arg.Serialize())
		}
		writer.WriteString("\n")
	}

	writer.Flush()
}

func ReadBytecodeListFromString(bytecodeListString string) ([]*Bytecode, error) {
	bytecodeList := []*Bytecode{}
	for i, line := range strings.Split(bytecodeListString, "\n") {
		lexer := NewLexer(line)
		instruction := lexer.NextToken()
		if instruction.Type == TOKEN_EOF {
			continue
		}

		if instruction.Type != TOKEN_SYMBOL {
			return nil, &BytecodeParseError{fmt.Sprintf("could not parse line %d", i+1)}
		}

		args := []VeniceValue{}
		token := lexer.NextToken()
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

		bytecodeList = append(bytecodeList, &Bytecode{instruction.Value, args})
	}

	return bytecodeList, nil
}

type BytecodeParseError struct {
	Message string
}

func (e *BytecodeParseError) Error() string {
	return e.Message
}
